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
	"golang.org/x/net/html"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

var capChat = uuid.MustParse("748F2420-6287-11D1-8222-444553540000")

type BOSProxy struct {
	AuthService       AuthService
	BuddyListRegistry BuddyListRegistry
	BuddyService      BuddyService
	ChatNavService    ChatNavService
	ICBMService       ICBMService
	LocateService     LocateService
	Logger            *slog.Logger
	OServiceService   OServiceService
	PermitDenyService PermitDenyService
}

func (b BOSProxy) ConsumeIncoming(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, ch chan []byte) {
	defer func() {
		fmt.Println("closing BOS ConsumeIncoming")
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-me.Closed():
			close(ch)
			fmt.Println("I got signed off")
			return
		case snac := <-me.ReceiveMessage():
			inFrame := snac.Frame
			switch inFrame.FoodGroup {
			case wire.Buddy:
				switch inFrame.SubGroup {
				case wire.BuddyArrived:
					// todo make these type assertions safe?
					select {
					case ch <- []byte(b.UpdateBuddyArrival(snac.Body.(wire.SNAC_0x03_0x0B_BuddyArrived))):
					case <-ctx.Done():
						return
					}
				case wire.BuddyDeparted:
					select {
					case ch <- []byte(b.UpdateBuddyDeparted(snac.Body.(wire.SNAC_0x03_0x0C_BuddyDeparted))):
					case <-ctx.Done():
						return
					}
				default:
					b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.ICBM:
				switch inFrame.SubGroup {
				case wire.ICBMChannelMsgToClient:
					select {
					case ch <- []byte(b.IMIn(chatRegistry, snac.Body.(wire.SNAC_0x04_0x07_ICBMChannelMsgToClient))):
					case <-ctx.Done():
						return
					}
				default:
					b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.OService:
				switch inFrame.SubGroup {
				case wire.OServiceEvilNotification:
					select {
					case ch <- []byte(b.Eviled(snac.Body.(wire.SNAC_0x01_0x10_OServiceEvilNotification))):
					case <-ctx.Done():
						return
					}
				default:
					b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			default:
				b.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
			}
		}
	}
}

func (b BOSProxy) Login(ctx context.Context, elems []string, registry *ChatRegistry, ch chan []byte) (*state.Session, []string) {
	username := elems[3]
	passwordHash, err := hex.DecodeString(elems[4][2:])
	if err != nil {
		b.Logger.Error("decode password hash failed", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	signonFrame := wire.FLAPSignonFrame{}
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsScreenName, username))
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsRoastedTOCPassword, passwordHash))

	block, err := b.AuthService.FLAPLogin(signonFrame, state.NewStubUser)
	if err != nil {
		b.Logger.Error("FLAP login failed", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	if block.HasTag(wire.LoginTLVTagsErrorSubcode) {
		b.Logger.Debug("login failed")
		return nil, []string{"ERROR:980"} // bad username/password
	}

	authCookie, ok := block.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		b.Logger.Error("unable to get session id from payload")
		return nil, []string{"ERROR:989:internal server error"}
	}

	sess, err := b.AuthService.RegisterBOSSession(ctx, authCookie)
	if err != nil {
		b.Logger.Error("register BOS session failed", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	// set chat capability so that... tk
	sess.SetCaps([][16]byte{capChat})

	if err := b.BuddyListRegistry.RegisterBuddyList(sess.IdentScreenName()); err != nil {
		b.Logger.Error("unable to init buddy list", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	go b.ConsumeIncoming(ctx, sess, registry, ch)

	return sess, []string{"SIGN_ON:TOC1.0", "CONFIG:"}
}

func (b BOSProxy) ClientReady(ctx context.Context, sess *state.Session) error {
	if err := b.OServiceService.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		return fmt.Errorf("client online failed: %v", err)
	}
	return nil
}

func (b BOSProxy) Profile(thisCtx context.Context, from, user string) (string, error) {
	sess := state.NewSession()
	sess.SetIdentScreenName(state.NewIdentScreenName(from))
	inBody := wire.SNAC_0x02_0x05_LocateUserInfoQuery{
		Type:       uint16(wire.LocateTypeSig),
		ScreenName: user,
	}

	info, err := b.LocateService.UserInfoQuery(thisCtx, sess, wire.SNACFrame{}, inBody)
	if err != nil {
		b.Logger.Error("user session failed", "err", err.Error())
		return "", nil
	}
	if !(info.Frame.FoodGroup == wire.Locate && info.Frame.SubGroup == wire.LocateUserInfoReply) {
		b.Logger.Error("didn't get expected locate response")
		return "", nil
	}

	locateInfoReply := info.Body.(wire.SNAC_0x02_0x06_LocateUserInfoReply)
	profile, hasProf := locateInfoReply.LocateInfo.Bytes(wire.LocateTLVTagsInfoSigData)
	if !hasProf {
		b.Logger.Error("didn't get expected location info")
		return "", nil
	}

	return string(profile), nil
}

func (b BOSProxy) SendIM(ctx context.Context, me *state.Session, params []string) error {
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
	caps = append(caps, capChat)

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

func (b BOSProxy) IMIn(chatRegistry *ChatRegistry, snac wire.SNAC_0x04_0x07_ICBMChannelMsgToClient) string {
	if snac.ChannelID == wire.ICBMChannelRendezvous {
		rdinfo, has := snac.TLVRestBlock.Bytes(0x05)
		if !has {
			fmt.Printf("doesn't have rendezvous block\n")
			return ""
		}
		frag := wire.ICBMCh2Fragment{}
		if err := wire.UnmarshalBE(&frag, bytes.NewBuffer(rdinfo)); err != nil {
			fmt.Printf("unmarshal ICBM channel message rdv apyload failed: %w", err)
			return ""
		}
		prompt, _ := frag.Bytes(12)

		svcData, _ := frag.Bytes(10001)

		roomInfo := wire.ICBMRoomInfo{}
		if err := wire.UnmarshalBE(&roomInfo, bytes.NewBuffer(svcData)); err != nil {
			fmt.Printf("unmarshal ICBM channel message rdv apyload failed: %w", err)
			return ""
		}

		name := strings.Split(roomInfo.Cookie, "-")[2]

		chatID := chatRegistry.Add(roomInfo.Cookie)
		return fmt.Sprintf("CHAT_INVITE:%s:%d:%s:%s", name, chatID, snac.ScreenName, prompt)
	}

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

func (b BOSProxy) SetConfig(ctx context.Context, me *state.Session, params []string) error {
	config := strings.Split(strings.TrimSpace(params[1]), "\n")

	var cfg [][2]string
	for _, item := range config {
		parts := strings.Split(item, " ")
		if len(parts) != 2 {
			b.Logger.Info("invalid config item", "item", item, "user", me.DisplayScreenName())
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
			b.Logger.Info("config: invalid mode", "val", c[1], "user", me.DisplayScreenName())
		}
		//break todo add
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
		if err := b.PermitDenyService.AddDenyListEntries(ctx, me, snac); err != nil {
			return fmt.Errorf("AddDenyListEntries: %w", err)
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
		if err := b.PermitDenyService.AddPermListEntries(ctx, me, snac); err != nil {
			return fmt.Errorf("AddPermListEntries: %w", err)
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
		if err := b.PermitDenyService.AddPermListEntries(ctx, me, snac); err != nil {
			return fmt.Errorf("AddPermListEntries: %w", err)
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
		if err := b.PermitDenyService.AddDenyListEntries(ctx, me, snac); err != nil {
			return fmt.Errorf("AddDenyListEntries: %w", err)
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

	if err := b.BuddyService.AddBuddies(ctx, me, snac); err != nil {
		return fmt.Errorf("AddBuddies: %w", err)
	}

	return nil
}

func (b BOSProxy) Signout(ctx context.Context, me *state.Session) {
	b.BuddyService.BroadcastBuddyDeparted(ctx, me)
	if err := b.BuddyListRegistry.UnregisterBuddyList(me.IdentScreenName()); err != nil {
		b.Logger.ErrorContext(ctx, "error removing buddy list entry", "err", err.Error())
	}
	b.AuthService.Signout(ctx, me)
}

func (b BOSProxy) ChatInvite(ctx context.Context, bos *state.Session, chatRegistry *ChatRegistry, params []string) error {
	chatID, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("ChatSend string to int: %w", err)
	}

	cookie := chatRegistry.Lookup(chatID)
	if cookie == "" {
		return fmt.Errorf("chat not found: %d", chatID)
	}

	snac := wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
		ChannelID:  wire.ICBMChannelRendezvous,
		ScreenName: params[3],
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(0x05, wire.ICBMCh2Fragment{
					Type:       0,
					Capability: capChat,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(10, uint16(1)),
							wire.NewTLVBE(12, params[2]),
							wire.NewTLVBE(13, "us-ascii"),
							wire.NewTLVBE(14, "en"),
							wire.NewTLVBE(10001, wire.ICBMRoomInfo{
								Exchange: 4, // todo add this to chat registry
								Cookie:   cookie,
							}),
						},
					},
				}),
			},
		},
	}

	if _, err := b.ICBMService.ChannelMsgToHost(ctx, bos, wire.SNACFrame{}, snac); err != nil {
		return fmt.Errorf("ChannelMsgToHost: %w", err)
	}

	return nil
}

type ChatProxy struct {
	AuthService         AuthService
	ChatNavService      ChatNavService
	Logger              *slog.Logger
	ChatService         ChatService
	OServiceServiceBOS  OServiceService
	OServiceServiceChat OServiceService
}

func (s ChatProxy) ConsumeIncoming(ctx context.Context, me *state.Session, chatID int, ch chan []byte) {
	defer func() {
		fmt.Println("closing chat ConsumeIncoming")
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
					select {
					case <-ctx.Done():
						return
					case ch <- []byte(s.ChatUpdateBuddyLeft(snac.Body.(wire.SNAC_0x0E_0x04_ChatUsersLeft), chatID)):
					}
				case wire.ChatUsersJoined:
					select {
					case <-ctx.Done():
						return
					case ch <- []byte(s.ChatUpdateBuddyArrived(snac.Body.(wire.SNAC_0x0E_0x03_ChatUsersJoined), chatID)):
					}
				case wire.ChatChannelMsgToClient:
					select {
					case <-ctx.Done():
						return
					case ch <- []byte(s.ChatIn(snac.Body.(wire.SNAC_0x0E_0x06_ChatChannelMsgToClient), chatID)):
					}
				default:
					s.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
				}
			default:
				s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
			}
		}
	}
}

func (s ChatProxy) ChatJoin(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, params []string, clientCh chan []byte) error {
	exchange, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("parse exchange failed: %w", err)
	}

	snac := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange: uint16(exchange),
		Cookie:   "create",
		TLVBlock: wire.TLVBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ChatRoomTLVRoomName, params[2]),
			},
		},
	}

	reply, err := s.ChatNavService.CreateRoom(ctx, me, wire.SNACFrame{}, snac)
	if err != nil {
		return fmt.Errorf("chat send failed: %v", err)
	}

	chatSNAC := reply.Body.(wire.SNAC_0x0D_0x09_ChatNavNavInfo)
	buf, ok := chatSNAC.Bytes(wire.ChatNavTLVRoomInfo)
	if !ok {
		return fmt.Errorf("retrieve chat s update: %v", err)
	}

	inBody := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := wire.UnmarshalBE(&inBody, bytes.NewBuffer(buf)); err != nil {
		return err
	}

	snac2 := wire.SNAC_0x01_0x04_OServiceServiceRequest{
		FoodGroup: wire.Chat,
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
					Cookie: inBody.Cookie,
				}),
			},
		},
	}
	rep, err := s.OServiceServiceBOS.ServiceRequest(ctx, me, wire.SNACFrame{}, snac2)
	if err != nil {
		return fmt.Errorf("service request failed: %v", err)
	}

	chatResp := rep.Body.(wire.SNAC_0x01_0x05_OServiceServiceResponse)

	cookie, hasCookie := chatResp.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !hasCookie {
		return fmt.Errorf("retrieve chat s update: %v", err)
	}

	sess, err := s.AuthService.RegisterChatSession(ctx, cookie)
	if err != nil {
		return fmt.Errorf("register chat session failed: %v", err)
	}

	chatID := chatRegistry.Add(inBody.Cookie)
	chatRegistry.Register(chatID, sess)

	go s.ConsumeIncoming(ctx, sess, chatID, clientCh)

	clientCh <- []byte(fmt.Sprintf("CHAT_JOIN:%d:%s", chatID, params[2]))

	if err := s.OServiceServiceChat.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		return fmt.Errorf("client online failed: %v", err)
	}

	return nil
}

func (s ChatProxy) ChatAccept(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, params []string, clientCh chan []byte) error {
	chatID, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("ChatSend string to int: %w", err)
	}

	cookie := chatRegistry.Lookup(chatID)
	if cookie == "" {
		return fmt.Errorf("chat not found: %d", chatID)
	}

	snac := wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
		Cookie:   cookie,
		Exchange: 4, // todo put this in session lookup
	}

	// begin
	info, err := s.ChatNavService.RequestRoomInfo(ctx, wire.SNACFrame{}, snac)
	if err != nil {
		return fmt.Errorf("chat request room info failed: %v", err)
	}

	infoSNAC := info.Body.(wire.SNAC_0x0D_0x09_ChatNavNavInfo)
	b, hasInfo := infoSNAC.Bytes(wire.ChatNavTLVRoomInfo)
	if !hasInfo {
		return fmt.Errorf("error getting room info from room info payload")
	}

	roomInfo := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := wire.UnmarshalBE(&roomInfo, bytes.NewBuffer(b)); err != nil {
		return fmt.Errorf("error unmarshalling room info: %w", err)
	}

	name, hasName := roomInfo.Bytes(wire.ChatRoomTLVRoomName)
	if !hasName {
		return fmt.Errorf("error getting room name from room info payload")
	}

	//end
	snac2 := wire.SNAC_0x01_0x04_OServiceServiceRequest{
		FoodGroup: wire.Chat,
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
					Cookie: cookie,
				}),
			},
		},
	}
	rep, err := s.OServiceServiceBOS.ServiceRequest(ctx, me, wire.SNACFrame{}, snac2)
	if err != nil {
		return fmt.Errorf("service request failed: %v", err)
	}

	chatResp := rep.Body.(wire.SNAC_0x01_0x05_OServiceServiceResponse)

	sessionCookie, hasCookie := chatResp.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !hasCookie {
		return fmt.Errorf("retrieve chat b update: %v", err)
	}

	sess, err := s.AuthService.RegisterChatSession(ctx, sessionCookie)
	if err != nil {
		return fmt.Errorf("register chat session failed: %v", err)
	}

	go s.ConsumeIncoming(ctx, sess, chatID, clientCh)

	chatRegistry.Register(chatID, sess)

	clientCh <- []byte(fmt.Sprintf("CHAT_JOIN:%d:%s", chatID, name))

	if err := s.OServiceServiceChat.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		return fmt.Errorf("client online failed: %v", err)
	}

	return nil
}

func (s ChatProxy) ChatSend(ctx context.Context, chatRegistry *ChatRegistry, params []string) (string, error) {
	chatID, err := strconv.Atoi(params[1])
	if err != nil {
		return "", fmt.Errorf("ChatSend string to int: %w", err)
	}

	me := chatRegistry.Retrieve(chatID)
	if me == nil {
		return "", fmt.Errorf("ChatSend session not found: %d", chatID)
	}

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
	if _, err := s.ChatService.ChannelMsgToHost(ctx, me, wire.SNACFrame{}, snac); err != nil {
		return "", fmt.Errorf("chat send failed: %v", err)
	}

	return fmt.Sprintf("CHAT_IN:%d:%s:F:%s", chatID, me.DisplayScreenName(), params[2]), nil
}

func (s ChatProxy) ChatUpdateBuddyArrived(snac wire.SNAC_0x0E_0x03_ChatUsersJoined, chatID int) string {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	return fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:T:%s", chatID, strings.Join(users, ":"))
}

func (s ChatProxy) ChatUpdateBuddyLeft(snac wire.SNAC_0x0E_0x04_ChatUsersLeft, chatID int) string {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	return fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:F:%s", chatID, strings.Join(users, ":"))
}

func (s ChatProxy) ChatIn(snac wire.SNAC_0x0E_0x06_ChatChannelMsgToClient, chatID int) string {
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

	return fmt.Sprintf("CHAT_IN:%d:%s:F:%s", chatID, u.ScreenName, text)
}

func (s ChatProxy) ChatLeave(ctx context.Context, chatRegistry *ChatRegistry, params []string) (string, error) {
	chatID, err := strconv.Atoi(params[1])
	if err != nil {
		return "", fmt.Errorf("ChatSend string to int: %w", err)
	}

	me := chatRegistry.Retrieve(chatID)
	if me == nil {
		return "", fmt.Errorf("ChatSend session not found: %d", chatID)
	}

	s.AuthService.SignoutChat(ctx, me)
	me.Close()

	return fmt.Sprintf("CHAT_LEFT:%d", chatID), nil
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
