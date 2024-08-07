package state

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
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

// FindByUIN returns a user where the UIN matches the ident screen name.
func (f SQLiteUserStore) FindByUIN(UIN uint32) (User, error) {
	users, err := getUsers(func() (string, []any) {
		return `identScreenName = ?`, []any{strconv.Itoa(int(UIN))}
	}, f.db)
	if err != nil {
		return User{}, fmt.Errorf("FindByUIN: %w", err)
	}

	if len(users) == 0 {
		return User{}, ErrNoUser
	}

	return users[0], nil
}

// FindByEmail returns a user with a matching email address.
func (f SQLiteUserStore) FindByEmail(email string) (User, error) {
	users, err := getUsers(func() (string, []any) {
		return `emailAddress = ?`, []any{email}
	}, f.db)
	if err != nil {
		return User{}, fmt.Errorf("FindByEmail: %w", err)
	}

	if len(users) == 0 {
		return User{}, ErrNoUser
	}

	return users[0], nil
}

// FindByDetails returns a user with either a matching first name, last name, or nickname.
func (f SQLiteUserStore) FindByDetails(firstName, lastName, nickName string) ([]User, error) {
	users, err := getUsers(func() (string, []any) {
		var conds []string
		var vals []any
		if firstName != "" {
			conds = append(conds, `LOWER(firstName) = LOWER(?)`)
			vals = append(vals, firstName)
		}
		if lastName != "" {
			conds = append(conds, `LOWER(lastName) = LOWER(?)`)
			vals = append(vals, lastName)
		}
		if nickName != "" {
			conds = append(conds, `LOWER(nickName) = LOWER(?)`)
			vals = append(vals, nickName)
		}
		return strings.Join(conds, " OR "), vals
	}, f.db)
	if err != nil {
		err = fmt.Errorf("FindByDetails: %w", err)
	}
	return users, nil
}

// FindByInterests returns a user who has at least one matching interest.
func (f SQLiteUserStore) FindByInterests(code uint16, keywords []string) ([]User, error) {
	users, err := getUsers(func() (string, []any) {
		var conds []string
		var vals []any

		for i := 1; i <= 4; i++ {
			var subConds []string
			vals = append(vals, code)
			for _, key := range keywords {
				subConds = append(subConds, fmt.Sprintf("interest%dKeyword LIKE ?", i))
				vals = append(vals, "%"+key+"%")
			}
			conds = append(conds, fmt.Sprintf("(interest%dCode = ? AND (%s))", i, strings.Join(subConds, " OR ")))
		}

		return strings.Join(conds, " OR "), vals
	}, f.db)
	if err != nil {
		err = fmt.Errorf("FindByInterests: %w", err)
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
			strongMD5Pass,
			confirmStatus,
			emailAddress,
			regStatus,
			firstName,
			lastName,
			authReq,
			gender
		FROM users
		WHERE identScreenName = ?
	`
	u := &User{
		IdentScreenName: screenName,
	}
	err := f.db.QueryRow(q, screenName.String()).Scan(
		&u.DisplayScreenName,
		&u.AuthKey,
		&u.WeakMD5Pass,
		&u.StrongMD5Pass,
		&u.ConfirmStatus,
		&u.EmailAddress,
		&u.RegStatus,
		&u.FirstName,
		&u.LastName,
		&u.AuthReq,
		&u.Gender,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

type filterFN func() (string, []any)

func filterByID(id IdentScreenName) filterFN {
	return func() (string, []any) {
		return `identScreenName = ?`, []any{id.String()}
	}
}

type queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// getUsers fetches users from the database by their screen name.
func getUsers(filterFN filterFN, tx queryer) ([]User, error) {

	cond, parms := filterFN()

	q := `
		SELECT
			identScreenName,
			affiliations1Code,
			affiliations1Keyword,
			affiliations2Code,
			affiliations2Keyword,
			affiliations3Code,
			affiliations3Keyword,
			authKey,
			authReq,
			birthDay,
			birthMonth,
			birthYear,
			cellPhone,
			company,
			confirmStatus,
			countryCode,
			department,
			displayScreenName,
			emailAddress,
			firstName,
			gmtOffset,
			gender,
			homeAddress,
			homeCity,
			homeFax,
			homePageAddr,
			homePhone,
			homeState,
			interest1Code,
			interest1Keyword,
			interest2Code,
			interest2Keyword,
			interest3Code,
			interest3Keyword,
			interest4Code,
			interest4Keyword,
			isICQ,
			lang1,
			lang2,
			lang3,
			lastName,
			nickname,
			notes,
			occupationCode,
			pastAffiliations1Code,
			pastAffiliations1Keyword,
			pastAffiliations2Code,
			pastAffiliations2Keyword,
			pastAffiliations3Code,
			pastAffiliations3Keyword,
			position,
			publishEmail,
			regStatus,
			strongMD5Pass,
			weakMD5Pass,
			workAddress,
			workCity,
			workCountryCode,
			workFax,
			workPhone,
			workState,
			workWebPage,
			workZIP,
			zipCode
		FROM users
		WHERE %s
	`
	q = fmt.Sprintf(q, cond)
	rows, err := tx.Query(q, parms...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var sn string
		err := rows.Scan(
			&sn,
			&u.Affiliations1Code,
			&u.Affiliations1Keyword,
			&u.Affiliations2Code,
			&u.Affiliations2Keyword,
			&u.Affiliations3Code,
			&u.Affiliations3Keyword,
			&u.AuthKey,
			&u.AuthReq,
			&u.BirthDay,
			&u.BirthMonth,
			&u.BirthYear,
			&u.CellPhone,
			&u.Company,
			&u.ConfirmStatus,
			&u.CountryCode,
			&u.Department,
			&u.DisplayScreenName,
			&u.EmailAddress,
			&u.FirstName,
			&u.GMTOffset,
			&u.Gender,
			&u.HomeAddress,
			&u.HomeCity,
			&u.HomeFax,
			&u.HomePageAddr,
			&u.HomePhone,
			&u.HomeState,
			&u.Interest1Code,
			&u.Interest1Keyword,
			&u.Interest2Code,
			&u.Interest2Keyword,
			&u.Interest3Code,
			&u.Interest3Keyword,
			&u.Interest4Code,
			&u.Interest4Keyword,
			&u.IsICQ,
			&u.Lang1,
			&u.Lang2,
			&u.Lang3,
			&u.LastName,
			&u.Nickname,
			&u.Notes,
			&u.OccupationCode,
			&u.PastAffiliations1Code,
			&u.PastAffiliations1Keyword,
			&u.PastAffiliations2Code,
			&u.PastAffiliations2Keyword,
			&u.PastAffiliations3Code,
			&u.PastAffiliations3Keyword,
			&u.Position,
			&u.PublishEmail,
			&u.RegStatus,
			&u.StrongMD5Pass,
			&u.WeakMD5Pass,
			&u.WorkAddress,
			&u.WorkCity,
			&u.WorkCountryCode,
			&u.WorkFax,
			&u.WorkPhone,
			&u.WorkState,
			&u.WorkWebPage,
			&u.WorkZIP,
			&u.ZipCode,
		)
		if err != nil {
			return nil, err
		}
		u.IdentScreenName = NewIdentScreenName(sn)
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// InsertUser inserts a user to the store. Return ErrDupUser if a user with the
// same screen name already exists.
func (f SQLiteUserStore) InsertUser(u User) error {
	q := `
		INSERT INTO users (identScreenName, displayScreenName, authKey, weakMD5Pass, strongMD5Pass, isICQ)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (identScreenName) DO NOTHING
	`
	result, err := f.db.Exec(q,
		u.IdentScreenName.String(),
		u.DisplayScreenName,
		u.AuthKey,
		u.WeakMD5Pass,
		u.StrongMD5Pass,
		u.IsICQ,
	)
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

// SetUserPassword sets the user's password hashes and auth key. The following
// fields must be set on u:
// - AuthKey
// - WeakMD5Pass
// - StrongMD5Pass
// - IdentScreenName
func (f SQLiteUserStore) SetUserPassword(screenName IdentScreenName, newPassword string) error {
	tx, err := f.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	q := `
		SELECT
			authKey,
			isICQ
		FROM users
		WHERE identScreenName = ?
	`

	u := User{}

	err = tx.QueryRow(q, screenName.String()).Scan(
		&u.AuthKey,
		&u.IsICQ,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoUser
	}

	if err = u.HashPassword(newPassword); err != nil {
		return err
	}

	q = `
		UPDATE users
		SET authKey = ?, weakMD5Pass = ?, strongMD5Pass = ?
		WHERE identScreenName = ?
	`
	result, err := tx.Exec(q, u.AuthKey, u.WeakMD5Pass, u.StrongMD5Pass, screenName.String())
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
		err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE identScreenName = ?", u.IdentScreenName.String()).Scan(&exists)
		if err != nil {
			return err // Handle possible SQL errors during the select
		}
		if exists == 0 {
			err = ErrNoUser // User does not exist
			return err
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
		if err := wire.UnmarshalBE(&item.TLVLBlock, bytes.NewBuffer(attrs)); err != nil {
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
		if err := wire.MarshalBE(item.TLVLBlock, buf); err != nil {
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
	chatRoom := ChatRoom{}

	q := `
		SELECT exchange, name, created, creator
		FROM chatRoom
		WHERE cookie = ?
	`
	var creator string
	err := f.db.QueryRow(q, cookie).Scan(
		&chatRoom.exchange,
		&chatRoom.name,
		&chatRoom.createTime,
		&creator,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrChatRoomNotFound
	}
	chatRoom.creator = NewIdentScreenName(creator)

	return chatRoom, err
}

// ChatRoomByName looks up a chat room by exchange and name. Returns
// ErrChatRoomNotFound if the room does not exist for exchange and name.
func (f SQLiteUserStore) ChatRoomByName(exchange uint16, name string) (ChatRoom, error) {
	chatRoom := ChatRoom{
		exchange: exchange,
		name:     name,
	}

	q := `
		SELECT created, creator
		FROM chatRoom
		WHERE exchange = ? AND name = ?
	`
	var creator string
	err := f.db.QueryRow(q, exchange, name).Scan(
		&chatRoom.createTime,
		&creator,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrChatRoomNotFound
	}
	chatRoom.creator = NewIdentScreenName(creator)

	return chatRoom, err
}

// CreateChatRoom creates a new chat room. It sets createTime on chatRoom to
// the current timestamp.
func (f SQLiteUserStore) CreateChatRoom(chatRoom *ChatRoom) error {
	chatRoom.createTime = time.Now().UTC()
	q := `
		INSERT INTO chatRoom (cookie, exchange, name, created, creator)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := f.db.Exec(
		q,
		chatRoom.Cookie(),
		chatRoom.Exchange(),
		chatRoom.Name(),
		chatRoom.createTime,
		chatRoom.Creator().String(),
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
		SELECT created, creator, name
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
			exchange: exchange,
		}
		var creator string
		if err := rows.Scan(&cr.createTime, &creator, &cr.name); err != nil {
			return nil, err
		}
		cr.creator = NewIdentScreenName(creator)
		users = append(users, cr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateDisplayScreenName updates the user's DisplayScreenName
func (f SQLiteUserStore) UpdateDisplayScreenName(displayScreenName DisplayScreenName) error {
	q := `
		UPDATE users
		SET displayScreenName = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q, displayScreenName.String(), displayScreenName.IdentScreenName().String())
	return err
}

// UpdateEmailAddress updates the user's EmailAddress
func (f SQLiteUserStore) UpdateEmailAddress(emailAddress *mail.Address, screenName IdentScreenName) error {
	q := `
		UPDATE users
		SET emailAddress = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q, emailAddress.Address, screenName.String())
	return err
}

// EmailAddressByName retrieves the user's EmailAddress
func (f SQLiteUserStore) EmailAddressByName(screenName IdentScreenName) (*mail.Address, error) {
	q := `
		SELECT emailAddress
		FROM users
		WHERE identScreenName = ?
	`
	var emailAddress string
	err := f.db.QueryRow(q, screenName.String()).Scan(&emailAddress)
	// username isn't found for some reason
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	e, err := mail.ParseAddress(emailAddress)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoEmailAddress, err)
	}
	return e, nil
}

// UpdateRegStatus updates the user's registration status preference
func (f SQLiteUserStore) UpdateRegStatus(regStatus uint16, screenName IdentScreenName) error {
	q := `
		UPDATE users
		SET regStatus = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q, regStatus, screenName.String())
	return err
}

// RegStatusByName retrieves the user's registration status preference
func (f SQLiteUserStore) RegStatusByName(screenName IdentScreenName) (uint16, error) {
	q := `
		SELECT regStatus
		FROM users
		WHERE identScreenName = ?
	`
	var regStatus uint16
	err := f.db.QueryRow(q, screenName.String()).Scan(&regStatus)
	// username isn't found for some reason
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return regStatus, nil
}

// UpdateConfirmStatus updates the user's confirmation status
func (f SQLiteUserStore) UpdateConfirmStatus(confirmStatus bool, screenName IdentScreenName) error {
	q := `
		UPDATE users
		SET confirmStatus = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q, confirmStatus, screenName.String())
	return err
}

// ConfirmStatusByName retrieves the user's confirmation status
func (f SQLiteUserStore) ConfirmStatusByName(screenName IdentScreenName) (bool, error) {
	q := `
		SELECT confirmStatus
		FROM users
		WHERE identScreenName = ?
	`
	var confirmStatus bool
	err := f.db.QueryRow(q, screenName.String()).Scan(&confirmStatus)
	// username isn't found for some reason
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return confirmStatus, nil
}

func (f SQLiteUserStore) SetWorkInfo(name IdentScreenName, data ICQWorkInfo) error {
	q := `
		UPDATE users SET 
			company = ?,
			department = ?,
			occupationCode = ?,
			position = ?,
			workAddress = ?,
			workCity = ?,
			workCountryCode = ?,
			workFax = ?,
			workPhone = ?,
			workState = ?,
			workWebPage = ?,
			workZIP = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q,
		data.Company,
		data.Department,
		data.OccupationCode,
		data.Position,
		data.WorkAddress,
		data.WorkCity,
		data.WorkCountryCode,
		data.WorkFax,
		data.WorkPhone,
		data.WorkState,
		data.WorkWebPage,
		data.WorkZIP,
		name.String(),
	)
	return err
}

func (f SQLiteUserStore) SetMoreInfo(name IdentScreenName, data ICQMoreInfo) error {
	q := `
		UPDATE users SET 
			birthDay = ?,
			birthMonth = ?,
			birthYear = ?,
			gender = ?,
			homePageAddr = ?,
			lang1 = ?,
			lang2 = ?,
			lang3 = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q,
		data.BirthDay,
		data.BirthMonth,
		data.BirthYear,
		data.Gender,
		data.HomePageAddr,
		data.Lang1,
		data.Lang2,
		data.Lang3,
		name.String(),
	)
	return err
}

func (f SQLiteUserStore) SetUserNotes(name IdentScreenName, data ICQUserNotes) error {
	q := `
		UPDATE users
		SET notes = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q,
		data.Notes,
		name.String(),
	)
	return err
}

func (f SQLiteUserStore) SetInterests(name IdentScreenName, data ICQInterests) error {
	q := `
		UPDATE users SET 
			interest1Code = ?,
			interest1Keyword = ?,
			interest2Code = ?,
			interest2Keyword = ?,
			interest3Code = ?,
			interest3Keyword = ?,
			interest4Code = ?,
			interest4Keyword = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q,
		data.Interest1Code,
		data.Interest1Keyword,
		data.Interest2Code,
		data.Interest2Keyword,
		data.Interest3Code,
		data.Interest3Keyword,
		data.Interest4Code,
		data.Interest4Keyword,
		name.String(),
	)
	return err
}

func (f SQLiteUserStore) SetAffiliations(name IdentScreenName, data ICQAffiliations) error {
	q := `
		UPDATE users SET 
			affiliations1Code = ?,
			affiliations1Keyword = ?,
			affiliations2Code = ?,
			affiliations2Keyword = ?,
			affiliations3Code = ?,
			affiliations3Keyword = ?,
			pastAffiliations1Code = ?,
			pastAffiliations1Keyword = ?,
			pastAffiliations2Code = ?,
			pastAffiliations2Keyword = ?,
			pastAffiliations3Code = ?,
			pastAffiliations3Keyword = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q,
		data.Affiliations1Code,
		data.Affiliations1Keyword,
		data.Affiliations2Code,
		data.Affiliations2Keyword,
		data.Affiliations3Code,
		data.Affiliations3Keyword,
		data.PastAffiliations1Code,
		data.PastAffiliations1Keyword,
		data.PastAffiliations2Code,
		data.PastAffiliations2Keyword,
		data.PastAffiliations3Code,
		data.PastAffiliations3Keyword,
		name.String(),
	)
	return err
}

func (f SQLiteUserStore) SetBasicInfo(name IdentScreenName, data ICQUserInfoBasic) error {
	q := `
		UPDATE users SET 
			cellPhone = ?,
			countryCode = ?,
			emailAddress = ?,
			firstName = ?,
			gmtOffset = ?,
			homeAddress = ?,
			homeCity = ?,
			homeFax = ?,
			homePhone = ?,
			homeState = ?,
			lastName = ?,
			nickName = ?,
			publishEmail = ?,
			zipCode = ?
		WHERE identScreenName = ?
	`
	_, err := f.db.Exec(q,
		data.CellPhone,
		data.CountryCode,
		data.EmailAddress,
		data.FirstName,
		data.GMTOffset,
		data.HomeAddress,
		data.HomeCity,
		data.HomeFax,
		data.HomePhone,
		data.HomeState,
		data.LastName,
		data.Nickname,
		data.PublishEmail,
		data.ZipCode,
		name.String(),
	)
	return err
}
