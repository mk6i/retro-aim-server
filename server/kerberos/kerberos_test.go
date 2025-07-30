package kerberos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
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
					KerberosLogin(mock.Anything, tt.request.Body, mock.Anything, "localhost:5190").
					Return(tt.response, tt.responseErr)
			}

			listenCfg := []config.Listener{
				{
					KerberosListenAddress: ":1088",
					BOSAdvertisedHost:     "localhost:5190",
				},
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			srv := NewKerberosServer(listenCfg, log, mockAuth)

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, srv.ListenAndServe())
			}()

			b := &bytes.Buffer{}
			assert.NoError(t, wire.MarshalBE(tt.request, b))

			req := httptest.NewRequest(http.MethodPost, "/", b)
			req.Header.Set("Content-Type", "application/x-snac")

			resp, err := http.Post(fmt.Sprintf("http://localhost:%s", "1088"), "application/x-snac", b)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.expectSNACResponse {
				respBytes, _ := io.ReadAll(resp.Body)
				reader := bytes.NewReader(respBytes)
				haveFrame := wire.SNACFrame{}
				assert.NoError(t, wire.UnmarshalBE(&haveFrame, reader))
				assert.Equal(t, tt.response.Frame, haveFrame)
				haveBody := wire.SNAC_0x050C_0x0003_KerberosLoginSuccessResponse{}
				assert.NoError(t, wire.UnmarshalBE(&haveBody, reader))
				assert.Equal(t, tt.response.Body, haveBody)
				assert.Equal(t, "application/x-snac", resp.Header.Get("Content-Type"))
			} else {
				assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
			}

			assert.NoError(t, srv.Shutdown(context.Background()))

			wg.Wait()
		})
	}
}
