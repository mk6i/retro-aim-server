package handler

import (
	"context"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type FeedbagManager interface {
	Blocked(screenName1, screenName2 string) (state.BlockedState, error)
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
	Empty() bool
	NewSessionWithSN(sessID string, screenName string) *state.Session
	Remove(sess *state.Session)
	Retrieve(ID string) *state.Session
}

type ProfileManager interface {
	RetrieveProfile(screenName string) (string, error)
	UpsertProfile(screenName string, body string) error
}

type MessageRelayer interface {
	BroadcastToScreenNames(ctx context.Context, screenNames []string, msg oscar.SNACMessage)
	RetrieveByScreenName(screenName string) *state.Session
	SendToScreenName(ctx context.Context, screenName string, msg oscar.SNACMessage)
}

type ChatMessageRelayer interface {
	MessageRelayer
	BroadcastExcept(ctx context.Context, except *state.Session, msg oscar.SNACMessage)
	Participants() []*state.Session
}

type ChatRegistry interface {
	Register(room state.ChatRoom, sessionManager any)
	Retrieve(chatID string) (state.ChatRoom, any, error)
	RemoveRoom(chatID string)
}
