package toc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

var (
	capChat = uuid.MustParse("748F2420-6287-11D1-8222-444553540000")
)

type OSCARProxy struct {
	AuthService         AuthService
	BuddyListRegistry   BuddyListRegistry
	BuddyService        BuddyService
	ChatNavService      ChatNavService
	ChatService         ChatService
	CookieBaker         CookieBaker
	DirSearchService    DirSearchService
	ICBMService         ICBMService
	LocateService       LocateService
	Logger              *slog.Logger
	OServiceServiceBOS  OServiceService
	OServiceServiceChat OServiceService
	PermitDenyService   PermitDenyService
	TOCConfigStore      TOCConfigStore
}

func newChatRegistry() *ChatRegistry {
	chatRegistry := &ChatRegistry{
		lookup:   make(map[int]wire.ICBMRoomInfo),
		sessions: make(map[int]*state.Session),
		m:        sync.RWMutex{},
	}
	return chatRegistry
}

type ChatRegistry struct {
	lookup   map[int]wire.ICBMRoomInfo
	sessions map[int]*state.Session
	nextID   int
	m        sync.RWMutex
}

func (c *ChatRegistry) Add(room wire.ICBMRoomInfo) int {
	c.m.Lock()
	defer c.m.Unlock()
	for chatID, r := range c.lookup {
		if r == room {
			return chatID
		}
	}
	id := c.nextID
	c.lookup[id] = room
	c.nextID++
	return id
}

func (c *ChatRegistry) LookupRoom(chatID int) (wire.ICBMRoomInfo, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	room, found := c.lookup[chatID]
	return room, found
}

func (c *ChatRegistry) RegisterSess(chatID int, sess *state.Session) {
	c.m.Lock()
	defer c.m.Unlock()
	c.sessions[chatID] = sess
}

func (c *ChatRegistry) RetrieveSess(chatID int) *state.Session {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.sessions[chatID]
}

// AddBuddy handles the toc_add_buddy TOC command.
//
// From the TiK documentation:
//
//	Add buddies to your buddy list. This does not change your saved config.
//
// Command syntax: toc_add_buddy <Buddy User 1> [<Buddy User2> [<Buddy User 3> [...]]]
func (s OSCARProxy) AddBuddy(ctx context.Context, me *state.Session, cmd []byte) string {
	users, err := parseArgs(cmd, "toc_add_buddy")
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x03_0x04_BuddyAddBuddies{}
	for _, sn := range users {
		snac.Buddies = append(snac.Buddies, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := s.BuddyService.AddBuddies(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("BuddyService.AddBuddies: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// AddPermit handles the toc_add_permit TOC command.
//
// From the TiK documentation:
//
//	ADD the following people to your permit mode. If you are in deny mode it
//	will switch you to permit mode first. With no arguments and in deny mode
//	this will switch you to permit none. If already in permit mode, no
//	arguments does nothing and your permit list remains the same.
//
// Command syntax: toc_add_permit [ <User 1> [<User 2> [...]]]
func (s OSCARProxy) AddPermit(ctx context.Context, me *state.Session, cmd []byte) string {
	users, err := parseArgs(cmd, "toc_add_permit")
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{}
	for _, sn := range users {
		snac.Users = append(snac.Users, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := s.PermitDenyService.AddPermListEntries(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("PermitDenyService.AddPermListEntries: %w", err))
		return cmdInternalSvcErr
	}
	return ""
}

// AddDeny handles the toc_chat_join TOC command.
//
// From the TiK documentation:
//
//	ADD the following people to your deny mode. If you are in permit mode it
//	will switch you to deny mode first. With no arguments and in permit mode,
//	this will switch you to deny none. If already in deny mode, no arguments
//	does nothing and your deny list remains unchanged.
//
// Command syntax: toc_add_deny [ <User 1> [<User 2> [...]]]
func (s OSCARProxy) AddDeny(ctx context.Context, me *state.Session, cmd []byte) string {
	users, err := parseArgs(cmd, "toc_add_deny")
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{}
	for _, sn := range users {
		snac.Users = append(snac.Users, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := s.PermitDenyService.AddDenyListEntries(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("PermitDenyService.AddDenyListEntries: %w", err))
		return cmdInternalSvcErr
	}
	return ""
}

// ChatAccept handles the toc_chat_accept TOC command.
//
// From the TiK documentation:
//
//	Accept a CHAT_INVITE message from TOC. The server will send a CHAT_JOIN in
//	response.
//
// Command syntax: toc_chat_accept <Chat Room ID>
func (s OSCARProxy) ChatAccept(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, cmd []byte) (int, string) {
	var chatIDStr string

	if _, err := parseArgs(cmd, "toc_chat_accept", &chatIDStr); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return 0, cmdInternalSvcErr
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		return 0, cmdInternalSvcErr
	}
	chatInfo, found := chatRegistry.LookupRoom(chatID)
	if !found {
		logErr(ctx, s.Logger, fmt.Errorf("chatRegistry.LookupRoom: no chat found for ID %d", chatID))
		return 0, cmdInternalSvcErr
	}

	reqRoomSNAC := wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
		Cookie:         chatInfo.Cookie,
		Exchange:       chatInfo.Exchange,
		InstanceNumber: chatInfo.Instance,
	}
	reqRoomReply, err := s.ChatNavService.RequestRoomInfo(ctx, wire.SNACFrame{}, reqRoomSNAC)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ChatNavService.RequestRoomInfo: %w", err))
		return 0, cmdInternalSvcErr
	}

	reqRoomReplyBody, ok := reqRoomReply.Body.(wire.SNAC_0x0D_0x09_ChatNavNavInfo)
	if !ok {
		logErr(ctx, s.Logger, fmt.Errorf("chatNavService.RequestRoomInfo: unexpected response type %v", reqRoomReplyBody))
		return 0, cmdInternalSvcErr
	}
	b, hasInfo := reqRoomReplyBody.Bytes(wire.ChatNavTLVRoomInfo)
	if !hasInfo {
		logErr(ctx, s.Logger, errors.New("reqRoomReplyBody.Bytes: missing wire.ChatNavTLVRoomInfo"))
		return 0, cmdInternalSvcErr
	}

	roomInfo := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := wire.UnmarshalBE(&roomInfo, bytes.NewReader(b)); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
		return 0, cmdInternalSvcErr
	}

	roomName, hasName := roomInfo.Bytes(wire.ChatRoomTLVRoomName)
	if !hasName {
		logErr(ctx, s.Logger, errors.New("roomInfo.Bytes: missing wire.ChatRoomTLVRoomName"))
		return 0, cmdInternalSvcErr
	}

	svcReqSNAC := wire.SNAC_0x01_0x04_OServiceServiceRequest{
		FoodGroup: wire.Chat,
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
					Cookie: chatInfo.Cookie,
				}),
			},
		},
	}
	svcReqReply, err := s.OServiceServiceBOS.ServiceRequest(ctx, me, wire.SNACFrame{}, svcReqSNAC)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ServiceRequest: %w", err))
		return 0, cmdInternalSvcErr
	}

	svcReqReplyBody, ok := svcReqReply.Body.(wire.SNAC_0x01_0x05_OServiceServiceResponse)
	if !ok {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ServiceRequest: unexpected response type %v", svcReqReplyBody))
		return 0, cmdInternalSvcErr
	}

	loginCookie, hasCookie := svcReqReplyBody.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !hasCookie {
		logErr(ctx, s.Logger, errors.New("missing wire.OServiceTLVTagsLoginCookie"))
		return 0, cmdInternalSvcErr
	}

	chatSess, err := s.AuthService.RegisterChatSession(ctx, loginCookie)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("AuthService.RegisterChatSession: %w", err))
		return 0, cmdInternalSvcErr
	}

	chatRegistry.RegisterSess(chatID, chatSess)

	if err := s.OServiceServiceChat.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, chatSess); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceChat.ClientOnline: %w", err))
		return 0, cmdInternalSvcErr
	}

	return chatID, fmt.Sprintf("CHAT_JOIN:%d:%s", chatID, roomName)
}

// ChatInvite handles the toc_chat_invite TOC command.
//
// From the TiK documentation:
//
//	Once you are inside a chat room you can invite other people into that room.
//	Remember to quote and encode the invite message.
//
// Command syntax: toc_chat_invite <Chat Room ID> <Invite Msg> <buddy1> [<buddy2> [<buddy3> [...]]]
func (s OSCARProxy) ChatInvite(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, cmd []byte) string {
	var chatRoomIDStr, msg string

	users, err := parseArgs(cmd, "toc_chat_invite", &chatRoomIDStr, &msg)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	chatID, err := strconv.Atoi(chatRoomIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		return cmdInternalSvcErr
	}

	roomInfo, found := chatRegistry.LookupRoom(chatID)
	if !found {
		logErr(ctx, s.Logger, fmt.Errorf("chatRegistry.LookupRoom: chat ID `%d` not found", chatID))
		return cmdInternalSvcErr
	}

	for _, guest := range users {
		snac := wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
			ChannelID:  wire.ICBMChannelRendezvous,
			ScreenName: guest,
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(0x05, wire.ICBMCh2Fragment{
						Type:       0,
						Capability: capChat,
						TLVRestBlock: wire.TLVRestBlock{
							TLVList: wire.TLVList{
								wire.NewTLVBE(10, uint16(1)),
								wire.NewTLVBE(12, msg),
								wire.NewTLVBE(13, "us-ascii"),
								wire.NewTLVBE(14, "en"),
								wire.NewTLVBE(10001, roomInfo),
							},
						},
					}),
				},
			},
		}

		if _, err := s.ICBMService.ChannelMsgToHost(ctx, me, wire.SNACFrame{}, snac); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("ICBMService.ChannelMsgToHost: %w", err))
			return cmdInternalSvcErr
		}
	}

	return ""
}

// ChatJoin handles the toc_chat_join TOC command.
//
// From the TiK documentation:
//
//	Join a chat room in the given exchange. Exchange is an integer that
//	represents a group of chat rooms. Different exchanges have different
//	properties. For example some exchanges might have room replication (ie a
//	room never fills up, there are just multiple instances.) and some exchanges
//	might have navigational information. Currently, exchange should always be
//	4, however this may change in the future. You will either receive an ERROR
//	if the room couldn't be joined or a CHAT_JOIN message. The Chat Room Name
//	is case-insensitive and consecutive spaces are removed.
//
// Command syntax: toc_chat_join <Exchange> <Chat Room Name>
func (s OSCARProxy) ChatJoin(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, cmd []byte) (int, string) {
	var exchangeStr, roomName string

	if _, err := parseArgs(cmd, "toc_chat_join", &exchangeStr, &roomName); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return 0, cmdInternalSvcErr
	}

	// create room or retrieve the room if it already exists
	exchange, err := strconv.Atoi(exchangeStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		return 0, cmdInternalSvcErr
	}

	mkRoomReq := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange: uint16(exchange),
		Cookie:   "create",
		TLVBlock: wire.TLVBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ChatRoomTLVRoomName, roomName),
			},
		},
	}
	mkRoomReply, err := s.ChatNavService.CreateRoom(ctx, me, wire.SNACFrame{}, mkRoomReq)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ChatNavService.CreateRoom: %w", err))
		return 0, cmdInternalSvcErr
	}

	mkRoomReplyBody, ok := mkRoomReply.Body.(wire.SNAC_0x0D_0x09_ChatNavNavInfo)
	if !ok {
		logErr(ctx, s.Logger, fmt.Errorf("chatNavService.CreateRoom: unexpected response type %v", mkRoomReplyBody))
		return 0, cmdInternalSvcErr
	}
	buf, ok := mkRoomReplyBody.Bytes(wire.ChatNavTLVRoomInfo)
	if !ok {
		logErr(ctx, s.Logger, errors.New("mkRoomReplyBody.Bytes: missing wire.ChatNavTLVRoomInfo"))
		return 0, cmdInternalSvcErr
	}

	inBody := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := wire.UnmarshalBE(&inBody, bytes.NewReader(buf)); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
		return 0, cmdInternalSvcErr
	}

	svcReqSNAC := wire.SNAC_0x01_0x04_OServiceServiceRequest{
		FoodGroup: wire.Chat,
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
					Cookie: inBody.Cookie,
				}),
			},
		},
	}
	svcReqReply, err := s.OServiceServiceBOS.ServiceRequest(ctx, me, wire.SNACFrame{}, svcReqSNAC)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ServiceRequest: %w", err))
		return 0, cmdInternalSvcErr
	}

	svcReqReplyBody, ok := svcReqReply.Body.(wire.SNAC_0x01_0x05_OServiceServiceResponse)
	if !ok {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ServiceRequest: unexpected response type %v", svcReqReplyBody))
		return 0, cmdInternalSvcErr
	}

	loginCookie, hasCookie := svcReqReplyBody.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !hasCookie {
		logErr(ctx, s.Logger, errors.New("svcReqReplyBody.Bytes: missing wire.OServiceTLVTagsLoginCookie"))
		return 0, cmdInternalSvcErr
	}

	chatSess, err := s.AuthService.RegisterChatSession(ctx, loginCookie)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("AuthService.RegisterChatSession: %w", err))
		return 0, cmdInternalSvcErr
	}

	roomInfo := wire.ICBMRoomInfo{
		Exchange: inBody.Exchange,
		Cookie:   inBody.Cookie,
		Instance: inBody.InstanceNumber,
	}
	chatID := chatRegistry.Add(roomInfo)
	chatRegistry.RegisterSess(chatID, chatSess)

	if err := s.OServiceServiceChat.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, chatSess); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceChat.ClientOnline: %w", err))
		return 0, cmdInternalSvcErr
	}

	return chatID, fmt.Sprintf("CHAT_JOIN:%d:%s", chatID, roomName)
}

// ChatLeave handles the toc_chat_leave TOC command.
//
// From the TiK documentation:
//
//	Leave the chat room.
//
// Command syntax: toc_chat_leave <Chat Room ID>
func (s OSCARProxy) ChatLeave(ctx context.Context, chatRegistry *ChatRegistry, cmd []byte) string {
	var chatIDStr string

	if _, err := parseArgs(cmd, "toc_chat_leave", &chatIDStr); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		return cmdInternalSvcErr
	}

	me := chatRegistry.RetrieveSess(chatID)
	if me == nil {
		logErr(ctx, s.Logger, fmt.Errorf("chatRegistry.RetrieveSess: chat session `%d` not found", chatID))
		return cmdInternalSvcErr
	}

	s.AuthService.SignoutChat(ctx, me)

	me.Close() // stop async server SNAC reply handler for this chat room

	return fmt.Sprintf("CHAT_LEFT:%d", chatID)
}

// ChatSend handles the toc_chat_send TOC command.
//
// From the TiK documentation:
//
//	Send a message in a chat room using the chat room id from CHAT_JOIN. Since
//	reflection is always on in TOC, you do not need to add the message to your
//	chat UI, since you will get a CHAT_IN with the message. Remember to quote
//	and encode the message.
//
// Command syntax: toc_chat_send <Chat Room ID> <Message>
func (s OSCARProxy) ChatSend(ctx context.Context, chatRegistry *ChatRegistry, cmd []byte) string {
	var chatIDStr, msg string

	if _, err := parseArgs(cmd, "toc_chat_send", &chatIDStr, &msg); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		return cmdInternalSvcErr
	}

	me := chatRegistry.RetrieveSess(chatID)
	if me == nil {
		logErr(ctx, s.Logger, fmt.Errorf("chatRegistry.RetrieveSess: session for chat ID `%d` not found", chatID))
		return cmdInternalSvcErr
	}

	block := wire.TLVRestBlock{}
	// the order of these TLVs matters for AIM 2.x. if out of order, screen
	// names do not appear with each chat message.
	block.Append(wire.NewTLVBE(wire.ChatTLVEnableReflectionFlag, uint8(1)))
	block.Append(wire.NewTLVBE(wire.ChatTLVSenderInformation, me.TLVUserInfo()))
	block.Append(wire.NewTLVBE(wire.ChatTLVPublicWhisperFlag, []byte{}))
	block.Append(wire.NewTLVBE(wire.ChatTLVMessageInfo, wire.TLVRestBlock{
		TLVList: wire.TLVList{
			wire.NewTLVBE(wire.ChatTLVMessageInfoText, msg),
		},
	}))

	snac := wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{
		Channel:      wire.ICBMChannelMIME,
		TLVRestBlock: block,
	}
	if _, err := s.ChatService.ChannelMsgToHost(ctx, me, wire.SNACFrame{}, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ChatService.ChannelMsgToHost: %w", err))
		return cmdInternalSvcErr
	}

	return fmt.Sprintf("CHAT_IN:%d:%s:F:%s", chatID, me.DisplayScreenName(), msg)
}

// Evil handles the toc_evil TOC command.
//
// From the TiK documentation:
//
//	Evil/Warn someone else. The 2nd argument is either the string "norm" for a
//	normal warning, or "anon" for an anonymous warning. You can only evil
//	people who have recently sent you ims. The higher someones evil level, the
//	slower they can send message.
//
// Command syntax: toc_evil <User> <norm|anon>
func (s OSCARProxy) Evil(ctx context.Context, me *state.Session, cmd []byte) string {
	var user, scope string

	if _, err := parseArgs(cmd, "toc_evil", &user, &scope); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x04_0x08_ICBMEvilRequest{
		ScreenName: user,
	}

	switch scope {
	case "anon":
		snac.SendAs = 1
	case "norm":
		snac.SendAs = 0
	default:
		s.Logger.Error("incorrect warning type. allowed values: anon, norm")
		return cmdInternalSvcErr
	}

	response, err := s.ICBMService.EvilRequest(ctx, me, wire.SNACFrame{}, snac)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ICBMService.EvilRequest: %w", err))
		return cmdInternalSvcErr
	}

	switch v := response.Body.(type) {
	case wire.SNAC_0x04_0x09_ICBMEvilReply:
		return ""
	case wire.SNACError:
		s.Logger.InfoContext(ctx, "unable to warn user", "code", v.Code)
	default:
		s.Logger.ErrorContext(ctx, "unexpected response")
		return cmdInternalSvcErr
	}

	return ""
}

// GetDirSearchURL handles the toc_dir_search TOC command.
//
// From the TiK documentation:
//
//	Perform a search of the Oscar Directory, using colon separated fields as in:
//
//		"first name":"middle name":"last name":"maiden name":"city":"state":"country":"email"
//
// You can search by keyword by setting search terms in the 11th position (this
// feature is not in the TiK docs but is present in the code):
//
//	::::::::::"search kw"
//
//	Returns either a GOTO_URL or ERROR msg.
//
// Command syntax: toc_dir_search <info information>
func (s OSCARProxy) GetDirSearchURL(ctx context.Context, me *state.Session, cmd []byte) string {
	var info string

	if _, err := parseArgs(cmd, "toc_dir_search", &info); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	params := strings.Split(info, ":")
	labels := []string{
		"first_name",
		"middle_name",
		"last_name",
		"maiden_name",
		"city",
		"state",
		"country",
		"email",
		"nop", // unused placeholder
		"nop",
		"keyword",
	}

	p := url.Values{}
	i := 0
	for i < len(params) && i < len(labels) {
		if len(params[i]) > 0 {
			p.Add(labels[i], strings.Trim(params[i], "\""))
		}
		i++
	}

	if len(p) == 0 {
		logErr(ctx, s.Logger, errors.New("no search fields found"))
		return cmdInternalSvcErr
	}

	if err := s.addCookie(me, p); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("addCookie: %w", err))
		return cmdInternalSvcErr
	}

	return fmt.Sprintf("GOTO_URL:search results:dir_search?%s", p.Encode())
}

// GetDirURL handles the toc_get_dir TOC command.
//
// From the TiK documentation:
//
//	Gets a user's dir info a GOTO_URL or ERROR message will be sent back to the client.
//
// Command syntax: toc_get_dir <username>
func (s OSCARProxy) GetDirURL(ctx context.Context, me *state.Session, cmd []byte) string {
	var user string

	if _, err := parseArgs(cmd, "toc_get_dir", &user); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	p := url.Values{}
	p.Add("user", user)

	if err := s.addCookie(me, p); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("addCookie: %w", err))
		return cmdInternalSvcErr
	}

	return fmt.Sprintf("GOTO_URL:directory info:dir_info?%s", p.Encode())
}

// GetInfoURL handles the toc_get_info TOC command.
//
// From the TiK documentation:
//
//	Gets a user's info a GOTO_URL or ERROR message will be sent back to the client.
//
// Command syntax: toc_get_info <username>
func (s OSCARProxy) GetInfoURL(ctx context.Context, me *state.Session, cmd []byte) string {
	var user string

	if _, err := parseArgs(cmd, "toc_get_info", &user); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	p := url.Values{}
	p.Add("from", me.IdentScreenName().String())
	p.Add("user", user)

	if err := s.addCookie(me, p); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("addCookie: %w", err))
		return cmdInternalSvcErr
	}

	return fmt.Sprintf("GOTO_URL:profile:info?%s", p.Encode())
}

// InitDone handles the toc_init_done TOC command.
//
// From the TiK documentation:
//
//	Tells TOC that we are ready to go online. TOC clients should first send TOC
//	the buddy list and any permit/deny lists. However, toc_init_done must be
//	called within 30 seconds after toc_signon, or the connection will be
//	dropped. Remember, it can't be called until after the SIGN_ON message is
//	received. Calling this before or multiple times after a SIGN_ON will cause
//	the connection to be dropped.
//
// Note: The business logic described in the last 3 sentences are not yet
// implemented.
//
// Command syntax: toc_init_done
func (s OSCARProxy) InitDone(ctx context.Context, sess *state.Session, cmd []byte) string {
	if _, err := parseArgs(cmd, "toc_init_done"); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}
	if err := s.OServiceServiceBOS.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ClientOnliney: %w", err))
		return cmdInternalSvcErr
	}
	return ""
}

// RemoveBuddy handles the toc_remove_buddy TOC command.
//
// From the TiK documentation:
//
//	Remove buddies from your buddy list. This does not change your saved config.
//
// Command syntax:
func (s OSCARProxy) RemoveBuddy(ctx context.Context, me *state.Session, cmd []byte) string {
	users, err := parseArgs(cmd, "toc_remove_buddy")
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x03_0x05_BuddyDelBuddies{}
	for _, sn := range users {
		snac.Buddies = append(snac.Buddies, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := s.BuddyService.DelBuddies(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("BuddyService.DelBuddies: %w", err))
		return cmdInternalSvcErr
	}
	return ""
}

// SendIM handles the toc_send_im TOC command.
//
// From the TiK documentation:
//
//	Send a message to a remote user. Remember to quote and encode the message.
//	If the optional string "auto" is the last argument, then the auto response
//	flag will be turned on for the IM.
//
// Command syntax: toc_send_im <Destination User> <Message> [auto]
func (s OSCARProxy) SendIM(ctx context.Context, sender *state.Session, cmd []byte) string {
	var recip, msg string

	autoReply, err := parseArgs(cmd, "toc_send_im", &recip, &msg)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	frags, err := wire.ICBMFragmentList(msg)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.ICBMFragmentList: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
		ChannelID:  wire.ICBMChannelIM,
		ScreenName: recip,
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ICBMTLVAOLIMData, frags),
			},
		},
	}

	if len(autoReply) > 0 && autoReply[0] == "auto" {
		snac.Append(wire.NewTLVBE(wire.ICBMTLVAutoResponse, []byte{}))
	}

	// send message and ignore response since there is no TOC error code to
	// handle errors such as "user is offline", etc.
	_, err = s.ICBMService.ChannelMsgToHost(ctx, sender, wire.SNACFrame{}, snac)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ICBMService.ChannelMsgToHost: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// SetAway handles the toc_chat_join TOC command.
//
// From the TiK documentation:
//
//	If the away message is present, then the unavailable status flag is set for
//	the user. If the away message is not present, then the unavailable status
//	flag is unset. The away message is basic HTML, remember to encode the
//	information.
//
// Command syntax: toc_set_away [<away message>]
func (s OSCARProxy) SetAway(ctx context.Context, me *state.Session, cmd []byte) string {
	maybeMsg, err := parseArgs(cmd, "toc_set_away")
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	var msg string
	if len(maybeMsg) > 0 {
		msg = maybeMsg[0]
	}

	snac := wire.SNAC_0x02_0x04_LocateSetInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.LocateTLVTagsInfoUnavailableData, msg),
			},
		},
	}

	if err := s.LocateService.SetInfo(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.SetInfo: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// SetCaps handles the toc_set_caps TOC command.
//
// From the TiK documentation:
//
//	Set my capabilities. All capabilities that we support need to be sent at
//	the same time. Capabilities are represented by UUIDs.
//
// This method automatically adds the "chat" capability since it doesn't seem
// to be sent explicitly by the official clients, even though they support
// chat.
//
// Command syntax: toc_set_caps [ <Capability 1> [<Capability 2> [...]]]
func (s OSCARProxy) SetCaps(ctx context.Context, me *state.Session, cmd []byte) string {
	params, err := parseArgs(cmd, "toc_set_caps")
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	caps := make([]uuid.UUID, 0, 16*(len(params)+1))
	for _, capStr := range params {
		uid, err := uuid.Parse(capStr)
		if err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("UUID.Parse: %w", err))
			return cmdInternalSvcErr
		}
		caps = append(caps, uid)
	}
	caps = append(caps, capChat)

	snac := wire.SNAC_0x02_0x04_LocateSetInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.LocateTLVTagsInfoCapabilities, caps),
			},
		},
	}

	if err := s.LocateService.SetInfo(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.SetInfo: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// SetConfig handles the toc_set_config TOC command.
//
// From the TiK documentation:
//
//	Set the config information for this user. The config information is line
//	oriented with the first character being the item type, followed by a space,
//	with the rest of the line being the item value. Only letters, numbers, and
//	spaces should be used. Remember you will have to enclose the entire config
//	in quotes.
//
//	Item Types:
//		- g - Buddy Group (All Buddies until the next g or the end of config are in this group.)
//		- b - A Buddy
//		- p - Person on permit list
//		- d - Person on deny list
//		- m - Permit/Deny Mode. Possible values are
//		- 1 - Permit All
//		- 2 - Deny All
//		- 3 - Permit Some
//		- 4 - Deny Some
//
// Command syntax: toc_set_config <Config Info>
func (s OSCARProxy) SetConfig(ctx context.Context, me *state.Session, cmd []byte) string {
	// replace curly braces with quotes so that the string can be properly
	// split up by the space-delimited reader
	for i, c := range cmd {
		if c == '{' || c == '}' {
			cmd[i] = '"'
		}
	}
	cmd = bytes.TrimSpace(cmd)

	var info string
	if _, err := parseArgs(cmd, "toc_set_config", &info); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	config := strings.Split(info, "\n")

	var cfg [][2]string
	for _, item := range config {
		parts := strings.Split(item, " ")
		if len(parts) != 2 {
			s.Logger.Info("invalid config item", "item", item, "user", me.DisplayScreenName())
			continue
		}
		cfg = append(cfg, [2]string{parts[0], parts[1]})
	}

	mode := wire.FeedbagPDModePermitAll
	for _, c := range cfg {
		if c[0] != "m" {
			continue
		}
		switch c[1] {
		case "1":
			mode = wire.FeedbagPDModePermitAll
		case "2":
			mode = wire.FeedbagPDModeDenyAll
		case "3":
			mode = wire.FeedbagPDModePermitSome
		case "4":
			mode = wire.FeedbagPDModeDenySome
		default:
			s.Logger.Info("config: invalid mode", "val", c[1], "user", me.DisplayScreenName())
			return cmdInternalSvcErr
		}
	}

	switch mode {
	case wire.FeedbagPDModePermitAll:
		snac := wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{
			Users: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: me.IdentScreenName().String(),
				},
			},
		}
		if err := s.PermitDenyService.AddDenyListEntries(ctx, me, snac); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("PermitDenyService.AddDenyListEntries: %w", err))
			return cmdInternalSvcErr
		}
	case wire.FeedbagPDModeDenyAll:
		snac := wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{
			Users: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: me.IdentScreenName().String(),
				},
			},
		}
		if err := s.PermitDenyService.AddPermListEntries(ctx, me, snac); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("PermitDenyService.AddPermListEntrie: %w", err))
			return cmdInternalSvcErr
		}
	case wire.FeedbagPDModePermitSome:
		snac := wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{}
		for _, c := range cfg {
			if c[0] != "p" {
				continue
			}
			snac.Users = append(snac.Users, struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{ScreenName: c[1]})
		}
		if err := s.PermitDenyService.AddPermListEntries(ctx, me, snac); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("PermitDenyService.AddPermListEntrie: %w", err))
			return cmdInternalSvcErr
		}
	case wire.FeedbagPDModeDenySome:
		snac := wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{}
		for _, c := range cfg {
			if c[0] != "d" {
				continue
			}
			snac.Users = append(snac.Users, struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{ScreenName: c[1]})
		}
		if err := s.PermitDenyService.AddDenyListEntries(ctx, me, snac); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("PermitDenyService.AddDenyListEntries: %w", err))
			return cmdInternalSvcErr
		}
	}

	snac := wire.SNAC_0x03_0x04_BuddyAddBuddies{}
	for _, c := range cfg {
		if c[0] != "b" {
			continue
		}
		snac.Buddies = append(snac.Buddies, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: c[1]})
	}

	if err := s.BuddyService.AddBuddies(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("BuddyService.AddBuddies: %w", err))
		return cmdInternalSvcErr
	}

	if err := s.TOCConfigStore.SetTOCConfig(me.IdentScreenName(), info); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("TOCConfigStore.SaveTOCConfig: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// SetDir handles the toc_set_dir TOC command.
//
// From the TiK documentation:
//
//	Set the DIR user information. This is a colon separated fields as in:
//
//		"first name":"middle name":"last name":"maiden name":"city":"state":"country":"email":"allow web searches".
//
//	Should return a DIR_STATUS msg. Having anything in the "allow web searches"
//	field allows people to use web-searches to find your directory info.
//	Otherwise, they'd have to use the client.
//
// The fields "email" and "allow web searches" are ignored by this method.
//
// Command syntax: toc_set_dir <info information>
func (s OSCARProxy) SetDir(ctx context.Context, me *state.Session, cmd []byte) string {
	var info string

	if _, err := parseArgs(cmd, "toc_set_dir", &info); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	rawFields := strings.Split(info, ":")

	var finalFields [9]string

	if len(rawFields) > len(finalFields) {
		logErr(ctx, s.Logger, fmt.Errorf("expected at most %d params, got %d", len(finalFields), len(rawFields)))
		return cmdInternalSvcErr
	}
	for i, a := range rawFields {
		finalFields[i] = strings.Trim(a, "\"")
	}

	snac := wire.SNAC_0x02_0x09_LocateSetDirInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ODirTLVFirstName, finalFields[0]),
				wire.NewTLVBE(wire.ODirTLVMiddleName, finalFields[1]),
				wire.NewTLVBE(wire.ODirTLVLastName, finalFields[2]),
				wire.NewTLVBE(wire.ODirTLVMaidenName, finalFields[3]),
				wire.NewTLVBE(wire.ODirTLVCountry, finalFields[6]),
				wire.NewTLVBE(wire.ODirTLVState, finalFields[5]),
				wire.NewTLVBE(wire.ODirTLVCity, finalFields[4]),
			},
		},
	}
	if _, err := s.LocateService.SetDirInfo(ctx, me, wire.SNACFrame{}, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.SetDirInfo: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// SetIdle handles the toc_set_idle TOC command.
//
// From the TiK documentation:
//
//	Set idle information. If <idle secs> is 0 then the user isn't idle at all.
//	If <idle secs> is greater than 0 then the user has already been idle for
//	<idle secs> number of seconds. The server will automatically keep
//	incrementing this number, so do not repeatedly call with new idle times.
//
// Command syntax: toc_set_idle <idle secs>
func (s OSCARProxy) SetIdle(ctx context.Context, me *state.Session, cmd []byte) string {
	var idleTimeStr string

	if _, err := parseArgs(cmd, "toc_set_idle", &idleTimeStr); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	time, err := strconv.Atoi(idleTimeStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x01_0x11_OServiceIdleNotification{
		IdleTime: uint32(time),
	}
	if err := s.OServiceServiceBOS.IdleNotification(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.IdleNotification: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// SetInfo handles the toc_set_info TOC command.
//
// From the TiK documentation:
//
//	Set the LOCATE user information. This is basic HTML. Remember to encode the info.
//
// Command syntax: toc_set_info <info information>
func (s OSCARProxy) SetInfo(ctx context.Context, me *state.Session, cmd []byte) string {
	var info string

	if _, err := parseArgs(cmd, "toc_set_info", &info); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}

	snac := wire.SNAC_0x02_0x04_LocateSetInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.LocateTLVTagsInfoSigData, info),
			},
		},
	}
	if err := s.LocateService.SetInfo(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.SetInfo: %w", err))
		return cmdInternalSvcErr
	}

	return ""
}

// Signon handles the toc_signon TOC command.
//
// From the TiK documentation:
//
//	The password needs to be roasted with the Roasting String if coming over a
//	FLAP connection, CP connections don't use roasted passwords. The language
//	specified will be used when generating web pages, such as the get info
//	pages. Currently, the only supported language is "english". If the language
//	sent isn't found, the default "english" language will be used. The version
//	string will be used for the client identity, and must be less than 50
//	characters.
//
//	Passwords are roasted when sent to the host. This is done so they aren't
//	sent in "clear text" over the wire, although they are still trivial to
//	decode. Roasting is performed by first xoring each byte in the password
//	with the equivalent modulo byte in the roasting string. The result is then
//	converted to ascii hex, and prepended with "0x". So for example the
//	password "password" roasts to "0x2408105c23001130".
//
//	The Roasting String is Tic/Toc.
//
// Command syntax: toc_signon <authorizer host> <authorizer port> <User Name> <Password> <language> <version>
func (s OSCARProxy) Signon(ctx context.Context, cmd []byte) (*state.Session, []string) {
	var userName, password string

	if _, err := parseArgs(cmd, "toc_signon", nil, nil, &userName, &password); err != nil {
		s.Logger.Error("parseArgs filed", "err", err.Error())
		return nil, []string{cmdInternalSvcErr}
	}

	passwordHash, err := hex.DecodeString(password[2:])
	if err != nil {
		s.Logger.Error("decode password hash failed", "err", err.Error())
		return nil, []string{cmdInternalSvcErr}
	}

	signonFrame := wire.FLAPSignonFrame{}
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsScreenName, userName))
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsRoastedTOCPassword, passwordHash))

	block, err := s.AuthService.FLAPLogin(signonFrame, state.NewStubUser)
	if err != nil {
		s.Logger.Error("FLAP login failed", "err", err.Error())
		return nil, []string{cmdInternalSvcErr}
	}

	if block.HasTag(wire.LoginTLVTagsErrorSubcode) {
		s.Logger.Debug("login failed")
		return nil, []string{"ERROR:980"} // bad username/password
	}

	authCookie, ok := block.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		s.Logger.Error("unable to get session id from payload")
		return nil, []string{cmdInternalSvcErr}
	}

	sess, err := s.AuthService.RegisterBOSSession(ctx, authCookie)
	if err != nil {
		s.Logger.Error("register BOS session failed", "err", err.Error())
		return nil, []string{cmdInternalSvcErr}
	}

	// set chat capability so that... tk
	sess.SetCaps([][16]byte{capChat})

	if err := s.BuddyListRegistry.RegisterBuddyList(sess.IdentScreenName()); err != nil {
		s.Logger.Error("unable to init buddy list", "err", err.Error())
		return nil, []string{cmdInternalSvcErr}
	}

	u, err := s.TOCConfigStore.User(sess.IdentScreenName())
	if err != nil {
		s.Logger.Error("TOCConfigStore.User retrieval error", "err", err.Error())
		return nil, []string{cmdInternalSvcErr}
	}
	if u == nil {
		s.Logger.Error("TOCConfigStore.User: user not found")
		return nil, []string{cmdInternalSvcErr}
	}

	return sess, []string{"SIGN_ON:TOC1.0", fmt.Sprintf("CONFIG:%s", u.TOCConfig)}
}

// addCookie appends an auth cookie to URL params.
func (s OSCARProxy) addCookie(me *state.Session, p url.Values) error {
	cookie, err := bakeCookie(s.CookieBaker, me.IdentScreenName())
	if err != nil {
		return err
	}
	p.Add("cookie", cookie)
	return nil
}

func bakeCookie(cookieBaker CookieBaker, me state.IdentScreenName) (string, error) {
	cookie, err := cookieBaker.Issue([]byte(me.String()))
	if err != nil {
		return "", err
	}
	return url.QueryEscape(base64.URLEncoding.EncodeToString(cookie)), nil
}

// parseArgs extracts arguments from a TOC command. Each positional argument is
// assigned to its corresponding args pointer. It returns the remaining
// arguments as varargs.
func parseArgs(payload []byte, cmd string, args ...*string) (varArgs []string, err error) {
	reader := csv.NewReader(bytes.NewReader(payload))
	reader.Comma = ' '
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	segs, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("CSV reader error: %w", err)
	}

	// sanity check the command name
	if segs[0] != cmd {
		return nil, fmt.Errorf("command mismatch. expected %s, got %s", cmd, segs[0])
	}

	// all elements after the command are arguments
	segs = segs[1:]
	if len(segs) < len(args) {
		return nil, fmt.Errorf("command contains fewer arguments than expected")
	}

	i := 0
	// populate placeholder pointers with their corresponding values
	for ; i < len(args); i++ {
		if args[i] == nil {
			// ignore argument, don't populate its corresponding pointer
			continue
		}
		*args[i] = strings.TrimSpace(segs[i])
	}

	// dump remaining arguments as varargs
	for _, param := range segs[i:] {
		varArgs = append(varArgs, param)
	}

	return varArgs, err
}
