package state

import (
	"os"
	"reflect"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestFeedbagStore(t *testing.T) {

	const testFile string = "aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	itemsIn := []oscar.FeedbagItem{
		{
			GroupID:   0,
			ItemID:    1805,
			ClassID:   3,
			Name:      "spimmer1234",
			TLVLBlock: oscar.TLVLBlock{},
		},
		{
			GroupID: 0x0A,
			ItemID:  0,
			ClassID: 1,
			Name:    "Friends",
		},
	}
	if err := f.Upsert(screenName, itemsIn); err != nil {
		t.Fatalf("failed to upsert: %s", err.Error())
	}

	itemsOut, err := f.Retrieve(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve: %s", err.Error())
	}

	if !reflect.DeepEqual(itemsIn, itemsOut) {
		t.Fatalf("items did not match:\n in: %v\n out: %v", itemsIn, itemsOut)
	}
}

func TestFeedbagDelete(t *testing.T) {

	const testFile string = "aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	itemsIn := []oscar.FeedbagItem{
		{
			GroupID: 0,
			ItemID:  1805,
			ClassID: 3,
			Name:    "spimmer1234",
			TLVLBlock: oscar.TLVLBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(0x01, uint16(1000)),
				},
			},
		},
		{
			GroupID: 0x0A,
			ItemID:  0,
			ClassID: 1,
			Name:    "Friends",
		},
		{
			GroupID: 0x0B,
			ItemID:  100,
			ClassID: 1,
			Name:    "co-workers",
		},
	}

	if err := f.Upsert(screenName, itemsIn); err != nil {
		t.Fatalf("failed to upsert: %s", err.Error())
	}

	if err := f.Delete(screenName, []oscar.FeedbagItem{itemsIn[0]}); err != nil {
		t.Fatalf("failed to delete: %s", err.Error())
	}

	itemsOut, err := f.Retrieve(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve: %s", err.Error())
	}

	expect := itemsIn[1:]

	if !reflect.DeepEqual(expect, itemsOut) {
		t.Fatalf("items did not match:\n in: %v\n out: %v", expect, itemsOut)
	}
}

func TestLastModifiedEmpty(t *testing.T) {

	const testFile string = "aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	_, err = f.LastModified(screenName)

	if err != nil {
		t.Fatalf("get error from last modified: %s", err.Error())
	}
}

func TestLastModifiedNotEmpty(t *testing.T) {

	const testFile string = "aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	itemsIn := []oscar.FeedbagItem{
		{
			GroupID: 0x0A,
			ItemID:  0,
			ClassID: 1,
			Name:    "Friends",
		},
	}
	if err := f.Upsert(screenName, itemsIn); err != nil {
		t.Fatalf("failed to upsert: %s", err.Error())
	}

	_, err = f.LastModified(screenName)

	if err != nil {
		t.Fatalf("get error from last modified: %s", err.Error())
	}
}

func TestProfile(t *testing.T) {

	const testFile string = "aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	u := User{
		ScreenName: screenName,
	}
	if err := f.UpsertUser(u); err != nil {
		t.Fatalf("failed to upsert new user: %s", err.Error())
	}

	profile, err := f.RetrieveProfile(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve profile: %s", err.Error())
	}

	if profile != "" {
		t.Fatalf("expected empty profile for %s", screenName)
	}

	newProfile := "here is my profile"
	if err := f.UpsertProfile(screenName, newProfile); err != nil {
		t.Fatalf("failed to create new profile: %s", err.Error())
	}

	profile, err = f.RetrieveProfile(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve profile: %s", err.Error())
	}

	if !reflect.DeepEqual(newProfile, profile) {
		t.Fatalf("profiles did not match:\n expected: %v\n actual: %v", newProfile, profile)
	}

	updatedProfile := "here is my profile [updated]"
	if err := f.UpsertProfile(screenName, updatedProfile); err != nil {
		t.Fatalf("failed to create new profile: %s", err.Error())
	}

	profile, err = f.RetrieveProfile(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve profile: %s", err.Error())
	}

	if !reflect.DeepEqual(updatedProfile, profile) {
		t.Fatalf("updated profiles did not match:\n expected: %v\n actual: %v", newProfile, profile)
	}
}

func TestProfileNonExistent(t *testing.T) {

	const testFile string = "aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	prof, err := f.RetrieveProfile(screenName)
	assert.NoError(t, err)
	assert.Empty(t, prof)
}

func TestInterestedUsers(t *testing.T) {
	const testFile string = "aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',0,13852,3,'userB',NULL,1691286176)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	users, err := f.InterestedUsers("userA")
	if len(users) != 0 {
		t.Fatalf("expected no interested users, got %v", users)
	}

	users, err = f.InterestedUsers("userB")
	if len(users) != 0 {
		t.Fatalf("expected no interested users, got %v", users)
	}
}

func TestFeedbagStoreBuddiesBlockedUser(t *testing.T) {
	const testFile string = "aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',0,13852,3,'userB',NULL,1691286176)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	users, err := f.Buddies("userA")
	if len(users) != 0 {
		t.Fatalf("expected no buddies, got %v", users)
	}

	users, err = f.Buddies("userB")
	if len(users) != 0 {
		t.Fatalf("expected no buddies, got %v", users)
	}
}

func TestFeedbagStoreBlockedA(t *testing.T) {
	const testFile string = "aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',0,13852,3,'userB',NULL,1691286176)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	sn1 := "userA"
	sn2 := "userB"
	blocked, err := f.Blocked(sn1, sn2)
	if err != nil {
		t.Fatalf("db err: %s", err.Error())
	}
	if blocked != BlockedA {
		t.Fatalf("expected A to be blocker")
	}
}

func TestFeedbagStoreBlockedB(t *testing.T) {
	const testFile string = "aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',0,13852,3,'userA',NULL,1691286176)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	sn1 := "userA"
	sn2 := "userB"
	blocked, err := f.Blocked(sn1, sn2)
	if err != nil {
		t.Fatalf("db err: %s", err.Error())
	}
	if blocked != BlockedB {
		t.Fatalf("expected B to be blocker")
	}
}

func TestFeedbagStoreBlockedNoBlocked(t *testing.T) {
	const testFile string = "aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	sn1 := "userA"
	sn2 := "userB"
	blocked, err := f.Blocked(sn1, sn2)
	if err != nil {
		t.Fatalf("db err: %s", err.Error())
	}
	if blocked != BlockedNo {
		t.Fatalf("expected no blocker")
	}
}

func TestGetUser(t *testing.T) {
	const testFile string = "aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	expectUser := &User{
		ScreenName: "testscreenname",
		AuthKey:    "theauthkey",
		PassHash:   []byte("thepasshash"),
	}
	_, err = f.db.Exec(`INSERT INTO user (ScreenName, authKey, passHash) VALUES(?, ?, ?)`,
		expectUser.ScreenName, expectUser.AuthKey, expectUser.PassHash)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err.Error())
	}

	actualUser, err := f.GetUser(expectUser.ScreenName)
	if err != nil {
		t.Fatalf("failed to get user: %s", err.Error())
	}

	if !reflect.DeepEqual(expectUser, actualUser) {
		t.Fatalf("users are not equal. expect: %v actual: %v", expectUser, actualUser)
	}
}

func TestGetUserNotFound(t *testing.T) {
	const testFile string = "aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewSQLiteFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	actualUser, err := f.GetUser("testscreenname")
	if err != nil {
		t.Fatalf("failed to get user: %s", err.Error())
	}

	if actualUser != nil {
		t.Fatal("expected user to not be found")
	}
}

func TestHashPassword(t *testing.T) {
	u := &User{
		AuthKey: "the_auth_key",
	}
	if err := u.HashPassword(""); err != nil {
		t.Fatalf("error hashing password: %s", err.Error())
	}
	t.Logf("password hash: %s", u.PassHash)
}
