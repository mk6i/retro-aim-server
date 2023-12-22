package state

import (
	"os"
	"reflect"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

const testFile string = "aim_test.db"

func TestUserStore(t *testing.T) {

	const screenName = "sn2day"

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

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
	if err := f.FeedbagUpsert(screenName, itemsIn); err != nil {
		t.Fatalf("failed to upsert: %s", err.Error())
	}

	itemsOut, err := f.Feedbag(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve: %s", err.Error())
	}

	if !reflect.DeepEqual(itemsIn, itemsOut) {
		t.Fatalf("items did not match:\n in: %v\n out: %v", itemsIn, itemsOut)
	}
}

func TestFeedbagDelete(t *testing.T) {

	const screenName = "sn2day"

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

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

	if err := f.FeedbagUpsert(screenName, itemsIn); err != nil {
		t.Fatalf("failed to upsert: %s", err.Error())
	}

	if err := f.FeedbagDelete(screenName, []oscar.FeedbagItem{itemsIn[0]}); err != nil {
		t.Fatalf("failed to delete: %s", err.Error())
	}

	itemsOut, err := f.Feedbag(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve: %s", err.Error())
	}

	expect := itemsIn[1:]

	if !reflect.DeepEqual(expect, itemsOut) {
		t.Fatalf("items did not match:\n in: %v\n out: %v", expect, itemsOut)
	}
}

func TestLastModifiedEmpty(t *testing.T) {

	const screenName = "sn2day"

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	_, err = f.FeedbagLastModified(screenName)

	if err != nil {
		t.Fatalf("get error from last modified: %s", err.Error())
	}
}

func TestLastModifiedNotEmpty(t *testing.T) {

	const screenName = "sn2day"

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	itemsIn := []oscar.FeedbagItem{
		{
			GroupID: 0x0A,
			ItemID:  0,
			ClassID: 1,
			Name:    "Friends",
		},
	}
	if err := f.FeedbagUpsert(screenName, itemsIn); err != nil {
		t.Fatalf("failed to upsert: %s", err.Error())
	}

	_, err = f.FeedbagLastModified(screenName)

	if err != nil {
		t.Fatalf("get error from last modified: %s", err.Error())
	}
}

func TestProfile(t *testing.T) {

	const screenName = "sn2day"

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	u := User{
		ScreenName: screenName,
	}
	if err := f.InsertUser(u); err != nil {
		t.Fatalf("failed to upsert new user: %s", err.Error())
	}

	profile, err := f.Profile(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve profile: %s", err.Error())
	}

	if profile != "" {
		t.Fatalf("expected empty profile for %s", screenName)
	}

	newProfile := "here is my profile"
	if err := f.SetProfile(screenName, newProfile); err != nil {
		t.Fatalf("failed to create new profile: %s", err.Error())
	}

	profile, err = f.Profile(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve profile: %s", err.Error())
	}

	if !reflect.DeepEqual(newProfile, profile) {
		t.Fatalf("profiles did not match:\n expected: %v\n actual: %v", newProfile, profile)
	}

	updatedProfile := "here is my profile [updated]"
	if err := f.SetProfile(screenName, updatedProfile); err != nil {
		t.Fatalf("failed to create new profile: %s", err.Error())
	}

	profile, err = f.Profile(screenName)
	if err != nil {
		t.Fatalf("failed to retrieve profile: %s", err.Error())
	}

	if !reflect.DeepEqual(updatedProfile, profile) {
		t.Fatalf("updated profiles did not match:\n expected: %v\n actual: %v", newProfile, profile)
	}
}

func TestProfileNonExistent(t *testing.T) {

	const screenName = "sn2day"

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	prof, err := f.Profile(screenName)
	assert.NoError(t, err)
	assert.Empty(t, prof)
}

func TestAdjacentUsers(t *testing.T) {

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',0,13852,3,'userB',NULL,1691286176)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	users, err := f.AdjacentUsers("userA")
	if len(users) != 0 {
		t.Fatalf("expected no interested users, got %v", users)
	}

	users, err = f.AdjacentUsers("userB")
	if len(users) != 0 {
		t.Fatalf("expected no interested users, got %v", users)
	}
}

func TestUserStoreBuddiesBlockedUser(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

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

func TestUserStoreBlockedA(t *testing.T) {

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',0,13852,3,'userB',NULL,1691286176)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	sn1 := "userA"
	sn2 := "userB"
	blocked, err := f.BlockedState(sn1, sn2)
	if err != nil {
		t.Fatalf("db err: %s", err.Error())
	}
	if blocked != BlockedA {
		t.Fatalf("expected A to be blocker")
	}
}

func TestUserStoreBlockedB(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',0,13852,3,'userA',NULL,1691286176)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	sn1 := "userA"
	sn2 := "userB"
	blocked, err := f.BlockedState(sn1, sn2)
	if err != nil {
		t.Fatalf("db err: %s", err.Error())
	}
	if blocked != BlockedB {
		t.Fatalf("expected B to be blocker")
	}
}

func TestUserStoreBlockedNoBlocked(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	f.db.Exec(`INSERT INTO "feedbag" VALUES('userA',27631,4016,0,'userB',NULL,1690508233)`)
	f.db.Exec(`INSERT INTO "feedbag" VALUES('userB',28330,8120,0,'userA',NULL,1691180328)`)

	sn1 := "userA"
	sn2 := "userB"
	blocked, err := f.BlockedState(sn1, sn2)
	if err != nil {
		t.Fatalf("db err: %s", err.Error())
	}
	if blocked != BlockedNo {
		t.Fatalf("expected no blocker")
	}
}

func TestGetUser(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

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

	actualUser, err := f.User(expectUser.ScreenName)
	if err != nil {
		t.Fatalf("failed to get user: %s", err.Error())
	}

	if !reflect.DeepEqual(expectUser, actualUser) {
		t.Fatalf("users are not equal. expect: %v actual: %v", expectUser, actualUser)
	}
}

func TestGetUserNotFound(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	actualUser, err := f.User("testscreenname")
	if err != nil {
		t.Fatalf("failed to get user: %s", err.Error())
	}

	if actualUser != nil {
		t.Fatal("expected user to not be found")
	}
}

func TestSQLiteUserStore_Users(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	want := []User{
		{ScreenName: "userA"},
		{ScreenName: "userB"},
		{ScreenName: "userC"},
	}

	for _, u := range want {
		err := f.InsertUser(u)
		assert.NoError(t, err)
	}

	have, err := f.AllUsers()
	assert.NoError(t, err)

	assert.Equal(t, want, have)
}

func TestSQLiteUserStore_Buddies(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	feedbagStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	assert.NoError(t, feedbagStore.FeedbagUpsert("userA", []oscar.FeedbagItem{
		{Name: "userB", ItemID: 1, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userC", ItemID: 2, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userD", ItemID: 3, ClassID: oscar.FeedbagClassIdBuddy},
	}))
	assert.NoError(t, feedbagStore.FeedbagUpsert("userB", []oscar.FeedbagItem{
		{Name: "userA", ItemID: 1, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userC", ItemID: 2, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userD", ItemID: 3, ClassID: oscar.FeedbagClassIdBuddy},
	}))

	want := []string{"userB", "userC", "userD"}
	have, err := feedbagStore.Buddies("userA")
	assert.NoError(t, err)

	assert.Equal(t, want, have)
}

func TestNewStubUser(t *testing.T) {
	have, err := NewStubUser("userA")
	assert.NoError(t, err)

	want := User{
		ScreenName: "userA",
		AuthKey:    have.AuthKey,
	}
	assert.NoError(t, want.HashPassword("welcome1"))

	assert.Equal(t, want, have)
}

func TestSQLiteUserStore_AdjacentUsers(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	feedbagStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	assert.NoError(t, feedbagStore.FeedbagUpsert("userA", []oscar.FeedbagItem{
		{Name: "userB", ItemID: 1, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userC", ItemID: 2, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userD", ItemID: 3, ClassID: oscar.FeedbagClassIdBuddy},
	}))
	assert.NoError(t, feedbagStore.FeedbagUpsert("userB", []oscar.FeedbagItem{
		{Name: "userA", ItemID: 1, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userC", ItemID: 2, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userD", ItemID: 3, ClassID: oscar.FeedbagClassIdBuddy},
	}))
	assert.NoError(t, feedbagStore.FeedbagUpsert("userC", []oscar.FeedbagItem{
		{Name: "userA", ItemID: 1, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userB", ItemID: 2, ClassID: oscar.FeedbagClassIdBuddy},
		{Name: "userD", ItemID: 3, ClassID: oscar.FeedbagClassIdBuddy},
	}))

	want := []string{"userB", "userC"}
	have, err := feedbagStore.AdjacentUsers("userA")
	assert.NoError(t, err)

	assert.Equal(t, want, have)
}
