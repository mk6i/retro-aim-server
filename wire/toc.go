package wire

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type TOCBuddyArrived struct {
	SNAC_0x03_0x0B_BuddyArrived
}

func (t TOCBuddyArrived) String() string {
	online, _ := t.Uint32BE(OServiceUserInfoSignonTOD)
	idle, _ := t.Uint16BE(OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if t.IsAway() {
		uc[2] = "U"
	}
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", t.ScreenName, "T", t.WarningLevel, online, idle, uc)
}

type TOCBuddyDeparted struct {
	SNAC_0x03_0x0C_BuddyDeparted
}

func (t TOCBuddyDeparted) String() string {
	online, _ := t.Uint32BE(OServiceUserInfoSignonTOD)
	idle, _ := t.Uint16BE(OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if t.IsAway() {
		uc[2] = "U"
	}
	class := strings.Join(uc[:], "")
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", t.ScreenName, "T", t.WarningLevel, online, idle, class)
}

type TOCIMIN struct {
	SNAC_0x04_0x07_ICBMChannelMsgToClient
}

func (sn TOCIMIN) String() string {

	switch sn.ChannelID {
	case ICBMChannelRendezvous:
		rdinfo, has := sn.TLVRestBlock.Bytes(0x05)
		if !has {
			fmt.Printf("doesn't have rendezvous block\n")
			return ""
		}
		frag := ICBMCh2Fragment{}
		if err := UnmarshalBE(&frag, bytes.NewBuffer(rdinfo)); err != nil {
			fmt.Printf("unmarshal ICBM channel message rdv apyload failed: %w", err)
			return ""
		}
		prompt, _ := frag.Bytes(12)

		svcData, _ := frag.Bytes(10001)

		roomInfo := ICBMRoomInfo{}
		if err := UnmarshalBE(&roomInfo, bytes.NewBuffer(svcData)); err != nil {
			fmt.Printf("unmarshal ICBM channel message rdv apyload failed: %w", err)
			return ""
		}

		name := strings.Split(roomInfo.Cookie, "-")[2]
		return fmt.Sprintf("CHAT_INVITE:%s:%s:%s:%s", name, "10", sn.ScreenName, prompt)
	default:
		b, ok := sn.TLVRestBlock.Bytes(ICBMTLVAOLIMData)
		if !ok {
			return ""
		}
		txt, err := UnmarshalICBMMessageText(b)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("IM_IN:%s:F:%s", sn.ScreenName, txt)
	}
	return ""
}

type TOCChatJoin struct {
	SNAC_0x0E_0x02_ChatRoomInfoUpdate
}

func (t TOCChatJoin) String(chatID string) string {
	name, _ := t.Bytes(ChatRoomTLVRoomName)
	return fmt.Sprintf("CHAT_JOIN:%s:%s", "10", name)
}

type TOCChatUsersJoined struct {
	SNAC_0x0E_0x03_ChatUsersJoined
}

func (t TOCChatUsersJoined) String(chatID string) string {
	users := make([]string, 0, len(t.Users))
	for _, u := range t.Users {
		users = append(users, u.ScreenName)
	}
	return fmt.Sprintf("CHAT_UPDATE_BUDDY:%s:T:%s", "10", "mike")
}

type TOCChatIn struct {
	SNAC_0x0E_0x06_ChatChannelMsgToClient
}

func (t TOCChatIn) String(chatID string) string {
	b, _ := t.Bytes(ChatTLVSenderInformation)

	u := TLVUserInfo{}
	err := UnmarshalBE(&u, bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}

	b, _ = t.Bytes(ChatTLVMessageInfo)
	text, err := textFromChatMsgBlob(b)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("CHAT_IN:%s:%s:F:%s", "10", u.ScreenName, text)
}

// textFromChatMsgBlob extracts plaintext message text from HTML located in
// chat message info TLV(0x05).
func textFromChatMsgBlob(msg []byte) ([]byte, error) {
	block := TLVRestBlock{}
	if err := UnmarshalBE(&block, bytes.NewBuffer(msg)); err != nil {
		return nil, err
	}

	b, hasMsg := block.Bytes(ChatTLVMessageInfoText)
	if !hasMsg {
		return nil, errors.New("SNAC(0x0E,0x05) has no chat msg text TLV")
	}

	tok := html.NewTokenizer(bytes.NewBuffer(b))
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

type TOC struct {
}

func (t TOC) String() string {
	return "TOC"
}
