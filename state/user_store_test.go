package state

import (
	"os"
	"reflect"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

const testFile string = "aim_test.db"

func TestUserStore(t *testing.T) {

	screenName := NewIdentScreenName("sn2day")

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	itemsIn := []wire.FeedbagItem{
		{
			GroupID:   0,
			ItemID:    1805,
			ClassID:   3,
			Name:      "spimmer1234",
			TLVLBlock: wire.TLVLBlock{},
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

	screenName := NewIdentScreenName("sn2day")

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	itemsIn := []wire.FeedbagItem{
		{
			GroupID: 0,
			ItemID:  1805,
			ClassID: 3,
			Name:    "spimmer1234",
			TLVLBlock: wire.TLVLBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, uint16(1000)),
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

	if err := f.FeedbagDelete(screenName, []wire.FeedbagItem{itemsIn[0]}); err != nil {
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

	screenName := NewIdentScreenName("sn2day")

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

	screenName := NewIdentScreenName("sn2day")

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	itemsIn := []wire.FeedbagItem{
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

	screenName := NewIdentScreenName("sn2day")

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	u := User{
		IdentScreenName: screenName,
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

	screenName := NewIdentScreenName("sn2day")

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

	users, err := f.AdjacentUsers(NewIdentScreenName("userA"))
	if len(users) != 0 {
		t.Fatalf("expected no interested users, got %v", users)
	}

	users, err = f.AdjacentUsers(NewIdentScreenName("userB"))
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

	users, err := f.Buddies(NewIdentScreenName("userA"))
	if len(users) != 0 {
		t.Fatalf("expected no buddies, got %v", users)
	}

	users, err = f.Buddies(NewIdentScreenName("userB"))
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

	sn1 := NewIdentScreenName("userA")
	sn2 := NewIdentScreenName("userB")

	err = f.FeedbagUpsert(sn1, []wire.FeedbagItem{
		{
			GroupID: 0,
			ItemID:  13852,
			ClassID: 3,
			Name:    "userB",
		},
		{
			GroupID: 27631,
			ItemID:  4016,
			ClassID: 0,
			Name:    "userB",
		},
	})
	assert.NoError(t, err)

	err = f.FeedbagUpsert(sn2, []wire.FeedbagItem{
		{
			GroupID: 28330,
			ItemID:  8120,
			ClassID: 0,
			Name:    "userA",
		},
	})
	assert.NoError(t, err)

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

	sn1 := NewIdentScreenName("userA")
	sn2 := NewIdentScreenName("userB")

	err = f.FeedbagUpsert(sn1, []wire.FeedbagItem{
		{
			GroupID: 27631,
			ItemID:  4016,
			ClassID: 0,
			Name:    "userB",
		},
	})
	assert.NoError(t, err)

	err = f.FeedbagUpsert(sn2, []wire.FeedbagItem{
		{
			GroupID: 0,
			ItemID:  13852,
			ClassID: 3,
			Name:    "userA",
		},
		{
			GroupID: 28330,
			ItemID:  8120,
			ClassID: 0,
			Name:    "userA",
		},
	})
	assert.NoError(t, err)

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

	sn1 := NewIdentScreenName("userA")
	sn2 := NewIdentScreenName("userB")
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

	screenName := NewIdentScreenName("testscreenname")

	insertedUser := &User{
		IdentScreenName:   screenName,
		DisplayScreenName: DisplayScreenName("testscreenname"),
		AuthKey:           "theauthkey",
		StrongMD5Pass:     []byte("thepasshash"),
	}
	err = f.InsertUser(*insertedUser)
	assert.NoError(t, err)

	actualUser, err := f.User(screenName)
	if err != nil {
		t.Fatalf("failed to get user: %s", err.Error())
	}

	if !reflect.DeepEqual(insertedUser, actualUser) {
		t.Fatalf("users are not equal. expect: %v actual: %v", insertedUser, actualUser)
	}
}

func TestGetUserNotFound(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	actualUser, err := f.User(NewIdentScreenName("testscreenname"))
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
		{
			IdentScreenName:   NewIdentScreenName("userA"),
			DisplayScreenName: "userA",
		},
		{
			IdentScreenName:   NewIdentScreenName("userB"),
			DisplayScreenName: "userB",
		},
		{
			IdentScreenName:   NewIdentScreenName("userC"),
			DisplayScreenName: "userC",
		},
	}

	for _, u := range want {
		err := f.InsertUser(u)
		assert.NoError(t, err)
	}

	have, err := f.AllUsers()
	assert.NoError(t, err)

	assert.Equal(t, want, have)
}

func TestSQLiteUserStore_DeleteUser_DeleteExistentUser(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	err = f.InsertUser(User{
		IdentScreenName:   NewIdentScreenName("userA"),
		DisplayScreenName: "userA",
	})
	assert.NoError(t, err)
	err = f.InsertUser(User{
		IdentScreenName:   NewIdentScreenName("userB"),
		DisplayScreenName: "userB",
	})
	assert.NoError(t, err)

	err = f.DeleteUser(NewIdentScreenName("userA"))
	assert.NoError(t, err)

	have, err := f.AllUsers()
	assert.NoError(t, err)

	want := []User{{
		IdentScreenName:   NewIdentScreenName("userB"),
		DisplayScreenName: "userB",
	}}
	assert.Equal(t, want, have)
}

func TestSQLiteUserStore_DeleteUser_DeleteNonExistentUser(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	err = f.DeleteUser(NewIdentScreenName("userA"))
	assert.ErrorIs(t, ErrNoUser, err)
}

func TestSQLiteUserStore_Buddies(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	feedbagStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	assert.NoError(t, feedbagStore.FeedbagUpsert(
		NewIdentScreenName("userA"),
		[]wire.FeedbagItem{
			{Name: "userB", ItemID: 1, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userC", ItemID: 2, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userD", ItemID: 3, ClassID: wire.FeedbagClassIdBuddy},
		}))
	assert.NoError(t, feedbagStore.FeedbagUpsert(
		NewIdentScreenName("userB"),
		[]wire.FeedbagItem{
			{Name: "userA", ItemID: 1, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userC", ItemID: 2, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userD", ItemID: 3, ClassID: wire.FeedbagClassIdBuddy},
		}))

	have, err := feedbagStore.Buddies(NewIdentScreenName("userA"))
	assert.NoError(t, err)

	want := []IdentScreenName{
		NewIdentScreenName("userB"),
		NewIdentScreenName("userC"),
		NewIdentScreenName("userD"),
	}
	assert.Equal(t, want, have)
}

func TestNewStubUser(t *testing.T) {
	have, err := NewStubUser("userA")
	assert.NoError(t, err)

	want := User{
		IdentScreenName:   NewIdentScreenName("userA"),
		DisplayScreenName: "userA",
		AuthKey:           have.AuthKey,
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

	assert.NoError(t, feedbagStore.FeedbagUpsert(
		NewIdentScreenName("userA"),
		[]wire.FeedbagItem{
			{Name: "userB", ItemID: 1, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userC", ItemID: 2, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userD", ItemID: 3, ClassID: wire.FeedbagClassIdBuddy},
		}))
	assert.NoError(t, feedbagStore.FeedbagUpsert(
		NewIdentScreenName("userB"),
		[]wire.FeedbagItem{
			{Name: "userA", ItemID: 1, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userC", ItemID: 2, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userD", ItemID: 3, ClassID: wire.FeedbagClassIdBuddy},
		}))
	assert.NoError(t, feedbagStore.FeedbagUpsert(
		NewIdentScreenName("userC"),
		[]wire.FeedbagItem{
			{Name: "userA", ItemID: 1, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userB", ItemID: 2, ClassID: wire.FeedbagClassIdBuddy},
			{Name: "userD", ItemID: 3, ClassID: wire.FeedbagClassIdBuddy},
		}))

	want := []IdentScreenName{
		NewIdentScreenName("userB"),
		NewIdentScreenName("userC"),
	}
	have, err := feedbagStore.AdjacentUsers(NewIdentScreenName("userA"))
	assert.NoError(t, err)

	assert.Equal(t, want, have)
}

func TestSQLiteUserStore_BARTUpsertAndRetrieve(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	feedbagStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	hash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	item := []byte{'a', 'b', 'c', 'd'}

	b, err := feedbagStore.BARTRetrieve(hash)
	assert.NoError(t, err)
	assert.Empty(t, b)

	err = feedbagStore.BARTUpsert(hash, item)
	assert.NoError(t, err)

	b, err = feedbagStore.BARTRetrieve(hash)
	assert.NoError(t, err)
	assert.Equal(t, item, b)
}

func TestSQLiteUserStore_SetUserPassword_UserExists(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	feedbagStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	u := User{
		IdentScreenName:   NewIdentScreenName("theuser"),
		DisplayScreenName: "theUser",
	}
	err = u.HashPassword("thepassword")
	assert.NoError(t, err)

	err = feedbagStore.InsertUser(u)
	assert.NoError(t, err)

	err = u.HashPassword("thenewpassword")
	assert.NoError(t, err)

	err = feedbagStore.SetUserPassword(u)
	assert.NoError(t, err)

	gotUser, err := feedbagStore.User(u.IdentScreenName)
	assert.NoError(t, err)

	valid := gotUser.ValidateHash(u.StrongMD5Pass)
	assert.True(t, valid)
}

func TestSQLiteUserStore_SetUserPassword_ErrNoUser(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	feedbagStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	u := User{
		IdentScreenName:   NewIdentScreenName("theuser"),
		DisplayScreenName: "theUser",
	}
	err = u.HashPassword("thepassword")
	assert.NoError(t, err)

	err = feedbagStore.SetUserPassword(u)
	assert.ErrorIs(t, err, ErrNoUser)

	// make sure previous transaction previously closed
	err = feedbagStore.SetUserPassword(u)
	assert.ErrorIs(t, err, ErrNoUser)
}
