package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/mkaminski/goaim/oscar"
)

type ChatHandler interface {
	ChannelMsgToHostHandler(ctx context.Context, sess *Session, room ChatSessionManager, snacPayloadIn oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*oscar.XMessage, error)
}

func NewChatRouter(logger *slog.Logger) ChatRouter {
	return ChatRouter{
		ChatHandler: ChatService{},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type ChatRouter struct {
	ChatHandler
	RouteLogger
}

func (rt *ChatRouter) RouteChat(ctx context.Context, sess *Session, chatSessMgr ChatSessionManager, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.ChatChannelMsgToHost:
		inSNAC := oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ChannelMsgToHostHandler(ctx, sess, chatSessMgr, inSNAC)
		if err != nil {
			return err
		}
		if outSNAC == nil {
			return nil
		}
		rt.Logger.InfoContext(ctx, "user sent a chat message")
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type ChatService struct {
}

func (s ChatService) ChannelMsgToHostHandler(ctx context.Context, sess *Session, chatSessMgr ChatSessionManager, snacPayloadIn oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*oscar.XMessage, error) {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: oscar.CHAT,
		SubGroup:  oscar.ChatChannelMsgToClient,
	}
	snacPayloadOut := oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
		Cookie:  snacPayloadIn.Cookie,
		Channel: snacPayloadIn.Channel,
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: snacPayloadIn.TLVList,
		},
	}
	snacPayloadOut.AddTLV(
		oscar.NewTLV(oscar.ChatTLVSenderInformation, oscar.TLVUserInfo{
			ScreenName:   sess.ScreenName(),
			WarningLevel: sess.Warning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: sess.UserInfo(),
			},
		}),
	)

	// send message to all the participants except sender
	chatSessMgr.BroadcastExcept(ctx, sess, oscar.XMessage{
		SnacFrame: snacFrameOut,
		SnacOut:   snacPayloadOut,
	})

	var ret *oscar.XMessage
	if _, ackMsg := snacPayloadIn.GetTLV(oscar.ChatTLVEnableReflectionFlag); ackMsg {
		// reflect the message back to the sender
		ret = &oscar.XMessage{
			SnacFrame: snacFrameOut,
			SnacOut:   snacPayloadOut,
		}
	}

	return ret, nil
}

func SetOnlineChatUsers(ctx context.Context, sess *Session, chatSessMgr ChatSessionManager) {
	snacPayloadOut := oscar.SNAC_0x0E_0x03_ChatUsersJoined{}
	sessions := chatSessMgr.Participants()

	for _, uSess := range sessions {
		snacPayloadOut.Users = append(snacPayloadOut.Users, oscar.TLVUserInfo{
			ScreenName:   uSess.ScreenName(),
			WarningLevel: uSess.Warning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: uSess.UserInfo(),
			},
		})
	}

	chatSessMgr.SendToScreenName(ctx, sess.ScreenName(), oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatUsersJoined,
		},
		SnacOut: snacPayloadOut,
	})
}

func AlertUserJoined(ctx context.Context, sess *Session, chatSessMgr ChatSessionManager) {
	chatSessMgr.BroadcastExcept(ctx, sess, oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatUsersJoined,
		},
		SnacOut: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
			Users: []oscar.TLVUserInfo{
				{
					ScreenName:   sess.ScreenName(),
					WarningLevel: sess.Warning(),
					TLVBlock: oscar.TLVBlock{
						TLVList: sess.UserInfo(),
					},
				},
			},
		},
	})
}

func AlertUserLeft(ctx context.Context, sess *Session, chatSessMgr ChatSessionManager) {
	chatSessMgr.BroadcastExcept(ctx, sess, oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatUsersLeft,
		},
		SnacOut: oscar.SNAC_0x0E_0x04_ChatUsersLeft{
			Users: []oscar.TLVUserInfo{
				{
					ScreenName:   sess.ScreenName(),
					WarningLevel: sess.Warning(),
					TLVBlock: oscar.TLVBlock{
						TLVList: sess.UserInfo(),
					},
				},
			},
		},
	})
}

func SendChatRoomInfoUpdate(ctx context.Context, sess *Session, chatSessMgr ChatSessionManager, room ChatRoom) {
	chatSessMgr.SendToScreenName(ctx, sess.ScreenName(), oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatRoomInfoUpdate,
		},
		SnacOut: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
			Exchange:       4,
			Cookie:         room.Cookie,
			InstanceNumber: 100,
			DetailLevel:    2,
			TLVBlock: oscar.TLVBlock{
				TLVList: room.TLVList(),
			},
		},
	})
}

type ChatRoom struct {
	CreateTime     time.Time
	DetailLevel    uint8
	Exchange       uint16
	Cookie         string
	InstanceNumber uint16
	Name           string
}

func (c ChatRoom) TLVList() []oscar.TLV {
	return []oscar.TLV{
		oscar.NewTLV(0x00c9, uint16(15)),
		oscar.NewTLV(0x00ca, uint32(c.CreateTime.Unix())),
		oscar.NewTLV(0x00d1, uint16(1024)),
		oscar.NewTLV(0x00d2, uint16(100)),
		oscar.NewTLV(0x00d5, uint8(2)),
		oscar.NewTLV(0x006a, c.Name),
		oscar.NewTLV(0x00d3, c.Name),
	}
}

type ChatRegistry struct {
	chatRoomStore map[string]ChatRoom
	smStore       map[string]ChatSessionManager
	mapMutex      sync.RWMutex
}

func NewChatRegistry() *ChatRegistry {
	return &ChatRegistry{
		chatRoomStore: make(map[string]ChatRoom),
		smStore:       make(map[string]ChatSessionManager),
	}
}

func (c *ChatRegistry) Register(room ChatRoom, sm ChatSessionManager) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	c.chatRoomStore[room.Cookie] = room
	c.smStore[room.Cookie] = sm
}

func (c *ChatRegistry) Retrieve(chatID string) (ChatRoom, ChatSessionManager, error) {
	c.mapMutex.RLock()
	defer c.mapMutex.RUnlock()
	cr, found := c.chatRoomStore[chatID]
	if !found {
		return ChatRoom{}, nil, errors.New("unable to find chat room")
	}
	sm, found := c.smStore[chatID]
	if !found {
		panic("unable to find session manager for chat")
	}
	return cr, sm, nil
}

func (c *ChatRegistry) MaybeRemoveRoom(chatID string) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	sm, found := c.smStore[chatID]
	if found && sm.Empty() {
		delete(c.chatRoomStore, chatID)
		delete(c.smStore, chatID)
	}
}
