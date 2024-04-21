package state

import "sync"

// AdjListBuddyListStore implements a buddy list using an adjacency list.
type AdjListBuddyListStore struct {
	buddies map[string]map[string]bool
	mu      sync.RWMutex // ensures thread-safe access
}

// NewAdjListBuddyListStore initializes a new instance of AdjListBuddyListStore.
func NewAdjListBuddyListStore() *AdjListBuddyListStore {
	return &AdjListBuddyListStore{
		buddies: make(map[string]map[string]bool),
	}
}

// AddBuddy adds buddyScreenName to userScreenName's buddy list.
func (store *AdjListBuddyListStore) AddBuddy(userScreenName, buddyScreenName string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.buddies[userScreenName]; !exists {
		store.buddies[userScreenName] = make(map[string]bool)
	}
	store.buddies[userScreenName][buddyScreenName] = true
}

// WhoAddedUser returns a list of screen names who have userScreenName in their buddy lists.
func (store *AdjListBuddyListStore) WhoAddedUser(userScreenName string) []string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var users []string
	for user, buddies := range store.buddies {
		if buddies[userScreenName] {
			users = append(users, user)
		}
	}
	return users
}

// Buddies returns a list of all buddies associated with the specified userScreenName.
func (store *AdjListBuddyListStore) Buddies(userScreenName string) []string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if buddies, exists := store.buddies[userScreenName]; exists {
		users := make([]string, 0, len(buddies))
		for buddy := range buddies {
			users = append(users, buddy)
		}
		return users
	}
	return nil
}

// DeleteBuddy removes buddyScreenName from userScreenName's buddy list.
func (store *AdjListBuddyListStore) DeleteBuddy(userScreenName, buddyScreenName string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if buddies, exists := store.buddies[userScreenName]; exists {
		delete(buddies, buddyScreenName)
		if len(buddies) == 0 {
			delete(store.buddies, userScreenName)
		}
	}
}

// DeleteUser removes userScreenName's buddy list.
func (store *AdjListBuddyListStore) DeleteUser(userScreenName string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.buddies, userScreenName)
}
