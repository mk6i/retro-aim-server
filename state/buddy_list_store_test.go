package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdjListBuddyListStore_AddBuddy_Buddies(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("bob"))
	store.AddBuddy(NewIdentScreenName("bob"), NewIdentScreenName("alice"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("charlie"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("dave"))
	store.AddBuddy(NewIdentScreenName("dave"), NewIdentScreenName("bob"))

	buddies := store.Buddies(NewIdentScreenName("alice"))
	assert.ElementsMatch(t, []IdentScreenName{
		NewIdentScreenName("bob"),
		NewIdentScreenName("charlie"),
		NewIdentScreenName("dave"),
	}, buddies)
}

func TestAdjListBuddyListStore_DeleteBuddy_OneBuddy(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("bob"))
	store.AddBuddy(NewIdentScreenName("bob"), NewIdentScreenName("alice"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("charlie"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("dave"))
	store.AddBuddy(NewIdentScreenName("dave"), NewIdentScreenName("bob"))

	store.DeleteBuddy(NewIdentScreenName("alice"), NewIdentScreenName("bob"))

	buddies := store.Buddies(NewIdentScreenName("alice"))
	assert.ElementsMatch(t, []IdentScreenName{
		NewIdentScreenName("charlie"),
		NewIdentScreenName("dave"),
	}, buddies)
}

func TestAdjListBuddyListStore_DeleteBuddy_AllBuddies(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("bob"))
	store.AddBuddy(NewIdentScreenName("bob"), NewIdentScreenName("alice"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("charlie"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("dave"))
	store.AddBuddy(NewIdentScreenName("dave"), NewIdentScreenName("bob"))

	store.DeleteBuddy(NewIdentScreenName("alice"), NewIdentScreenName("bob"))
	store.DeleteBuddy(NewIdentScreenName("alice"), NewIdentScreenName("charlie"))
	store.DeleteBuddy(NewIdentScreenName("alice"), NewIdentScreenName("dave"))

	buddies := store.Buddies(NewIdentScreenName("alice"))
	assert.Nil(t, buddies)
}

func TestAdjListBuddyListStore_DeleteUser(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("bob"))
	store.AddBuddy(NewIdentScreenName("bob"), NewIdentScreenName("alice"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("charlie"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("dave"))
	store.AddBuddy(NewIdentScreenName("dave"), NewIdentScreenName("bob"))

	store.DeleteUser(NewIdentScreenName("alice"))

	buddies := store.Buddies(NewIdentScreenName("alice"))
	assert.Nil(t, buddies)
}

func TestAdjListBuddyListStore_WhoAddedUser(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("bob"))
	store.AddBuddy(NewIdentScreenName("alice"), NewIdentScreenName("charlie"))
	store.AddBuddy(NewIdentScreenName("charlie"), NewIdentScreenName("bob"))
	store.AddBuddy(NewIdentScreenName("dave"), NewIdentScreenName("bob"))
	store.AddBuddy(NewIdentScreenName("dave"), NewIdentScreenName("alive"))

	whoAddedBob := store.WhoAddedUser(NewIdentScreenName("bob"))
	assert.ElementsMatch(t, []IdentScreenName{
		NewIdentScreenName("alice"),
		NewIdentScreenName("charlie"),
		NewIdentScreenName("dave"),
	}, whoAddedBob)
}
