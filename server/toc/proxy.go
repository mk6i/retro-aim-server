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
func (s OSCARProxy) InitDone(ctx context.Context, sess *state.Session, cmd []byte) []byte {
	if _, err := parseArgs(cmd, "toc_init_done"); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("parseArgs: %w", err))
		return cmdInternalSvcErr
	}
	if err := s.OServiceServiceBOS.ClientOnline(ctx, wire.SNAC_0x01_0x02_OServiceClientOnline{}, sess); err != nil {
		logErr(ctx, s.Logger, fmt.Errorf("OServiceServiceBOS.ClientOnliney: %w", err))
		return cmdInternalSvcErr
	}
	return nil
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
