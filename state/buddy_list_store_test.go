package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdjListBuddyListStore_AddBuddy_Buddies(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy("alice", "bob")
	store.AddBuddy("bob", "alice")
	store.AddBuddy("alice", "charlie")
	store.AddBuddy("alice", "dave")
	store.AddBuddy("dave", "bob")

	buddies := store.Buddies("alice")
	assert.ElementsMatch(t, []string{"bob", "charlie", "dave"}, buddies)
}

func TestAdjListBuddyListStore_DeleteBuddy_OneBuddy(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy("alice", "bob")
	store.AddBuddy("bob", "alice")
	store.AddBuddy("alice", "charlie")
	store.AddBuddy("alice", "dave")
	store.AddBuddy("dave", "bob")

	store.DeleteBuddy("alice", "bob")

	buddies := store.Buddies("alice")
	assert.ElementsMatch(t, []string{"charlie", "dave"}, buddies)
}

func TestAdjListBuddyListStore_DeleteBuddy_AllBuddies(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy("alice", "bob")
	store.AddBuddy("bob", "alice")
	store.AddBuddy("alice", "charlie")
	store.AddBuddy("alice", "dave")
	store.AddBuddy("dave", "bob")

	store.DeleteBuddy("alice", "bob")
	store.DeleteBuddy("alice", "charlie")
	store.DeleteBuddy("alice", "dave")

	buddies := store.Buddies("alice")
	assert.Nil(t, buddies)
}

func TestAdjListBuddyListStore_DeleteUser(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy("alice", "bob")
	store.AddBuddy("bob", "alice")
	store.AddBuddy("alice", "charlie")
	store.AddBuddy("alice", "dave")
	store.AddBuddy("dave", "bob")

	store.DeleteUser("alice")

	buddies := store.Buddies("alice")
	assert.Nil(t, buddies)
}

func TestAdjListBuddyListStore_WhoAddedUser(t *testing.T) {
	store := NewAdjListBuddyListStore()
	store.AddBuddy("alice", "bob")
	store.AddBuddy("alice", "charlie")
	store.AddBuddy("charlie", "bob")
	store.AddBuddy("dave", "bob")
	store.AddBuddy("dave", "alive")

	whoAddedBob := store.WhoAddedUser("bob")
	assert.ElementsMatch(t, []string{"alice", "charlie", "dave"}, whoAddedBob)
}
