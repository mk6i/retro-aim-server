package toc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
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

type OSCARProxy struct {
	AuthService         AuthService
	BuddyListRegistry   BuddyListRegistry
	BuddyService        BuddyService
	ChatNavService      ChatNavService
	ChatService         ChatService
	DirSearchService    DirSearchService
	ICBMService         ICBMService
	LocateService       LocateService
	Logger              *slog.Logger
	OServiceServiceBOS  OServiceService
	OServiceServiceChat OServiceService
	PermitDenyService   PermitDenyService
	TOCConfigStore      TOCConfigStore
	CookieBaker         state.HMACCookieBaker
}

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
					s.UpdateBuddyArrival(ctx, snac.Body.(wire.SNAC_0x03_0x0B_BuddyArrived), ch)
				case wire.BuddyDeparted:
					s.UpdateBuddyDeparted(ctx, snac.Body.(wire.SNAC_0x03_0x0C_BuddyDeparted), ch)
				default:
					// don't return error because they could be booted by malicious actor?
					s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.ICBM:
				switch inFrame.SubGroup {
				case wire.ICBMChannelMsgToClient:
					s.IMIn(ctx, chatRegistry, snac.Body.(wire.SNAC_0x04_0x07_ICBMChannelMsgToClient), ch)
				default:
					s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
				}
			case wire.OService:
				switch inFrame.SubGroup {
				case wire.OServiceEvilNotification:
					s.Eviled(ctx, snac.Body.(wire.SNAC_0x01_0x10_OServiceEvilNotification), ch)
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

func (s OSCARProxy) Login(ctx context.Context, cmd []byte) (*state.Session, []string) {
	var userName, password string

	if _, err := parseArgs(cmd, "toc_signon", nil, nil, &userName, &password); err != nil {
		s.Logger.Error("parseArgs filed", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	passwordHash, err := hex.DecodeString(password[2:])
	if err != nil {
		s.Logger.Error("decode password hash failed", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	signonFrame := wire.FLAPSignonFrame{}
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsScreenName, userName))
	signonFrame.Append(wire.NewTLVBE(wire.LoginTLVTagsRoastedTOCPassword, passwordHash))

	block, err := s.AuthService.FLAPLogin(signonFrame, state.NewStubUser)
	if err != nil {
		s.Logger.Error("FLAP login failed", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	if block.HasTag(wire.LoginTLVTagsErrorSubcode) {
		s.Logger.Debug("login failed")
		return nil, []string{"ERROR:980"} // bad username/password
	}

	authCookie, ok := block.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		s.Logger.Error("unable to get session id from payload")
		return nil, []string{"ERROR:989:internal server error"}
	}

	sess, err := s.AuthService.RegisterBOSSession(ctx, authCookie)
	if err != nil {
		s.Logger.Error("register BOS session failed", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	// set chat capability so that... tk
	sess.SetCaps([][16]byte{capChat})

	if err := s.BuddyListRegistry.RegisterBuddyList(sess.IdentScreenName()); err != nil {
		s.Logger.Error("unable to init buddy list", "err", err.Error())
		return nil, []string{"ERROR:989:internal server error"}
	}

	u, err := s.TOCConfigStore.User(sess.IdentScreenName())
	if err != nil {
		s.Logger.Error("TOCConfigStore.User retrieval error", "err", err.Error())
	}

	var cfg string
	if u != nil {
		cfg = u.TOCConfig
	} else {
		s.Logger.Error("TOCConfigStore.User: user not found")
	}

	return sess, []string{"SIGN_ON:TOC1.0", fmt.Sprintf("CONFIG:%s", cfg)}
}

func (s OSCARProxy) BOSReady(ctx context.Context, sess *state.Session, ch chan<- []byte) {
	if err := s.OServiceServiceBOS.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ClientOnliney: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}
}

func (s OSCARProxy) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := url.QueryUnescape(r.URL.Query().Get("cookie"))
		if err != nil {
			http.Error(w, "unable to read auth cookie", http.StatusBadRequest)
			return
		}

		data, err := base64.StdEncoding.DecodeString(cookie)
		if err != nil {
			http.Error(w, "unable to read auth cookie", http.StatusBadRequest)
			return
		}

		if _, err = s.CookieBaker.Crack(data); err != nil {
			http.Error(w, "unable to crack auth cookie", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

const profileTpl = `
<HTML><HEAD><TITLE>Profile Lookup</TITLE></HEAD><BODY>
Username : <B>{{- .ScreenName -}}</B><BR><BR>
{{ .Profile }}
</BODY></HTML>`

func (s OSCARProxy) Profile(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	if from == "" {
		http.Error(w, "user does not exist", http.StatusBadRequest)
		return
	}
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "user does not exist", http.StatusBadRequest)
		return
	}

	sess := state.NewSession()
	sess.SetIdentScreenName(state.NewIdentScreenName(from))
	inBody := wire.SNAC_0x02_0x05_LocateUserInfoQuery{
		Type:       uint16(wire.LocateTypeSig),
		ScreenName: user,
	}

	ctx := r.Context()
	info, err := s.LocateService.UserInfoQuery(ctx, sess, wire.SNACFrame{}, inBody)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.UserInfoQuery: %w", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	switch v := info.Body.(type) {
	case wire.SNACError:
		if v.Code == wire.ErrorCodeNotLoggedOn {
			if _, err = w.Write([]byte("user is unavailable")); err != nil {
				s.Logger.Error("error writing response", "err", err.Error())
			}
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
	case wire.SNAC_0x02_0x06_LocateUserInfoReply:
		profile, hasProf := v.LocateInfo.Bytes(wire.LocateTLVTagsInfoSigData)
		if !hasProf {
			logErr(ctx, s.Logger, errors.New("LocateInfo.Bytes: missing wire.LocateTLVTagsInfoSigData"))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		t, err := template.New("results").Parse(profileTpl)
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		pd := struct {
			ScreenName string
			Profile    string
		}{
			ScreenName: user,
			Profile:    extractBodyContent(profile),
		}

		buf := &bytes.Buffer{}
		if err := t.Execute(buf, pd); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(buf.Bytes()); err != nil {
			s.Logger.Error("error writing response", "err", err.Error())
		}
	default:
		logErr(ctx, s.Logger, fmt.Errorf("unknown response type: %T", v))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// extractBodyContent parses an HTML string and extracts the content within <BODY>...</BODY> tags.
func extractBodyContent(htmlContent []byte) string {
	tokenizer := html.NewTokenizer(bytes.NewReader(htmlContent))
	var bodyContent bytes.Buffer
	inBody := false

	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != nil && err != io.EOF {
				return "unable to read profile"
			}
			return bodyContent.String()
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "body" {
				inBody = true
			}
		case html.EndTagToken:
			token := tokenizer.Token()
			if token.Data == "body" {
				inBody = false
			}
		case html.TextToken:
			if inBody {
				bodyContent.Write(tokenizer.Text())
			}
		}
	}

	return ""
}

// SendIM handles the toc_send_im TOC command, which sends instant messages. It
// returns a TOC internal error if there's a problem performing the operation.
//
// Command syntax: toc_send_im <Destination User> <Message> [auto]
func (s OSCARProxy) SendIM(ctx context.Context, sender *state.Session, cmd []byte) []byte {
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

	return nil
}

// AddBuddy handles the toc_add_buddy TOC command, which adds buddies to your
// buddy list. It returns a TOC internal error if there's a problem performing
// the operation.
//
// Command syntax: toc_add_buddy <Buddy User 1> [<Buddy User2> [<Buddy User 3> [...]]]
func (s OSCARProxy) AddBuddy(ctx context.Context, me *state.Session, cmd []byte) []byte {
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

	return nil
}

// RemoveBuddy handles the toc_remove_buddy TOC command, which removes buddies
// from your buddy list. It returns a TOC internal error if there's a problem
// performing the operation.
//
// Command syntax: toc_remove_buddy <Buddy User 1> [<Buddy User2> [<Buddy User 3> [...]]]
func (s OSCARProxy) RemoveBuddy(ctx context.Context, me *state.Session, cmd []byte) []byte {
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
	return nil
}

// AddPermit handles the toc_add_permit TOC command, which adds buddies to your
// list of allowed buddies. It returns a TOC internal error if there's a
// problem performing the operation.
//
// Command syntax: toc_add_permit [ <User 1> [<User 2> [...]]]
func (s OSCARProxy) AddPermit(ctx context.Context, me *state.Session, cmd []byte) []byte {
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
	return nil
}

// AddDeny handles the toc_add_deny TOC command, which adds buddies to your
// list of denied buddies. It returns a TOC internal error if there's a problem
// performing the operation.
//
// Command syntax: toc_add_deny [ <User 1> [<User 2> [...]]]
func (s OSCARProxy) AddDeny(ctx context.Context, me *state.Session, cmd []byte) []byte {
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
	return nil
}

// SetCaps handles the toc_set_caps TOC command, which informs the server which
// capabilities the client supports. It returns a TOC internal error if there's
// a problem performing the operation.
//
// Command syntax: toc_set_caps [ <Capability 1> [<Capability 2> [...]]]
func (s OSCARProxy) SetCaps(ctx context.Context, me *state.Session, cmd []byte) []byte {
	params, err := parseArgs(cmd, "toc_set_caps")
	if err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
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

	return nil
}

// SetAway handles the toc_set_away TOC command, which sets an away message. If
// the message parameter is present, set the user as away, otherwise clear away
// status. It returns a TOC internal error if there's a problem performing the
// operation.
//
// Command syntax: toc_set_away [<away message>]
func (s OSCARProxy) SetAway(ctx context.Context, me *state.Session, cmd []byte) []byte {
	maybeMsg, err := parseArgs(cmd, "toc_set_away")
	if err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
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

	return nil
}

func (s OSCARProxy) Evil(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var user, scope string

	if _, err := parseArgs(cmd, "toc_evil", &user, &scope); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	snac := wire.SNAC_0x04_0x08_ICBMEvilRequest{
		SendAs:     0,
		ScreenName: user,
	}
	if scope == "anon" {
		snac.SendAs = 1
	}
	response, err := s.ICBMService.EvilRequest(ctx, me, wire.SNACFrame{}, snac)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ICBMService.EvilRequest: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("EVILED:%d:%s", response.Body.(wire.SNAC_0x04_0x09_ICBMEvilReply).UpdatedEvilValue, user)))
}

func (s OSCARProxy) SetInfo(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var info string

	if _, err := parseArgs(cmd, "toc_set_info", &info); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
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
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}
}

func (s OSCARProxy) SetDir(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var info string

	if _, err := parseArgs(cmd, "toc_set_dir", &info); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	attrs := strings.Split(info, ":")
	if len(attrs) != 7 {
		logErr(ctx, s.Logger, fmt.Errorf("expected 7 params, got %d", len(attrs)))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	snac := wire.SNAC_0x02_0x09_LocateSetDirInfo{
		TLVRestBlock: wire.TLVRestBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ODirTLVFirstName, attrs[0]),
				wire.NewTLVBE(wire.ODirTLVMiddleName, attrs[1]),
				wire.NewTLVBE(wire.ODirTLVLastName, attrs[2]),
				wire.NewTLVBE(wire.ODirTLVMaidenName, attrs[3]),
				wire.NewTLVBE(wire.ODirTLVCountry, attrs[6]),
				wire.NewTLVBE(wire.ODirTLVState, attrs[5]),
				wire.NewTLVBE(wire.ODirTLVCity, attrs[4]),
			},
		},
	}
	if _, err := s.LocateService.SetDirInfo(ctx, me, wire.SNACFrame{}, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.SetDirInfo: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}
}

func (s OSCARProxy) GetDirURL(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var user string

	if _, err := parseArgs(cmd, "toc_get_dir", &user); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	p := url.Values{}
	p.Add("user", user)

	if err := s.addCookie(me, p); err != nil {
		s.Logger.Error("unable to generate cookie", "err", err.Error())
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("GOTO_URL:directory info:dir_info?%s", p.Encode())))
}

func (s OSCARProxy) GetDirSearchURL(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var info string

	if _, err := parseArgs(cmd, "toc_dir_search", &info); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
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
			p.Add(labels[i], params[i])
		}
		i++
	}

	if err := s.addCookie(me, p); err != nil {
		s.Logger.Error("unable to generate cookie", "err", err.Error())
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("GOTO_URL:search results:dir_search?%s", p.Encode())))
}

func (s OSCARProxy) SetIdle(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var idleTimeStr string

	if _, err := parseArgs(cmd, "toc_set_idle", &idleTimeStr); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	time, err := strconv.Atoi(idleTimeStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	snac := wire.SNAC_0x01_0x11_OServiceIdleNotification{
		IdleTime: uint32(time),
	}
	if err := s.OServiceServiceBOS.IdleNotification(ctx, me, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.IdleNotification: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}
}

func (s OSCARProxy) UpdateBuddyArrival(ctx context.Context, snac wire.SNAC_0x03_0x0B_BuddyArrived, ch chan<- []byte) {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "T", snac.WarningLevel, online, idle, uc)))
}

func (s OSCARProxy) UpdateBuddyDeparted(ctx context.Context, snac wire.SNAC_0x03_0x0C_BuddyDeparted, ch chan<- []byte) {
	online, _ := snac.Uint32BE(wire.OServiceUserInfoSignonTOD)
	idle, _ := snac.Uint16BE(wire.OServiceUserInfoIdleTime)
	uc := [3]string{" ", "O", " "}
	if snac.IsAway() {
		uc[2] = "U"
	}
	class := strings.Join(uc[:], "")
	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("UPDATE_BUDDY:%s:%s:%d:%d:%d:%s", snac.ScreenName, "F", snac.WarningLevel, online, idle, class)))
}

func (s OSCARProxy) IMIn(ctx context.Context, chatRegistry *ChatRegistry, snac wire.SNAC_0x04_0x07_ICBMChannelMsgToClient, ch chan<- []byte) {
	if snac.ChannelID == wire.ICBMChannelRendezvous {
		rdinfo, has := snac.TLVRestBlock.Bytes(0x05)
		if !has {
			logErr(ctx, s.Logger, errors.New("TLVRestBlock.Bytes: missing rendezvous block"))
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
		}
		frag := wire.ICBMCh2Fragment{}
		if err := wire.UnmarshalBE(&frag, bytes.NewBuffer(rdinfo)); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
		}
		prompt, ok := frag.Bytes(12)
		if !ok {
			logErr(ctx, s.Logger, errors.New("frag.Bytes: missing prompt"))
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
		}

		svcData, ok := frag.Bytes(10001)
		if !ok {
			logErr(ctx, s.Logger, errors.New("frag.Bytes: missing room info"))
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
		}

		roomInfo := wire.ICBMRoomInfo{}
		if err := wire.UnmarshalBE(&roomInfo, bytes.NewBuffer(svcData)); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
		}

		name := strings.Split(roomInfo.Cookie, "-")[2]

		chatID := chatRegistry.Add(roomInfo.Cookie)
		sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_INVITE:%s:%d:%s:%s", name, chatID, snac.ScreenName, prompt)))
		return
	}

	buf, ok := snac.TLVRestBlock.Bytes(wire.ICBMTLVAOLIMData)
	if !ok {
		logErr(ctx, s.Logger, errors.New("TLVRestBlock.Bytes: missing wire.ICBMTLVAOLIMData"))
		return
	}
	txt, err := wire.UnmarshalICBMMessageText(buf)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalICBMMessageText: %w", err))
		return
	}
	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("IM_IN:%s:F:%s", snac.ScreenName, txt)))
	return
}

func (s OSCARProxy) Eviled(ctx context.Context, snac wire.SNAC_0x01_0x10_OServiceEvilNotification, ch chan<- []byte) {
	who := ""
	if snac.Snitcher != nil {
		who = snac.Snitcher.ScreenName
	}
	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("EVILED:%d:%s", snac.NewEvil, who)))
}

func (s OSCARProxy) SetConfig(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var info string

	if _, err := parseArgs(cmd, "toc_set_config", &info); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	// gaim uses braces instead of quotes for some reason
	info = strings.Replace(info, "{", "\"", 1)
	info = strings.Replace(info, "}", "\"", 1)
	info = strings.TrimSpace(info)

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
		if err := s.PermitDenyService.AddDenyListEntries(ctx, me, snac); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("PermitDenyService.AddDenyListEntries: %w", err))
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
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
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
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
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
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
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
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
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	if err := s.TOCConfigStore.SetTOCConfig(me.IdentScreenName(), info); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("TOCConfigStore.SaveTOCConfig: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}
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

func (s OSCARProxy) ChatInvite(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, cmd []byte, ch chan<- []byte) {
	var chatRoomIDStr, msg string

	users, err := parseArgs(cmd, "toc_chat_invite", &chatRoomIDStr, &msg)
	if err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	chatID, err := strconv.Atoi(chatRoomIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	cookie := chatRegistry.Lookup(chatID)
	if cookie == "" {
		logErr(ctx, s.Logger, fmt.Errorf("chatRegistry.Lookup: chat ID `%d` not found", chatID))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
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

		if _, err := s.ICBMService.ChannelMsgToHost(ctx, me, wire.SNACFrame{}, snac); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("ICBMService.ChannelMsgToHost: %w", err))
			sendOrCancel(ctx, ch, cmdInternalSvcErr)
			return
		}
	}
}

func (s OSCARProxy) GetInfoURL(ctx context.Context, me *state.Session, cmd []byte, ch chan<- []byte) {
	var user string
	if _, err := parseArgs(cmd, "toc_get_info", &user); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	p := url.Values{}
	p.Add("from", me.IdentScreenName().String())
	p.Add("user", user)

	if err := s.addCookie(me, p); err != nil {
		s.Logger.Error("unable to generate cookie", "err", err.Error())
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("GOTO_URL:profile:info?%s", p.Encode())))
}

func (s OSCARProxy) addCookie(me *state.Session, p url.Values) error {
	cookie, err := s.CookieBaker.Issue([]byte(me.IdentScreenName().String()))
	if err != nil {
		return err
	}
	p.Add("cookie", url.QueryEscape(base64.StdEncoding.EncodeToString(cookie)))
	return nil
}

func (s OSCARProxy) DirInfoHTTP(w http.ResponseWriter, request *http.Request) {
	user := request.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "user does not exist", http.StatusBadRequest)
		return
	}

	inBody := wire.SNAC_0x02_0x0B_LocateGetDirInfo{
		ScreenName: user,
	}

	ctx := request.Context()
	info, err := s.LocateService.DirInfo(ctx, wire.SNACFrame{}, inBody)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.UserInfoQuery: %w", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !(info.Frame.FoodGroup == wire.Locate && info.Frame.SubGroup == wire.LocateGetDirReply) {
		logErr(ctx, s.Logger, fmt.Errorf("LocateService.DirInfo: expected response SNAC(%d,%d), got SNAC(%d,%d)",
			wire.Locate, wire.LocateGetDirReply, info.Frame.FoodGroup, info.Frame.SubGroup))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	locateInfoReply := info.Body.(wire.SNAC_0x02_0x0C_LocateGetDirReply)

	if len(locateInfoReply.TLVList) == 0 {
		if _, err = w.Write([]byte("no user directory info")); err != nil {
			s.Logger.Error("error writing response", "err", err.Error())
		}
		return
	}

	outputSearchResults(w, s.Logger, locateInfoReply.TLVBlock)
}

func (s OSCARProxy) DirSearchHTTP(w http.ResponseWriter, r *http.Request) {
	inBody := wire.SNAC_0x0F_0x02_InfoQuery{}

	q := r.URL.Query()
	switch {
	case q.Has("first_name") || q.Has("last_name"):
		if val := q.Get("first_name"); val != "" {
			inBody.Append(wire.NewTLVBE(wire.ODirTLVFirstName, val))
		}
		if val := q.Get("middle_name"); val != "" {
			inBody.Append(wire.NewTLVBE(wire.ODirTLVMiddleName, val))
		}
		if val := q.Get("last_name"); val != "" {
			inBody.Append(wire.NewTLVBE(wire.ODirTLVLastName, val))
		}
		if val := q.Get("maiden_name"); val != "" {
			inBody.Append(wire.NewTLVBE(wire.ODirTLVMaidenName, val))
		}
		if val := q.Get("city"); val != "" {
			inBody.Append(wire.NewTLVBE(wire.ODirTLVCity, val))
		}
		if val := q.Get("state"); val != "" {
			inBody.Append(wire.NewTLVBE(wire.ODirTLVState, val))
		}
		if val := q.Get("country"); val != "" {
			inBody.Append(wire.NewTLVBE(wire.ODirTLVCountry, val))
		}
	case q.Has("email"):
		inBody.Append(wire.NewTLVBE(wire.ODirTLVEmailAddress, q.Get("email")))
	case q.Has("keyword"):
		inBody.Append(wire.NewTLVBE(wire.ODirTLVInterest, q.Get("keyword")))
	}

	ctx := r.Context()
	info, err := s.DirSearchService.InfoQuery(ctx, wire.SNACFrame{}, inBody)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("DirSearchService.InfoQuery: %w", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !(info.Frame.FoodGroup == wire.ODir && info.Frame.SubGroup == wire.ODirInfoReply) {
		logErr(ctx, s.Logger, fmt.Errorf("DirSearchService.InfoQuery: expected response SNAC(%d,%d), got SNAC(%d,%d)",
			wire.ODir, wire.ODirInfoReply, info.Frame.FoodGroup, info.Frame.SubGroup))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	locateInfoReply := info.Body.(wire.SNAC_0x0F_0x03_InfoReply)

	switch {
	case locateInfoReply.Status == wire.ODirSearchResponseNameMissing:
		if _, err = w.Write([]byte("search must contain first or last name")); err != nil {
			s.Logger.Error("error writing response", "err", err.Error())
		}
		return
	case locateInfoReply.Status != wire.ODirSearchResponseOK:
		if _, err = w.Write([]byte("search failed")); err != nil {
			s.Logger.Error("error writing response", "err", err.Error())
		}
		return
	}

	outputSearchResults(w, s.Logger, locateInfoReply.Results.List...)
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
					s.ChatUpdateBuddyLeft(ctx, snac.Body.(wire.SNAC_0x0E_0x04_ChatUsersLeft), chatID, ch)
				case wire.ChatUsersJoined:
					s.ChatUpdateBuddyArrived(ctx, snac.Body.(wire.SNAC_0x0E_0x03_ChatUsersJoined), chatID, ch)
				case wire.ChatChannelMsgToClient:
					s.ChatIn(ctx, snac.Body.(wire.SNAC_0x0E_0x06_ChatChannelMsgToClient), chatID, ch)
				default:
					s.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
				}
			default:
				s.Logger.Info(fmt.Sprintf("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup)))
			}
		}
	}
}

func (s OSCARProxy) ChatJoin(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, cmd []byte, ch chan<- []byte) (int, bool) {
	var exchangeStr, roomName string

	if _, err := parseArgs(cmd, "toc_chat_join", &exchangeStr, &roomName); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	exchange, err := strconv.Atoi(exchangeStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	snac := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange: uint16(exchange),
		Cookie:   "create",
		TLVBlock: wire.TLVBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ChatRoomTLVRoomName, roomName),
			},
		},
	}

	reply, err := s.ChatNavService.CreateRoom(ctx, me, wire.SNACFrame{}, snac)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ChatNavService.CreateRoom: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	chatSNAC := reply.Body.(wire.SNAC_0x0D_0x09_ChatNavNavInfo)
	buf, ok := chatSNAC.Bytes(wire.ChatNavTLVRoomInfo)
	if !ok {
		logErr(ctx, s.Logger, errors.New("chatSNAC.Bytes: missing wire.ChatNavTLVRoomInfo"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	inBody := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := wire.UnmarshalBE(&inBody, bytes.NewBuffer(buf)); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
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
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ServiceRequest: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	chatResp := rep.Body.(wire.SNAC_0x01_0x05_OServiceServiceResponse)

	cookie, hasCookie := chatResp.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !hasCookie {
		logErr(ctx, s.Logger, errors.New("chatResp.Bytes: missing wire.OServiceTLVTagsLoginCookie"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	sess, err := s.AuthService.RegisterChatSession(ctx, cookie)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("AuthService.RegisterChatSession: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	chatID := chatRegistry.Add(inBody.Cookie)
	chatRegistry.Register(chatID, sess)

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_JOIN:%d:%s", chatID, roomName)))

	if err := s.OServiceServiceChat.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceChat.ClientOnline: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	return chatID, true
}

func (s OSCARProxy) ChatAccept(ctx context.Context, me *state.Session, chatRegistry *ChatRegistry, cmd []byte, ch chan<- []byte) (int, bool) {
	var chatIDStr string

	if _, err := parseArgs(cmd, "toc_chat_accept", &chatIDStr); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	cookie := chatRegistry.Lookup(chatID)
	if cookie == "" {
		logErr(ctx, s.Logger, errors.New("chatRegistry.Lookup: no chat found"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	snac := wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
		Cookie:   cookie,
		Exchange: 4, // todo put this in session lookup
	}

	// begin
	info, err := s.ChatNavService.RequestRoomInfo(ctx, wire.SNACFrame{}, snac)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ChatNavService.RequestRoomInfo: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	infoSNAC := info.Body.(wire.SNAC_0x0D_0x09_ChatNavNavInfo)
	b, hasInfo := infoSNAC.Bytes(wire.ChatNavTLVRoomInfo)
	if !hasInfo {
		logErr(ctx, s.Logger, errors.New("infoSNAC.Bytes: missing wire.ChatNavTLVRoomInfo"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	roomInfo := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := wire.UnmarshalBE(&roomInfo, bytes.NewBuffer(b)); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	name, hasName := roomInfo.Bytes(wire.ChatRoomTLVRoomName)
	if !hasName {
		logErr(ctx, s.Logger, errors.New("roomInfo.Bytes: missing wire.ChatRoomTLVRoomName"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
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
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ServiceRequest: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	chatResp := rep.Body.(wire.SNAC_0x01_0x05_OServiceServiceResponse)

	sessionCookie, hasCookie := chatResp.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !hasCookie {
		logErr(ctx, s.Logger, errors.New("missing wire.OServiceTLVTagsLoginCookie"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	sess, err := s.AuthService.RegisterChatSession(ctx, sessionCookie)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("AuthService.RegisterChatSession: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	chatRegistry.Register(chatID, sess)

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_JOIN:%d:%s", chatID, name)))

	if err := s.OServiceServiceChat.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceChat.ClientOnline: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return 0, false
	}

	return chatID, true
}

func (s OSCARProxy) ChatReady(ctx context.Context, me *state.Session) bool {
	if err := s.OServiceServiceChat.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, me); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceChat.ClientOnline: %w", err))
		return false
	}
	return true
}

func (s OSCARProxy) ChatSend(ctx context.Context, chatRegistry *ChatRegistry, cmd []byte, ch chan<- []byte) {
	var chatIDStr, msg string

	if _, err := parseArgs(cmd, "toc_chat_send", &chatIDStr, &msg); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	me := chatRegistry.Retrieve(chatID)
	if me == nil {
		logErr(ctx, s.Logger, fmt.Errorf("chatRegistry.Retrieve: chat session `%d` not found", chatID))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
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
		Channel:      3,
		TLVRestBlock: block,
	}
	if _, err := s.ChatService.ChannelMsgToHost(ctx, me, wire.SNACFrame{}, snac); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("ChatService.ChannelMsgToHost: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_IN:%d:%s:F:%s", chatID, me.DisplayScreenName(), msg))) // todo do we reflect this?
}

func (s OSCARProxy) ChatUpdateBuddyArrived(ctx context.Context, snac wire.SNAC_0x0E_0x03_ChatUsersJoined, chatID int, ch chan<- []byte) {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:T:%s", chatID, strings.Join(users, ":"))))
}

func (s OSCARProxy) ChatUpdateBuddyLeft(ctx context.Context, snac wire.SNAC_0x0E_0x04_ChatUsersLeft, chatID int, ch chan<- []byte) {
	users := make([]string, 0, len(snac.Users))
	for _, u := range snac.Users {
		users = append(users, u.ScreenName)
	}
	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_UPDATE_BUDDY:%d:F:%s", chatID, strings.Join(users, ":"))))
}

func (s OSCARProxy) ChatIn(ctx context.Context, snac wire.SNAC_0x0E_0x06_ChatChannelMsgToClient, chatID int, ch chan<- []byte) {
	b, ok := snac.Bytes(wire.ChatTLVSenderInformation)
	if !ok {
		logErr(ctx, s.Logger, errors.New("snac.Bytes: missing wire.ChatTLVSenderInformation"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	u := wire.TLVUserInfo{}
	err := wire.UnmarshalBE(&u, bytes.NewBuffer(b))
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("wire.UnmarshalBE: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	b, ok = snac.Bytes(wire.ChatTLVMessageInfo)
	if !ok {
		logErr(ctx, s.Logger, errors.New("snac.Bytes: missing wire.ChatTLVMessageInfo"))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	text, err := textFromChatMsgBlob(b)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("textFromChatMsgBlob: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_IN:%d:%s:F:%s", chatID, u.ScreenName, text)))
}

func (s OSCARProxy) ChatLeave(ctx context.Context, chatRegistry *ChatRegistry, cmd []byte, ch chan<- []byte) {
	var chatIDStr string

	if _, err := parseArgs(cmd, "toc_chat_leave", &chatIDStr); err != nil {
		s.Logger.Error("error parsing TOC command", "givenPayload", string(cmd), "err", err)
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("strconv.Atoi: %w", err))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	me := chatRegistry.Retrieve(chatID)
	if me == nil {
		logErr(ctx, s.Logger, fmt.Errorf("chatRegistry.Retrieve: chat session `%d` not found", chatID))
		sendOrCancel(ctx, ch, cmdInternalSvcErr)
		return
	}

	s.AuthService.SignoutChat(ctx, me)
	me.Close()

	sendOrCancel(ctx, ch, []byte(fmt.Sprintf("CHAT_LEFT:%d", chatID)))
}

const directoryTpl = `
<HTML><HEAD><TITLE>Retro AIM Server</TITLE></HEAD><BODY><H3>Dir Results</H3>
{{- if .Results -}}
<TABLE>
{{- range .Results -}}
<TR><TD>
{{- if .FirstName}}<B>First Name:</B> {{.FirstName}}<BR>{{- end -}}
{{- if .MiddleName}}<B>Middle Name:</B> {{.MiddleName}}<BR>{{- end -}}
{{- if .LastName}}<B>Last Name:</B> {{.LastName}}<BR>{{- end -}}
{{- if .MaidenName}}<B>Maiden Name:</B> {{.MaidenName}}<BR>{{- end -}}
{{- if .Country}}<B>Country:</B> {{.Country}}<BR>{{- end -}}
{{- if .State}}<B>State:</B> {{.State}}<BR>{{- end -}}
{{- if .City}}<B>City:</B> {{.City}}<BR>{{- end -}}
{{- if .NickName}}<B>Nick Name:</B> {{.NickName}}<BR>{{- end -}}
{{- if .ZIP}}<B>ZIP Code:</B> {{.ZIP}}<BR>{{- end -}}
{{- if .Address}}<B>Address :</B> {{.Address}}<BR>{{- end -}}
</TD></TR>
{{- end -}}
</TABLE>
{{- else -}}
<BR>No results found.
{{- end -}}
</BODY></HTML>`

func outputSearchResults(w http.ResponseWriter, logger *slog.Logger, users ...wire.TLVBlock) {
	type DirSearchResult struct {
		FirstName  string
		MiddleName string
		LastName   string
		MaidenName string
		Country    string
		State      string
		City       string
		NickName   string
		ZIP        string
		Address    string
	}
	type PageData struct {
		Results []DirSearchResult
	}

	results := make([]DirSearchResult, 0, len(users))
	for _, result := range users {
		rec := DirSearchResult{}
		rec.FirstName, _ = result.String(wire.ODirTLVFirstName)
		rec.MiddleName, _ = result.String(wire.ODirTLVMiddleName)
		rec.LastName, _ = result.String(wire.ODirTLVLastName)
		rec.MaidenName, _ = result.String(wire.ODirTLVMaidenName)
		rec.Country, _ = result.String(wire.ODirTLVCountry)
		rec.State, _ = result.String(wire.ODirTLVState)
		rec.City, _ = result.String(wire.ODirTLVCity)
		rec.NickName, _ = result.String(wire.ODirTLVNickName)
		rec.ZIP, _ = result.String(wire.ODirTLVZIP)
		rec.Address, _ = result.String(wire.ODirTLVAddress)
		results = append(results, rec)
	}

	t, err := template.New("results").Parse(directoryTpl)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, PageData{Results: results}); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(buf.Bytes()); err != nil {
		logger.Error("error writing response", "err", err.Error())
	}
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
