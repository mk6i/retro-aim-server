package toc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/html"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

var (
	cmdInternalSvcErr = []byte("ERROR:989:internal server error")
	capChat           = uuid.MustParse("748F2420-6287-11D1-8222-444553540000")
)

var errDisconnect = errors.New("got booted by another session")

func (s OSCARProxy) ConsumeIncomingBOS(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, ch chan<- []byte) error {
	defer func() {
		fmt.Println("closing ConsumeIncomingBOS")
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-me.Closed():
			fmt.Println("I got signed off")
			return errDisconnect
		case snac := <-me.ReceiveMessage():
			inFrame := snac.Frame
			switch inFrame.FoodGroup {
			case wire.Buddy:
				switch inFrame.SubGroup {
				case wire.BuddyArrived:
					// todo make these type assertions safe?
					sendOrCancel(ctx, ch, s.UpdateBuddyArrival(ctx, snac.Body.(wire.SNAC_0x03_0x0B_BuddyArrived)))
				case wire.BuddyDeparted:
					sendOrCancel(ctx, ch, s.UpdateBuddyDeparted(ctx, snac.Body.(wire.SNAC_0x03_0x0C_BuddyDeparted)))
				default:
					// don't return error because they could be booted by malicious actor?
					s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.ICBM:
				switch inFrame.SubGroup {
				case wire.ICBMChannelMsgToClient:
					sendOrCancel(ctx, ch, s.IMIn(ctx, chatRegistry, snac.Body.(wire.SNAC_0x04_0x07_ICBMChannelMsgToClient)))
				default:
					s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.OService:
				switch inFrame.SubGroup {
				case wire.OServiceEvilNotification:
					sendOrCancel(ctx, ch, s.Eviled(ctx, snac.Body.(wire.SNAC_0x01_0x10_OServiceEvilNotification)))
				default:
					s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			default:
				s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
			}
		}
	}

	return nil
}

func (s OSCARProxy) ConsumeIncomingChat(ctx context.Context, me *state.Session, chatID int, ch chan<- []byte) {
	defer func() {
		fmt.Println("closing chat ConsumeIncomingChat")
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-me.Closed():
			return
		case snac := <-me.ReceiveMessage():
			inFrame := snac.Frame
			switch inFrame.FoodGroup {
			case wire.Chat:
				switch inFrame.SubGroup {
				case wire.ChatUsersLeft:
					sendOrCancel(ctx, ch, s.ChatUpdateBuddyLeft(ctx, snac.Body.(wire.SNAC_0x0E_0x04_ChatUsersLeft), chatID))
				case wire.ChatUsersJoined:
					sendOrCancel(ctx, ch, s.ChatUpdateBuddyArrived(ctx, snac.Body.(wire.SNAC_0x0E_0x03_ChatUsersJoined), chatID))
				case wire.ChatChannelMsgToClient:
					sendOrCancel(ctx, ch, s.ChatIn(ctx, snac.Body.(wire.SNAC_0x0E_0x06_ChatChannelMsgToClient), chatID))
				default:
					s.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
				}
			default:
				s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
			}
		}
	}
}

func (s OSCARProxy) UpdateBuddyArrival(ctx context.Context, snac wire.SNAC_0x03_0x0B_BuddyArrived) []byte {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	return []byte(fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "T", snac.WarningLevel, online, idle, uc))
}

func (s OSCARProxy) UpdateBuddyDeparted(ctx context.Context, snac wire.SNAC_0x03_0x0C_BuddyDeparted) []byte {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	class := strings.Join(uc[:], "")
	return []byte(fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "F", snac.WarningLevel, online, idle, class))
}

func (s OSCARProxy) IMIn(ctx context.Context, chatRegistry *ChatRegistry, snac wire.SNAC_0x04_0x07_ICBMChannelMsgToClient) []byte {
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
		return []byte(fmt.Sprintf("CHAT_INVITE:%s:%d:%s:%s", name, chatID, snac.ScreenName, prompt))
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

	return []byte(fmt.Sprintf("IM_IN:%s:F:%s", snac.ScreenName, txt))
}

func (s OSCARProxy) Eviled(ctx context.Context, snac wire.SNAC_0x01_0x10_OServiceEvilNotification) []byte {
	who := ""
	if snac.Snitcher != nil {
		who = snac.Snitcher.ScreenName
	}
	return []byte(fmt.Sprintf("EVILED:%d:%s", snac.NewEvil, who))
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

func (s OSCARProxy) ChatUpdateBuddyArrived(ctx context.Context, snac wire.SNAC_0x0E_0x03_ChatUsersJoined, chatID int) []byte {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	return []byte(fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:T:%s", chatID, strings.Join(users, ":")))
}

func (s OSCARProxy) ChatUpdateBuddyLeft(ctx context.Context, snac wire.SNAC_0x0E_0x04_ChatUsersLeft, chatID int) []byte {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	return []byte(fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:F:%s", chatID, strings.Join(users, ":")))
}

func (s OSCARProxy) ChatIn(ctx context.Context, snac wire.SNAC_0x0E_0x06_ChatChannelMsgToClient, chatID int) []byte {
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

	return []byte(fmt.Sprintf("CHAT_IN:%d:%s:F:%s", chatID, u.ScreenName, text))
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

func sendOrCancel(ctx context.Context, ch chan<- []byte, msg []byte) {
	select {
	case <-ctx.Done():
		return
	case ch <- msg:
		return
	}
}
