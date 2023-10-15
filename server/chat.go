package server

import (
	"github.com/mkaminski/goaim/oscar"
	"io"
)

const (
	ChatErr                uint16 = 0x0001
	ChatRoomInfoUpdate            = 0x0002
	ChatUsersJoined               = 0x0003
	ChatUsersLeft                 = 0x0004
	ChatChannelMsgTohost          = 0x0005
	ChatChannelMsgToclient        = 0x0006
	ChatEvilRequest               = 0x0007
	ChatEvilReply                 = 0x0008
	ChatClientErr                 = 0x0009
	ChatPauseRoomReq              = 0x000A
	ChatPauseRoomAck              = 0x000B
	ChatResumeRoom                = 0x000C
	ChatShowMyRow                 = 0x000D
	ChatShowRowByUsername         = 0x000E
	ChatShowRowByNumber           = 0x000F
	ChatShowRowByName             = 0x0010
	ChatRowInfo                   = 0x0011
	ChatListRows                  = 0x0012
	ChatRowListInfo               = 0x0013
	ChatMoreRows                  = 0x0014
	ChatMoveToRow                 = 0x0015
	ChatToggleChat                = 0x0016
	ChatSendQuestion              = 0x0017
	ChatSendComment               = 0x0018
	ChatTallyVote                 = 0x0019
	ChatAcceptBid                 = 0x001A
	ChatSendInvite                = 0x001B
	ChatDeclineInvite             = 0x001C
	ChatAcceptInvite              = 0x001D
	ChatNotifyMessage             = 0x001E
	ChatGotoRow                   = 0x001F
	ChatStageUserJoin             = 0x0020
	ChatStageUserLeft             = 0x0021
	ChatUnnamedSnac22             = 0x0022
	ChatClose                     = 0x0023
	ChatUserBan                   = 0x0024
	ChatUserUnban                 = 0x0025
	ChatJoined                    = 0x0026
	ChatUnnamedSnac27             = 0x0027
	ChatUnnamedSnac28             = 0x0028
	ChatUnnamedSnac29             = 0x0029
	ChatRoomInfoOwner             = 0x0030
)

func routeChat(sess *Session, sm SessionManager, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case ChatChannelMsgTohost:
		return SendAndReceiveChatChannelMsgTohost(sess, sm, snac, r, w, sequence)
	default:
		return handleUnimplementedSNAC(snac, w, sequence)
	}
}

func SendAndReceiveChatChannelMsgTohost(sess *Session, sm SessionManager, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	snacPayloadIn := oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: CHAT,
		SubGroup:  ChatChannelMsgToclient,
	}
	snacPayloadOut := oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
		Cookie:  snacPayloadIn.Cookie,
		Channel: snacPayloadIn.Channel,
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: snacPayloadIn.TLVList,
		},
	}

	snacPayloadOut.AddTLV(oscar.TLV{
		TType: oscar.ChatTLVSenderInformation,
		Val: oscar.TLVUserInfo{
			ScreenName:   sess.ScreenName,
			WarningLevel: sess.GetWarning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: sess.GetUserInfo(),
			},
		},
	})

	// send message to all the participants except sender
	sm.BroadcastExcept(sess, XMessage{
		snacFrame: snacFrameOut,
		snacOut:   snacPayloadOut,
	})

	if _, ackMsg := snacPayloadIn.GetTLV(oscar.ChatTLVEnableReflectionFlag); !ackMsg {
		return nil
	}

	// reflect the message back to the sender
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func SetOnlineChatUsers(sm SessionManager, w io.Writer, sequence *uint32) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: CHAT,
		SubGroup:  ChatUsersJoined,
	}
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

	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func AlertUserJoined(sess *Session, sm SessionManager) {
	sm.BroadcastExcept(sess, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: CHAT,
			SubGroup:  ChatUsersJoined,
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
			FoodGroup: CHAT,
			SubGroup:  ChatUsersLeft,
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

func SendChatRoomInfoUpdate(room ChatRoom, w io.Writer, sequence *uint32) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: CHAT,
		SubGroup:  ChatRoomInfoUpdate,
	}
	snacPayloadOut := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange:       4,
		Cookie:         room.Cookie,
		InstanceNumber: 100,
		DetailLevel:    2,
		TLVBlock: oscar.TLVBlock{
			TLVList: room.TLVList(),
		},
	}
	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}
