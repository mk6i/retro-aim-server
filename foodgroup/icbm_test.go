package foodgroup

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestICBMService_ChannelMsgToHost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// senderSession is the session of the user sending the message
		senderSession *state.Session
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		inputSNAC wire.SNACMessage
		// expectOutput is the expected return SNAC value.
		expectOutput *wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// timeNow returns the current time
		timeNow func() time.Time
	}{
		{
			name:          "transmit message from sender to recipient, ack message back to sender",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10)),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     newTestSession("recipient-screen-name", sessOptWarning(20)),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICBM,
									SubGroup:  wire.ICBMChannelMsgToClient,
									RequestID: wire.ReqIDFromServer,
								},
								Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
									ChannelID:   wire.ICBMChannelIM,
									TLVUserInfo: newTestSession("sender-screen-name", sessOptWarning(10)).TLVUserInfo(),
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											{
												Tag:   wire.ICBMTLVData,
												Value: []byte{1, 2, 3, 4},
											},
											{
												Tag:   wire.ICBMTLVWantEvents,
												Value: []byte{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
							{
								Tag:   wire.ICBMTLVData,
								Value: []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMHostAck,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x0C_ICBMHostAck{
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: "recipient-screen-name",
				},
			},
		},
		{
			name:          "transmit message from sender to recipient, don't ack message back to sender",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10)),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     newTestSession("recipient-screen-name", sessOptWarning(20)),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICBM,
									SubGroup:  wire.ICBMChannelMsgToClient,
									RequestID: wire.ReqIDFromServer,
								},
								Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
									ChannelID:   wire.ICBMChannelIM,
									TLVUserInfo: newTestSession("sender-screen-name", sessOptWarning(10)).TLVUserInfo(),
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											{
												Tag:   wire.ICBMTLVData,
												Value: []byte{1, 2, 3, 4},
											},
											{
												Tag:   wire.ICBMTLVWantEvents,
												Value: []byte{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVData,
								Value: []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			expectOutput: nil,
		},
		{
			name:          "don't transmit message from sender to recipient because sender has blocked recipient",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10)),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      true,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeInLocalPermitDeny,
				},
			},
		},
		{
			name:          "don't transmit message from sender to recipient because recipient has blocked sender",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10)),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     true,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:          "don't transmit message from sender to recipient because recipient doesn't exist",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10)),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     nil,
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:          "send offline message to ICQ recipient",
			senderSession: newTestSession("11111111"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: "22222222",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.ICBMTLVRequestHostAck, []byte{}),
							wire.NewTLVBE(wire.ICBMTLVStore, []byte{}),
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
			},
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("11111111"),
							them: state.NewIdentScreenName("22222222"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("22222222"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("22222222"),
							result:     nil,
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{},
				},
				offlineMessageManagerParams: offlineMessageManagerParams{
					saveMessageParams: saveMessageParams{
						{
							offlineMessageIn: state.OfflineMessage{
								Message: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
									ChannelID:  wire.ICBMChannelIM,
									ScreenName: "22222222",
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.ICBMTLVRequestHostAck, []byte{}),
											wire.NewTLVBE(wire.ICBMTLVStore, []byte{}),
										},
									},
								},
								Recipient: state.NewIdentScreenName("22222222"),
								Sender:    state.NewIdentScreenName("11111111"),
								Sent:      time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC),
							},
						},
					},
				},
			},
		},
		{
			name: "send rendezvous request for file transfer, expect IP TLV override",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10),
				sessRemoteAddr(netip.AddrPortFrom(netip.MustParseAddr("129.168.0.1"), 0))),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     newTestSession("recipient-screen-name", sessOptWarning(20)),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICBM,
									SubGroup:  wire.ICBMChannelMsgToClient,
									RequestID: wire.ReqIDFromServer,
								},
								Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
									ChannelID:   wire.ICBMChannelRendezvous,
									TLVUserInfo: newTestSession("sender-screen-name", sessOptWarning(10)).TLVUserInfo(),
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.ICBMTLVData, wire.ICBMCh2Fragment{
												Type:       wire.ICBMRdvMessagePropose,
												Capability: wire.CapFileTransfer,
												TLVRestBlock: wire.TLVRestBlock{
													TLVList: wire.TLVList{
														wire.NewTLVBE(wire.ICBMRdvTLVTagsPort, uint16(4000)),
														wire.NewTLVBE(wire.ICBMRdvTLVTagsRequesterIP, net.ParseIP("129.168.0.1").To4()),
														wire.NewTLVBE(wire.ICBMRdvTLVTagsVerifiedIP, net.ParseIP("129.168.0.1").To4()),
													},
												},
											}),
											wire.NewTLVBE(wire.ICBMTLVWantEvents, []byte{}),
										},
									},
								},
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelRendezvous,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.ICBMTLVData, wire.ICBMCh2Fragment{
								Type:       wire.ICBMRdvMessagePropose,
								Capability: wire.CapFileTransfer,
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.ICBMRdvTLVTagsPort, uint16(4000)),
										wire.NewTLVBE(wire.ICBMRdvTLVTagsRequesterIP, net.ParseIP("127.0.0.1").To4()),
									},
								},
							}),
						},
					},
				},
			},
			expectOutput: nil,
		},
		{
			name: "send rendezvous rejection for file transfer, expect no IP TLV override",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10),
				sessRemoteAddr(netip.AddrPortFrom(netip.MustParseAddr("129.168.0.1"), 0))),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     newTestSession("recipient-screen-name", sessOptWarning(20)),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICBM,
									SubGroup:  wire.ICBMChannelMsgToClient,
									RequestID: wire.ReqIDFromServer,
								},
								Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
									ChannelID:   wire.ICBMChannelRendezvous,
									TLVUserInfo: newTestSession("sender-screen-name", sessOptWarning(10)).TLVUserInfo(),
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.ICBMTLVData, wire.ICBMCh2Fragment{
												Type:       wire.ICBMRdvMessageCancel,
												Capability: wire.CapFileTransfer,
											}),
											wire.NewTLVBE(wire.ICBMTLVWantEvents, []byte{}),
										},
									},
								},
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelRendezvous,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.ICBMTLVData, wire.ICBMCh2Fragment{
								Type:       wire.ICBMRdvMessageCancel,
								Capability: wire.CapFileTransfer,
							}),
						},
					},
				},
			},
			expectOutput: nil,
		},
		{
			name:          "send rendezvous request for file transfer without IP in session, expect no IP TLV override",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(10)),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     newTestSession("recipient-screen-name", sessOptWarning(20)),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICBM,
									SubGroup:  wire.ICBMChannelMsgToClient,
									RequestID: wire.ReqIDFromServer,
								},
								Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
									ChannelID:   wire.ICBMChannelRendezvous,
									TLVUserInfo: newTestSession("sender-screen-name", sessOptWarning(10)).TLVUserInfo(),
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.ICBMTLVData, wire.ICBMCh2Fragment{
												Type:       wire.ICBMRdvMessagePropose,
												Capability: wire.CapFileTransfer,
												TLVRestBlock: wire.TLVRestBlock{
													TLVList: wire.TLVList{
														wire.NewTLVBE(wire.ICBMRdvTLVTagsPort, uint16(4000)),
														wire.NewTLVBE(wire.ICBMRdvTLVTagsRequesterIP, net.ParseIP("127.0.0.1").To4()),
													},
												},
											}),
											wire.NewTLVBE(wire.ICBMTLVWantEvents, []byte{}),
										},
									},
								},
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ChannelID:  wire.ICBMChannelRendezvous,
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.ICBMTLVData, wire.ICBMCh2Fragment{
								Type:       wire.ICBMRdvMessagePropose,
								Capability: wire.CapFileTransfer,
								TLVRestBlock: wire.TLVRestBlock{
									TLVList: wire.TLVList{
										wire.NewTLVBE(wire.ICBMRdvTLVTagsPort, uint16(4000)),
										wire.NewTLVBE(wire.ICBMRdvTLVTagsRequesterIP, net.ParseIP("127.0.0.1").To4()),
									},
								},
							}),
						},
					},
				},
			},
			expectOutput: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buddyListRetriever := newMockBuddyListRetriever(t)
			for _, item := range tc.mockParams.buddyListRetrieverParams.relationshipParams {
				buddyListRetriever.EXPECT().
					Relationship(item.me, item.them).
					Return(item.result, item.err)
			}
			sessionRetriever := newMockSessionRetriever(t)
			for _, item := range tc.mockParams.sessionRetrieverParams.retrieveSessionParams {
				sessionRetriever.EXPECT().
					RetrieveSession(item.screenName).
					Return(item.result)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, item := range tc.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, item.screenName, item.message)
			}
			offlineMessageManager := newMockOfflineMessageManager(t)
			for _, params := range tc.mockParams.saveMessageParams {
				offlineMessageManager.EXPECT().
					SaveMessage(params.offlineMessageIn).
					Return(params.err)
			}

			svc := ICBMService{
				buddyListRetriever:  buddyListRetriever,
				messageRelayer:      messageRelayer,
				offlineMessageSaver: offlineMessageManager,
				sessionRetriever:    sessionRetriever,
				timeNow:             tc.timeNow,
			}

			outputSNAC, err := svc.ChannelMsgToHost(nil, tc.senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x04_0x06_ICBMChannelMsgToHost))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestICBMService_ClientEvent(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// senderScreenName is the screen name of the user sending the event
		senderScreenName state.DisplayScreenName
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:             "transmit message from sender to recipient",
			senderScreenName: "sender-screen-name",
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICBM,
									SubGroup:  wire.ICBMClientEvent,
									RequestID: 1234,
								},
								Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
									Cookie:     12345678,
									ChannelID:  42,
									ScreenName: "sender-screen-name",
									Event:      12,
								},
							},
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
					Cookie:     12345678,
					ChannelID:  42,
					ScreenName: "recipient-screen-name",
					Event:      12,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because sender has blocked recipient",
			senderScreenName: "sender-screen-name",
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      true,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
					ScreenName: "recipient-screen-name",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buddyListRetriever := newMockBuddyListRetriever(t)
			for _, item := range tc.mockParams.buddyListRetrieverParams.relationshipParams {
				buddyListRetriever.EXPECT().
					Relationship(item.me, item.them).
					Return(item.result, item.err)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, item := range tc.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, item.screenName, item.message)
			}

			senderSession := newTestSession(tc.senderScreenName)
			svc := ICBMService{
				buddyListRetriever: buddyListRetriever,
				messageRelayer:     messageRelayer,
			}
			assert.NoError(t, svc.ClientEvent(nil, senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x04_0x14_ICBMClientEvent)))
		})
	}
}

func TestICBMService_EvilRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// senderScreenName is the session of the user sending the EvilRequest
		senderSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:          "transmit anonymous warning from sender to recipient",
			senderSession: newTestSession("sender-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     1, // make it anonymous
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMEvilReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 30,
					UpdatedEvilValue: 30,
				},
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
						},
					},
				},
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     newTestSession("recipient-screen-name", sessOptCannedSignonTime),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.OService,
									SubGroup:  wire.OServiceEvilNotification,
								},
								Body: wire.SNAC_0x01_0x10_OServiceEvilNotification{
									NewEvil: evilDeltaAnon,
								},
							},
						},
					},
				},
			},
		},
		{
			name:          "transmit non-anonymous warning from sender to recipient",
			senderSession: newTestSession("sender-screen-name", sessOptWarning(110)),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMEvilReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 100,
					UpdatedEvilValue: 100,
				},
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
						},
					},
				},
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     newTestSession("recipient-screen-name", sessOptCannedSignonTime),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.OService,
									SubGroup:  wire.OServiceEvilNotification,
								},
								Body: wire.SNAC_0x01_0x10_OServiceEvilNotification{
									NewEvil: evilDelta,
									Snitcher: &struct {
										wire.TLVUserInfo
									}{
										wire.TLVUserInfo{
											ScreenName:   "sender-screen-name",
											WarningLevel: 110,
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
			name:          "don't transmit non-anonymous warning from sender to recipient because sender has blocked recipient",
			senderSession: newTestSession("sender-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      true,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
			},
		},
		{
			name:          "don't transmit non-anonymous warning from sender to recipient because recipient has blocked sender",
			senderSession: newTestSession("sender-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     true,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
			},
		},
		{
			name:          "don't let users warn themselves",
			senderSession: newTestSession("sender-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "sender-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			},
		},
		{
			name:          "don't transmit non-anonymous warning from sender to recipient because recipient is offline",
			senderSession: newTestSession("sender-screen-name"),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     nil,
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:          "don't transmit anonymous warning from sender to recipient because recipient is offline",
			senderSession: newTestSession("sender-screen-name"),
			mockParams: mockParams{
				buddyListRetrieverParams: buddyListRetrieverParams{
					relationshipParams: relationshipParams{
						{
							me:   state.NewIdentScreenName("sender-screen-name"),
							them: state.NewIdentScreenName("recipient-screen-name"),
							result: state.Relationship{
								User:          state.NewIdentScreenName("recipient-screen-name"),
								BlocksYou:     false,
								YouBlock:      false,
								IsOnTheirList: false,
								IsOnYourList:  false,
							},
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
							result:     nil,
						},
					},
				},
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     1, // make it anonymous
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockBuddyBroadcaster := newMockbuddyBroadcaster(t)
			for _, item := range tc.mockParams.broadcastBuddyArrivedParams {
				mockBuddyBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, matchSession(item.screenName)).
					Return(item.err)
			}
			buddyListRetriever := newMockBuddyListRetriever(t)
			for _, item := range tc.mockParams.buddyListRetrieverParams.relationshipParams {
				buddyListRetriever.EXPECT().
					Relationship(item.me, item.them).
					Return(item.result, item.err)
			}
			sessionRetriever := newMockSessionRetriever(t)
			for _, item := range tc.mockParams.sessionRetrieverParams.retrieveSessionParams {
				sessionRetriever.EXPECT().
					RetrieveSession(item.screenName).
					Return(item.result)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, item := range tc.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, item.screenName, item.message)
			}
			offlineMessageManager := newMockOfflineMessageManager(t)
			for _, params := range tc.mockParams.saveMessageParams {
				offlineMessageManager.EXPECT().
					SaveMessage(params.offlineMessageIn).
					Return(params.err)
			}

			svc := ICBMService{
				buddyBroadcaster:    mockBuddyBroadcaster,
				buddyListRetriever:  buddyListRetriever,
				messageRelayer:      messageRelayer,
				offlineMessageSaver: offlineMessageManager,
				sessionRetriever:    sessionRetriever,
			}

			outputSNAC, err := svc.EvilRequest(nil, tc.senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x04_0x08_ICBMEvilRequest))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestICBMService_ParameterQuery(t *testing.T) {
	svc := NewICBMService(nil, nil, nil, nil)

	have := svc.ParameterQuery(nil, wire.SNACFrame{RequestID: 1234})
	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMParameterReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x04_0x05_ICBMParameterReply{
			MaxSlots:             100,
			ICBMFlags:            3,
			MaxIncomingICBMLen:   512,
			MaxSourceEvil:        999,
			MaxDestinationEvil:   999,
			MinInterICBMInterval: 0,
		},
	}

	assert.Equal(t, want, have)
}

func TestICBMService_ClientErr(t *testing.T) {
	sess := newTestSession("theScreenName")

	inBody := wire.SNAC_0x04_0x0B_ICBMClientErr{
		Cookie:     1234,
		ChannelID:  wire.ICBMChannelMIME,
		ScreenName: "recipientScreenName",
		Code:       10,
		ErrInfo:    []byte{1, 2, 3, 4},
	}

	expect := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMClientErr,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x04_0x0B_ICBMClientErr{
			Cookie:     inBody.Cookie,
			ChannelID:  inBody.ChannelID,
			ScreenName: sess.DisplayScreenName().String(),
			Code:       inBody.Code,
			ErrInfo:    inBody.ErrInfo,
		},
	}

	messageRelayer := newMockMessageRelayer(t)
	messageRelayer.EXPECT().
		RelayToScreenName(mock.Anything, state.NewIdentScreenName("recipientScreenName"), expect)

	svc := NewICBMService(messageRelayer, nil, nil, nil)

	err := svc.ClientErr(nil, sess, wire.SNACFrame{RequestID: 1234}, inBody)
	assert.NoError(t, err)
}
