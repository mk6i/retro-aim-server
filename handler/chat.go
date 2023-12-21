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

func (s ChatService) ChannelMsgToHostHandler(ctx context.Context, sess *state.Session, chatID string, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*oscar.SNACMessage, error) {
	frameOut := oscar.SNACFrame{
		FoodGroup: oscar.Chat,
		SubGroup:  oscar.ChatChannelMsgToClient,
	}
	bodyOut := oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
		Cookie:  inBody.Cookie,
		Channel: inBody.Channel,
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: inBody.TLVList,
		},
	}
	bodyOut.AddTLV(oscar.NewTLV(oscar.ChatTLVSenderInformation, sess.TLVUserInfo()))

	_, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		return nil, err
	}

	// send message to all the participants except sender
	chatSessMgr.(ChatMessageRelayer).RelayToAllExcept(ctx, sess, oscar.SNACMessage{
		Frame: frameOut,
		Body:  bodyOut,
	})

	var ret *oscar.SNACMessage
	if _, ackMsg := inBody.GetTLV(oscar.ChatTLVEnableReflectionFlag); ackMsg {
		// reflect the message back to the sender
		ret = &oscar.SNACMessage{
			Frame: frameOut,
			Body:  bodyOut,
		}
		ret.Frame.RequestID = inFrame.RequestID
	}

	return ret, nil
}

func setOnlineChatUsers(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer) {
	snacPayloadOut := oscar.SNAC_0x0E_0x03_ChatUsersJoined{}
	sessions := chatMessageRelayer.AllSessions()

	for _, uSess := range sessions {
		snacPayloadOut.Users = append(snacPayloadOut.Users, uSess.TLVUserInfo())
	}

	chatMessageRelayer.RelayToScreenName(ctx, sess.ScreenName(), oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Chat,
			SubGroup:  oscar.ChatUsersJoined,
		},
		Body: snacPayloadOut,
	})
}

func alertUserJoined(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer) {
	chatMessageRelayer.RelayToAllExcept(ctx, sess, oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Chat,
			SubGroup:  oscar.ChatUsersJoined,
		},
		Body: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
			Users: []oscar.TLVUserInfo{
				sess.TLVUserInfo(),
			},
		},
	})
}

func alertUserLeft(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer) {
	chatMessageRelayer.RelayToAllExcept(ctx, sess, oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Chat,
			SubGroup:  oscar.ChatUsersLeft,
		},
		Body: oscar.SNAC_0x0E_0x04_ChatUsersLeft{
			Users: []oscar.TLVUserInfo{
				sess.TLVUserInfo(),
			},
		},
	})
}

func sendChatRoomInfoUpdate(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer, room state.ChatRoom) {
	chatMessageRelayer.RelayToScreenName(ctx, sess.ScreenName(), oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Chat,
			SubGroup:  oscar.ChatRoomInfoUpdate,
		},
		Body: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
			Exchange:       room.Exchange,
			Cookie:         room.Cookie,
			InstanceNumber: room.InstanceNumber,
			DetailLevel:    room.DetailLevel,
			TLVBlock: oscar.TLVBlock{
				TLVList: room.TLVList(),
			},
		},
	})
}
