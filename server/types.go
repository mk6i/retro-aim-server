package server

import (
	"context"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/user"
	"time"
)

type FeedbagManager interface {
	Blocked(sn1, sn2 string) (user.BlockedState, error)
	Buddies(screenName string) ([]string, error)
	Delete(screenName string, items []oscar.FeedbagItem) error
	InterestedUsers(screenName string) ([]string, error)
	LastModified(screenName string) (time.Time, error)
	Retrieve(screenName string) ([]oscar.FeedbagItem, error)
	Upsert(screenName string, items []oscar.FeedbagItem) error
}

type SessionManager interface {
	Broadcast(ctx context.Context, msg oscar.XMessage)
	BroadcastExcept(ctx context.Context, except *user.Session, msg oscar.XMessage)
	BroadcastToScreenNames(ctx context.Context, screenNames []string, msg oscar.XMessage)
	Empty() bool
	NewSessionWithSN(sessID string, screenName string) *user.Session
	Participants() []*user.Session
	Remove(sess *user.Session)
	Retrieve(ID string) (*user.Session, bool)
	RetrieveByScreenName(screenName string) *user.Session
	SendToScreenName(ctx context.Context, screenName string, msg oscar.XMessage)
}

type ProfileManager interface {
	RetrieveProfile(screenName string) (string, error)
	UpsertProfile(screenName string, body string) error
}
