package server

import (
	"context"
	"github.com/mkaminski/goaim/oscar"
	"time"
)

type UserManager interface {
	GetUser(screenName string) (*User, error)
	UpsertUser(u User) error
}

type FeedbagManager interface {
	Blocked(sn1, sn2 string) (BlockedState, error)
	Buddies(screenName string) ([]string, error)
	Delete(screenName string, items []oscar.FeedbagItem) error
	InterestedUsers(screenName string) ([]string, error)
	LastModified(screenName string) (time.Time, error)
	Retrieve(screenName string) ([]oscar.FeedbagItem, error)
	Upsert(screenName string, items []oscar.FeedbagItem) error
}

type SessionManager interface {
	Broadcast(ctx context.Context, msg oscar.XMessage)
	BroadcastToScreenNames(ctx context.Context, screenNames []string, msg oscar.XMessage)
	Empty() bool
	NewSessionWithSN(sessID string, screenName string) *Session
	Remove(sess *Session)
	Retrieve(ID string) (*Session, bool)
	RetrieveByScreenName(screenName string) *Session
	SendToScreenName(ctx context.Context, screenName string, msg oscar.XMessage)
}

type ChatSessionManager interface {
	SessionManager
	BroadcastExcept(ctx context.Context, except *Session, msg oscar.XMessage)
	Participants() []*Session
}

type ProfileManager interface {
	RetrieveProfile(screenName string) (string, error)
	UpsertProfile(screenName string, body string) error
}
