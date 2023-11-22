package handler

import (
	"context"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type FeedbagManager interface {
	Blocked(sn1, sn2 string) (state.BlockedState, error)
	Buddies(screenName string) ([]string, error)
	Delete(screenName string, items []oscar.FeedbagItem) error
	InterestedUsers(screenName string) ([]string, error)
	LastModified(screenName string) (time.Time, error)
	Retrieve(screenName string) ([]oscar.FeedbagItem, error)
	Upsert(screenName string, items []oscar.FeedbagItem) error
}

type UserManager interface {
	GetUser(screenName string) (*state.User, error)
	UpsertUser(u state.User) error
}

type SessionManager interface {
	Broadcast(ctx context.Context, msg oscar.SNACMessage)
	BroadcastToScreenNames(ctx context.Context, screenNames []string, msg oscar.SNACMessage)
	Empty() bool
	NewSessionWithSN(sessID string, screenName string) *state.Session
	Remove(sess *state.Session)
	Retrieve(ID string) (*state.Session, bool)
	RetrieveByScreenName(screenName string) *state.Session
	SendToScreenName(ctx context.Context, screenName string, msg oscar.SNACMessage)
}

type ChatSessionManager interface {
	SessionManager
	BroadcastExcept(ctx context.Context, except *state.Session, msg oscar.SNACMessage)
	Participants() []*state.Session
}

type ProfileManager interface {
	RetrieveProfile(screenName string) (string, error)
	UpsertProfile(screenName string, body string) error
}
