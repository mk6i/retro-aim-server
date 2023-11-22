package handler

import (
	"context"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

func NewChatService(chatRegistry *state.ChatRegistry) *ChatService {
	return &ChatService{
		chatRegistry: chatRegistry,
	}
}

type ChatService struct {
	chatRegistry *state.ChatRegistry
}

func (s ChatService) ChannelMsgToHostHandler(ctx context.Context, sess *state.Session, chatID string, snacPayloadIn oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*oscar.XMessage, error) {
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

	_, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		return nil, err
	}

	// send message to all the participants except sender
	chatSessMgr.(ChatSessionManager).BroadcastExcept(ctx, sess, oscar.XMessage{
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

func setOnlineChatUsers(ctx context.Context, sess *state.Session, chatSessMgr ChatSessionManager) {
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

func alertUserJoined(ctx context.Context, sess *state.Session, chatSessMgr ChatSessionManager) {
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

func alertUserLeft(ctx context.Context, sess *state.Session, chatSessMgr ChatSessionManager) {
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

func sendChatRoomInfoUpdate(ctx context.Context, sess *state.Session, chatSessMgr ChatSessionManager, room state.ChatRoom) {
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
