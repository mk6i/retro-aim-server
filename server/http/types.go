package http

import (
	"context"
	"net/mail"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type ChatRoomRetriever interface {
	AllChatRooms(exchange uint16) ([]state.ChatRoom, error)
}

type ChatRoomCreator interface {
	CreateChatRoom(chatRoom *state.ChatRoom) error
}

type ChatSessionRetriever interface {
	AllSessions(cookie string) []*state.Session
}

type SessionRetriever interface {
	AllSessions() []*state.Session
	RetrieveByScreenName(screenName state.IdentScreenName) *state.Session
}

type UserManager interface {
	AllUsers() ([]state.User, error)
	DeleteUser(screenName state.IdentScreenName) error
	InsertUser(u state.User) error
	SetUserPassword(screenName state.IdentScreenName, newPassword string) error
	User(screenName state.IdentScreenName) (*state.User, error)
}

type MessageRelayer interface {
	RelayToScreenName(ctx context.Context, screenName state.IdentScreenName, msg wire.SNACMessage)
}

type AccountRetriever interface {
	EmailAddressByName(screenName state.IdentScreenName) (*mail.Address, error)
	RegStatusByName(screenName state.IdentScreenName) (uint16, error)
	ConfirmStatusByName(screnName state.IdentScreenName) (bool, error)
}

type BARTRetriever interface {
	BARTRetrieve(itemHash []byte) ([]byte, error)
}

type FeedBagRetriever interface {
	BuddyIconRefByName(screenName state.IdentScreenName) (*wire.BARTID, error)
}

type userWithPassword struct {
	ScreenName string `json:"screen_name"`
	Password   string `json:"password,omitempty"`
}

type onlineUsers struct {
	Count    int             `json:"count"`
	Sessions []sessionHandle `json:"sessions"`
}

type userHandle struct {
	ID         string `json:"id"`
	ScreenName string `json:"screen_name"`
	IsICQ      bool   `json:"is_icq"`
}

type aimChatUserHandle struct {
	ID         string `json:"id"`
	ScreenName string `json:"screen_name"`
}

type userAccountHandle struct {
	ID           string `json:"id"`
	ScreenName   string `json:"screen_name"`
	Profile      string `json:"profile"`
	EmailAddress string `json:"email_address"`
	RegStatus    uint16 `json:"reg_status"`
	Confirmed    bool   `json:"confirmed"`
	IsICQ        bool   `json:"is_icq"`
}

type sessionHandle struct {
	ID            string  `json:"id"`
	ScreenName    string  `json:"screen_name"`
	OnlineSeconds float64 `json:"online_seconds"`
	AwayMessage   string  `json:"away_message"`
	IdleSeconds   float64 `json:"idle_seconds"`
	IsICQ         bool    `json:"is_icq"`
}

type ProfileRetriever interface {
	Profile(screenName state.IdentScreenName) (string, error)
}

type chatRoomCreate struct {
	Name string `json:"name"`
}

type chatRoom struct {
	Name         string              `json:"name"`
	CreateTime   time.Time           `json:"create_time"`
	CreatorID    string              `json:"creator_id,omitempty"`
	URL          string              `json:"url"`
	Participants []aimChatUserHandle `json:"participants"`
}

type instantMessage struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}
