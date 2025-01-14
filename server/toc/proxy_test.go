package toc

import (
	"context"
	"io"
	"log/slog"
	"testing"

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
