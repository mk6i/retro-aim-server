package foodgroup

import (
	"context"
	"errors"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewChatService creates a new instance of ChatService.
func NewChatService(chatMessageRelayer ChatMessageRelayer) *ChatService {
	return &ChatService{
		chatMessageRelayer: chatMessageRelayer,
	}
}

// ChatService provides functionality for the Chat food group, which is
// responsible for sending and receiving chat messages.
type ChatService struct {
	chatMessageRelayer ChatMessageRelayer
}

// ChannelMsgToHost relays wire.ChatChannelMsgToClient SNAC sent from a user
// to the other chat room participants. It returns the same
// wire.ChatChannelMsgToClient message back to the user if the chat reflection
// TLV flag is set, otherwise return nil.
func (s ChatService) ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*wire.SNACMessage, error) {
	frameOut := wire.SNACFrame{
		FoodGroup: wire.Chat,
		SubGroup:  wire.ChatChannelMsgToClient,
	}

	msg, hasMessage := inBody.Slice(wire.ChatTLVMessageInformation)
	if !hasMessage {
		return nil, errors.New("SNAC(0x0E,0x05) does not contain a message TLV")
	}

	bodyOut := wire.SNAC_0x0E_0x06_ChatChannelMsgToClient{
		Cookie:  inBody.Cookie,
		Channel: inBody.Channel,
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				// The order of these TLVs matters for AIM 2.x. if out of
				// order, screen names do not appear with each chat message.
				wire.NewTLV(wire.ChatTLVSenderInformation, sess.TLVUserInfo()),
				wire.NewTLV(wire.ChatTLVPublicWhisperFlag, []byte{}),
				wire.NewTLV(wire.ChatTLVMessageInformation, msg),
			},
		},
	}

	// send message to all the participants except sender
	s.chatMessageRelayer.RelayToAllExcept(ctx, sess.ChatRoomCookie(), sess.IdentScreenName(), wire.SNACMessage{
		Frame: frameOut,
		Body:  bodyOut,
	})

	var ret *wire.SNACMessage
	if _, ackMsg := inBody.Slice(wire.ChatTLVEnableReflectionFlag); ackMsg {
		// reflect the message back to the sender
		ret = &wire.SNACMessage{
			Frame: frameOut,
			Body:  bodyOut,
		}
		ret.Frame.RequestID = inFrame.RequestID
	}

	return ret, nil
}

func setOnlineChatUsers(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer) {
	snacPayloadOut := wire.SNAC_0x0E_0x03_ChatUsersJoined{}
	sessions := chatMessageRelayer.AllSessions(sess.ChatRoomCookie())

	for _, uSess := range sessions {
		snacPayloadOut.Users = append(snacPayloadOut.Users, uSess.TLVUserInfo())
	}

	chatMessageRelayer.RelayToScreenName(ctx, sess.ChatRoomCookie(), sess.IdentScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatUsersJoined,
		},
		Body: snacPayloadOut,
	})
}

func alertUserJoined(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer) {
	chatMessageRelayer.RelayToAllExcept(ctx, sess.ChatRoomCookie(), sess.IdentScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatUsersJoined,
		},
		Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
			Users: []wire.TLVUserInfo{
				sess.TLVUserInfo(),
			},
		},
	})
}

func alertUserLeft(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer) {
	chatMessageRelayer.RelayToAllExcept(ctx, sess.ChatRoomCookie(), sess.IdentScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatUsersLeft,
		},
		Body: wire.SNAC_0x0E_0x04_ChatUsersLeft{
			Users: []wire.TLVUserInfo{
				sess.TLVUserInfo(),
			},
		},
	})
}

func sendChatRoomInfoUpdate(ctx context.Context, sess *state.Session, chatMessageRelayer ChatMessageRelayer, room state.ChatRoom) {
	chatMessageRelayer.RelayToScreenName(ctx, sess.ChatRoomCookie(), sess.IdentScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatRoomInfoUpdate,
		},
		Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
			Exchange:       room.Exchange,
			Cookie:         room.Cookie,
			InstanceNumber: room.InstanceNumber,
			DetailLevel:    detailLevel,
			TLVBlock: wire.TLVBlock{
				TLVList: room.TLVList(),
			},
		},
	})
}
