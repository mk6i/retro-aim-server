package toc

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestOSCARProxy_SendIM(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully send instant message",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_send_im chattingChuck "hello world!"`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					channelMsgToHostParamsICBM: channelMsgToHostParamsICBM{
						{
							sender: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
								ChannelID:  wire.ICBMChannelIM,
								ScreenName: "chattingChuck",
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.ICBMTLVAOLIMData, []wire.ICBMCh1Fragment{
											{
												ID:      5,
												Version: 1,
												Payload: []byte{1, 1, 2},
											},
											{
												ID:      1,
												Version: 1,
												Payload: []byte{
													0x00, 0x00,
													0x00, 0x00,
													'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '!',
												},
											},
										}),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "successfully auto-reply send instant message",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_send_im chattingChuck "hello world!" auto`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					channelMsgToHostParamsICBM: channelMsgToHostParamsICBM{
						{
							sender: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
								ChannelID:  wire.ICBMChannelIM,
								ScreenName: "chattingChuck",
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.ICBMTLVAOLIMData, []wire.ICBMCh1Fragment{
											{
												ID:      5,
												Version: 1,
												Payload: []byte{1, 1, 2},
											},
											{
												ID:      1,
												Version: 1,
												Payload: []byte{
													0x00, 0x00,
													0x00, 0x00,
													'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '!',
												},
											},
										}),
										wire.NewTLVBE(wire.ICBMTLVAutoResponse, []byte{}),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "send instant message, receive error from ICBM service",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_send_im chattingChuck "hello world!"`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					channelMsgToHostParamsICBM: channelMsgToHostParamsICBM{
						{
							sender: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
								ChannelID:  wire.ICBMChannelIM,
								ScreenName: "chattingChuck",
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.ICBMTLVAOLIMData, []wire.ICBMCh1Fragment{
											{
												ID:      5,
												Version: 1,
												Payload: []byte{1, 1, 2},
											},
											{
												ID:      1,
												Version: 1,
												Payload: []byte{
													0x00, 0x00,
													0x00, 0x00,
													'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '!',
												},
											},
										}),
									},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_send_im`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			icbmSvc := newMockICBMService(t)
			for _, params := range tc.mockParams.channelMsgToHostParamsICBM {
				icbmSvc.EXPECT().
					ChannelMsgToHost(ctx, matchSession(params.sender), params.inFrame, params.inBody).
					Return(params.result, params.err)
			}

			svc := OSCARProxy{
				Logger:      slog.Default(),
				ICBMService: icbmSvc,
			}
			msg := svc.SendIM(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_AddBuddy(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully add buddies",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_add_buddy friend1 friend2 friend3"),
			mockParams: mockParams{
				buddyParams: buddyParams{
					addBuddiesParams: addBuddiesParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x03_0x04_BuddyAddBuddies{
								Buddies: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
									{ScreenName: "friend2"},
									{ScreenName: "friend3"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "add buddies, receive error from buddy service",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_add_buddy friend1"),
			mockParams: mockParams{
				buddyParams: buddyParams{
					addBuddiesParams: addBuddiesParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x03_0x04_BuddyAddBuddies{
								Buddies: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_add_buddy_bad`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			buddySvc := newMockBuddyService(t)
			for _, params := range tc.mockParams.addBuddiesParams {
				buddySvc.EXPECT().
					AddBuddies(ctx, matchSession(params.me), params.inBody).
					Return(params.err)
			}

			svc := OSCARProxy{
				Logger:       slog.Default(),
				BuddyService: buddySvc,
			}
			msg := svc.AddBuddy(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_RemoveBuddy(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully remove buddies",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_remove_buddy friend1 friend2 friend3"),
			mockParams: mockParams{
				buddyParams: buddyParams{
					delBuddiesParams: delBuddiesParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x03_0x05_BuddyDelBuddies{
								Buddies: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
									{ScreenName: "friend2"},
									{ScreenName: "friend3"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "remove buddies, receive error from buddy service",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_remove_buddy friend1"),
			mockParams: mockParams{
				buddyParams: buddyParams{
					delBuddiesParams: delBuddiesParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x03_0x05_BuddyDelBuddies{
								Buddies: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_remove_buddy_bad`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			buddySvc := newMockBuddyService(t)
			for _, params := range tc.mockParams.delBuddiesParams {
				buddySvc.EXPECT().
					DelBuddies(ctx, matchSession(params.me), params.inBody).
					Return(params.err)
			}

			svc := OSCARProxy{
				Logger:       slog.Default(),
				BuddyService: buddySvc,
			}
			msg := svc.RemoveBuddy(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_AddPermit(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully permit buddies",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_add_permit friend1 friend2 friend3"),
			mockParams: mockParams{
				permitDenyParams: permitDenyParams{
					addPermListEntriesParams: addPermListEntriesParams{
						{
							me: state.NewIdentScreenName("me"),
							body: wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{
								Users: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
									{ScreenName: "friend2"},
									{ScreenName: "friend3"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "permit buddies, receive error from buddy service",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_add_permit friend1"),
			mockParams: mockParams{
				permitDenyParams: permitDenyParams{
					addPermListEntriesParams: addPermListEntriesParams{
						{
							me: state.NewIdentScreenName("me"),
							body: wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{
								Users: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_remove_buddy_bad`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			pdSvc := newMockPermitDenyService(t)
			for _, params := range tc.mockParams.addPermListEntriesParams {
				pdSvc.EXPECT().
					AddPermListEntries(ctx, matchSession(params.me), params.body).
					Return(params.err)
			}

			svc := OSCARProxy{
				Logger:            slog.Default(),
				PermitDenyService: pdSvc,
			}
			msg := svc.AddPermit(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_AddDeny(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully deny buddies",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_add_deny friend1 friend2 friend3"),
			mockParams: mockParams{
				permitDenyParams: permitDenyParams{
					addDenyListEntriesParams: addDenyListEntriesParams{
						{
							me: state.NewIdentScreenName("me"),
							body: wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{
								Users: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
									{ScreenName: "friend2"},
									{ScreenName: "friend3"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "deny buddies, receive error from buddy service",
			me:       newTestSession("me"),
			givenCmd: []byte("toc_add_deny friend1"),
			mockParams: mockParams{
				permitDenyParams: permitDenyParams{
					addDenyListEntriesParams: addDenyListEntriesParams{
						{
							me: state.NewIdentScreenName("me"),
							body: wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{
								Users: []struct {
									ScreenName string `oscar:"len_prefix=uint8"`
								}{
									{ScreenName: "friend1"},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_add_deny_bad`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			pdSvc := newMockPermitDenyService(t)
			for _, params := range tc.mockParams.addDenyListEntriesParams {
				pdSvc.EXPECT().
					AddDenyListEntries(ctx, matchSession(params.me), params.body).
					Return(params.err)
			}

			svc := OSCARProxy{
				Logger:            slog.Default(),
				PermitDenyService: pdSvc,
			}
			msg := svc.AddDeny(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_SetAway(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully set away with message",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_set_away "I'm away from my computer right now."`),
			mockParams: mockParams{
				locateParams: locateParams{
					setInfoParams: setInfoParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x04_LocateSetInfo{
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.LocateTLVTagsInfoUnavailableData, "I'm away from my computer right now."),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "successfully set away without message",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_set_away`),
			mockParams: mockParams{
				locateParams: locateParams{
					setInfoParams: setInfoParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x04_LocateSetInfo{
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.LocateTLVTagsInfoUnavailableData, ""),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "set away message, receive error from locate service",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_set_away "I'm away from my computer right now."`),
			mockParams: mockParams{
				locateParams: locateParams{
					setInfoParams: setInfoParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x04_LocateSetInfo{
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.LocateTLVTagsInfoUnavailableData, "I'm away from my computer right now."),
									},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_set_away_bad`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			locateSvc := newMockLocateService(t)
			for _, params := range tc.mockParams.setInfoParams {
				locateSvc.EXPECT().
					SetInfo(ctx, matchSession(params.me), params.inBody).
					Return(params.err)
			}

			svc := OSCARProxy{
				Logger:        slog.Default(),
				LocateService: locateSvc,
			}
			msg := svc.SetAway(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_SetCaps(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully set capabilities",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_set_caps 09460000-4C7F-11D1-8222-444553540000 09460001-4C7F-11D1-8222-444553540000`),
			mockParams: mockParams{
				locateParams: locateParams{
					setInfoParams: setInfoParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x04_LocateSetInfo{
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.LocateTLVTagsInfoCapabilities, []uuid.UUID{
											uuid.MustParse("09460000-4C7F-11D1-8222-444553540000"),
											uuid.MustParse("09460001-4C7F-11D1-8222-444553540000"),
											capChat,
										}),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "set capability, receive error from locate service",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_set_caps 09460000-4C7F-11D1-8222-444553540000`),
			mockParams: mockParams{
				locateParams: locateParams{
					setInfoParams: setInfoParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x02_0x04_LocateSetInfo{
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.LocateTLVTagsInfoCapabilities, []uuid.UUID{
											uuid.MustParse("09460000-4C7F-11D1-8222-444553540000"),
											capChat,
										}),
									},
								},
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "set malformed capability UUID",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_set_caps 09460000-`),
			wantMsg:  cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_set_caps_bad`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			locateSvc := newMockLocateService(t)
			for _, params := range tc.mockParams.setInfoParams {
				locateSvc.EXPECT().
					SetInfo(ctx, matchSession(params.me), params.inBody).
					Return(params.err)
			}

			svc := OSCARProxy{
				Logger:        slog.Default(),
				LocateService: locateSvc,
			}
			msg := svc.SetCaps(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_Evil(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully warn normally",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_evil them norm`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					evilRequestParams: evilRequestParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x08_ICBMEvilRequest{
								SendAs:     0,
								ScreenName: "them",
							},
							msg: wire.SNACMessage{
								Body: wire.SNAC_0x04_0x09_ICBMEvilReply{},
							},
						},
					},
				},
			},
		},
		{
			name:     "successfully warn anonymously",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_evil them anon`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					evilRequestParams: evilRequestParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x08_ICBMEvilRequest{
								SendAs:     1,
								ScreenName: "them",
							},
							msg: wire.SNACMessage{
								Body: wire.SNAC_0x04_0x09_ICBMEvilReply{},
							},
						},
					},
				},
			},
		},
		{
			name:     "warn, receive error from ICBM service",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_evil them anon`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					evilRequestParams: evilRequestParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x08_ICBMEvilRequest{
								SendAs:     1,
								ScreenName: "them",
							},
							err: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "warn, receive snac err",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_evil them anon`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					evilRequestParams: evilRequestParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x08_ICBMEvilRequest{
								SendAs:     1,
								ScreenName: "them",
							},
							msg: wire.SNACMessage{
								Body: wire.SNACError{},
							},
						},
					},
				},
			},
		},
		{
			name:     "warn, ICBM svc returns unexpected snac type",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_evil them anon`),
			mockParams: mockParams{
				icbmParams: icbmParams{
					evilRequestParams: evilRequestParams{
						{
							me: state.NewIdentScreenName("me"),
							inBody: wire.SNAC_0x04_0x08_ICBMEvilRequest{
								SendAs:     1,
								ScreenName: "them",
							},
							msg: wire.SNACMessage{
								Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{},
							},
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "warn with incorrect type",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_evil them blah`),
			wantMsg:  cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_evil`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			icbmSvc := newMockICBMService(t)
			for _, params := range tc.mockParams.evilRequestParams {
				icbmSvc.EXPECT().
					EvilRequest(ctx, matchSession(params.me), wire.SNACFrame{}, params.inBody).
					Return(params.msg, params.err)
			}

			svc := OSCARProxy{
				Logger:      slog.Default(),
				ICBMService: icbmSvc,
			}
			msg := svc.Evil(ctx, tc.me, tc.givenCmd)

			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}

func TestOSCARProxy_GetInfoURL(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// me is the TOC user session
		me *state.Session
		// givenCmd is the TOC command
		givenCmd []byte
		// wantMsg is the expected TOC response
		wantMsg []byte
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:     "successfully request user info",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_get_info them`),
			mockParams: mockParams{
				cookieBakerParams: cookieBakerParams{
					issueParams: issueParams{
						{
							data:       []byte("me"),
							returnData: []byte("monster"),
						},
					},
				},
			},
			wantMsg: []byte("GOTO_URL:profile:info?cookie=bW9uc3Rlcg%253D%253D&from=me&user=them"),
		},
		{
			name:     "request user info, get cookie issue error",
			me:       newTestSession("me"),
			givenCmd: []byte(`toc_get_info them`),
			mockParams: mockParams{
				cookieBakerParams: cookieBakerParams{
					issueParams: issueParams{
						{
							data:      []byte("me"),
							returnErr: io.EOF,
						},
					},
				},
			},
			wantMsg: cmdInternalSvcErr,
		},
		{
			name:     "bad command",
			givenCmd: []byte(`toc_get_info`),
			wantMsg:  cmdInternalSvcErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			cookieBaker := newMockCookieBaker(t)
			for _, params := range tc.mockParams.issueParams {
				cookieBaker.EXPECT().
					Issue(params.data).
					Return(params.returnData, params.returnErr)
			}

			svc := OSCARProxy{
				Logger:      slog.Default(),
				CookieBaker: cookieBaker,
			}
			msg := svc.GetInfoURL(ctx, tc.me, tc.givenCmd)

			fmt.Println(string(msg))
			assert.Equal(t, tc.wantMsg, msg)
		})
	}
}
