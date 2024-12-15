package wire

import (
	"fmt"
	"strings"
)

type TOCBuddyArrived struct {
	SNAC_0x03_0x0B_BuddyArrived
}

func (t TOCBuddyArrived) String() string {
	online, _ := t.Uint32BE(OServiceUserInfoSignonTOD)
	idle, _ := t.Uint16BE(OServiceUserInfoIdleTime)
	unavailable := ""
	if t.IsAway() {
		unavailable = "U"
	}
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s%s", t.ScreenName, "T", t.WarningLevel, online, idle, " O", unavailable)
}

type TOCBuddyDeparted struct {
	SNAC_0x03_0x0C_BuddyDeparted
}

func (t TOCBuddyDeparted) String() string {
	online, _ := t.Uint32BE(OServiceUserInfoSignonTOD)
	idle, _ := t.Uint16BE(OServiceUserInfoIdleTime)
	unavailable := ""
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s%s", t.ScreenName, "F", t.WarningLevel, online, idle, " O", unavailable)
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

func (t TOCChatJoin) String() string {
	name, _ := t.Bytes(ChatRoomTLVRoomName)
	return fmt.Sprintf("CHAT_JOIN:%s:%s", t.Cookie, name)
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

type TOC struct {
}

func (t TOC) String() string {
	return "TOC"
}
