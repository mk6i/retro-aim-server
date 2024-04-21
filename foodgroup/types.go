// Package foodgroup implements OSCAR food group business logic.
//
// The OSCAR protocol passes messages in SNAC (Simple Network Atomic
// Communication) format. SNAC messages are grouped by "food groups" (get it?
// snack, snac, foodgroup...). Each food group is responsible for a discrete
// piece of functionality, such as buddy list management (Feedbag), instant
// messaging (ICBM), and chat messaging (Chat).
//
// Each food group operation is represented by a struct type. The methods
// correspond 1:1 to each food group operation. Each food group operation is
// typically triggered by a client request. The operation may return a
// response. As such, methods receive client requests via SNAC frame and
// body parameters and send responses via returned SNAC objects.
//
// The following is a typical food group method signature. This example
// illustrates the ICBM ChannelMsgToHost operation.
//
//	ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*wire.SNACMessage, error)
//
// Params:
//   - ctx context.Context is the client request context.
//   - sess *state.Session is the client's session object.
//   - inFrame wire.SNACFrame is the request SNAC frame that contains the food group and subgroup parameters.
//   - inBody wire.SNAC_0x04_0x06_ICBMChannelMsgToHost contains the body of the SNAC message. In this case, it contains instant message text and metadata.
//
// ChannelMsgToHost optionally sends a client response by returning
// *wire.SNACMessage. For operations that always send client responses,
// the methods return wire.SNACMessage value types (not pointer types).
// Methods for operations that never send client responses do not return
// wire.SNACMessage values.
//
// The foodgroup package delegates responsibility for message transport, user
// retrieval, and session management to callers via several interface types.
package foodgroup

import (
	"context"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type FeedbagManager interface {
	BlockedState(screenName1, screenName2 string) (state.BlockedState, error)
	Buddies(screenName string) ([]string, error)
	FeedbagDelete(screenName string, items []wire.FeedbagItem) error
	AdjacentUsers(screenName string) ([]string, error)
	FeedbagLastModified(screenName string) (time.Time, error)
	Feedbag(screenName string) ([]wire.FeedbagItem, error)
	FeedbagUpsert(screenName string, items []wire.FeedbagItem) error
}

// LegacyBuddyListManager defines operations for tracking user relationships
// for the client-side buddy list system used by clients prior to AIM version
// 4.3.
type LegacyBuddyListManager interface {
	// AddBuddy adds buddyScreenName to userScreenName's buddy list.
	AddBuddy(userScreenName, buddyScreenName string)

	// Buddies returns a list of all buddies associated with the specified
	// userScreenName.
	Buddies(userScreenName string) []string

	// DeleteBuddy removes buddyScreenName from userScreenName's buddy list.
	DeleteBuddy(userScreenName, buddyScreenName string)

	// DeleteUser removes userScreenName's buddy list.
	DeleteUser(userScreenName string)

	// WhoAddedUser returns a list of screen names who have userScreenName in
	// their buddy lists.
	WhoAddedUser(userScreenName string) []string
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
	RelayToScreenNames(ctx context.Context, screenNames []string, msg wire.SNACMessage)
	RetrieveByScreenName(screenName string) *state.Session
	RelayToScreenName(ctx context.Context, screenName string, msg wire.SNACMessage)
}

type ChatMessageRelayer interface {
	MessageRelayer
	RelayToAllExcept(ctx context.Context, except *state.Session, msg wire.SNACMessage)
	AllSessions() []*state.Session
}

type ChatRegistry interface {
	Register(room state.ChatRoom, sessionManager any)
	Retrieve(cookie string) (state.ChatRoom, any, error)
	Remove(cookie string)
}

type BARTManager interface {
	BARTUpsert(itemHash []byte, payload []byte) error
	BARTRetrieve(itemHash []byte) ([]byte, error)
}
