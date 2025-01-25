package toc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"golang.org/x/net/html"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

var (
	cmdInternalSvcErr = "ERROR:989:internal server error"
	errDisconnect     = errors.New("got booted by another session")
)

func (s OSCARProxy) RecvBOS(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, ch chan<- []byte) error {
	defer func() {
		fmt.Println("closing RecvBOS")
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-me.Closed():
			fmt.Println("I got signed off")
			return errDisconnect
		case snac := <-me.ReceiveMessage():
			switch v := snac.Body.(type) {
			case wire.SNAC_0x03_0x0B_BuddyArrived:
				sendOrCancel(ctx, ch, s.UpdateBuddyArrival(v))
			case wire.SNAC_0x03_0x0C_BuddyDeparted:
				sendOrCancel(ctx, ch, s.UpdateBuddyDeparted(v))
			case wire.SNAC_0x04_0x07_ICBMChannelMsgToClient:
				sendOrCancel(ctx, ch, s.IMIn(ctx, chatRegistry, v))
			case wire.SNAC_0x01_0x10_OServiceEvilNotification:
				sendOrCancel(ctx, ch, s.Eviled(ctx, v))
			default:
				s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s",
					wire.FoodGroupName(snac.Frame.FoodGroup),
					wire.SubGroupName(snac.Frame.FoodGroup, snac.Frame.SubGroup)))
			}
		}
	}

	return nil
}

func (s OSCARProxy) RecvChat(ctx context.Context, me *state.Session, chatID int, ch chan<- []byte) {
	defer func() {
		fmt.Println("closing chat RecvChat")
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-me.Closed():
			return
		case snac := <-me.ReceiveMessage():
			switch v := snac.Body.(type) {
			case wire.SNAC_0x0E_0x04_ChatUsersLeft:
				sendOrCancel(ctx, ch, s.ChatUpdateBuddyLeft(ctx, v, chatID))
			case wire.SNAC_0x0E_0x03_ChatUsersJoined:
				sendOrCancel(ctx, ch, s.ChatUpdateBuddyArrived(ctx, v, chatID))
			case wire.SNAC_0x0E_0x06_ChatChannelMsgToClient:
				sendOrCancel(ctx, ch, s.ChatIn(ctx, v, chatID))
			default:
				s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s",
					wire.FoodGroupName(snac.Frame.FoodGroup),
					wire.SubGroupName(snac.Frame.FoodGroup, snac.Frame.SubGroup)))
			}
		}
	}
}

// UpdateBuddyArrival handles the UPDATE_BUDDY TOC command for buddy arrival events.
//
// From the TiK documentation:
//
//	This one command handles arrival/depart/updates. Evil Amount is a percentage, Signon Time is UNIX epoc, idle time is in minutes, UC (User Class) is a two/three character string.
//		- uc[0]
//			- ' ' - Ignore
//			- 'A' - On AOL
//		- uc[1]
//			- ' ' - Ignore
//			- 'A' - Oscar Admin
//			- 'U' - Oscar Unconfirmed
//			- 'O' - Oscar Normal
//		- uc[2]
//			- '\0' - Ignore
//			- ' ' - Ignore
//			- 'U' - The user has set their unavailable flag.
//
// Command syntax: UPDATE_BUDDY:<Buddy User>:<Online? T/F>:<Evil Amount>:<Signon Time>:<IdleTime>:<UC>
func (s OSCARProxy) UpdateBuddyArrival(snac wire.SNAC_0x03_0x0B_BuddyArrived) string {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "T", snac.WarningLevel, online, idle, uc)
}

// UpdateBuddyDeparted handles the UPDATE_BUDDY TOC command for buddy departure events.
//
// From the TiK documentation:
//
//	This one command handles arrival/depart/updates. Evil Amount is a
//	percentage, Signon Time is UNIX epoc, idle time is in minutes, UC (User
//	Class) is a two/three character string.
//		- uc[0]
//			- ' ' - Ignore
//			- 'A' - On AOL
//		- uc[1]
//			- ' ' - Ignore
//			- 'A' - Oscar Admin
//			- 'U' - Oscar Unconfirmed
//			- 'O' - Oscar Normal
//		- uc[2]
//			- '\0' - Ignore
//			- ' ' - Ignore
//			- 'U' - The user has set their unavailable flag.
//
// Command syntax: UPDATE_BUDDY:<Buddy User>:<Online? T/F>:<Evil Amount>:<Signon Time>:<IdleTime>:<UC>
func (s OSCARProxy) UpdateBuddyDeparted(snac wire.SNAC_0x03_0x0C_BuddyDeparted) string {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	class := strings.Join(uc[:], "")
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "F", snac.WarningLevel, online, idle, class)
}

// IMIn handles the IM_IN TOC command.
//
// From the TiK documentation:
//
//	Receive an IM from someone. Everything after the third colon is the
//	incoming message, including other colons.
//
// Command syntax: IM_IN:<Source User>:<Auto Response T/F?>:<Message>
func (s OSCARProxy) IMIn(ctx context.Context, chatRegistry *ChatRegistry, snac wire.SNAC_0x04_0x07_ICBMChannelMsgToClient) string {
	if snac.ChannelID == wire.ICBMChannelRendezvous {
		rdinfo, has := snac.TLVRestBlock.Bytes(0x05)
		if !has {
			logErr(ctx, s.Logger, errors.New("TLVRestBlock.Bytes: missing rendezvous block"))
			return cmdInternalSvcErr
		}
		frag := wire.ICBMCh2Fragment{}
		if err := wire.UnmarshalBE(&frag, bytes.NewReader(rdinfo)); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
			return cmdInternalSvcErr
		}
		prompt, ok := frag.Bytes(12)
		if !ok {
			logErr(ctx, s.Logger, errors.New("frag.Bytes: missing prompt"))
			return cmdInternalSvcErr
		}

		svcData, ok := frag.Bytes(10001)
		if !ok {
			logErr(ctx, s.Logger, errors.New("frag.Bytes: missing room info"))
			return cmdInternalSvcErr
		}

		roomInfo := wire.ICBMRoomInfo{}
		if err := wire.UnmarshalBE(&roomInfo, bytes.NewReader(svcData)); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
			return cmdInternalSvcErr
		}

		name := strings.Split(roomInfo.Cookie, "-")[2]

		chatID := chatRegistry.Add(roomInfo)
		return fmt.Sprintf("CHAT_INVITE:%s:%d:%s:%s", name, chatID, snac.ScreenName, prompt)
	}

	buf, ok := snac.TLVRestBlock.Bytes(wire.ICBMTLVAOLIMData)
	if !ok {
		logErr(ctx, s.Logger, errors.New("TLVRestBlock.Bytes: missing wire.ICBMTLVAOLIMData"))
		return cmdInternalSvcErr
	}
	txt, err := wire.UnmarshalICBMMessageText(buf)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalICBMMessageText: %w", err))
		return cmdInternalSvcErr
	}

	return fmt.Sprintf("IM_IN:%s:F:%s", snac.ScreenName, txt)
}

// Eviled handles the EVILED TOC command.
//
// From the TiK documentation:
//
//	The user was just eviled.
//
// Command syntax: EVILED:<new evil>:<name of eviler, blank if anonymous>
func (s OSCARProxy) Eviled(ctx context.Context, snac wire.SNAC_0x01_0x10_OServiceEvilNotification) string {
	who := ""
	if snac.Snitcher != nil {
		who = snac.Snitcher.ScreenName
	}
	return fmt.Sprintf("EVILED:%d:%s", snac.NewEvil, who)
}

// ChatUpdateBuddyArrived handles the CHAT_UPDATE_BUDDY TOC command for chat
// room arrival events.
//
// From the TiK documentation:
//
//	This one command handles arrival/departs from a chat room. The very first
//	message of this type for each chat room contains the users already in the
//	room.
//
// Command syntax: CHAT_UPDATE_BUDDY:<Chat Room Id>:<Inside? T/F>:<User 1>:<User 2>...
func (s OSCARProxy) ChatUpdateBuddyArrived(ctx context.Context, snac wire.SNAC_0x0E_0x03_ChatUsersJoined, chatID int) string {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	return fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:T:%s", chatID, strings.Join(users, ":"))
}

// ChatUpdateBuddyLeft handles the CHAT_UPDATE_BUDDY TOC command for chat
// room departure events.
//
// From the TiK documentation:
//
//	This one command handles arrival/departs from a chat room. The very first
//	message of this type for each chat room contains the users already in the
//	room.
//
// Command syntax: CHAT_UPDATE_BUDDY:<Chat Room Id>:<Inside? T/F>:<User 1>:<User 2>...
func (s OSCARProxy) ChatUpdateBuddyLeft(ctx context.Context, snac wire.SNAC_0x0E_0x04_ChatUsersLeft, chatID int) string {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	return fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:F:%s", chatID, strings.Join(users, ":"))
}

// ChatIn handles the CHAT_IN TOC command.
//
// From the TiK documentation:
//
//	A chat message was sent in a chat room.
//
// Command syntax: CHAT_IN:<Chat Room Id>:<Source User>:<Whisper? T/F>:<Message>
func (s OSCARProxy) ChatIn(ctx context.Context, snac wire.SNAC_0x0E_0x06_ChatChannelMsgToClient, chatID int) string {
	b, ok := snac.Bytes(wire.ChatTLVSenderInformation)
	if !ok {
		logErr(ctx, s.Logger, errors.New("snac.Bytes: missing wire.ChatTLVSenderInformation"))
		return cmdInternalSvcErr
	}

	u := wire.TLVUserInfo{}
	err := wire.UnmarshalBE(&u, bytes.NewReader(b))
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
		return cmdInternalSvcErr
	}

	b, ok = snac.Bytes(wire.ChatTLVMessageInfo)
	if !ok {
		logErr(ctx, s.Logger, errors.New("snac.Bytes: missing wire.ChatTLVMessageInfo"))
		return cmdInternalSvcErr
	}

	text, err := textFromChatMsgBlob(b)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("textFromChatMsgBlob: %w", err))
		return cmdInternalSvcErr
	}

	return fmt.Sprintf("CHAT_IN:%d:%s:F:%s", chatID, u.ScreenName, text)
}

func (s OSCARProxy) Signout(ctx context.Context, me *state.Session) {
	if err := s.BuddyService.BroadcastBuddyDeparted(ctx, me); err != nil {
		s.Logger.ErrorContext(ctx, "error sending departure notifications", "err", err.Error())
	}
	if err := s.BuddyListRegistry.UnregisterBuddyList(me.IdentScreenName()); err != nil {
		s.Logger.ErrorContext(ctx, "error removing buddy list entry", "err", err.Error())
	}
	s.AuthService.Signout(ctx, me)
}

// textFromChatMsgBlob extracts plaintext message text from HTML located in
// chat message info TLV(0x05).
func textFromChatMsgBlob(msg []byte) ([]byte, error) {
	block := wire.TLVRestBlock{}
	if err := wire.UnmarshalBE(&block, bytes.NewReader(msg)); err != nil {
		return nil, err
	}

	b, hasMsg := block.Bytes(wire.ChatTLVMessageInfoText)
	if !hasMsg {
		return nil, errors.New("SNAC(0x0E,0x05) has no chat msg text TLV")
	}

	tok := html.NewTokenizer(bytes.NewReader(b))
	for {
		switch tok.Next() {
		case html.TextToken:
			return tok.Text(), nil
		case html.ErrorToken:
			err := tok.Err()
			if err == io.EOF {
				err = nil
			}
			return nil, err
		}
	}
}

func logErr(ctx context.Context, logger *slog.Logger, err error) {
	logger.ErrorContext(ctx, "internal service error", "err", err.Error())
}

func sendOrCancel(ctx context.Context, ch chan<- []byte, msg string) {
	select {
	case <-ctx.Done():
		return
	case ch <- []byte(msg):
		return
	}
}
