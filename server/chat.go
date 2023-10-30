package server

import (
	"io"

	"github.com/mkaminski/goaim/oscar"
)

type ChatHandler interface {
	ChannelMsgToHostHandler(sess *Session, sm SessionManager, snacPayloadIn oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*XMessage, error)
}

func NewChatRouter() ChatRouter {
	return ChatRouter{
		ChatHandler: ChatService{},
	}
}

type ChatRouter struct {
	ChatHandler
}

func (rt *ChatRouter) RouteChat(sess *Session, sm SessionManager, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.ChatChannelMsgToHost:
		inSNAC := oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ChannelMsgToHostHandler(sess, sm, inSNAC)
		if err != nil || outSNAC == nil {
			return err
		}
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type ChatService struct {
}

func (s ChatService) ChannelMsgToHostHandler(sess *Session, sm SessionManager, snacPayloadIn oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*XMessage, error) {
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
			ScreenName:   sess.ScreenName,
			WarningLevel: sess.GetWarning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: sess.GetUserInfo(),
			},
		}),
	)

	// send message to all the participants except sender
	sm.BroadcastExcept(sess, XMessage{
		snacFrame: snacFrameOut,
		snacOut:   snacPayloadOut,
	})

	var ret *XMessage
	if _, ackMsg := snacPayloadIn.GetTLV(oscar.ChatTLVEnableReflectionFlag); ackMsg {
		// reflect the message back to the sender
		ret = &XMessage{
			snacFrame: snacFrameOut,
			snacOut:   snacPayloadOut,
		}
	}

	return ret, nil
}

func SetOnlineChatUsers(sess *Session, sm SessionManager) {
	snacPayloadOut := oscar.SNAC_0x0E_0x03_ChatUsersJoined{}
	sessions := sm.Participants()

	for _, uSess := range sessions {
		snacPayloadOut.Users = append(snacPayloadOut.Users, oscar.TLVUserInfo{
			ScreenName:   uSess.ScreenName,
			WarningLevel: uSess.GetWarning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: uSess.GetUserInfo(),
			},
		})
	}

	sm.SendToScreenName(sess.ScreenName, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatUsersJoined,
		},
		snacOut: snacPayloadOut,
	})
}

func AlertUserJoined(sess *Session, sm SessionManager) {
	sm.BroadcastExcept(sess, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatUsersJoined,
		},
		snacOut: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
			Users: []oscar.TLVUserInfo{
				{
					ScreenName:   sess.ScreenName,
					WarningLevel: sess.GetWarning(),
					TLVBlock: oscar.TLVBlock{
						TLVList: sess.GetUserInfo(),
					},
				},
			},
		},
	})
}

func AlertUserLeft(sess *Session, sm SessionManager) {
	sm.BroadcastExcept(sess, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatUsersLeft,
		},
		snacOut: oscar.SNAC_0x0E_0x04_ChatUsersLeft{
			Users: []oscar.TLVUserInfo{
				{
					ScreenName:   sess.ScreenName,
					WarningLevel: sess.GetWarning(),
					TLVBlock: oscar.TLVBlock{
						TLVList: sess.GetUserInfo(),
					},
				},
			},
		},
	})
}

func SendChatRoomInfoUpdate(sess *Session, sm SessionManager, room ChatRoom) {
	sm.SendToScreenName(sess.ScreenName, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT,
			SubGroup:  oscar.ChatRoomInfoUpdate,
		},
		snacOut: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
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
