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

func (t TOCIMIN) String() string {
	b, ok := t.TLVRestBlock.Bytes(ICBMTLVAOLIMData)
	if !ok {
		return ""
	}
	txt, err := UnmarshalICBMMessageText(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("IM_IN:%s:F:%s", t.ScreenName, txt)
}

type TOCChatJoin struct {
	SNAC_0x0E_0x02_ChatRoomInfoUpdate
}

func (t TOCChatJoin) String(chatID string) string {
	name, _ := t.Bytes(ChatRoomTLVRoomName)
	return fmt.Sprintf("CHAT_JOIN:%s:%s", chatID, name)
}

type TOCChatUsersJoined struct {
	SNAC_0x0E_0x03_ChatUsersJoined
}

func (t TOCChatUsersJoined) String(chatID string) string {
	users := make([]string, 0, len(t.Users))
	for _, u := range t.Users {
		users = append(users, u.ScreenName)
	}
	return fmt.Sprintf("CHAT_UPDATE_BUDDY:%s:T:%s", chatID, strings.Join(users, ":"))
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

	return fmt.Sprintf("CHAT_IN:%s:%s:F:%s", chatID, u.ScreenName, text)
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
