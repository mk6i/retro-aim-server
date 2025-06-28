package oscar

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKerberosLoginHandler(t *testing.T) {
	tests := []struct {
		name               string
		request            wire.SNACMessage
		response           wire.SNACMessage
		responseErr        error
		expectLogin        bool
		expectSNACResponse bool
		wantStatus         int
	}{
		{
			name: "successful login",
			request: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Kerberos,
					SubGroup:  wire.KerberosLoginRequest,
				},
				Body: wire.SNAC_0x050C_0x0002_KerberosLoginRequest{
					RequestID: 4321,
				},
			},
			response: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Kerberos,
					SubGroup:  wire.KerberosLoginSuccessResponse,
				},
				Body: wire.SNAC_0x050C_0x0003_KerberosLoginSuccessResponse{
					RequestID: 4321,
				},
			},
			expectLogin:        true,
			expectSNACResponse: true,
			wantStatus:         http.StatusOK,
		},
		{
			name: "invalid request SNAC type",
			request: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMChannelMsgToHost,
				},
				Body: wire.SNAC_0x050C_0x0002_KerberosLoginRequest{
					RequestID: 4321,
				},
			},
			expectLogin:        false,
			expectSNACResponse: false,
			wantStatus:         http.StatusBadRequest,
		},
		{
			name: "login runtime error",
			request: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Kerberos,
					SubGroup:  wire.KerberosLoginRequest,
				},
				Body: wire.SNAC_0x050C_0x0002_KerberosLoginRequest{
					RequestID: 4321,
				},
			},
			response:           wire.SNACMessage{},
			responseErr:        io.EOF,
			expectLogin:        true,
			expectSNACResponse: false,
			wantStatus:         http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := newMockAuthService(t)
			if tt.expectLogin {
				mockAuth.EXPECT().
					KerberosLogin(mock.Anything, tt.request.Body, mock.Anything).
					Return(tt.response, tt.responseErr)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			srv := NewKerberosServer(config.Config{KerberosPort: "0"}, log, mockAuth)

			b := &bytes.Buffer{}
			assert.NoError(t, wire.MarshalBE(tt.request, b))

			req := httptest.NewRequest(http.MethodPost, "/", b)
			req.Header.Set("Content-Type", "application/x-snac")
			w := httptest.NewRecorder()

			srv.Handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)

			if tt.expectSNACResponse {
				respBytes, _ := io.ReadAll(w.Result().Body)
				reader := bytes.NewReader(respBytes)
				haveFrame := wire.SNACFrame{}
				assert.NoError(t, wire.UnmarshalBE(&haveFrame, reader))
				assert.Equal(t, tt.response.Frame, haveFrame)
				haveBody := wire.SNAC_0x050C_0x0003_KerberosLoginSuccessResponse{}
				assert.NoError(t, wire.UnmarshalBE(&haveBody, reader))
				assert.Equal(t, tt.response.Body, haveBody)
				assert.Equal(t, "application/x-snac", w.Result().Header.Get("Content-Type"))
			} else {
				assert.Equal(t, "text/plain; charset=utf-8", w.Result().Header.Get("Content-Type"))
			}
		})
	}
}
