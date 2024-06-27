package http

import (
	"time"

	"github.com/mk6i/retro-aim-server/state"
)

type ChatRoomRetriever interface {
	AllChatRooms(exchange uint16) ([]state.ChatRoom, error)
}

type ChatRoomCreator interface {
	CreateChatRoom(chatRoom state.ChatRoom) error
}

type ChatSessionRetriever interface {
	AllSessions(cookie string) []*state.Session
}

type SessionRetriever interface {
	AllSessions() []*state.Session
}

type UserManager interface {
	AllUsers() ([]state.User, error)
	DeleteUser(screenName state.IdentScreenName) error
	InsertUser(u state.User) error
	SetUserPassword(u state.User) error
	User(screenName state.IdentScreenName) (*state.User, error)
}

type userWithPassword struct {
	state.User
	Password string `json:"password,omitempty"`
}

type userSession struct {
	ScreenName string `json:"screen_name"`
}

type onlineUsers struct {
	Count    int           `json:"count"`
	Sessions []userSession `json:"sessions"`
}

type userHandle struct {
	ID         string `json:"id"`
	ScreenName string `json:"screen_name"`
}

type chatRoomCreate struct {
	Name string `json:"name"`
}

type chatRoom struct {
	Name         string       `json:"name"`
	CreateTime   time.Time    `json:"create_time"`
	CreatorID    string       `json:"creator_id,omitempty"`
	URL          string       `json:"url"`
	Participants []userHandle `json:"participants"`
}
