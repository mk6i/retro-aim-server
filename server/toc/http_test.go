package toc

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOSCARProxy_NewServeMux(t *testing.T) {
	cookieBaker, err := state.NewHMACCookieBaker()
	assert.NoError(t, err)

	cookie, err := bakeCookie(cookieBaker, state.NewIdentScreenName("me"))
	assert.NoError(t, err)

	cases := []struct {
		// name is the unit test name
		name string
		// path is the HTTP request path
		path string
		// expectedStatus is the expected HTTP response code
		expectedStatus int
		// expectedBody is the expected body payload
		expectedBody string
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:           "Successfully retrieve profile",
			path:           "/info?from=me&user=them&cookie=" + cookie,
			expectedStatus: http.StatusOK,
			expectedBody:   "My profile!",
			mockParams: mockParams{
				locateParams: locateParams{
					userInfoQueryParams: userInfoQueryParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x05_LocateUserInfoQuery{
								Type:       uint16(wire.LocateTypeSig),
								ScreenName: "them",
							},
							msg: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Locate,
									SubGroup:  wire.LocateUserInfoReply,
								},
								Body: wire.SNAC_0x02_0x06_LocateUserInfoReply{
									TLVUserInfo: newTestSession("them").TLVUserInfo(),
									LocateInfo: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.LocateTLVTagsInfoSigData, "<HTML><BODY>My profile!</BODY></HTML>"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:           "Retrieve profile with missing `from` query param",
			path:           "/info?user=them&cookie=" + cookie,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "required `from` param is missing`",
		},
		{
			name:           "Retrieve profile with missing `from` query param",
			path:           "/info?from=me&cookie=" + cookie,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "required `user` param is missing",
		},
		{
			name:           "Retrieve profile, receive error from locate svc",
			path:           "/info?from=me&user=them&cookie=" + cookie,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
			mockParams: mockParams{
				locateParams: locateParams{
					userInfoQueryParams: userInfoQueryParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x05_LocateUserInfoQuery{
								Type:       uint16(wire.LocateTypeSig),
								ScreenName: "them",
							},
							msg: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Locate,
									SubGroup:  wire.LocateUserInfoReply,
								},
								Body: wire.SNAC_0x02_0x06_LocateUserInfoReply{
									TLVUserInfo: newTestSession("them").TLVUserInfo(),
									LocateInfo: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.LocateTLVTagsInfoSigData, "<HTML><BODY>My profile!</BODY></HTML>"),
										},
									},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
		},
		{
			name:           "Retrieve profile, user offline",
			path:           "/info?from=me&user=them&cookie=" + cookie,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "user is unavailable",
			mockParams: mockParams{
				locateParams: locateParams{
					userInfoQueryParams: userInfoQueryParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x05_LocateUserInfoQuery{
								Type:       uint16(wire.LocateTypeSig),
								ScreenName: "them",
							},
							msg: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Locate,
									SubGroup:  wire.LocateUserInfoReply,
								},
								Body: wire.SNACError{
									Code: wire.ErrorCodeNotLoggedOn,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			locateSvc := newMockLocateService(t)
			for _, params := range tc.mockParams.userInfoQueryParams {
				locateSvc.EXPECT().
					UserInfoQuery(mock.Anything, matchSession(params.me), wire.SNACFrame{}, params.inBody).
					Return(params.msg, params.err)
			}

			svc := OSCARProxy{
				CookieBaker:   cookieBaker,
				LocateService: locateSvc,
				Logger:        slog.Default(),
			}

			req, err := http.NewRequest(http.MethodGet, tc.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()

			svc.NewServeMux().ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}
