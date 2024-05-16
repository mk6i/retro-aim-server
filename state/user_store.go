package state

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/google/uuid"
)

// BlockedState represents the blocked status between two users
type BlockedState int

//go:embed migrations/*
var migrations embed.FS

var (
	// ErrDupUser indicates that a user already exists.
	ErrDupUser = errors.New("user already exists")
	// ErrNoUser indicates that a user does not exist.
	ErrNoUser = errors.New("user does not exist")
)

const (
	// BlockedNo indicates that neither user blocks the other.
	BlockedNo BlockedState = iota
	// BlockedA indicates that user A blocks user B.
	BlockedA
	// BlockedB indicates that user B blocks user A.
	BlockedB
)

// User represents a user account.
type User struct {
	// ScreenName is the AIM screen name.
	ScreenName string `json:"screen_name" db:"screenName"`
	// AuthKey is the salt for the MD5 password hash.
	AuthKey string `json:"-" db:"authKey"`
	// StrongMD5Pass is the MD5 password hash format used by AIM v4.8-v5.9.
	StrongMD5Pass []byte `json:"-" db:"strongMD5Pass"`
	// WeakMD5Pass is the MD5 password hash format used by AIM v3.5-v4.7. This
	// hash is used to authenticate roasted passwords for AIM v1.0-v3.0.
	WeakMD5Pass []byte `json:"-" db:"weakMD5Pass"`
}

// ValidateHash checks if md5Hash is identical to one of the password hashes.
func (u *User) ValidateHash(md5Hash []byte) bool {
	return bytes.Equal(u.StrongMD5Pass, md5Hash) || bytes.Equal(u.WeakMD5Pass, md5Hash)
}

// ValidateRoastedPass checks if the provided roasted password matches the MD5
// hash of the user's actual password. A roasted password is a XOR-obfuscated
// form of the real password, intended to add a simple layer of security.
func (u *User) ValidateRoastedPass(roastedPass []byte) bool {
	var roastTable = [16]byte{
		0xF3, 0x26, 0x81, 0xC4, 0x39, 0x86, 0xDB, 0x92,
		0x71, 0xA3, 0xB9, 0xE6, 0x53, 0x7A, 0x95, 0x7C,
	}
	clearPass := make([]byte, len(roastedPass))
	for i := range roastedPass {
		clearPass[i] = roastedPass[i] ^ roastTable[i%len(roastTable)]
	}
	md5Hash := weakMD5PasswordHash(string(clearPass), u.AuthKey) // todo remove string conversion
	return bytes.Equal(u.WeakMD5Pass, md5Hash)
}

// HashPassword computes MD5 hashes of the user's password. It computes both
// weak and strong variants and stores them in the struct.
func (u *User) HashPassword(passwd string) error {
	u.WeakMD5Pass = weakMD5PasswordHash(passwd, u.AuthKey)
	u.StrongMD5Pass = strongMD5PasswordHash(passwd, u.AuthKey)
	return nil
}

//goland:noinspection GoUnhandledErrorResult
func weakMD5PasswordHash(pass, authKey string) []byte {
	hash := md5.New()
	io.WriteString(hash, authKey)
	io.WriteString(hash, pass)
	io.WriteString(hash, "AOL Instant Messenger (SM)")
	return hash.Sum(nil)
}

//goland:noinspection GoUnhandledErrorResult
func strongMD5PasswordHash(pass, authKey string) []byte {
	top := md5.New()
	io.WriteString(top, pass)
	bottom := md5.New()
	io.WriteString(bottom, authKey)
	bottom.Write(top.Sum(nil))
	io.WriteString(bottom, "AOL Instant Messenger (SM)")
	return bottom.Sum(nil)
}

// UserStore stores user feedbag (buddy list), profile, and
// authentication credentials information in a database.
type UserStore struct {
	db     *sqlx.DB
	dbType string
}

// NewUserStore creates a new instance of UserStore. If the
// database does not already exist, a new one is created with the required
// schema.
func NewUserStore(dbType, connString string) (*UserStore, error) {
	db, err := sqlx.Open(dbType, connString)
	if err != nil {
		return nil, err
	}
	store := &UserStore{db: db, dbType: dbType}
	return store, store.runMigrations()
}

func (f UserStore) runMigrations() error {
	migrationFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to prepare migration subdirectory: %v", err)
	}

	sourceInstance, err := httpfs.New(http.FS(migrationFS), ".")
	if err != nil {
		return fmt.Errorf("failed to create source instance from embedded filesystem: %v", err)
	}

	var driver database.Driver
	if f.dbType == "sqlite3" {
		driver, err = sqlite3.WithInstance(f.db.DB, &sqlite3.Config{})
	} else if f.dbType == "postgres" {
		driver, err = postgres.WithInstance(f.db.DB, &postgres.Config{})
	} else {
		return fmt.Errorf("unsupported database type: %s", f.dbType)
	}
	if err != nil {
		return fmt.Errorf("cannot create database driver: %v", err)
	}

	m, err := migrate.NewWithInstance("httpfs", sourceInstance, f.dbType, driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

// AllUsers returns all stored users. It only populates the User.ScreenName field
// populated in the returned slice.
func (f UserStore) AllUsers() ([]User, error) {
	q := `SELECT screenName FROM user`
	rows, err := f.db.Queryx(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		u := User{}
		if err := rows.StructScan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// User looks up a user by screen name. It populates the User record with
// credentials that can be used to validate the user's password.
func (f UserStore) User(screenName string) (*User, error) {
	q := `
		SELECT
			screenName, 
			authKey, 
			weakMD5Pass,
			strongMD5Pass
		FROM user
		WHERE screenName = ?
	`
	u := &User{}
	err := f.db.QueryRowx(q, screenName).StructScan(u)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

// InsertUser inserts a user to the store. Return ErrDupUser if a user with the
// same screen name already exists.
func (f UserStore) InsertUser(u User) error {
	q := `
		INSERT INTO user (screenName, authKey, weakMD5Pass, strongMD5Pass)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (screenName) DO NOTHING
	`
	result, err := f.db.Exec(q, u.ScreenName, u.AuthKey, u.WeakMD5Pass, u.StrongMD5Pass)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrDupUser
	}

	return nil
}

// SetUserPassword sets the user's password hashes and auth key.
func (f UserStore) SetUserPassword(u User) error {
	tx, err := f.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	q := `
		UPDATE user
		SET authKey = ?, weakMD5Pass = ?, strongMD5Pass = ?
		WHERE screenName = ?
	`
	result, err := tx.Exec(q, u.AuthKey, u.WeakMD5Pass, u.StrongMD5Pass, u.ScreenName)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// it's possible the user didn't change OR the user doesn't exist.
		// check if the user exists.
		var exists int
		err = tx.QueryRow("SELECT COUNT(*) FROM user WHERE screenName = ?", u.ScreenName).Scan(&exists)
		if err != nil {
			return err // Handle possible SQL errors during the select
		}
		if exists == 0 {
			return ErrNoUser // User does not exist
		}
	}

	return tx.Commit()
}

// Feedbag fetches the contents of a user's feedbag (buddy list).
func (f UserStore) Feedbag(screenName string) ([]wire.FeedbagItem, error) {
	q := `
		SELECT 
			groupID,
			itemID,
			classID,
			name,
			attributes
		FROM feedbag
		WHERE screenName = ?
	`

	rows, err := f.db.Queryx(q, screenName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []wire.FeedbagItem
	for rows.Next() {
		var item wire.FeedbagItem
		var attrs []byte
		if err := rows.Scan(&item.GroupID, &item.ItemID, &item.ClassID, &item.Name, &attrs); err != nil {
			return nil, err
		}
		if err := wire.Unmarshal(&item.TLVLBlock, bytes.NewBuffer(attrs)); err != nil {
			return items, err
		}
		items = append(items, item)
	}

	return items, nil
}

// FeedbagLastModified returns the last time a user's feedbag (buddy list) was
// updated.
func (f UserStore) FeedbagLastModified(screenName string) (time.Time, error) {
	var lastModified sql.NullInt64
	q := `SELECT MAX(lastModified) FROM feedbag WHERE screenName = ?`
	err := f.db.QueryRow(q, screenName).Scan(&lastModified)
	return time.Unix(lastModified.Int64, 0), err
}

// FeedbagDelete deletes an entry from a user's feedbag (buddy list).
func (f UserStore) FeedbagDelete(screenName string, items []wire.FeedbagItem) error {
	// todo add transaction
	q := `DELETE FROM feedbag WHERE screenName = ? AND itemID = ?`

	for _, item := range items {
		if _, err := f.db.Exec(q, screenName, item.ItemID); err != nil {
			return err
		}
	}

	return nil
}

// FeedbagUpsert upserts an entry to a user's feedbag (buddy list). An entry is
// created if it doesn't already exist, or modified if it already exists.
func (f UserStore) FeedbagUpsert(screenName string, items []wire.FeedbagItem) error {
	q := `
		INSERT INTO feedbag (screenName, groupID, itemID, classID, name, attributes, lastModified)
		VALUES (?, ?, ?, ?, ?, ?, UNIXEPOCH())
		ON CONFLICT (screenName, groupID, itemID)
			DO UPDATE SET classID      = excluded.classID,
						  name         = excluded.name,
						  attributes   = excluded.attributes,
						  lastModified = UNIXEPOCH()
	`

	for _, item := range items {
		buf := &bytes.Buffer{}
		if err := wire.Marshal(item.TLVLBlock, buf); err != nil {
			return err
		}

		_, err := f.db.Exec(q,
			screenName,
			item.GroupID,
			item.ItemID,
			item.ClassID,
			item.Name,
			buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

// AdjacentUsers returns all users who have screenName in their buddy list.
// Exclude users who are on screenName's block list.
func (f UserStore) AdjacentUsers(screenName string) ([]string, error) {
	q := `
		SELECT f.screenName
		FROM feedbag f
		WHERE f.name = ?
		  AND f.classID = 0
		-- Don't show screenName that its blocked buddy is online
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE screenName = ? AND name = f.screenName AND classID = 3)
		-- Don't show blocked buddy that screenName is online
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE screenName = f.screenName AND name = f.name AND classID = 3)
	`

	rows, err := f.db.Queryx(q, screenName, screenName, screenName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var sn string
		if err := rows.Scan(&sn); err != nil {
			return nil, err
		}
		items = append(items, sn)
	}

	return items, nil
}

// Buddies returns all user's buddies. Don't return a buddy if the user has
// them on their block list.
func (f UserStore) Buddies(screenName string) ([]string, error) {
	q := `
		SELECT f.name
		FROM feedbag f
		WHERE f.screenName = ? AND f.classID = 0
		-- Don't include buddy if they blocked screenName
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE screenName = f.name AND name = ? AND classID = 3)
		-- Don't include buddy if screen name blocked them
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE screenName = ? AND name = f.name AND classID = 3)
	`

	rows, err := f.db.Queryx(q, screenName, screenName, screenName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var sn string
		if err := rows.Scan(&sn); err != nil {
			return nil, err
		}
		items = append(items, sn)
	}

	return items, nil
}

// BlockedState returns the BlockedState between two users.
func (f UserStore) BlockedState(screenNameA, screenNameB string) (BlockedState, error) {
	q := `
		SELECT EXISTS(SELECT 1
					  FROM feedbag f
					  WHERE f.classID = 3
						AND f.screenName = ?
						AND f.name = ?)
		UNION ALL
		SELECT EXISTS(SELECT 1
					  FROM feedbag f
					  WHERE f.classID = 3
						AND f.screenName = ?
						AND f.name = ?)
	`
	row, err := f.db.Queryx(q, screenNameA, screenNameB, screenNameB, screenNameA)
	if err != nil {
		return BlockedNo, err
	}
	defer row.Close()

	var blockedA bool
	if row.Next() {
		if err := row.Scan(&blockedA); err != nil {
			return BlockedNo, err
		}
	}

	var blockedB bool
	if row.Next() {
		if err := row.Scan(&blockedB); err != nil {
			return BlockedNo, err
		}
	}

	switch {
	case blockedA:
		return BlockedA, nil
	case blockedB:
		return BlockedB, nil
	default:
		return BlockedNo, nil
	}
}

// Profile fetches a user profile. Return empty string if the user
// does not exist or has no profile.
func (f UserStore) Profile(screenName string) (string, error) {
	q := `
		SELECT IFNULL(body, '')
		FROM profile
		WHERE screenName = ?
	`
	var profile string
	err := f.db.QueryRow(q, screenName).Scan(&profile)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return profile, nil
}

// SetProfile sets the text contents of a user's profile.
func (f UserStore) SetProfile(screenName string, body string) error {
	q := `
		INSERT INTO profile (screenName, body)
		VALUES (?, ?)
		ON CONFLICT (screenName)
			DO UPDATE SET body = excluded.body
	`
	_, err := f.db.Exec(q, screenName, body)
	return err
}

func (f UserStore) BARTUpsert(itemHash []byte, body []byte) error {
	q := `
		INSERT INTO bartItem (hash, body)
		VALUES (?, ?)
		ON CONFLICT DO NOTHING
	`
	_, err := f.db.Exec(q, itemHash, body)
	return err
}

func (f UserStore) BARTRetrieve(hash []byte) ([]byte, error) {
	q := `
		SELECT body
		FROM bartItem
		WHERE hash = ?
	`
	var body []byte
	err := f.db.QueryRow(q, hash).Scan(&body)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return body, nil
}

// NewStubUser creates a new user with canned credentials. The default password
// is "welcome1". This is typically used for development purposes.
func NewStubUser(screenName string) (User, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return User{}, err
	}
	u := User{
		ScreenName: screenName,
		AuthKey:    uid.String(),
	}
	err = u.HashPassword("welcome1")
	return u, err
}
