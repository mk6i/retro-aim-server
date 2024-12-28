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
	"net/mail"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type AccountManager interface {
	UpdateDisplayScreenName(displayScreenName state.DisplayScreenName) error
	UpdateEmailAddress(emailAddress *mail.Address, screenName state.IdentScreenName) error
	EmailAddressByName(screenName state.IdentScreenName) (*mail.Address, error)
	UpdateRegStatus(regStatus uint16, screenName state.IdentScreenName) error
	RegStatusByName(screenName state.IdentScreenName) (uint16, error)
	UpdateConfirmStatus(confirmStatus bool, screenName state.IdentScreenName) error
	ConfirmStatusByName(screnName state.IdentScreenName) (bool, error)
}

type BARTManager interface {
	BARTUpsert(itemHash []byte, payload []byte) error
	BARTRetrieve(itemHash []byte) ([]byte, error)
}

type buddyBroadcaster interface {
	BroadcastBuddyArrived(ctx context.Context, sess *state.Session) error
	BroadcastBuddyDeparted(ctx context.Context, sess *state.Session) error
	BroadcastVisibility(ctx context.Context, you *state.Session, filter []state.IdentScreenName, sendDepartures bool) error
}

type BuddyListRetriever interface {
	AllRelationships(screenName state.IdentScreenName, filter []state.IdentScreenName) ([]state.Relationship, error)
	BuddyIconRefByName(screenName state.IdentScreenName) (*wire.BARTID, error)
	Relationship(me state.IdentScreenName, them state.IdentScreenName) (state.Relationship, error)
}

// ChatMessageRelayer defines the interface for sending messages to chat room
// participants.
type ChatMessageRelayer interface {
	// AllSessions returns all chat room participants. Returns
	// ErrChatRoomNotFound if the room does not exist.
	AllSessions(chatCookie string) []*state.Session

	// RelayToAllExcept sends a message to all chat room participants except
	// for the participant with a particular screen name. Returns
	// ErrChatRoomNotFound if the room does not exist for cookie.
	RelayToAllExcept(ctx context.Context, chatCookie string, except state.IdentScreenName, msg wire.SNACMessage)

	// RelayToScreenName sends a message to a chat room user. Returns
	// ErrChatRoomNotFound if the room does not exist for cookie.
	RelayToScreenName(ctx context.Context, chatCookie string, recipient state.IdentScreenName, msg wire.SNACMessage)
}

// ChatRoomRegistry defines the interface for storing and retrieving chat
// rooms in a persistent store. The persistent store has two purposes:
// - Remember user-created chat rooms (exchange 4) so that clients can
// reconnect to the rooms following server restarts.
// - Keep track of public chat room created by the server operator (exchange
// 5). User's can only join public chat rooms that exist in the room registry.
type ChatRoomRegistry interface {
	// ChatRoomByCookie looks up a chat room by exchange. Returns
	// ErrChatRoomNotFound if the room does not exist for cookie.
	ChatRoomByCookie(chatCookie string) (state.ChatRoom, error)

	// ChatRoomByName looks up a chat room by exchange and name. Returns
	// ErrChatRoomNotFound if the room does not exist for exchange and name.
	ChatRoomByName(exchange uint16, name string) (state.ChatRoom, error)

	// CreateChatRoom creates a new chat room.
	CreateChatRoom(chatRoom *state.ChatRoom) error
}

// ChatSessionRegistry defines the interface for adding and removing chat
// sessions.
type ChatSessionRegistry interface {
	// AddSession adds a session to the chat session manager. The chatCookie
	// param identifies the chat room to which screenName is added. It returns
	// the newly created session instance registered in the chat session
	// manager.
	AddSession(ctx context.Context, chatCookie string, screenName state.DisplayScreenName) (*state.Session, error)

	// RemoveSession removes a session from the chat session manager.
	RemoveSession(sess *state.Session)
}

type CookieBaker interface {
	Crack(data []byte) ([]byte, error)
	Issue(data []byte) ([]byte, error)
}

type FeedbagManager interface {
	Feedbag(screenName state.IdentScreenName) ([]wire.FeedbagItem, error)
	FeedbagDelete(screenName state.IdentScreenName, items []wire.FeedbagItem) error
	FeedbagLastModified(screenName state.IdentScreenName) (time.Time, error)
	FeedbagUpsert(screenName state.IdentScreenName, items []wire.FeedbagItem) error
	UseFeedbag(user state.IdentScreenName) error
}

type ICQUserFinder interface {
	// FindByUIN returns a user with a matching UIN.
	FindByUIN(UIN uint32) (state.User, error)
	// FindByICQEmail returns a user with a matching email address.
	FindByICQEmail(email string) (state.User, error)
	// FindByICQName returns users with matching first name, last name, and
	// nickname. Empty values are not included in the search parameters.
	FindByICQName(firstName, lastName, nickName string) ([]state.User, error)
	// FindByICQInterests returns users who have at least one matching interest
	// for a given category code.
	FindByICQInterests(code uint16, keywords []string) ([]state.User, error)
	// FindByICQKeyword returns users with matching interest keyword across all
	// interest categories.
	FindByICQKeyword(keyword string) ([]state.User, error)
}

type ICQUserUpdater interface {
	SetAffiliations(name state.IdentScreenName, data state.ICQAffiliations) error
	SetBasicInfo(name state.IdentScreenName, data state.ICQBasicInfo) error
	SetInterests(name state.IdentScreenName, data state.ICQInterests) error
	SetMoreInfo(name state.IdentScreenName, data state.ICQMoreInfo) error
	SetUserNotes(name state.IdentScreenName, data state.ICQUserNotes) error
	SetWorkInfo(name state.IdentScreenName, data state.ICQWorkInfo) error
}

type LocalBuddyListManager interface {
	AddBuddy(me state.IdentScreenName, them state.IdentScreenName) error
	RemoveBuddy(me state.IdentScreenName, them state.IdentScreenName) error
	DenyBuddy(me state.IdentScreenName, them state.IdentScreenName) error
	PermitBuddy(me state.IdentScreenName, them state.IdentScreenName) error
	RemoveDenyBuddy(me state.IdentScreenName, them state.IdentScreenName) error
	RemovePermitBuddy(me state.IdentScreenName, them state.IdentScreenName) error
	SetPDMode(user state.IdentScreenName, pdMode wire.FeedbagPDMode) error
}

type MessageRelayer interface {
	RelayToScreenNames(ctx context.Context, screenNames []state.IdentScreenName, msg wire.SNACMessage)
	RelayToScreenName(ctx context.Context, screenName state.IdentScreenName, msg wire.SNACMessage)
}

type OfflineMessageManager interface {
	DeleteMessages(recip state.IdentScreenName) error
	RetrieveMessages(recip state.IdentScreenName) ([]state.OfflineMessage, error)
	SaveMessage(offlineMessage state.OfflineMessage) error
}

type ProfileManager interface {
	FindByAIMEmail(email string) (state.User, error)
	FindByAIMKeyword(keyword string) ([]state.User, error)
	FindByAIMNameAndAddr(info state.AIMNameAndAddr) ([]state.User, error)
	InterestList() ([]wire.ODirKeywordListItem, error)
	Profile(screenName state.IdentScreenName) (string, error)
	SetDirectoryInfo(name state.IdentScreenName, info state.AIMNameAndAddr) error
	SetKeywords(name state.IdentScreenName, keywords [5]string) error
	SetProfile(screenName state.IdentScreenName, body string) error
	User(screenName state.IdentScreenName) (*state.User, error)
}

type SessionRegistry interface {
	AddSession(ctx context.Context, screenName state.DisplayScreenName) (*state.Session, error)
	RemoveSession(sess *state.Session)
}

type SessionRetriever interface {
	RetrieveSession(screenName state.IdentScreenName) *state.Session
}

type UserManager interface {
	User(screenName state.IdentScreenName) (*state.User, error)
	InsertUser(u state.User) error
}
