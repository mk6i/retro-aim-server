package toc

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"golang.org/x/net/html"
)

const profileTpl = `
<HTML><HEAD><TITLE>Profile Lookup</TITLE></HEAD><BODY>
Username : <B>{{- .ScreenName -}}</B><BR><BR>
{{ .Profile }}
</BODY></HTML>`

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

func (s OSCARProxy) NewServeMux() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /info", s.AuthMiddleware(http.HandlerFunc(s.ProfileHandler)))
	mux.Handle("GET /dir_info", s.AuthMiddleware(http.HandlerFunc(s.DirInfoHandler)))
	mux.Handle("GET /dir_search", s.AuthMiddleware(http.HandlerFunc(s.DirSearchHandler)))
	return mux
}

func (s OSCARProxy) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := url.QueryUnescape(r.URL.Query().Get("cookie"))
		if err != nil {
			http.Error(w, "unable to read auth cookie", http.StatusBadRequest)
			return
		}

		data, err := base64.URLEncoding.DecodeString(cookie)
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

func (s OSCARProxy) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	if from == "" {
		http.Error(w, "required `from` param is missing`", http.StatusBadRequest)
		return
	}
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "required `user` param is missing", http.StatusBadRequest)
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
			http.Error(w, "user is unavailable", http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	case wire.SNAC_0x02_0x06_LocateUserInfoReply:
		profile, hasProf := v.LocateInfo.Bytes(wire.LocateTLVTagsInfoSigData)
		if !hasProf {
			logErr(ctx, s.Logger, errors.New("LocateInfo.Bytes: missing wire.LocateTLVTagsInfoSigData"))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		t, err := template.New("results").Parse(profileTpl)
		if err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("template.New: %w", err))
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

		if err := t.Execute(w, pd); err != nil {
			logErr(ctx, s.Logger, fmt.Errorf("t.Execute: %w", err))
			http.Error(w, "internal Server Error", http.StatusInternalServerError)
			return
		}
	default:
		logErr(ctx, s.Logger, fmt.Errorf("unknown response type: %T", v))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (s OSCARProxy) DirInfoHandler(w http.ResponseWriter, request *http.Request) {
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

func (s OSCARProxy) DirSearchHandler(w http.ResponseWriter, r *http.Request) {
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
