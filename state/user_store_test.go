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
		RegStatus:         3,
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
	assert.NoError(t, err)

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
		{
			IdentScreenName:   NewIdentScreenName("100003"),
			DisplayScreenName: "100003",
			IsICQ:             true,
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

func TestSQLiteUserStore_InsertUser_UINButNotIsICQ(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	user := User{
		IdentScreenName:   NewIdentScreenName("100003"),
		DisplayScreenName: "100003",
	}

	err = f.InsertUser(user)
	assert.ErrorContains(t, err, "inserting user with UIN and isICQ=false")
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

	err = feedbagStore.SetUserPassword(u.IdentScreenName, "theNEWpassword")
	assert.NoError(t, err)

	gotUser, err := feedbagStore.User(u.IdentScreenName)
	assert.NoError(t, err)

	wantUser := User{}
	wantUser.HashPassword("theNEWpassword")

	valid := gotUser.ValidateHash(wantUser.StrongMD5Pass)
	assert.True(t, valid)
}

func TestSQLiteUserStore_SetUserPassword_ErrNoUser(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	feedbagStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	err = feedbagStore.SetUserPassword(NewIdentScreenName("some_user"), "thepassword")
	assert.ErrorIs(t, err, ErrNoUser)
}

func TestSQLiteUserStore_ChatRoomByCookie_RoomFound(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	userStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	chatRoom := NewChatRoom("my new chat room!", NewIdentScreenName("the-screen-name"), PrivateExchange)

	err = userStore.CreateChatRoom(&chatRoom)
	assert.NoError(t, err)

	gotRoom, err := userStore.ChatRoomByCookie(chatRoom.Cookie())
	assert.NoError(t, err)
	assert.Equal(t, chatRoom, gotRoom)
}

func TestSQLiteUserStore_ChatRoomByCookie_RoomNotFound(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	userStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	_, err = userStore.ChatRoomByCookie("the-chat-cookie")
	assert.ErrorIs(t, err, ErrChatRoomNotFound)
}

func TestSQLiteUserStore_ChatRoomByName_RoomFound(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	userStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	chatRoom := NewChatRoom("my new chat room!", NewIdentScreenName("the-screen-name"), PrivateExchange)

	err = userStore.CreateChatRoom(&chatRoom)
	assert.NoError(t, err)

	gotRoom, err := userStore.ChatRoomByName(chatRoom.Exchange(), chatRoom.Name())
	assert.NoError(t, err)
	assert.Equal(t, chatRoom, gotRoom)
}

func TestSQLiteUserStore_ChatRoomByName_RoomNotFound(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	userStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	_, err = userStore.ChatRoomByName(4, "the-chat-room")
	assert.ErrorIs(t, err, ErrChatRoomNotFound)
}

func TestSQLiteUserStore_AllChatRooms(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	userStore, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	chatRooms := []ChatRoom{
		NewChatRoom("chat room 1", NewIdentScreenName("creator"), PrivateExchange),
		NewChatRoom("chat room 2", NewIdentScreenName("creator"), PrivateExchange),
		NewChatRoom("chat room 3", NewIdentScreenName("creator"), PublicExchange),
	}

	for i := range chatRooms {
		tBefore := (&chatRooms[i]).CreateTime()
		err = userStore.CreateChatRoom(&chatRooms[i])
		assert.NoError(t, err)
		assert.True(t, chatRooms[i].CreateTime().After(tBefore))
	}

	// public exchange
	gotRooms, err := userStore.AllChatRooms(5)
	assert.NoError(t, err)

	assert.Equal(t, chatRooms[2:], gotRooms)

	// private exchange
	gotRooms, err = userStore.AllChatRooms(4)
	assert.NoError(t, err)

	assert.Equal(t, chatRooms[0:2], gotRooms)
}

func TestSQLiteUserStore_CreateChatRoom_ErrChatRoomExists(t *testing.T) {

	tt := []struct {
		name         string
		firstInsert  ChatRoom
		secondInsert ChatRoom
		wantErr      error
	}{
		{
			name:         "create two rooms with different cookie/exchange, same name",
			firstInsert:  NewChatRoom("chat room", NewIdentScreenName("creator"), PublicExchange),
			secondInsert: NewChatRoom("chat room", NewIdentScreenName("creator"), PrivateExchange),
			wantErr:      nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				assert.NoError(t, os.Remove(testFile))
			}()

			userStore, err := NewSQLiteUserStore(testFile)
			assert.NoError(t, err)

			err = userStore.CreateChatRoom(&tc.firstInsert)
			assert.NoError(t, err)

			err = userStore.CreateChatRoom(&tc.secondInsert)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestUpdateDisplayScreenName(t *testing.T) {

	screenNameOriginal := DisplayScreenName("chattingchuck")
	screenNameFormatted := DisplayScreenName("Chatting Chuck")

	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	user := User{
		DisplayScreenName: screenNameOriginal,
		IdentScreenName:   screenNameOriginal.IdentScreenName(),
		RegStatus:         3,
	}
	if err := f.InsertUser(user); err != nil {
		t.Fatalf("failed to upsert new user: %s", err.Error())
	}
	err = f.UpdateDisplayScreenName(screenNameFormatted)
	if err != nil {
		t.Fatalf("failed to update display screen name: %s", err.Error())
	}

	dbUser, err := f.User(screenNameOriginal.IdentScreenName())
	if err != nil {
		t.Fatalf("failed to retrieve screen name: %s", err.Error())
	}

	assert.Equal(t, dbUser.DisplayScreenName, screenNameFormatted)
}

func TestSQLiteUserStore_SetWorkInfo(t *testing.T) {
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	screenName := NewIdentScreenName("testuser")
	user := User{
		IdentScreenName: screenName,
	}
	err = f.InsertUser(user)
	assert.NoError(t, err)

	// Define the work info to set
	workInfo := ICQWorkInfo{
		Company:        "Test Company",
		Department:     "Test Department",
		OccupationCode: 10,
		Position:       "Test Position",
		Address:        "123 Test St",
		City:           "Test City",
		CountryCode:    1,
		Fax:            "123-456-7890",
		Phone:          "098-765-4321",
		State:          "Test State",
		WebPage:        "http://www.test.com",
		ZIPCode:        "12345",
	}

	t.Run("Successful Update", func(t *testing.T) {
		// Call SetWorkInfo
		err := f.SetWorkInfo(screenName, workInfo)
		assert.NoError(t, err)

		// Retrieve the user and verify the work info was set correctly
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Equal(t, workInfo.Company, updatedUser.ICQWorkInfo.Company)
		assert.Equal(t, workInfo.Department, updatedUser.ICQWorkInfo.Department)
		assert.Equal(t, workInfo.OccupationCode, updatedUser.ICQWorkInfo.OccupationCode)
		assert.Equal(t, workInfo.Position, updatedUser.ICQWorkInfo.Position)
		assert.Equal(t, workInfo.Address, updatedUser.ICQWorkInfo.Address)
		assert.Equal(t, workInfo.City, updatedUser.ICQWorkInfo.City)
		assert.Equal(t, workInfo.CountryCode, updatedUser.ICQWorkInfo.CountryCode)
		assert.Equal(t, workInfo.Fax, updatedUser.ICQWorkInfo.Fax)
		assert.Equal(t, workInfo.Phone, updatedUser.ICQWorkInfo.Phone)
		assert.Equal(t, workInfo.State, updatedUser.ICQWorkInfo.State)
		assert.Equal(t, workInfo.WebPage, updatedUser.ICQWorkInfo.WebPage)
		assert.Equal(t, workInfo.ZIPCode, updatedUser.ICQWorkInfo.ZIPCode)
	})

	t.Run("Update Non-Existing User", func(t *testing.T) {
		// Try to set work info for a non-existing user
		nonExistingScreenName := NewIdentScreenName("nonexistentuser")
		err := f.SetWorkInfo(nonExistingScreenName, workInfo)

		// This should return ErrNoUser, as the user does not exist
		assert.ErrorIs(t, err, ErrNoUser)
	})

	t.Run("Empty Work Info", func(t *testing.T) {
		// Test updating with empty work info (depending on business rules, this might be allowed or not)
		emptyWorkInfo := ICQWorkInfo{}
		err := f.SetWorkInfo(screenName, emptyWorkInfo)
		assert.NoError(t, err)

		// Retrieve the user and verify that fields are empty or have default values
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.ICQWorkInfo.Company)
		assert.Empty(t, updatedUser.ICQWorkInfo.Department)
		assert.Empty(t, updatedUser.ICQWorkInfo.OccupationCode)
		assert.Empty(t, updatedUser.ICQWorkInfo.Position)
		assert.Empty(t, updatedUser.ICQWorkInfo.Address)
		assert.Empty(t, updatedUser.ICQWorkInfo.City)
		assert.Empty(t, updatedUser.ICQWorkInfo.CountryCode)
		assert.Empty(t, updatedUser.ICQWorkInfo.Fax)
		assert.Empty(t, updatedUser.ICQWorkInfo.Phone)
		assert.Empty(t, updatedUser.ICQWorkInfo.State)
		assert.Empty(t, updatedUser.ICQWorkInfo.WebPage)
		assert.Empty(t, updatedUser.ICQWorkInfo.ZIPCode)
	})
}

func TestSQLiteUserStore_SetMoreInfo(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create a test user
	screenName := NewIdentScreenName("testuser")
	user := User{
		IdentScreenName: screenName,
	}
	err = f.InsertUser(user)
	assert.NoError(t, err)

	// Define the more info data to set
	moreInfo := ICQMoreInfo{
		BirthDay:     15,
		BirthMonth:   8,
		BirthYear:    1990,
		Gender:       1,
		HomePageAddr: "http://example.com",
		Lang1:        1,
		Lang2:        2,
		Lang3:        3,
	}

	t.Run("Successful Update", func(t *testing.T) {
		// Call SetMoreInfo
		err := f.SetMoreInfo(screenName, moreInfo)
		assert.NoError(t, err)

		// Retrieve the user and verify the more info was set correctly
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Equal(t, moreInfo.BirthDay, updatedUser.ICQMoreInfo.BirthDay)
		assert.Equal(t, moreInfo.BirthMonth, updatedUser.ICQMoreInfo.BirthMonth)
		assert.Equal(t, moreInfo.BirthYear, updatedUser.ICQMoreInfo.BirthYear)
		assert.Equal(t, moreInfo.Gender, updatedUser.ICQMoreInfo.Gender)
		assert.Equal(t, moreInfo.HomePageAddr, updatedUser.ICQMoreInfo.HomePageAddr)
		assert.Equal(t, moreInfo.Lang1, updatedUser.ICQMoreInfo.Lang1)
		assert.Equal(t, moreInfo.Lang2, updatedUser.ICQMoreInfo.Lang2)
		assert.Equal(t, moreInfo.Lang3, updatedUser.ICQMoreInfo.Lang3)
	})

	t.Run("Update Non-Existing User", func(t *testing.T) {
		// Try to set more info for a non-existing user
		nonExistingScreenName := NewIdentScreenName("nonexistentuser")
		err := f.SetMoreInfo(nonExistingScreenName, moreInfo)

		// This should return ErrNoUser, as the user does not exist
		assert.ErrorIs(t, err, ErrNoUser)
	})

	t.Run("Empty More Info", func(t *testing.T) {
		// Test updating with empty more info
		emptyMoreInfo := ICQMoreInfo{}
		err := f.SetMoreInfo(screenName, emptyMoreInfo)
		assert.NoError(t, err)

		// Retrieve the user and verify that fields are empty or have default values
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.ICQMoreInfo.BirthDay)
		assert.Empty(t, updatedUser.ICQMoreInfo.BirthMonth)
		assert.Empty(t, updatedUser.ICQMoreInfo.BirthYear)
		assert.Empty(t, updatedUser.ICQMoreInfo.Gender)
		assert.Empty(t, updatedUser.ICQMoreInfo.HomePageAddr)
		assert.Empty(t, updatedUser.ICQMoreInfo.Lang1)
		assert.Empty(t, updatedUser.ICQMoreInfo.Lang2)
		assert.Empty(t, updatedUser.ICQMoreInfo.Lang3)
	})
}

func TestSQLiteUserStore_SetUserNotes(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create a test user
	screenName := NewIdentScreenName("testuser")
	user := User{
		IdentScreenName: screenName,
	}
	err = f.InsertUser(user)
	assert.NoError(t, err)

	// Define the user notes to set
	userNotes := ICQUserNotes{
		Notes: "This is a test note.",
	}

	t.Run("Successful Update", func(t *testing.T) {
		// Call SetUserNotes
		err := f.SetUserNotes(screenName, userNotes)
		assert.NoError(t, err)

		// Retrieve the user and verify the notes were set correctly
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Equal(t, userNotes.Notes, updatedUser.ICQNotes.Notes)
	})

	t.Run("Update Non-Existing User", func(t *testing.T) {
		// Try to set notes for a non-existing user
		nonExistingScreenName := NewIdentScreenName("nonexistentuser")
		err := f.SetUserNotes(nonExistingScreenName, userNotes)

		// This should return ErrNoUser, as the user does not exist
		assert.ErrorIs(t, err, ErrNoUser)
	})

	t.Run("Empty Notes", func(t *testing.T) {
		// Test updating with empty notes
		emptyNotes := ICQUserNotes{Notes: ""}
		err := f.SetUserNotes(screenName, emptyNotes)
		assert.NoError(t, err)

		// Retrieve the user and verify that notes are empty
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.ICQNotes.Notes)
	})
}

func TestSQLiteUserStore_SetInterests(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create a test user
	screenName := NewIdentScreenName("testuser")
	user := User{
		IdentScreenName: screenName,
	}
	err = f.InsertUser(user)
	assert.NoError(t, err)

	// Define the interests data to set
	interests := ICQInterests{
		Code1:    1,
		Keyword1: "Coding",
		Code2:    2,
		Keyword2: "Music",
		Code3:    3,
		Keyword3: "Gaming",
		Code4:    4,
		Keyword4: "Travel",
	}

	t.Run("Successful Update", func(t *testing.T) {
		// Call SetInterests
		err := f.SetInterests(screenName, interests)
		assert.NoError(t, err)

		// Retrieve the user and verify the interests were set correctly
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Equal(t, interests.Code1, updatedUser.ICQInterests.Code1)
		assert.Equal(t, interests.Keyword1, updatedUser.ICQInterests.Keyword1)
		assert.Equal(t, interests.Code2, updatedUser.ICQInterests.Code2)
		assert.Equal(t, interests.Keyword2, updatedUser.ICQInterests.Keyword2)
		assert.Equal(t, interests.Code3, updatedUser.ICQInterests.Code3)
		assert.Equal(t, interests.Keyword3, updatedUser.ICQInterests.Keyword3)
		assert.Equal(t, interests.Code4, updatedUser.ICQInterests.Code4)
		assert.Equal(t, interests.Keyword4, updatedUser.ICQInterests.Keyword4)
	})

	t.Run("Update Non-Existing User", func(t *testing.T) {
		// Try to set interests for a non-existing user
		nonExistingScreenName := NewIdentScreenName("nonexistentuser")
		err := f.SetInterests(nonExistingScreenName, interests)

		// This should return ErrNoUser, as the user does not exist
		assert.ErrorIs(t, err, ErrNoUser)
	})

	t.Run("Empty Interests", func(t *testing.T) {
		// Test updating with empty interests
		emptyInterests := ICQInterests{}
		err := f.SetInterests(screenName, emptyInterests)
		assert.NoError(t, err)

		// Retrieve the user and verify that interests fields are empty or have default values
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.ICQInterests.Code1)
		assert.Empty(t, updatedUser.ICQInterests.Keyword1)
		assert.Empty(t, updatedUser.ICQInterests.Code2)
		assert.Empty(t, updatedUser.ICQInterests.Keyword2)
		assert.Empty(t, updatedUser.ICQInterests.Code3)
		assert.Empty(t, updatedUser.ICQInterests.Keyword3)
		assert.Empty(t, updatedUser.ICQInterests.Code4)
		assert.Empty(t, updatedUser.ICQInterests.Keyword4)
	})
}

func TestSQLiteUserStore_SetAffiliations(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create a test user
	screenName := NewIdentScreenName("testuser")
	user := User{
		IdentScreenName: screenName,
	}
	err = f.InsertUser(user)
	assert.NoError(t, err)

	// Define the affiliations data to set
	affiliations := ICQAffiliations{
		CurrentCode1:    1,
		CurrentKeyword1: "Sports",
		CurrentCode2:    2,
		CurrentKeyword2: "Science",
		CurrentCode3:    3,
		CurrentKeyword3: "Arts",
		PastCode1:       4,
		PastKeyword1:    "Literature",
		PastCode2:       5,
		PastKeyword2:    "Music",
		PastCode3:       6,
		PastKeyword3:    "Technology",
	}

	t.Run("Successful Update", func(t *testing.T) {
		// Call SetAffiliations
		err := f.SetAffiliations(screenName, affiliations)
		assert.NoError(t, err)

		// Retrieve the user and verify the affiliations were set correctly
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Equal(t, affiliations.CurrentCode1, updatedUser.ICQAffiliations.CurrentCode1)
		assert.Equal(t, affiliations.CurrentKeyword1, updatedUser.ICQAffiliations.CurrentKeyword1)
		assert.Equal(t, affiliations.CurrentCode2, updatedUser.ICQAffiliations.CurrentCode2)
		assert.Equal(t, affiliations.CurrentKeyword2, updatedUser.ICQAffiliations.CurrentKeyword2)
		assert.Equal(t, affiliations.CurrentCode3, updatedUser.ICQAffiliations.CurrentCode3)
		assert.Equal(t, affiliations.CurrentKeyword3, updatedUser.ICQAffiliations.CurrentKeyword3)
		assert.Equal(t, affiliations.PastCode1, updatedUser.ICQAffiliations.PastCode1)
		assert.Equal(t, affiliations.PastKeyword1, updatedUser.ICQAffiliations.PastKeyword1)
		assert.Equal(t, affiliations.PastCode2, updatedUser.ICQAffiliations.PastCode2)
		assert.Equal(t, affiliations.PastKeyword2, updatedUser.ICQAffiliations.PastKeyword2)
		assert.Equal(t, affiliations.PastCode3, updatedUser.ICQAffiliations.PastCode3)
		assert.Equal(t, affiliations.PastKeyword3, updatedUser.ICQAffiliations.PastKeyword3)
	})

	t.Run("Update Non-Existing User", func(t *testing.T) {
		// Try to set affiliations for a non-existing user
		nonExistingScreenName := NewIdentScreenName("nonexistentuser")
		err := f.SetAffiliations(nonExistingScreenName, affiliations)

		// This should return ErrNoUser, as the user does not exist
		assert.ErrorIs(t, err, ErrNoUser)
	})

	t.Run("Empty Affiliations", func(t *testing.T) {
		// Test updating with empty affiliations
		emptyAffiliations := ICQAffiliations{}
		err := f.SetAffiliations(screenName, emptyAffiliations)
		assert.NoError(t, err)

		// Retrieve the user and verify that affiliations fields are empty or have default values
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.ICQAffiliations.CurrentCode1)
		assert.Empty(t, updatedUser.ICQAffiliations.CurrentKeyword1)
		assert.Empty(t, updatedUser.ICQAffiliations.CurrentCode2)
		assert.Empty(t, updatedUser.ICQAffiliations.CurrentKeyword2)
		assert.Empty(t, updatedUser.ICQAffiliations.CurrentCode3)
		assert.Empty(t, updatedUser.ICQAffiliations.CurrentKeyword3)
		assert.Empty(t, updatedUser.ICQAffiliations.PastCode1)
		assert.Empty(t, updatedUser.ICQAffiliations.PastKeyword1)
		assert.Empty(t, updatedUser.ICQAffiliations.PastCode2)
		assert.Empty(t, updatedUser.ICQAffiliations.PastKeyword2)
		assert.Empty(t, updatedUser.ICQAffiliations.PastCode3)
		assert.Empty(t, updatedUser.ICQAffiliations.PastKeyword3)
	})
}

func TestSQLiteUserStore_SetBasicInfo(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create a test user
	screenName := NewIdentScreenName("testuser")
	user := User{
		IdentScreenName: screenName,
	}
	err = f.InsertUser(user)
	assert.NoError(t, err)

	// Define the basic info data to set
	basicInfo := ICQBasicInfo{
		CellPhone:    "123-456-7890",
		CountryCode:  1,
		EmailAddress: "test@example.com",
		FirstName:    "John",
		GMTOffset:    5,
		Address:      "123 Main St",
		City:         "Test City",
		Fax:          "098-765-4321",
		Phone:        "111-222-3333",
		State:        "Test State",
		LastName:     "Doe",
		Nickname:     "Johnny",
		PublishEmail: true,
		ZIPCode:      "12345",
	}

	t.Run("Successful Update", func(t *testing.T) {
		// Call SetBasicInfo
		err := f.SetBasicInfo(screenName, basicInfo)
		assert.NoError(t, err)

		// Retrieve the user and verify the basic info was set correctly
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Equal(t, basicInfo.CellPhone, updatedUser.ICQBasicInfo.CellPhone)
		assert.Equal(t, basicInfo.CountryCode, updatedUser.ICQBasicInfo.CountryCode)
		assert.Equal(t, basicInfo.EmailAddress, updatedUser.ICQBasicInfo.EmailAddress)
		assert.Equal(t, basicInfo.FirstName, updatedUser.ICQBasicInfo.FirstName)
		assert.Equal(t, basicInfo.GMTOffset, updatedUser.ICQBasicInfo.GMTOffset)
		assert.Equal(t, basicInfo.Address, updatedUser.ICQBasicInfo.Address)
		assert.Equal(t, basicInfo.City, updatedUser.ICQBasicInfo.City)
		assert.Equal(t, basicInfo.Fax, updatedUser.ICQBasicInfo.Fax)
		assert.Equal(t, basicInfo.Phone, updatedUser.ICQBasicInfo.Phone)
		assert.Equal(t, basicInfo.State, updatedUser.ICQBasicInfo.State)
		assert.Equal(t, basicInfo.LastName, updatedUser.ICQBasicInfo.LastName)
		assert.Equal(t, basicInfo.Nickname, updatedUser.ICQBasicInfo.Nickname)
		assert.Equal(t, basicInfo.PublishEmail, updatedUser.ICQBasicInfo.PublishEmail)
		assert.Equal(t, basicInfo.ZIPCode, updatedUser.ICQBasicInfo.ZIPCode)
	})

	t.Run("Update Non-Existing User", func(t *testing.T) {
		// Try to set basic info for a non-existing user
		nonExistingScreenName := NewIdentScreenName("nonexistentuser")
		err := f.SetBasicInfo(nonExistingScreenName, basicInfo)

		// This should return ErrNoUser, as the user does not exist
		assert.ErrorIs(t, err, ErrNoUser)
	})

	t.Run("Empty Basic Info", func(t *testing.T) {
		// Test updating with empty basic info
		emptyBasicInfo := ICQBasicInfo{}
		err := f.SetBasicInfo(screenName, emptyBasicInfo)
		assert.NoError(t, err)

		// Retrieve the user and verify that basic info fields are empty or have default values
		updatedUser, err := f.User(screenName)
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.ICQBasicInfo.CellPhone)
		assert.Empty(t, updatedUser.ICQBasicInfo.CountryCode)
		assert.Empty(t, updatedUser.ICQBasicInfo.EmailAddress)
		assert.Empty(t, updatedUser.ICQBasicInfo.FirstName)
		assert.Empty(t, updatedUser.ICQBasicInfo.GMTOffset)
		assert.Empty(t, updatedUser.ICQBasicInfo.Address)
		assert.Empty(t, updatedUser.ICQBasicInfo.City)
		assert.Empty(t, updatedUser.ICQBasicInfo.Fax)
		assert.Empty(t, updatedUser.ICQBasicInfo.Phone)
		assert.Empty(t, updatedUser.ICQBasicInfo.State)
		assert.Empty(t, updatedUser.ICQBasicInfo.LastName)
		assert.Empty(t, updatedUser.ICQBasicInfo.Nickname)
		assert.Empty(t, updatedUser.ICQBasicInfo.PublishEmail)
		assert.Empty(t, updatedUser.ICQBasicInfo.ZIPCode)
	})
}

func TestSQLiteUserStore_FindByInterests(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create and set up test users with different interests
	user1 := User{
		IdentScreenName: NewIdentScreenName("user1"),
	}
	err = f.InsertUser(user1)
	assert.NoError(t, err)
	interests1 := ICQInterests{
		Code1:    1,
		Keyword1: "Coding",
		Code2:    2,
		Keyword2: "Music",
	}
	err = f.SetInterests(user1.IdentScreenName, interests1)
	assert.NoError(t, err)

	user2 := User{
		IdentScreenName: NewIdentScreenName("user2"),
	}
	err = f.InsertUser(user2)
	assert.NoError(t, err)
	interests2 := ICQInterests{
		Code1:    1,
		Keyword1: "Coding",
		Code3:    3,
		Keyword3: "Gaming",
	}
	err = f.SetInterests(user2.IdentScreenName, interests2)
	assert.NoError(t, err)

	user3 := User{
		IdentScreenName: NewIdentScreenName("user3"),
	}
	err = f.InsertUser(user3)
	assert.NoError(t, err)
	interests3 := ICQInterests{
		Code2:    2,
		Keyword2: "Music",
		Code4:    4,
		Keyword4: "Travel",
	}
	err = f.SetInterests(user3.IdentScreenName, interests3)
	assert.NoError(t, err)

	// Helper function to check if a user with a specific IdentScreenName exists in the results
	containsUserWithScreenName := func(users []User, screenName IdentScreenName) bool {
		for _, user := range users {
			if user.IdentScreenName == screenName {
				return true
			}
		}
		return false
	}

	t.Run("Find Users by Single Keyword", func(t *testing.T) {
		// Search for users interested in "Music"
		users, err := f.FindByInterests(2, []string{"Music"})
		assert.NoError(t, err)
		assert.Len(t, users, 2)

		// Check that the correct users are returned by IdentScreenName
		assert.True(t, containsUserWithScreenName(users, user1.IdentScreenName))
		assert.True(t, containsUserWithScreenName(users, user3.IdentScreenName))
	})

	t.Run("Find Users by Multiple Keywords", func(t *testing.T) {
		// Search for users interested in "Coding" or "Gaming"
		users, err := f.FindByInterests(1, []string{"Coding", "Gaming"})
		assert.NoError(t, err)
		assert.Len(t, users, 2)

		// Check that the correct users are returned by IdentScreenName
		assert.True(t, containsUserWithScreenName(users, user1.IdentScreenName))
		assert.True(t, containsUserWithScreenName(users, user2.IdentScreenName))
	})

	t.Run("Find Users by Multiple Codes and Keywords", func(t *testing.T) {
		// Search for users interested in "Coding"
		users, err := f.FindByInterests(1, []string{"Coding"})
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.True(t, containsUserWithScreenName(users, user1.IdentScreenName))
		assert.True(t, containsUserWithScreenName(users, user2.IdentScreenName))

		// Search for users interested in "Travel"
		users, err = f.FindByInterests(4, []string{"Travel"})
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.True(t, containsUserWithScreenName(users, user3.IdentScreenName))
	})

	t.Run("No Users Found", func(t *testing.T) {
		// Search for users interested in a keyword that no user has
		users, err := f.FindByInterests(1, []string{"Unknown"})
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestSQLiteUserStore_FindByDetails(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create and set up test users with different details using SetBasicInfo
	user1 := User{
		IdentScreenName: NewIdentScreenName("user1"),
	}
	err = f.InsertUser(user1)
	assert.NoError(t, err)
	basicInfo1 := ICQBasicInfo{
		FirstName: "John",
		LastName:  "Doe",
		Nickname:  "Johnny",
	}
	err = f.SetBasicInfo(user1.IdentScreenName, basicInfo1)
	assert.NoError(t, err)

	user2 := User{
		IdentScreenName: NewIdentScreenName("user2"),
	}
	err = f.InsertUser(user2)
	assert.NoError(t, err)
	basicInfo2 := ICQBasicInfo{
		FirstName: "Jane",
		LastName:  "Smith",
		Nickname:  "Janey",
	}
	err = f.SetBasicInfo(user2.IdentScreenName, basicInfo2)
	assert.NoError(t, err)

	user3 := User{
		IdentScreenName: NewIdentScreenName("user3"),
	}
	err = f.InsertUser(user3)
	assert.NoError(t, err)
	basicInfo3 := ICQBasicInfo{
		FirstName: "John",
		LastName:  "Smith",
		Nickname:  "JohnnyS",
	}
	err = f.SetBasicInfo(user3.IdentScreenName, basicInfo3)
	assert.NoError(t, err)

	// Helper function to check if a user with a specific IdentScreenName exists in the results
	containsUserWithScreenName := func(users []User, screenName IdentScreenName) bool {
		for _, user := range users {
			if user.IdentScreenName == screenName {
				return true
			}
		}
		return false
	}

	t.Run("Find Users by First Name", func(t *testing.T) {
		// Search for users with the first name "John"
		users, err := f.FindByDetails("John", "", "")
		assert.NoError(t, err)
		assert.Len(t, users, 2)

		// Check that the correct users are returned by IdentScreenName
		assert.True(t, containsUserWithScreenName(users, user1.IdentScreenName))
		assert.True(t, containsUserWithScreenName(users, user3.IdentScreenName))
	})

	t.Run("Find Users by Last Name", func(t *testing.T) {
		// Search for users with the last name "Smith"
		users, err := f.FindByDetails("", "Smith", "")
		assert.NoError(t, err)
		assert.Len(t, users, 2)

		// Check that the correct users are returned by IdentScreenName
		assert.True(t, containsUserWithScreenName(users, user2.IdentScreenName))
		assert.True(t, containsUserWithScreenName(users, user3.IdentScreenName))
	})

	t.Run("Find Users by Nickname", func(t *testing.T) {
		// Search for users with the nickname "Johnny"
		users, err := f.FindByDetails("", "", "Johnny")
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		// Check that the correct user is returned by IdentScreenName
		assert.True(t, containsUserWithScreenName(users, user1.IdentScreenName))
	})

	t.Run("Find Users by Multiple Fields", func(t *testing.T) {
		// Search for users with the first name "Jane" and last name "Smith"
		users, err := f.FindByDetails("Jane", "Smith", "")
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		// Check that the correct user is returned by IdentScreenName
		assert.True(t, containsUserWithScreenName(users, user2.IdentScreenName))
	})

	t.Run("No Users Found", func(t *testing.T) {
		// Search for users with a first name that no user has
		users, err := f.FindByDetails("NonExistent", "", "")
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestSQLiteUserStore_FindByEmail(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create and set up test users with different email addresses using SetBasicInfo
	user1 := User{
		IdentScreenName: NewIdentScreenName("user1"),
	}
	err = f.InsertUser(user1)
	assert.NoError(t, err)
	basicInfo1 := ICQBasicInfo{
		EmailAddress: "user1@example.com",
	}
	err = f.SetBasicInfo(user1.IdentScreenName, basicInfo1)
	assert.NoError(t, err)

	user2 := User{
		IdentScreenName: NewIdentScreenName("user2"),
	}
	err = f.InsertUser(user2)
	assert.NoError(t, err)
	basicInfo2 := ICQBasicInfo{
		EmailAddress: "user2@example.com",
	}
	err = f.SetBasicInfo(user2.IdentScreenName, basicInfo2)
	assert.NoError(t, err)

	user3 := User{
		IdentScreenName: NewIdentScreenName("user3"),
	}
	err = f.InsertUser(user3)
	assert.NoError(t, err)
	basicInfo3 := ICQBasicInfo{
		EmailAddress: "user3@example.com",
	}
	err = f.SetBasicInfo(user3.IdentScreenName, basicInfo3)
	assert.NoError(t, err)

	t.Run("Find User by Email", func(t *testing.T) {
		// Search for user with email "user1@example.com"
		user, err := f.FindByEmail("user1@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user1.IdentScreenName, user.IdentScreenName)

		// Search for user with email "user2@example.com"
		user, err = f.FindByEmail("user2@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user2.IdentScreenName, user.IdentScreenName)

		// Search for user with email "user3@example.com"
		user, err = f.FindByEmail("user3@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user3.IdentScreenName, user.IdentScreenName)
	})

	t.Run("User Not Found", func(t *testing.T) {
		// Search for an email that doesn't exist
		_, err := f.FindByEmail("nonexistent@example.com")
		assert.ErrorIs(t, err, ErrNoUser)
	})
}

func TestSQLiteUserStore_FindByUIN(t *testing.T) {
	// Cleanup after test
	defer func() {
		assert.NoError(t, os.Remove(testFile))
	}()

	// Initialize the SQLiteUserStore with a test database file
	f, err := NewSQLiteUserStore(testFile)
	assert.NoError(t, err)

	// Create and set up test users where UIN is the same as IdentScreenName
	user1 := User{
		IdentScreenName: NewIdentScreenName("12345"),
	}
	err = f.InsertUser(user1)
	assert.NoError(t, err)

	user2 := User{
		IdentScreenName: NewIdentScreenName("67890"),
	}
	err = f.InsertUser(user2)
	assert.NoError(t, err)

	user3 := User{
		IdentScreenName: NewIdentScreenName("54321"),
	}
	err = f.InsertUser(user3)
	assert.NoError(t, err)

	t.Run("Find User by UIN", func(t *testing.T) {
		// Search for user with UIN 12345
		user, err := f.FindByUIN(12345)
		assert.NoError(t, err)
		assert.Equal(t, user1.IdentScreenName, user.IdentScreenName)

		// Search for user with UIN 67890
		user, err = f.FindByUIN(67890)
		assert.NoError(t, err)
		assert.Equal(t, user2.IdentScreenName, user.IdentScreenName)

		// Search for user with UIN 54321
		user, err = f.FindByUIN(54321)
		assert.NoError(t, err)
		assert.Equal(t, user3.IdentScreenName, user.IdentScreenName)
	})

	t.Run("User Not Found", func(t *testing.T) {
		// Search for a UIN that doesn't exist
		_, err := f.FindByUIN(99999)
		assert.ErrorIs(t, err, ErrNoUser)
	})
}
