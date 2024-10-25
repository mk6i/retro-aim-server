package state

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"modernc.org/sqlite"
	lib "modernc.org/sqlite/lib"

	"github.com/mk6i/retro-aim-server/wire"
)

var (
	ErrKeywordCategoryExists   = errors.New("keyword category already exists")
	ErrKeywordCategoryNotFound = errors.New("keyword category not found")
	ErrKeywordExists           = errors.New("keyword already exists")
	ErrKeywordInUse            = errors.New("can't delete keyword that is associated with a user")
	ErrKeywordNotFound         = errors.New("keyword not found")
	errTooManyCategories       = errors.New("there are too many keyword categories")
	errTooManyKeywords         = errors.New("there are too many keywords")
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
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys=on", dbFilePath))
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open connections to 1.
	// This is crucial to prevent SQLITE_BUSY errors, which occur when the database
	// is locked due to concurrent access. By limiting the number of open connections
	// to 1, we ensure that all database operations are serialized, thus avoiding
	// any potential locking issues.
	db.SetMaxOpenConns(1)

	store := &SQLiteUserStore{db: db}

	if err := store.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
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

	driver, err := migratesqlite.WithInstance(f.db, &migratesqlite.Config{})
	if err != nil {
		return fmt.Errorf("cannot create database driver: %v", err)
	}

	m, err := migrate.NewWithInstance("httpfs", sourceInstance, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

// AllUsers returns all stored users. It only populates the following fields:
// - IdentScreenName
// - DisplayScreenName
// - IsICQ
func (f SQLiteUserStore) AllUsers() ([]User, error) {
	q := `SELECT identScreenName, displayScreenName, isICQ FROM users`
	rows, err := f.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var identSN, displaySN string
		var isICQ bool
		if err := rows.Scan(&identSN, &displaySN, &isICQ); err != nil {
			return nil, err
		}
		users = append(users, User{
			IdentScreenName:   NewIdentScreenName(identSN),
			DisplayScreenName: DisplayScreenName(displaySN),
			IsICQ:             isICQ,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// FindByUIN returns a user with a matching UIN.
func (f SQLiteUserStore) FindByUIN(UIN uint32) (User, error) {
	users, err := f.queryUsers(`identScreenName = ?`, []any{strconv.Itoa(int(UIN))})
	if err != nil {
		return User{}, fmt.Errorf("FindByUIN: %w", err)
	}

	if len(users) == 0 {
		return User{}, ErrNoUser
	}

	return users[0], nil
}

// FindByICQEmail returns a user with a matching email address.
func (f SQLiteUserStore) FindByICQEmail(email string) (User, error) {
	users, err := f.queryUsers(`icq_basicInfo_emailAddress = ?`, []any{email})
	if err != nil {
		return User{}, fmt.Errorf("FindByICQEmail: %w", err)
	}

	if len(users) == 0 {
		return User{}, ErrNoUser
	}

	return users[0], nil
}

// FindByAIMEmail returns a user with a matching email address.
func (f SQLiteUserStore) FindByAIMEmail(email string) (User, error) {
	users, err := f.queryUsers(`emailAddress = ?`, []any{email})
	if err != nil {
		return User{}, fmt.Errorf("FindByAIMEmail: %w", err)
	}

	if len(users) == 0 {
		return User{}, ErrNoUser
	}

	return users[0], nil
}

// FindByAIMKeyword returns users who have a matching keyword.
func (f SQLiteUserStore) FindByAIMKeyword(keyword string) ([]User, error) {
	where := `
		(SELECT id FROM aimKeyword WHERE name = ?) IN
		(aim_keyword1, aim_keyword2, aim_keyword3, aim_keyword4, aim_keyword5)
	`
	users, err := f.queryUsers(where, []any{keyword})
	if err != nil {
		return nil, err
	}

	return users, nil
}

// FindByICQName returns users with matching first name, last name, and
// nickname. Empty values are not included in the search parameters.
func (f SQLiteUserStore) FindByICQName(firstName, lastName, nickName string) ([]User, error) {
	var args []any
	var clauses []string

	if firstName != "" {
		args = append(args, firstName)
		clauses = append(clauses, `LOWER(icq_basicInfo_firstName) = LOWER(?)`)
	}

	if lastName != "" {
		args = append(args, lastName)
		clauses = append(clauses, `LOWER(icq_basicInfo_lastName) = LOWER(?)`)
	}

	if nickName != "" {
		args = append(args, nickName)
		clauses = append(clauses, `LOWER(icq_basicInfo_nickName) = LOWER(?)`)
	}

	whereClause := strings.Join(clauses, " AND ")

	users, err := f.queryUsers(whereClause, args)
	if err != nil {
		err = fmt.Errorf("FindByICQName: %w", err)
	}

	return users, nil
}

// FindByAIMNameAndAddr returns users with all matching non-empty directory info
// fields. Empty values are not included in the search parameters.
func (f SQLiteUserStore) FindByAIMNameAndAddr(info AIMNameAndAddr) ([]User, error) {
	var args []any
	var clauses []string

	if info.FirstName != "" {
		args = append(args, info.FirstName)
		clauses = append(clauses, `LOWER(aim_firstName) = LOWER(?)`)
	}

	if info.LastName != "" {
		args = append(args, info.LastName)
		clauses = append(clauses, `LOWER(aim_lastName) = LOWER(?)`)
	}

	if info.MiddleName != "" {
		args = append(args, info.MiddleName)
		clauses = append(clauses, `LOWER(aim_middleName) = LOWER(?)`)
	}

	if info.MaidenName != "" {
		args = append(args, info.MaidenName)
		clauses = append(clauses, `LOWER(aim_maidenName) = LOWER(?)`)
	}

	if info.Country != "" {
		args = append(args, info.Country)
		clauses = append(clauses, `LOWER(aim_country) = LOWER(?)`)
	}

	if info.State != "" {
		args = append(args, info.State)
		clauses = append(clauses, `LOWER(aim_state) = LOWER(?)`)
	}

	if info.City != "" {
		args = append(args, info.City)
		clauses = append(clauses, `LOWER(aim_city) = LOWER(?)`)
	}

	if info.NickName != "" {
		args = append(args, info.NickName)
		clauses = append(clauses, `LOWER(aim_nickName) = LOWER(?)`)
	}

	if info.ZIPCode != "" {
		args = append(args, info.ZIPCode)
		clauses = append(clauses, `LOWER(aim_zipCode) = LOWER(?)`)
	}

	if info.Address != "" {
		args = append(args, info.Address)
		clauses = append(clauses, `LOWER(aim_address) = LOWER(?)`)
	}

	whereClause := strings.Join(clauses, " AND ")

	users, err := f.queryUsers(whereClause, args)
	if err != nil {
		err = fmt.Errorf("FindByAIMNameAndAddr: %w", err)
	}

	return users, nil
}

// FindByICQInterests returns users who have at least one matching interest.
func (f SQLiteUserStore) FindByICQInterests(code uint16, keywords []string) ([]User, error) {
	var args []any
	var clauses []string

	for i := 1; i <= 4; i++ {
		var subClauses []string
		args = append(args, code)
		for _, key := range keywords {
			subClauses = append(subClauses, fmt.Sprintf("icq_interests_keyword%d LIKE ?", i))
			args = append(args, "%"+key+"%")
		}
		clauses = append(clauses, fmt.Sprintf("(icq_interests_code%d = ? AND (%s))", i, strings.Join(subClauses, " OR ")))
	}

	cond := strings.Join(clauses, " OR ")

	users, err := f.queryUsers(cond, args)
	if err != nil {
		err = fmt.Errorf("FindByICQInterests: %w", err)
	}

	return users, nil
}

// FindByICQKeyword returns users with matching interest keyword across all
// interest categories.
func (f SQLiteUserStore) FindByICQKeyword(keyword string) ([]User, error) {
	var args []any
	var clauses []string

	for i := 1; i <= 4; i++ {
		args = append(args, "%"+keyword+"%")
		clauses = append(clauses, fmt.Sprintf("icq_interests_keyword%d LIKE ?", i))
	}

	whereClause := strings.Join(clauses, " OR ")

	users, err := f.queryUsers(whereClause, args)
	if err != nil {
		err = fmt.Errorf("FindByICQKeyword: %w", err)
	}

	return users, nil
}

// User looks up a user by screen name. It populates the User record with
// credentials that can be used to validate the user's password.
func (f SQLiteUserStore) User(screenName IdentScreenName) (*User, error) {
	users, err := f.queryUsers(`identScreenName = ?`, []any{screenName.String()})
	if err != nil {
		return nil, fmt.Errorf("User: %w", err)
	}

	if len(users) == 0 {
		return nil, nil
	}

	return &users[0], nil
}

// queryUsers retrieves a list of users from the database based on the
// specified WHERE clause and query parameters. Returns a slice of User objects
// or an error if the query fails.
func (f SQLiteUserStore) queryUsers(whereClause string, queryParams []any) ([]User, error) {
	q := `
		SELECT
			identScreenName,
			displayScreenName,
			emailAddress,
			authKey,
			strongMD5Pass,
			weakMD5Pass,
			confirmStatus,
			regStatus,
			isICQ,
			icq_affiliations_currentCode1,
			icq_affiliations_currentCode2,
			icq_affiliations_currentCode3,
			icq_affiliations_currentKeyword1,
			icq_affiliations_currentKeyword2,
			icq_affiliations_currentKeyword3,
			icq_affiliations_pastCode1,
			icq_affiliations_pastCode2,
			icq_affiliations_pastCode3,
			icq_affiliations_pastKeyword1,
			icq_affiliations_pastKeyword2,
			icq_affiliations_pastKeyword3,
			icq_basicInfo_address,
			icq_basicInfo_cellPhone,
			icq_basicInfo_city,
			icq_basicInfo_countryCode,
			icq_basicInfo_emailAddress,
			icq_basicInfo_fax,
			icq_basicInfo_firstName,
			icq_basicInfo_gmtOffset,
			icq_basicInfo_lastName,
			icq_basicInfo_nickName,
			icq_basicInfo_phone,
			icq_basicInfo_publishEmail,
			icq_basicInfo_state,
			icq_basicInfo_zipCode,
			icq_interests_code1,
			icq_interests_code2,
			icq_interests_code3,
			icq_interests_code4,
			icq_interests_keyword1,
			icq_interests_keyword2,
			icq_interests_keyword3,
			icq_interests_keyword4,
			icq_moreInfo_birthDay,
			icq_moreInfo_birthMonth,
			icq_moreInfo_birthYear,
			icq_moreInfo_gender,
			icq_moreInfo_homePageAddr,
			icq_moreInfo_lang1,
			icq_moreInfo_lang2,
			icq_moreInfo_lang3,
			icq_notes,
			icq_permissions_authRequired,
			icq_workInfo_address,
			icq_workInfo_city,
			icq_workInfo_company,
			icq_workInfo_countryCode,
			icq_workInfo_department,
			icq_workInfo_fax,
			icq_workInfo_occupationCode,
			icq_workInfo_phone,
			icq_workInfo_position,
			icq_workInfo_state,
			icq_workInfo_webPage,
			icq_workInfo_zipCode,
			aim_firstName,
			aim_lastName,
			aim_middleName,
			aim_maidenName,
			aim_country,
			aim_state,
			aim_city,
			aim_nickName,
			aim_zipCode,
			aim_address
		FROM users
		WHERE %s
	`
	q = fmt.Sprintf(q, whereClause)
	rows, err := f.db.Query(q, queryParams...)
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
			&u.DisplayScreenName,
			&u.EmailAddress,
			&u.AuthKey,
			&u.StrongMD5Pass,
			&u.WeakMD5Pass,
			&u.ConfirmStatus,
			&u.RegStatus,
			&u.IsICQ,
			&u.ICQAffiliations.CurrentCode1,
			&u.ICQAffiliations.CurrentCode2,
			&u.ICQAffiliations.CurrentCode3,
			&u.ICQAffiliations.CurrentKeyword1,
			&u.ICQAffiliations.CurrentKeyword2,
			&u.ICQAffiliations.CurrentKeyword3,
			&u.ICQAffiliations.PastCode1,
			&u.ICQAffiliations.PastCode2,
			&u.ICQAffiliations.PastCode3,
			&u.ICQAffiliations.PastKeyword1,
			&u.ICQAffiliations.PastKeyword2,
			&u.ICQAffiliations.PastKeyword3,
			&u.ICQBasicInfo.Address,
			&u.ICQBasicInfo.CellPhone,
			&u.ICQBasicInfo.City,
			&u.ICQBasicInfo.CountryCode,
			&u.ICQBasicInfo.EmailAddress,
			&u.ICQBasicInfo.Fax,
			&u.ICQBasicInfo.FirstName,
			&u.ICQBasicInfo.GMTOffset,
			&u.ICQBasicInfo.LastName,
			&u.ICQBasicInfo.Nickname,
			&u.ICQBasicInfo.Phone,
			&u.ICQBasicInfo.PublishEmail,
			&u.ICQBasicInfo.State,
			&u.ICQBasicInfo.ZIPCode,
			&u.ICQInterests.Code1,
			&u.ICQInterests.Code2,
			&u.ICQInterests.Code3,
			&u.ICQInterests.Code4,
			&u.ICQInterests.Keyword1,
			&u.ICQInterests.Keyword2,
			&u.ICQInterests.Keyword3,
			&u.ICQInterests.Keyword4,
			&u.ICQMoreInfo.BirthDay,
			&u.ICQMoreInfo.BirthMonth,
			&u.ICQMoreInfo.BirthYear,
			&u.ICQMoreInfo.Gender,
			&u.ICQMoreInfo.HomePageAddr,
			&u.ICQMoreInfo.Lang1,
			&u.ICQMoreInfo.Lang2,
			&u.ICQMoreInfo.Lang3,
			&u.ICQNotes.Notes,
			&u.ICQPermissions.AuthRequired,
			&u.ICQWorkInfo.Address,
			&u.ICQWorkInfo.City,
			&u.ICQWorkInfo.Company,
			&u.ICQWorkInfo.CountryCode,
			&u.ICQWorkInfo.Department,
			&u.ICQWorkInfo.Fax,
			&u.ICQWorkInfo.OccupationCode,
			&u.ICQWorkInfo.Phone,
			&u.ICQWorkInfo.Position,
			&u.ICQWorkInfo.State,
			&u.ICQWorkInfo.WebPage,
			&u.ICQWorkInfo.ZIPCode,
			&u.AIMDirectoryInfo.FirstName,
			&u.AIMDirectoryInfo.LastName,
			&u.AIMDirectoryInfo.MiddleName,
			&u.AIMDirectoryInfo.MaidenName,
			&u.AIMDirectoryInfo.Country,
			&u.AIMDirectoryInfo.State,
			&u.AIMDirectoryInfo.City,
			&u.AIMDirectoryInfo.NickName,
			&u.AIMDirectoryInfo.ZIPCode,
			&u.AIMDirectoryInfo.Address,
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
	if u.DisplayScreenName.IsUIN() && !u.IsICQ {
		return errors.New("inserting user with UIN and isICQ=false")
	}
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

// SetDirectoryInfo sets an AIM user's directory information.
func (f SQLiteUserStore) SetDirectoryInfo(screenName IdentScreenName, info AIMNameAndAddr) error {
	q := `
		UPDATE users SET 
			aim_firstName = ?,
			aim_lastName = ?,
			aim_middleName = ?,
			aim_maidenName = ?,
			aim_country = ?,
			aim_state = ?,
			aim_city = ?,
			aim_nickName = ?,
			aim_zipCode = ?,
			aim_address = ?
		WHERE identScreenName = ?
	`
	res, err := f.db.Exec(q,
		info.FirstName,
		info.LastName,
		info.MiddleName,
		info.MaidenName,
		info.Country,
		info.State,
		info.City,
		info.NickName,
		info.ZIPCode,
		info.Address,
		screenName.String(),
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if c == 0 {
		return ErrNoUser
	}

	return nil
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
		WHERE lower(cookie) = lower(?)
	`
	var creator string
	err := f.db.QueryRow(q, cookie).Scan(
		&chatRoom.exchange,
		&chatRoom.name,
		&chatRoom.createTime,
		&creator,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("%w: %s", ErrChatRoomNotFound, cookie)
	}
	chatRoom.creator = NewIdentScreenName(creator)

	return chatRoom, err
}

// ChatRoomByName looks up a chat room by exchange and name. Returns
// ErrChatRoomNotFound if the room does not exist for exchange and name.
func (f SQLiteUserStore) ChatRoomByName(exchange uint16, name string) (ChatRoom, error) {
	chatRoom := ChatRoom{
		exchange: exchange,
	}

	q := `
		SELECT name, created, creator
		FROM chatRoom
		WHERE exchange = ? AND lower(name) = lower(?)
	`
	var creator string
	err := f.db.QueryRow(q, exchange, name).Scan(
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
		if strings.Contains(err.Error(), "constraint failed") {
			err = ErrDupChatRoom
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

// SetWorkInfo updates the work-related information for an ICQ user.
func (f SQLiteUserStore) SetWorkInfo(name IdentScreenName, data ICQWorkInfo) error {
	q := `
		UPDATE users SET 
			icq_workInfo_company = ?,
			icq_workInfo_department = ?,
			icq_workInfo_occupationCode = ?,
			icq_workInfo_position = ?,
			icq_workInfo_address = ?,
			icq_workInfo_city = ?,
			icq_workInfo_countryCode = ?,
			icq_workInfo_fax = ?,
			icq_workInfo_phone = ?,
			icq_workInfo_state = ?,
			icq_workInfo_webPage = ?,
			icq_workInfo_zipCode = ?
		WHERE identScreenName = ?
	`
	res, err := f.db.Exec(q,
		data.Company,
		data.Department,
		data.OccupationCode,
		data.Position,
		data.Address,
		data.City,
		data.CountryCode,
		data.Fax,
		data.Phone,
		data.State,
		data.WebPage,
		data.ZIPCode,
		name.String(),
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if c == 0 {
		return ErrNoUser
	}
	return nil
}

func (f SQLiteUserStore) SetMoreInfo(name IdentScreenName, data ICQMoreInfo) error {
	q := `
		UPDATE users SET 
			icq_moreInfo_birthDay = ?,
			icq_moreInfo_birthMonth = ?,
			icq_moreInfo_birthYear = ?,
			icq_moreInfo_gender = ?,
			icq_moreInfo_homePageAddr = ?,
			icq_moreInfo_lang1 = ?,
			icq_moreInfo_lang2 = ?,
			icq_moreInfo_lang3 = ?
		WHERE identScreenName = ?
	`
	res, err := f.db.Exec(q,
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
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if c == 0 {
		return ErrNoUser
	}
	return nil
}

func (f SQLiteUserStore) SetUserNotes(name IdentScreenName, data ICQUserNotes) error {
	q := `
		UPDATE users
		SET icq_notes = ?
		WHERE identScreenName = ?
	`
	res, err := f.db.Exec(q,
		data.Notes,
		name.String(),
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if c == 0 {
		return ErrNoUser
	}
	return nil
}

func (f SQLiteUserStore) SetInterests(name IdentScreenName, data ICQInterests) error {
	q := `
		UPDATE users SET 
			icq_interests_code1 = ?,
			icq_interests_keyword1 = ?,
			icq_interests_code2 = ?,
			icq_interests_keyword2 = ?,
			icq_interests_code3 = ?,
			icq_interests_keyword3 = ?,
			icq_interests_code4 = ?,
			icq_interests_keyword4 = ?
		WHERE identScreenName = ?
	`
	res, err := f.db.Exec(q,
		data.Code1,
		data.Keyword1,
		data.Code2,
		data.Keyword2,
		data.Code3,
		data.Keyword3,
		data.Code4,
		data.Keyword4,
		name.String(),
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if c == 0 {
		return ErrNoUser
	}
	return nil
}

func (f SQLiteUserStore) SetAffiliations(name IdentScreenName, data ICQAffiliations) error {
	q := `
		UPDATE users SET 
			icq_affiliations_currentCode1 = ?,
			icq_affiliations_currentKeyword1 = ?,
			icq_affiliations_currentCode2 = ?,
			icq_affiliations_currentKeyword2 = ?,
			icq_affiliations_currentCode3 = ?,
			icq_affiliations_currentKeyword3 = ?,
			icq_affiliations_pastCode1 = ?,
			icq_affiliations_pastKeyword1 = ?,
			icq_affiliations_pastCode2 = ?,
			icq_affiliations_pastKeyword2 = ?,
			icq_affiliations_pastCode3 = ?,
			icq_affiliations_pastKeyword3 = ?
		WHERE identScreenName = ?
	`
	res, err := f.db.Exec(q,
		data.CurrentCode1,
		data.CurrentKeyword1,
		data.CurrentCode2,
		data.CurrentKeyword2,
		data.CurrentCode3,
		data.CurrentKeyword3,
		data.PastCode1,
		data.PastKeyword1,
		data.PastCode2,
		data.PastKeyword2,
		data.PastCode3,
		data.PastKeyword3,
		name.String(),
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if c == 0 {
		return ErrNoUser
	}
	return nil
}

func (f SQLiteUserStore) SetBasicInfo(name IdentScreenName, data ICQBasicInfo) error {
	q := `
		UPDATE users SET 
			icq_basicInfo_cellPhone = ?,
			icq_basicInfo_countryCode = ?,
			icq_basicInfo_emailAddress = ?,
			icq_basicInfo_firstName = ?,
			icq_basicInfo_gmtOffset = ?,
			icq_basicInfo_address = ?,
			icq_basicInfo_city = ?,
			icq_basicInfo_fax = ?,
			icq_basicInfo_phone = ?,
			icq_basicInfo_state = ?,
			icq_basicInfo_lastName = ?,
			icq_basicInfo_nickName = ?,
			icq_basicInfo_publishEmail = ?,
			icq_basicInfo_zipCode = ?
		WHERE identScreenName = ?
	`
	res, err := f.db.Exec(q,
		data.CellPhone,
		data.CountryCode,
		data.EmailAddress,
		data.FirstName,
		data.GMTOffset,
		data.Address,
		data.City,
		data.Fax,
		data.Phone,
		data.State,
		data.LastName,
		data.Nickname,
		data.PublishEmail,
		data.ZIPCode,
		name.String(),
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if c == 0 {
		return ErrNoUser
	}
	return nil
}

// SaveMessage saves an offline message for later retrieval.
func (f SQLiteUserStore) SaveMessage(offlineMessage OfflineMessage) error {
	buf := &bytes.Buffer{}
	if err := wire.MarshalBE(offlineMessage.Message, buf); err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	q := `
		INSERT INTO offlineMessage (sender, recipient, message, sent)
		VALUES (?, ?, ?, ?)
	`
	_, err := f.db.Exec(
		q,
		offlineMessage.Sender.String(),
		offlineMessage.Recipient.String(),
		buf.Bytes(),
		offlineMessage.Sent,
	)
	return err
}

// RetrieveMessages retrieves all offline messages sent to recipient.
func (f SQLiteUserStore) RetrieveMessages(recip IdentScreenName) ([]OfflineMessage, error) {
	q := `
		SELECT 
		    sender, 
		    message,
		    sent
		FROM offlineMessage
		WHERE recipient = ?
	`
	rows, err := f.db.Query(q, recip.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []OfflineMessage

	for rows.Next() {
		var sender string
		var buf []byte
		var sent time.Time
		if err := rows.Scan(&sender, &buf, &sent); err != nil {
			return nil, err
		}

		var msg wire.SNAC_0x04_0x06_ICBMChannelMsgToHost
		if err := wire.UnmarshalBE(&msg, bytes.NewBuffer(buf)); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		messages = append(messages, OfflineMessage{
			Sender:    NewIdentScreenName(sender),
			Recipient: recip,
			Message:   msg,
			Sent:      sent,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// DeleteMessages deletes all offline messages sent to recipient.
func (f SQLiteUserStore) DeleteMessages(recip IdentScreenName) error {
	q := `
		DELETE FROM offlineMessage WHERE recipient = ?
	`
	_, err := f.db.Exec(q, recip.String())
	return err
}

// BuddyIconRefByName retrieves the buddy icon reference for a given user
func (f SQLiteUserStore) BuddyIconRefByName(screenName IdentScreenName) (*wire.BARTID, error) {
	q := `
		SELECT
			groupID,
			itemID,
			classID,
			name,
			attributes
		FROM feedBag
		WHERE screenname = ? AND name = ? AND classID = ?
	`
	var item wire.FeedbagItem
	var attrs []byte
	err := f.db.QueryRow(q, screenName.String(), wire.BARTTypesBuddyIcon, wire.FeedbagClassIdBart).Scan(&item.GroupID, &item.ItemID, &item.ClassID, &item.Name, &attrs)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := wire.UnmarshalBE(&item.TLVLBlock, bytes.NewBuffer(attrs)); err != nil {
		return nil, err
	}
	b, hasBuf := item.Bytes(wire.FeedbagAttributesBartInfo)
	if !hasBuf {
		return nil, errors.New("unable to extract icon payload")
	}
	bartInfo := wire.BARTInfo{}
	if err := wire.UnmarshalBE(&bartInfo, bytes.NewBuffer(b)); err != nil {
		return nil, err
	}
	return &wire.BARTID{
		Type: wire.BARTTypesBuddyIcon,
		BARTInfo: wire.BARTInfo{
			Flags: bartInfo.Flags,
			Hash:  bartInfo.Hash,
		},
	}, nil

}

func (f SQLiteUserStore) SetKeywords(name IdentScreenName, keywords [5]string) error {
	q := `
		WITH interests AS (SELECT CASE WHEN name = ? THEN id ELSE NULL END AS aim_keyword1,
								  CASE WHEN name = ? THEN id ELSE NULL END AS aim_keyword2,
								  CASE WHEN name = ? THEN id ELSE NULL END AS aim_keyword3,
								  CASE WHEN name = ? THEN id ELSE NULL END AS aim_keyword4,
								  CASE WHEN name = ? THEN id ELSE NULL END AS aim_keyword5
						   FROM aimKeyword
						   WHERE name IN (?, ?, ?, ?, ?))
		UPDATE users
		SET aim_keyword1 = (SELECT aim_keyword1 FROM interests WHERE aim_keyword1 IS NOT NULL),
			aim_keyword2 = (SELECT aim_keyword2 FROM interests WHERE aim_keyword2 IS NOT NULL),
			aim_keyword3 = (SELECT aim_keyword3 FROM interests WHERE aim_keyword3 IS NOT NULL),
			aim_keyword4 = (SELECT aim_keyword4 FROM interests WHERE aim_keyword4 IS NOT NULL),
			aim_keyword5 = (SELECT aim_keyword5 FROM interests WHERE aim_keyword5 IS NOT NULL)
		WHERE identScreenName = ?
	`

	_, err := f.db.Exec(q,
		keywords[0], keywords[1], keywords[2], keywords[3], keywords[4],
		keywords[0], keywords[1], keywords[2], keywords[3], keywords[4],
		name.String())
	return err
}

// Categories returns a list of keyword categories.
func (f SQLiteUserStore) Categories() ([]Category, error) {
	q := `SELECT id, name FROM aimKeywordCategory ORDER BY name`

	rows, err := f.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		category := Category{}
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// CreateCategory creates a new keyword category.
func (f SQLiteUserStore) CreateCategory(name string) (Category, error) {
	tx, err := f.db.Begin()
	if err != nil {
		return Category{}, err
	}

	defer tx.Rollback()

	q := `INSERT INTO aimKeywordCategory (name) VALUES (?)`
	res, err := tx.Exec(q, name)
	if err != nil {
		if sqliteErr, ok := err.(*sqlite.Error); ok && sqliteErr.Code() == lib.SQLITE_CONSTRAINT_UNIQUE {
			err = ErrKeywordCategoryExists
		}
		return Category{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Category{}, err
	}

	if id > math.MaxUint8 {
		return Category{}, errTooManyCategories
	}

	if err := tx.Commit(); err != nil {
		return Category{}, err
	}

	return Category{
		ID:   uint8(id),
		Name: name,
	}, nil
}

// DeleteCategory deletes a keyword category and all of its associated
// keywords.
func (f SQLiteUserStore) DeleteCategory(categoryID uint8) error {
	q := `DELETE FROM aimKeywordCategory WHERE id = ?`
	res, err := f.db.Exec(q, categoryID)
	if err != nil {
		// Check if the error is a foreign key constraint violation
		if sqliteErr, ok := err.(*sqlite.Error); ok && sqliteErr.Code() == lib.SQLITE_CONSTRAINT_FOREIGNKEY {
			return ErrKeywordInUse
		}
	}

	c, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if c == 0 {
		return ErrKeywordCategoryNotFound
	}

	return nil
}

// KeywordsByCategory returns all keywords for a given category.
func (f SQLiteUserStore) KeywordsByCategory(categoryID uint8) ([]Keyword, error) {
	q := `SELECT id, name FROM aimKeyword WHERE parent = ? ORDER BY name`
	if categoryID == 0 {
		q = `SELECT id, name FROM aimKeyword WHERE parent IS NULL ORDER BY name`
	}

	rows, err := f.db.Query(q, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []Keyword
	for rows.Next() {
		keyword := Keyword{}
		if err := rows.Scan(&keyword.ID, &keyword.Name); err != nil {
			return nil, err
		}
		keywords = append(keywords, keyword)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(keywords) == 0 {
		var exists int
		err = f.db.QueryRow("SELECT COUNT(*) FROM aimKeywordCategory WHERE id = ?", categoryID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, ErrKeywordCategoryNotFound
		}
	}

	return keywords, nil
}

// CreateKeyword creates a new keyword. If categoryID is 0, it has no category.
func (f SQLiteUserStore) CreateKeyword(name string, categoryID uint8) (Keyword, error) {
	tx, err := f.db.Begin()
	if err != nil {
		return Keyword{}, err
	}

	defer tx.Rollback()

	q := `INSERT INTO aimKeyword (name, parent) VALUES (?, ?)`
	var parent interface{} = nil
	if categoryID != 0 {
		parent = categoryID
	}

	res, err := tx.Exec(q, name, parent)
	if err != nil {
		if sqliteErr, ok := err.(*sqlite.Error); ok && sqliteErr.Code() == lib.SQLITE_CONSTRAINT_UNIQUE {
			err = ErrKeywordExists
		} else if sqliteErr, ok := err.(*sqlite.Error); ok && sqliteErr.Code() == lib.SQLITE_CONSTRAINT_FOREIGNKEY {
			err = ErrKeywordCategoryNotFound
		}
		return Keyword{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Keyword{}, err
	}

	if id > math.MaxUint8 {
		return Keyword{}, errTooManyKeywords
	}

	if err := tx.Commit(); err != nil {
		return Keyword{}, err
	}

	return Keyword{
		ID:   uint8(id),
		Name: name,
	}, nil
}

// DeleteKeyword deletes a keyword.
func (f SQLiteUserStore) DeleteKeyword(id uint8) error {
	q := `DELETE FROM aimKeyword WHERE id = ?`
	res, err := f.db.Exec(q, id)

	if err != nil {
		// Check if the error is a foreign key constraint violation
		if sqliteErr, ok := err.(*sqlite.Error); ok && sqliteErr.Code() == lib.SQLITE_CONSTRAINT_FOREIGNKEY {
			return ErrKeywordInUse
		}
	}

	c, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if c == 0 {
		return ErrKeywordNotFound
	}

	return nil
}

// InterestList returns a list of keywords grouped by category used to render
// the AIM directory interests list. The list is made up of 3 types of elements:
//
// Categories
//
//	ID: The category ID
//	Name: The category name
//	Type: [wire.ODirKeywordCategory]
//
// Keywords
//
//	ID: The parent category ID
//	Name: The keyword name
//	Type: [wire.ODirKeyword]
//
// Top-level Keywords
//
//	ID: 0 (does not have a parent category)
//	Name: The keyword name
//	Type: [wire.ODirKeyword]
//
// Keywords are grouped contiguously by category and preceded by the category
// name. Top-level keywords appear by themselves. Categories and top-level
// keywords are sorted alphabetically. Keyword groups are sorted alphabetically.
//
// Conceptually, the list looks like this:
//
//	> Animals (top-level keyword, id=0)
//	> Artificial Intelligence (keyword, id=3)
//		> Cybersecurity (keyword, id=3)
//	> Music (category, id=1)
//		> Jazz (keyword, id=1)
//		> Rock (keyword, id=1)
//	> Sports (category, id=2)
//		> Basketball (keyword, id=2)
//		> Soccer (keyword, id=2)
//		> Tennis (keyword, id=2)
//	> Technology (category, id=3)
//	> Zoology (top-level keyword, id=0)
func (f SQLiteUserStore) InterestList() ([]wire.ODirKeywordListItem, error) {
	q := `
		WITH categories AS (
			SELECT
				name AS grouping,
				id,
				0 AS sortPrio,
				name
			FROM aimKeywordCategory
			UNION
			SELECT
				IFNULL(akc.name, ak.name) AS grouping,
				IFNULL(ak.parent, 0) AS id,
				CASE WHEN ak.parent IS NULL THEN 1 ELSE 2 END AS sortPrio,
				ak.name
			FROM aimKeyword ak
			LEFT JOIN aimKeywordCategory akc ON akc.id = ak.parent
			ORDER BY 1, 3, 4
		)
		SELECT
			id,
			sortPrio,
			name
		FROM categories
	`

	rows, err := f.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []wire.ODirKeywordListItem
	for rows.Next() {
		msg := wire.ODirKeywordListItem{}

		var sortPrio int
		if err := rows.Scan(&msg.ID, &sortPrio, &msg.Name); err != nil {
			return nil, err
		}
		switch sortPrio {
		case 0:
			msg.Type = wire.ODirKeywordCategory
		case 1, 2:
			msg.Type = wire.ODirKeyword
		}

		list = append(list, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
