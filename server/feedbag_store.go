package server

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mkaminski/goaim/oscar"
)

var feedbagDDL = `
	CREATE TABLE IF NOT EXISTS user
	(
		ScreenName VARCHAR(16) PRIMARY KEY,
		authKey    TEXT,
		passHash   TEXT
	);
	CREATE TABLE IF NOT EXISTS feedbag
	(
		ScreenName   VARCHAR(16),
		groupID      INTEGER,
		itemID       INTEGER,
		classID      INTEGER,
		name         TEXT,
		attributes   BLOB,
		lastModified INTEGER,
		UNIQUE (ScreenName, groupID, itemID)
	);
	CREATE TABLE IF NOT EXISTS profile
	(
		ScreenName VARCHAR(16) PRIMARY KEY,
		body  TEXT
	);
`

type BlockedState int

const (
	BlockedNo BlockedState = iota
	BlockedA
	BlockedB
)

func NewSQLiteFeedbagStore(dbFile string) (*SQLiteFeedbagStore, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(feedbagDDL); err != nil {
		return nil, err
	}
	return &SQLiteFeedbagStore{db: db}, nil
}

func NewStubUser(screenName string) (User, error) {
	u := User{ScreenName: screenName}

	uid, err := uuid.NewRandom()
	if err != nil {
		return u, err
	}
	u.AuthKey = uid.String()

	if err := u.HashPassword("welcome1"); err != nil {
		return u, err
	}
	return u, u.HashPassword("welcome1")
}

type User struct {
	ScreenName string `json:"screen_name"`
	AuthKey    string `json:"-"`
	PassHash   []byte `json:"-"`
}

func (u *User) HashPassword(passwd string) error {
	top := md5.New()
	if _, err := io.WriteString(top, passwd); err != nil {
		return err
	}
	bottom := md5.New()
	if _, err := io.WriteString(bottom, u.AuthKey); err != nil {
		return err
	}
	if _, err := bottom.Write(top.Sum(nil)); err != nil {
		return err
	}
	if _, err := io.WriteString(bottom, "AOL Instant Messenger (SM)"); err != nil {
		return err
	}
	u.PassHash = bottom.Sum(nil)
	return nil
}

type SQLiteFeedbagStore struct {
	db *sql.DB
}

func (f *SQLiteFeedbagStore) Users() ([]*User, error) {
	q := `SELECT ScreenName FROM user`
	rows, err := f.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ScreenName); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (f *SQLiteFeedbagStore) GetUser(screenName string) (*User, error) {
	q := `
		SELECT 
			ScreenName, 
			authKey, 
			passHash
		FROM user
		WHERE ScreenName = ?
	`
	u := &User{}
	err := f.db.QueryRow(q, screenName).Scan(&u.ScreenName, &u.AuthKey, &u.PassHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (f *SQLiteFeedbagStore) InsertUser(u User) error {
	q := `
		INSERT INTO user (ScreenName, authKey, passHash)
		VALUES (?, ?, ?)
	`
	_, err := f.db.Exec(q, u.ScreenName, u.AuthKey, u.PassHash)
	return err
}

func (f *SQLiteFeedbagStore) UpsertUser(u User) error {
	q := `
		INSERT INTO user (ScreenName, authKey, passHash)
		VALUES (?, ?, ?)
		ON CONFLICT DO NOTHING
	`
	_, err := f.db.Exec(q, u.ScreenName, u.AuthKey, u.PassHash)
	return err
}

func (f *SQLiteFeedbagStore) Delete(screenName string, items []oscar.FeedbagItem) error {
	// todo add transaction
	q := `DELETE FROM feedbag WHERE ScreenName = ? AND itemID = ?`

	for _, item := range items {
		if _, err := f.db.Exec(q, screenName, item.ItemID); err != nil {
			return err
		}
	}

	return nil
}

func (f *SQLiteFeedbagStore) Retrieve(screenName string) ([]oscar.FeedbagItem, error) {
	q := `
		SELECT 
			groupID,
			itemID,
			classID,
			name,
			attributes
		FROM feedbag
		WHERE ScreenName = ?
	`

	rows, err := f.db.Query(q, screenName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []oscar.FeedbagItem
	for rows.Next() {
		var item oscar.FeedbagItem
		var attrs []byte
		if err := rows.Scan(&item.GroupID, &item.ItemID, &item.ClassID, &item.Name, &attrs); err != nil {
			return nil, err
		}
		err = oscar.Unmarshal(&item.TLVLBlock, bytes.NewBuffer(attrs))
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (f *SQLiteFeedbagStore) LastModified(screenName string) (time.Time, error) {
	var lastModified sql.NullInt64
	q := `SELECT MAX(lastModified) FROM feedbag WHERE ScreenName = ?`
	err := f.db.QueryRow(q, screenName).Scan(&lastModified)
	return time.Unix(lastModified.Int64, 0), err
}

func (f *SQLiteFeedbagStore) Upsert(screenName string, items []oscar.FeedbagItem) error {

	q := `
		INSERT INTO feedbag (ScreenName, groupID, itemID, classID, name, attributes, lastModified)
		VALUES (?, ?, ?, ?, ?, ?, UNIXEPOCH())
		ON CONFLICT (ScreenName, groupID, itemID)
			DO UPDATE SET classID      = excluded.classID,
						  name         = excluded.name,
						  attributes   = excluded.attributes,
						  lastModified = UNIXEPOCH()
	`

	for _, item := range items {

		buf := &bytes.Buffer{}
		if err := oscar.Marshal(item.TLVLBlock, buf); err != nil {
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

// InterestedUsers returns all users who have screenName in their buddy list.
// Exclude users who are on screenName's block list.
func (f *SQLiteFeedbagStore) InterestedUsers(screenName string) ([]string, error) {
	q := `
		SELECT f.ScreenName
		FROM feedbag f
		WHERE f.name = ?
		  AND f.classID = 0
		-- Don't show screenName that its blocked buddy is online
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE ScreenName = ? AND name = f.ScreenName AND classID = 3)
		-- Don't show blocked buddy that screenName is online
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE ScreenName = f.ScreenName AND name = f.name AND classID = 3)
	`

	rows, err := f.db.Query(q, screenName, screenName, screenName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var screenName string
		if err := rows.Scan(&screenName); err != nil {
			return nil, err
		}
		items = append(items, screenName)
	}

	return items, nil
}

// Buddies returns all user's buddies. Don't return a buddy if screenName
// blocked them.
func (f *SQLiteFeedbagStore) Buddies(screenName string) ([]string, error) {
	q := `
		SELECT f.name
		FROM feedbag f
		WHERE f.ScreenName = ? AND f.classID = 0
		-- Don't include buddy if they blocked screenName
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE ScreenName = f.name AND name = ? AND classID = 3)
		-- Don't include buddy if screen name blocked them
		AND NOT EXISTS(SELECT 1 FROM feedbag WHERE ScreenName = ? AND name = f.name AND classID = 3)
	`

	rows, err := f.db.Query(q, screenName, screenName, screenName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var screenName string
		if err := rows.Scan(&screenName); err != nil {
			return nil, err
		}
		items = append(items, screenName)
	}

	return items, nil
}

// Blocked informs whether there is a blocking relationship between sn1 and
// sn2. Return BlockedA if sn1 blocked sn2, BlockedB if sn2 blocked sn1, or
// BlockedNo if neither screen name blocked the other.
func (f *SQLiteFeedbagStore) Blocked(sn1, sn2 string) (BlockedState, error) {
	q := `
		SELECT EXISTS(SELECT 1
					  FROM feedbag f
					  WHERE f.classID = 3
						AND f.ScreenName = ?
						AND f.name = ?)
		UNION ALL
		SELECT EXISTS(SELECT 1
					  FROM feedbag f
					  WHERE f.classID = 3
						AND f.ScreenName = ?
						AND f.name = ?)
	`
	var blockedA bool
	row, err := f.db.Query(q, sn1, sn2, sn2, sn1)
	if err != nil {
		// todo check to make sure there's no runtime error here...
		return BlockedNo, err
	}
	defer row.Close()

	row.Next()
	err = row.Scan(&blockedA)
	if err != nil {
		return BlockedNo, err
	}

	row.Next()
	var blockedB bool
	err = row.Scan(&blockedB)
	if err != nil {
		return BlockedNo, err
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

// RetrieveProfile fetches a user profile. Return empty string if the user
// does not exist or has no profile.
func (f *SQLiteFeedbagStore) RetrieveProfile(screenName string) (string, error) {
	q := `
		SELECT IFNULL(body, '')
		FROM profile
		WHERE ScreenName = ?
	`
	var profile string
	err := f.db.QueryRow(q, screenName).Scan(&profile)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return profile, nil
}

func (f *SQLiteFeedbagStore) UpsertProfile(screenName string, body string) error {
	q := `
		INSERT INTO profile (ScreenName, body)
		VALUES (?, ?)
		ON CONFLICT (ScreenName)
			DO UPDATE SET body = excluded.body
	`
	_, err := f.db.Exec(q, screenName, body)
	return err
}
