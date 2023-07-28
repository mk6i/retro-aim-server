package oscar

import (
	"bytes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	"time"
)

const file string = "/Users/mike/dev/goaim/aim.db"

var feedbagDDL = `
	CREATE TABLE IF NOT EXISTS user
	(
		ScreenName VARCHAR(16) PRIMARY KEY
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

func NewFeedbagStore(dbFile string) (*FeedbagStore, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(feedbagDDL); err != nil {
		return nil, err
	}
	return &FeedbagStore{db: db}, nil
}

type FeedbagStore struct {
	db *sql.DB
}

func (f *FeedbagStore) Retrieve(screenName string) ([]*feedbagItem, error) {
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

	var items []*feedbagItem
	for rows.Next() {
		var item feedbagItem
		var attrs []byte
		if err := rows.Scan(&item.groupID, &item.itemID, &item.classID, &item.name, &attrs); err != nil {
			return nil, err
		}
		err = item.TLVPayload.read(bytes.NewBuffer(attrs), map[uint16]reflect.Kind{
			0xC8: reflect.Slice,
			//FeedbagClassIdBuddy:            reflect.Slice,
			FeedbagClassIdGroup: reflect.Uint16,
			//FeedbagClassIdPermit:           reflect.Slice,
			//FeedbagClassIdDeny:             reflect.Slice,
			//FeedbagClassIdPdinfo:           reflect.Slice,
			//FeedbagClassIdBuddyPrefs:       reflect.Slice,
			//FeedbagClassIdNonbuddy:         reflect.Slice,
			//FeedbagClassIdTpaProvider:      reflect.Slice,
			//FeedbagClassIdTpaSubscription:  reflect.Slice,
			//FeedbagClassIdClientPrefs:      reflect.Slice,
			//FeedbagClassIdStock:            reflect.Slice,
			//FeedbagClassIdWeather:          reflect.Slice,
			//FeedbagClassIdWatchList:        reflect.Slice,
			//FeedbagClassIdIgnoreList:       reflect.Slice,
			//FeedbagClassIdDateTime:         reflect.Slice,
			//FeedbagClassIdExternalUser:     reflect.Slice,
			//FeedbagClassIdRootCreator:      reflect.Slice,
			//FeedbagClassIdFish:             reflect.Slice,
			//FeedbagClassIdImportTimestamp:  reflect.Slice,
			//FeedbagClassIdBart:             reflect.Slice,
			FeedbagClassIdRbOrder: reflect.Uint16,
			//FeedbagClassIdPersonality:      reflect.Slice,
			//FeedbagClassIdAlProf:           reflect.Slice,
			//FeedbagClassIdAlInfo:           reflect.Slice,
			//FeedbagClassIdInteraction:      reflect.Slice,
			//FeedbagClassIdVanityInfo:       reflect.Slice,
			//FeedbagClassIdFavoriteLocation: reflect.Slice,
			//FeedbagClassIdBartPdinfo:       reflect.Slice,
			//FeedbagClassIdXIcqStatusNote:   reflect.Slice,
			//FeedbagClassIdMin:              reflect.Slice,
		})
		if err != nil {
			return items, err
		}
		items = append(items, &item)
	}

	return items, nil
}

func (f *FeedbagStore) LastModified(screenName string) (time.Time, error) {
	var lastModified sql.NullInt64
	sql := `SELECT MAX(lastModified) FROM feedbag WHERE ScreenName = ?`
	err := f.db.QueryRow(sql, screenName).Scan(&lastModified)
	return time.Unix(lastModified.Int64, 0), err
}

func (f *FeedbagStore) Upsert(screenName string, items []*feedbagItem) error {

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
		if err := item.TLVPayload.write(buf); err != nil {
			return err
		}

		_, err := f.db.Exec(q,
			screenName,
			item.groupID,
			item.itemID,
			item.classID,
			item.name,
			buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FeedbagStore) InterestedUsers(screenName string) ([]string, error) {
	q := `
		SELECT name
		FROM feedbag
		WHERE ScreenName = ? AND classID = 0
	`

	rows, err := f.db.Query(q, screenName)
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

func (f *FeedbagStore) Buddies(screenName string) ([]string, error) {
	q := `
		SELECT name
		FROM feedbag
		WHERE ScreenName = ? AND classID = 0
	`

	rows, err := f.db.Query(q, screenName)
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

func (f *FeedbagStore) RetrieveProfile(screenName string) (string, error) {
	q := `
		SELECT body
		FROM profile
		WHERE ScreenName = ?
	`
	var profile string
	err := f.db.QueryRow(q, screenName).Scan(&profile)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return profile, err
}

func (f *FeedbagStore) UpsertProfile(screenName string, body string) error {
	sql := `
		INSERT INTO profile (ScreenName, body)
		VALUES (?, ?)
		ON CONFLICT (ScreenName)
			DO UPDATE SET body = excluded.body
	`
	_, err := f.db.Exec(sql, screenName, body)
	return err
}
