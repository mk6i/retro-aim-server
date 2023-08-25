package oscar

import (
	"os"
	"reflect"
	"testing"
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

	f, err := NewFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	itemsIn := []*feedbagItem{
		{
			groupID:    0,
			itemID:     1805,
			classID:    3,
			name:       "spimmer1234",
			TLVPayload: TLVPayload{},
		},
		{
			groupID: 0x0A,
			itemID:  0,
			classID: 1,
			name:    "Friends",
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

	const testFile string = "/Users/mike/dev/goaim/aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	itemsIn := []*feedbagItem{
		{
			groupID: 0,
			itemID:  1805,
			classID: 3,
			name:    "spimmer1234",
			TLVPayload: TLVPayload{
				TLVs: []*TLV{
					{
						tType: 0x01,
						val:   uint16(1000),
					},
				},
			},
		},
		{
			groupID: 0x0A,
			itemID:  0,
			classID: 1,
			name:    "Friends",
		},
		{
			groupID: 0x0B,
			itemID:  100,
			classID: 1,
			name:    "co-workers",
		},
	}

	if err := f.Upsert(screenName, itemsIn); err != nil {
		t.Fatalf("failed to upsert: %s", err.Error())
	}

	if err := f.Delete(screenName, []*feedbagItem{itemsIn[0]}); err != nil {
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

	const testFile string = "/Users/mike/dev/goaim/aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	_, err = f.LastModified(screenName)

	if err != nil {
		t.Fatalf("get error from last modified: %s", err.Error())
	}
}

func TestLastModifiedNotEmpty(t *testing.T) {

	const testFile string = "/Users/mike/dev/goaim/aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	itemsIn := []*feedbagItem{
		{
			groupID: 0x0A,
			itemID:  0,
			classID: 1,
			name:    "Friends",
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

	const testFile string = "/Users/mike/dev/goaim/aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	if err := f.UpsertUser(screenName); err != nil {
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

	const testFile string = "/Users/mike/dev/goaim/aim_test.db"
	const screenName = "sn2day"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
	if err != nil {
		t.Fatalf("failed to create new feedbag store: %s", err.Error())
	}

	_, err = f.RetrieveProfile(screenName)
	if err != errUserNotExist {
		t.Fatalf("failed to get error on non-existing profile: %v", err)
	}
}

func TestInterestedUsers(t *testing.T) {
	const testFile string = "/Users/mike/dev/goaim/aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
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
	const testFile string = "/Users/mike/dev/goaim/aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
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
	const testFile string = "/Users/mike/dev/goaim/aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
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
	const testFile string = "/Users/mike/dev/goaim/aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
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
	const testFile string = "/Users/mike/dev/goaim/aim_test.db"

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error("unable to clean up test file")
		}
	}()

	f, err := NewFeedbagStore(testFile)
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
