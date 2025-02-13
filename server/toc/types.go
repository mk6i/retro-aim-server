package toc

import (
	"context"

	"github.com/google/uuid"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type BuddyService interface {
	AddBuddies(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x04_BuddyAddBuddies) error
	BroadcastBuddyDeparted(ctx context.Context, sess *state.Session) error
	DelBuddies(_ context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x05_BuddyDelBuddies) error
	RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
}

type ChatService interface {
	ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*wire.SNACMessage, error)
}

type ChatNavService interface {
	CreateRoom(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (wire.SNACMessage, error)
	ExchangeInfo(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo) (wire.SNACMessage, error)
	RequestChatRights(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	RequestRoomInfo(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (wire.SNACMessage, error)
}

type ICBMService interface {
	ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*wire.SNACMessage, error)
	ClientEvent(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x14_ICBMClientEvent) error
	EvilRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x08_ICBMEvilRequest) (wire.SNACMessage, error)
	ParameterQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	ClientErr(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x04_0x0B_ICBMClientErr) error
}

type OServiceService interface {
	ClientOnline(ctx context.Context, _ wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error
	IdleNotification(ctx context.Context, sess *state.Session, bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification) error
	ServiceRequest(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x04_OServiceServiceRequest) (wire.SNACMessage, error)
}

type AuthService interface {
	BUCPChallenge(bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (wire.SNACMessage, error)
	BUCPLogin(bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.SNACMessage, error)
	FLAPLogin(frame wire.FLAPSignonFrame, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.TLVRestBlock, error)
	RegisterBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	RetrieveBOSSession(authCookie []byte) (*state.Session, error)
	RegisterChatSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session)
	SignoutChat(ctx context.Context, sess *state.Session)
}

type LocateService interface {
	SetDirInfo(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x09_LocateSetDirInfo) (wire.SNACMessage, error)
	SetInfo(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x02_0x04_LocateSetInfo) error
	UserInfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery) (wire.SNACMessage, error)
	DirInfo(ctx context.Context, inFrame wire.SNACFrame, body wire.SNAC_0x02_0x0B_LocateGetDirInfo) (wire.SNACMessage, error)
}

type DirSearchService interface {
	InfoQuery(_ context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0F_0x02_InfoQuery) (wire.SNACMessage, error)
}

type PermitDenyService interface {
	AddDenyListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries) error
	AddPermListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries) error
	DelDenyListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries) error
	DelPermListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries) error
	RightsQuery(_ context.Context, frame wire.SNACFrame) wire.SNACMessage
}

// BuddyListRegistry is the interface for keeping track of users with active
// buddy lists. Once registered, a user becomes visible to other users' buddy
// lists and vice versa.
type BuddyListRegistry interface {
	RegisterBuddyList(user state.IdentScreenName) error
	UnregisterBuddyList(user state.IdentScreenName) error
}

type TOCConfigStore interface {
	SetTOCConfig(user state.IdentScreenName, config string) error
	User(screenName state.IdentScreenName) (*state.User, error)
}

type CookieBaker interface {
	Crack(data []byte) ([]byte, error)
	Issue(data []byte) ([]byte, error)
}

type AdminService interface {
	InfoChangeRequest(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x07_0x04_AdminInfoChangeRequest) (wire.SNACMessage, error)
}
