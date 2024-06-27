package state

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	sqlite "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*
var migrations embed.FS

// SQLiteUserStore stores user feedbag (buddy list), profile, and
// authentication credentials information in a SQLite database.
type SQLiteUserStore struct {
	db *sql.DB
}

// NewSQLiteUserStore creates a new instance of SQLiteUserStore. If the
// database does not already exist, a new one is created with the required
// schema.
func NewSQLiteUserStore(dbFilePath string) (*SQLiteUserStore, error) {
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}
	store := &SQLiteUserStore{db: db}
	return store, store.runMigrations()
}

func (f SQLiteUserStore) runMigrations() error {
	migrationFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to prepare migration subdirectory: %v", err)
	}

	sourceInstance, err := httpfs.New(http.FS(migrationFS), ".")
	if err != nil {
		return fmt.Errorf("failed to create source instance from embedded filesystem: %v", err)
	}

	driver, err := sqlite3.WithInstance(f.db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("cannot create database driver: %v", err)
	}

	m, err := migrate.NewWithInstance("httpfs", sourceInstance, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

// AllUsers returns all stored users. It only populates the User.IdentScreenName and
// User.DisplayScreenName fields in the return slice.
func (f SQLiteUserStore) AllUsers() ([]User, error) {
	q := `SELECT identScreenName, displayScreenName FROM users`
	rows, err := f.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var identSN, displaySN string
		if err := rows.Scan(&identSN, &displaySN); err != nil {
			return nil, err
		}
		users = append(users, User{
			IdentScreenName:   NewIdentScreenName(identSN),
			DisplayScreenName: DisplayScreenName(displaySN),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// User looks up a user by screen name. It populates the User record with
// credentials that can be used to validate the user's password.
func (f SQLiteUserStore) User(screenName IdentScreenName) (*User, error) {
	q := `
		SELECT
			displayScreenName,
			authKey,
			weakMD5Pass,
			strongMD5Pass
		FROM users
		WHERE identScreenName = ?
	`
	u := &User{
		IdentScreenName: screenName,
	}
	err := f.db.QueryRow(q, screenName.String()).
		Scan(&u.DisplayScreenName, &u.AuthKey, &u.WeakMD5Pass, &u.StrongMD5Pass)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

// InsertUser inserts a user to the store. Return ErrDupUser if a user with the
// same screen name already exists.
func (f SQLiteUserStore) InsertUser(u User) error {
	q := `
		INSERT INTO users (identScreenName, displayScreenName, authKey, weakMD5Pass, strongMD5Pass)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (identScreenName) DO NOTHING
	`
	result, err := f.db.Exec(q, u.IdentScreenName.String(), u.DisplayScreenName, u.AuthKey, u.WeakMD5Pass, u.StrongMD5Pass)
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

// DeleteUser deletes a user from the store. Return ErrNoUser if the user did
// not exist prior to deletion.
func (f SQLiteUserStore) DeleteUser(screenName IdentScreenName) error {
	q := `
		DELETE FROM users WHERE identScreenName = ?
	`
	result, err := f.db.Exec(q, screenName.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoUser
	}

	return nil
}

// SetUserPassword sets the user's password hashes and auth key.
func (f SQLiteUserStore) SetUserPassword(u User) (err error) {
	tx, err := f.db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	q := `
		UPDATE users
		SET authKey = ?, weakMD5Pass = ?, strongMD5Pass = ?
		WHERE identScreenName = ?
	`
	result, err := tx.Exec(q, u.AuthKey, u.WeakMD5Pass, u.StrongMD5Pass, u.IdentScreenName.String())
	if err != nil {
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return
	}

	if rowsAffected == 0 {
		// it's possible the user didn't change OR the user doesn't exist.
		// check if the user exists.
		var exists int
		err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE identScreenName = ?", u.IdentScreenName.String()).Scan(&exists)
		if err != nil {
			return // Handle possible SQL errors during the select
		}
		if exists == 0 {
			err = ErrNoUser // User does not exist
			return
		}
	}

	return tx.Commit()
}

// Feedbag fetches the contents of a user's feedbag (buddy list).
func (f SQLiteUserStore) Feedbag(screenName IdentScreenName) ([]wire.FeedbagItem, error) {
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

	rows, err := f.db.Query(q, screenName.String())
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
func (f SQLiteUserStore) FeedbagLastModified(screenName IdentScreenName) (time.Time, error) {
	var lastModified sql.NullInt64
	q := `SELECT MAX(lastModified) FROM feedbag WHERE screenName = ?`
	err := f.db.QueryRow(q, screenName.String()).Scan(&lastModified)
	return time.Unix(lastModified.Int64, 0), err
}

// FeedbagDelete deletes an entry from a user's feedbag (buddy list).
func (f SQLiteUserStore) FeedbagDelete(screenName IdentScreenName, items []wire.FeedbagItem) error {
	// todo add transaction
	q := `DELETE FROM feedbag WHERE screenName = ? AND itemID = ?`

	for _, item := range items {
		if _, err := f.db.Exec(q, screenName.String(), item.ItemID); err != nil {
			return err
		}
	}

	return nil
}

// FeedbagUpsert upserts an entry to a user's feedbag (buddy list). An entry is
// created if it doesn't already exist, or modified if it already exists.
func (f SQLiteUserStore) FeedbagUpsert(screenName IdentScreenName, items []wire.FeedbagItem) error {
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

		// todo are these all the right items?
		if item.ClassID == wire.FeedbagClassIdBuddy ||
			item.ClassID == wire.FeedbagClassIDPermit ||
			item.ClassID == wire.FeedbagClassIDDeny {
			// insert screen name identifier
			item.Name = NewIdentScreenName(item.Name).String()
		}
		_, err := f.db.Exec(q,
			screenName.String(),
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
func (f SQLiteUserStore) AdjacentUsers(screenName IdentScreenName) ([]IdentScreenName, error) {
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

	rows, err := f.db.Query(q, screenName.String(), screenName.String(), screenName.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []IdentScreenName
	for rows.Next() {
		var sn string
		if err := rows.Scan(&sn); err != nil {
			return nil, err
		}
		items = append(items, NewIdentScreenName(sn))
	}

	return items, nil
}

// Buddies returns all user's buddies. Don't return a buddy if the user has
// them on their block list.
func (f SQLiteUserStore) Buddies(screenName IdentScreenName) ([]IdentScreenName, error) {
	q := `
		SELECT f.name
		FROM feedbag f
		WHERE f.screenName = ? AND f.classID = 0
		-- Don't include buddy if they blocked screenName
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE screenName = f.name AND name = ? AND classID = 3)
		-- Don't include buddy if screen name blocked them
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE screenName = ? AND name = f.name AND classID = 3)
	`

	rows, err := f.db.Query(q, screenName.String(), screenName.String(), screenName.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []IdentScreenName
	for rows.Next() {
		var sn string
		if err := rows.Scan(&sn); err != nil {
			return nil, err
		}
		items = append(items, NewIdentScreenName(sn))
	}

	return items, nil
}

// BlockedState returns the BlockedState between two users.
func (f SQLiteUserStore) BlockedState(screenName1, screenName2 IdentScreenName) (BlockedState, error) {
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
	row, err := f.db.Query(q, screenName1.String(), screenName2.String(), screenName2.String(), screenName1.String())
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
func (f SQLiteUserStore) Profile(screenName IdentScreenName) (string, error) {
	q := `
		SELECT IFNULL(body, '')
		FROM profile
		WHERE screenName = ?
	`
	var profile string
	err := f.db.QueryRow(q, screenName.String()).Scan(&profile)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return profile, nil
}

// SetProfile sets the text contents of a user's profile.
func (f SQLiteUserStore) SetProfile(screenName IdentScreenName, body string) error {
	q := `
		INSERT INTO profile (screenName, body)
		VALUES (?, ?)
		ON CONFLICT (screenName)
			DO UPDATE SET body = excluded.body
	`
	_, err := f.db.Exec(q, screenName.String(), body)
	return err
}

func (f SQLiteUserStore) BARTUpsert(itemHash []byte, body []byte) error {
	q := `
		INSERT INTO bartItem (hash, body)
		VALUES (?, ?)
		ON CONFLICT DO NOTHING
	`
	_, err := f.db.Exec(q, itemHash, body)
	return err
}

func (f SQLiteUserStore) BARTRetrieve(hash []byte) ([]byte, error) {
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
	return body, err
}

// ChatRoomByCookie looks up a chat room by cookie. Returns
// ErrChatRoomNotFound if the room does not exist for cookie.
func (f SQLiteUserStore) ChatRoomByCookie(cookie string) (ChatRoom, error) {
	chatRoom := ChatRoom{
		Cookie: cookie,
	}

	q := `
		SELECT exchange, name, created, creator
		FROM chatRoom
		WHERE cookie = ?
	`
	var creator string
	err := f.db.QueryRow(q, cookie).Scan(
		&chatRoom.Exchange,
		&chatRoom.Name,
		&chatRoom.CreateTime,
		&creator,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrChatRoomNotFound
	}
	chatRoom.Creator = NewIdentScreenName(creator)

	return chatRoom, err
}

// ChatRoomByName looks up a chat room by exchange and name. Returns
// ErrChatRoomNotFound if the room does not exist for exchange and name.
func (f SQLiteUserStore) ChatRoomByName(exchange uint16, name string) (ChatRoom, error) {
	chatRoom := ChatRoom{
		Exchange: exchange,
		Name:     name,
	}

	q := `
		SELECT cookie, created, creator
		FROM chatRoom
		WHERE exchange = ? AND name = ?
	`
	var creator string
	err := f.db.QueryRow(q, exchange, name).Scan(
		&chatRoom.Cookie,
		&chatRoom.CreateTime,
		&creator,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrChatRoomNotFound
	}
	chatRoom.Creator = NewIdentScreenName(creator)

	return chatRoom, err
}

// CreateChatRoom creates a new chat room.
func (f SQLiteUserStore) CreateChatRoom(chatRoom ChatRoom) error {
	q := `
		INSERT INTO chatRoom (cookie, exchange, name, created, creator)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := f.db.Exec(
		q,
		chatRoom.Cookie,
		chatRoom.Exchange,
		chatRoom.Name,
		chatRoom.CreateTime,
		chatRoom.Creator.String(),
	)

	if err != nil {
		if sqliteErr, ok := err.(sqlite.Error); ok {
			if sqliteErr.ExtendedCode == sqlite.ErrConstraintUnique || sqliteErr.ExtendedCode == sqlite.ErrConstraintPrimaryKey {
				err = ErrDupChatRoom
			}
		}
		err = fmt.Errorf("CreateChatRoom: %w", err)
	}
	return err
}

func (f SQLiteUserStore) AllChatRooms(exchange uint16) ([]ChatRoom, error) {
	q := `
		SELECT cookie, created, creator, name
		FROM chatRoom
		WHERE exchange = ?
		ORDER BY created ASC
	`
	rows, err := f.db.Query(q, exchange)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []ChatRoom
	for rows.Next() {
		cr := ChatRoom{
			Exchange: exchange,
		}
		var creator string
		if err := rows.Scan(&cr.Cookie, &cr.CreateTime, &creator, &cr.Name); err != nil {
			return nil, err
		}
		cr.Creator = NewIdentScreenName(creator)
		users = append(users, cr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
