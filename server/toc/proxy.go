package toc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"golang.org/x/net/html"
)

type BOSProxy struct {
	AuthService       AuthService
	BuddyService      BuddyService
	ICBMService       ICBMService
	LocateService     LocateService
	Logger            *slog.Logger
	OServiceService   OServiceService
	PermitDenyService PermitDenyService
}

func (b BOSProxy) ConsumeIncoming(ctx context.Context, me *state.Session, ch chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case snac := <-me.ReceiveMessage():
			inFrame := snac.Frame
			switch inFrame.FoodGroup {
			case wire.Buddy:
				switch inFrame.SubGroup {
				case wire.BuddyArrived:
					// todo make these type assertions safe?
					ch <- []byte(b.UpdateBuddyArrival(snac.Body.(wire.SNAC_0x03_0x0B_BuddyArrived)))
				case wire.BuddyDeparted:
					ch <- []byte(b.UpdateBuddyDeparted(snac.Body.(wire.SNAC_0x03_0x0C_BuddyDeparted)))
				default:
					b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.ICBM:
				switch inFrame.SubGroup {
				case wire.ICBMChannelMsgToClient:
					ch <- []byte(b.IMIn(snac.Body.(wire.SNAC_0x04_0x07_ICBMChannelMsgToClient)))
				default:
					b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.OService:
				switch inFrame.SubGroup {
				case wire.OServiceEvilNotification:
					ch <- []byte(b.Eviled(snac.Body.(wire.SNAC_0x01_0x10_OServiceEvilNotification)))
				default:
					b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			default:
				b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
			}
		}
	}
}

func (b BOSProxy) Login(elems []string) (*state.Session, error) {
	username := elems[3]
	passwordHash, err := hex.DecodeString(elems[4][2:])
	if err != nil {
		return nil, fmt.Errorf("decode password hash failed: %w", err)
	}

	passwordHash = wire.RoastTOCPassword(passwordHash)

	signonFrame := wire.FLAPSignonFrame{}
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsScreenName, username))
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, passwordHash))

	block, err := b.AuthService.FLAPLogin(signonFrame, state.NewStubUser)
	if err != nil {
		return nil, fmt.Errorf("FLAP login failed: %v", err)
	}

	authCookie, ok := block.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, errors.New("unable to get session id from payload")
	}

	sess, err := b.AuthService.RegisterBOSSession(authCookie)
	if err != nil {
		return nil, fmt.Errorf("register BOS session failed: %v", err)
	}
	if sess == nil {
		return nil, errors.New("BOS session not found")
	}

	return sess, nil
}

func (b BOSProxy) ClientReady(ctx context.Context, sess *state.Session) error {
	if err := b.OServiceService.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		return fmt.Errorf("client online failed: %v", err)
	}
	return nil
}

func (b BOSProxy) SendIM(ctx context.Context, me *state.Session, params []string) error {
	//message = strings.ReplaceAll("@MsgContent@", "@MsgContent@", message)

	frags, err := wire.ICBMFragmentList(params[2])
	if err != nil {
		return fmt.Errorf("unable to create ICBM fragment list: %w", err)
	}

	frame := wire.SNACFrame{
		FoodGroup: wire.ICBM,
		SubGroup:  wire.ICBMChannelMsgToHost,
	}
	snac := wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
		Cookie:     0,
		ChannelID:  1,
		ScreenName: params[1],
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ICBMTLVAOLIMData, frags),
			},
		},
	}

	if _, err := b.ICBMService.ChannelMsgToHost(ctx, me, frame, snac); err != nil {
		return fmt.Errorf("ChannelMsgToHost: %w", err)
	}

	return nil
}

func (b BOSProxy) AddBuddy(ctx context.Context, me *state.Session, params []string) error {
	buddies := params[1:]

	snac := wire.SNAC_0x03_0x04_BuddyAddBuddies{}
	for _, sn := range buddies {
		snac.Buddies = append(snac.Buddies, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := b.BuddyService.AddBuddies(ctx, me, snac); err != nil {
		return fmt.Errorf("BuddyService add buddies: %w", err)
	}

	return nil
}

func (b BOSProxy) RemoveBuddy(ctx context.Context, me *state.Session, params []string) error {
	buddies := params[1:]

	snac := wire.SNAC_0x03_0x05_BuddyDelBuddies{}
	for _, sn := range buddies {
		snac.Buddies = append(snac.Buddies, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := b.BuddyService.DelBuddies(ctx, me, snac); err != nil {
		return fmt.Errorf("BuddyService add buddies: %w", err)
	}

	return nil
}

func (b BOSProxy) AddPermit(ctx context.Context, me *state.Session, params []string) error {
	buddies := params[1:]

	snac := wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{}
	for _, sn := range buddies {
		snac.Users = append(snac.Users, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := b.PermitDenyService.AddPermListEntries(ctx, me, snac); err != nil {
		return fmt.Errorf("BuddyService add buddies: %w", err)
	}

	return nil
}

func (b BOSProxy) AddDeny(ctx context.Context, me *state.Session, params []string) error {
	buddies := params[1:]

	snac := wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{}
	for _, sn := range buddies {
		snac.Users = append(snac.Users, struct {
			ScreenName string `oscar:"len_prefix=uint8"`
		}{ScreenName: sn})
	}

	if err := b.PermitDenyService.AddDenyListEntries(ctx, me, snac); err != nil {
		return fmt.Errorf("BuddyService add buddies: %w", err)
	}

	return nil
}

func (b BOSProxy) SetCaps(ctx context.Context, me *state.Session, params []string) error {
	params = params[1:]

	caps := make([]uuid.UUID, 0, len(params))
	for _, capStr := range params {
		uid, err := uuid.Parse(capStr)
		if err != nil {
			return fmt.Errorf("parse caps failed: %w", err)
		}
		caps = append(caps, uid)
	}

	chatuid, err := uuid.Parse("748F2420-6287-11D1-8222-444553540000")
	if err != nil {
		return fmt.Errorf("parse caps failed: %w", err)
	}
	caps = append(caps, chatuid)

	snac := wire.SNAC_0x02_0x04_LocateSetInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.LocateTLVTagsInfoCapabilities, caps),
			},
		},
	}

	if err := b.LocateService.SetInfo(ctx, me, snac); err != nil {
		return fmt.Errorf("SetInfo: %w", err)
	}

	return nil
}

func (b BOSProxy) SetAway(ctx context.Context, me *state.Session, awayMessage string) error {
	snac := wire.SNAC_0x02_0x04_LocateSetInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.LocateTLVTagsInfoUnavailableData, awayMessage),
			},
		},
	}

	if err := b.LocateService.SetInfo(ctx, me, snac); err != nil {
		return fmt.Errorf("SetInfo: %w", err)
	}

	return nil
}

func (b BOSProxy) Evil(ctx context.Context, me *state.Session, params []string) (string, error) {
	snac := wire.SNAC_0x04_0x08_ICBMEvilRequest{
		SendAs:     0,
		ScreenName: params[1],
	}
	if params[2] == "anon" {
		snac.SendAs = 1
	}
	response, err := b.ICBMService.EvilRequest(ctx, me, wire.SNACFrame{}, snac)
	if err != nil {
		return "", fmt.Errorf("EvilRequest: %w", err)
	}

	return fmt.Sprintf("EVILED:%d:jon", response.Body.(wire.SNAC_0x04_0x09_ICBMEvilReply).UpdatedEvilValue), nil
}

func (b BOSProxy) SetInfo(ctx context.Context, me *state.Session, params []string) error {
	snac := wire.SNAC_0x02_0x04_LocateSetInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.LocateTLVTagsInfoSigData, params[1]),
			},
		},
	}
	if err := b.LocateService.SetInfo(ctx, me, snac); err != nil {
		return fmt.Errorf("SetInfo: %w", err)
	}

	return nil
}

func (b BOSProxy) SetDir(ctx context.Context, me *state.Session, params []string) error {
	info := strings.Split(params[1], ":")

	snac := wire.SNAC_0x02_0x09_LocateSetDirInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ODirTLVFirstName, info[0]),
				wire.NewTLVBE(wire.ODirTLVMiddleName, info[1]),
				wire.NewTLVBE(wire.ODirTLVLastName, info[2]),
				wire.NewTLVBE(wire.ODirTLVMaidenName, info[3]),
				wire.NewTLVBE(wire.ODirTLVCountry, info[6]),
				wire.NewTLVBE(wire.ODirTLVState, info[5]),
				wire.NewTLVBE(wire.ODirTLVCity, info[4]),
			},
		},
	}
	if _, err := b.LocateService.SetDirInfo(ctx, me, wire.SNACFrame{}, snac); err != nil {
		return fmt.Errorf("SetDirInfo: %w", err)
	}

	return nil
}

func (b BOSProxy) SetIdle(ctx context.Context, me *state.Session, params []string) error {
	time, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("SetIdle string to int: %w", err)
	}

	snac := wire.SNAC_0x01_0x11_OServiceIdleNotification{
		IdleTime: uint32(time),
	}
	if err := b.OServiceService.IdleNotification(ctx, me, snac); err != nil {
		return fmt.Errorf("SetIdle: %w", err)
	}

	return nil
}

func (b BOSProxy) UpdateBuddyArrival(snac wire.SNAC_0x03_0x0B_BuddyArrived) string {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "T", snac.WarningLevel, online, idle, uc)
}

func (b BOSProxy) UpdateBuddyDeparted(snac wire.SNAC_0x03_0x0C_BuddyDeparted) string {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	class := strings.Join(uc[:], "")
	return fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "F", snac.WarningLevel, online, idle, class)
}

func (b BOSProxy) IMIn(snac wire.SNAC_0x04_0x07_ICBMChannelMsgToClient) string {
	buf, ok := snac.TLVRestBlock.Bytes(wire.ICBMTLVAOLIMData)
	if !ok {
		return ""
	}
	txt, err := wire.UnmarshalICBMMessageText(buf)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("IM_IN:%s:F:%s", snac.ScreenName, txt)
}

func (b BOSProxy) Eviled(snac wire.SNAC_0x01_0x10_OServiceEvilNotification) string {
	who := ""
	if snac.Snitcher != nil {
		who = snac.Snitcher.ScreenName
	}
	return fmt.Sprintf("EVILED:%d:%s", snac.NewEvil, who)
}

type ChatProxy struct {
	//AuthService       AuthService
	//BuddyService      BuddyService
	//ICBMService       ICBMService
	//LocateService     LocateService
	ChatNavService  ChatNavService
	Logger          *slog.Logger
	ChatService     ChatService
	OServiceService OServiceService
	//PermitDenyService PermitDenyService
}

func (c ChatProxy) ConsumeIncoming(ctx context.Context, me *state.Session, ch chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case snac := <-me.ReceiveMessage():
			inFrame := snac.Frame
			switch inFrame.FoodGroup {
			case wire.Chat:
				switch inFrame.SubGroup {
				case wire.ChatUsersJoined:
					ch <- []byte(c.ChatUpdateBuddy(snac.Body.(wire.SNAC_0x0E_0x03_ChatUsersJoined)))
				case wire.ChatChannelMsgToClient:
					ch <- []byte(c.ChatIn(snac.Body.(wire.SNAC_0x0E_0x06_ChatChannelMsgToClient)))
				default:
					c.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
				}
			default:
				c.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
			}
		}
	}
}

func (c ChatProxy) ClientReady(ctx context.Context, sess *state.Session) error {
	if err := c.OServiceService.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		return fmt.Errorf("client online failed: %v", err)
	}
	return nil
}

func (c ChatProxy) ChatSend(ctx context.Context, me *state.Session, params []string) error {
	block := wire.TLVRestBlock{}
	// the order of these TLVs matters for AIM 2.x. if out of order, screen
	// names do not appear with each chat message.
	block.Append(wire.NewTLVBE(wire.ChatTLVEnableReflectionFlag, uint8(1)))
	block.Append(wire.NewTLVBE(wire.ChatTLVSenderInformation, me.TLVUserInfo()))
	block.Append(wire.NewTLVBE(wire.ChatTLVPublicWhisperFlag, []byte{}))
	block.Append(wire.NewTLVBE(wire.ChatTLVMessageInfo, wire.TLVRestBlock{
		TLVList: wire.TLVList{
			wire.NewTLVBE(wire.ChatTLVMessageInfoText, params[2]),
		},
	}))

	snac := wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{
		Channel:      3,
		TLVRestBlock: block,
	}
	if _, err := c.ChatService.ChannelMsgToHost(ctx, me, wire.SNACFrame{}, snac); err != nil {
		return fmt.Errorf("chat send failed: %v", err)
	}

	return nil
}

func (c ChatProxy) ChatJoin(ctx context.Context, me *state.Session, roomName string) (string, error) {
	snac := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange: 4, // todo
		Cookie:   "create",
		TLVBlock: wire.TLVBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ChatRoomTLVRoomName, roomName),
			},
		},
	}

	reply, err := c.ChatNavService.CreateRoom(ctx, me, wire.SNACFrame{}, snac)
	if err != nil {
		return "", fmt.Errorf("chat send failed: %v", err)
	}

	snac2 := wire.SNAC_0x01_0x04_OServiceServiceRequest{
		FoodGroup: wire.Chat,
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
					Cookie: roomInfo.Cookie,
				}),
			},
		},
	}
	rep, err := c.OServiceService.ServiceRequest(ctx, me, wire.SNACFrame{}, snac2)

	//	chatNavCh <- wire.SNACMessage{
	//		Frame: wire.SNACFrame{
	//			FoodGroup: wire.ChatNav,
	//			SubGroup:  wire.ChatNavCreateRoom,
	//		},
	//		Body: ,
	//	}
	//
	//	if err := clientFlap.SendDataFrame([]byte(fmt.Sprintf("CHAT_JOIN:%s:%s", "10", "haha"))); err != nil {
	//		return fmt.Errorf("send sign on data frame failed: %w", err)
	//	}
}

func (c ChatProxy) ChatUpdateBuddy(snac wire.SNAC_0x0E_0x03_ChatUsersJoined) string {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	return fmt.Sprintf("CHAT_UPDATE_BUDDY:%s:T:%s", "10", "mike")
}

func (c ChatProxy) ChatIn(snac wire.SNAC_0x0E_0x06_ChatChannelMsgToClient) string {
	b, _ := snac.Bytes(wire.ChatTLVSenderInformation)

	u := wire.TLVUserInfo{}
	err := wire.UnmarshalBE(&u, bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}

	b, _ = snac.Bytes(wire.ChatTLVMessageInfo)
	text, err := textFromChatMsgBlob(b)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("CHAT_IN:%s:%s:F:%s", "10", u.ScreenName, text)
}

// textFromChatMsgBlob extracts plaintext message text from HTML located in
// chat message info TLV(0x05).
func textFromChatMsgBlob(msg []byte) ([]byte, error) {
	block := wire.TLVRestBlock{}
	if err := wire.UnmarshalBE(&block, bytes.NewBuffer(msg)); err != nil {
		return nil, err
	}

	b, hasMsg := block.Bytes(wire.ChatTLVMessageInfoText)
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
