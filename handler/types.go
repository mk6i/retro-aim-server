package handler

import (
	"context"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type FeedbagManager interface {
	BlockedState(screenName1, screenName2 string) (state.BlockedState, error)
	Buddies(screenName string) ([]string, error)
	FeedbagDelete(screenName string, items []oscar.FeedbagItem) error
	AdjacentUsers(screenName string) ([]string, error)
	FeedbagLastModified(screenName string) (time.Time, error)
	Feedbag(screenName string) ([]oscar.FeedbagItem, error)
	FeedbagUpsert(screenName string, items []oscar.FeedbagItem) error
}

type UserManager interface {
	User(screenName string) (*state.User, error)
	InsertUser(u state.User) error
}

type SessionManager interface {
	Empty() bool
	AddSession(sessID string, screenName string) *state.Session
	RemoveSession(sess *state.Session)
	RetrieveSession(ID string) *state.Session
}

type ProfileManager interface {
	Profile(screenName string) (string, error)
	SetProfile(screenName string, body string) error
}

type MessageRelayer interface {
	RelayToScreenNames(ctx context.Context, screenNames []string, msg oscar.SNACMessage)
	RetrieveByScreenName(screenName string) *state.Session
	RelayToScreenName(ctx context.Context, screenName string, msg oscar.SNACMessage)
}

type ChatMessageRelayer interface {
	MessageRelayer
	RelayToAllExcept(ctx context.Context, except *state.Session, msg oscar.SNACMessage)
	AllSessions() []*state.Session
}

type ChatRegistry interface {
	Register(room state.ChatRoom, sessionManager any)
	Retrieve(chatID string) (state.ChatRoom, any, error)
	Remove(chatID string)
}
