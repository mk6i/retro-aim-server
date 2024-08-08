package foodgroup

import (
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestICQService_UpdateAffiliations(t *testing.T) {
	tests := []struct {
		name       string
		seq        uint16
		sess       *state.Session
		req        wire.ICQAffiliations
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "update affiliations - happy path",
			seq:  1,
			sess: newTestSession("100003", sessOptUIN(100003)),
			req: wire.ICQAffiliations{
				PastAffiliations: []struct {
					Code    uint16
					Keyword string `oscar:"len_prefix=uint16,nullterm"`
				}{
					{
						Code:    1,
						Keyword: "kw1",
					},
					{
						Code:    2,
						Keyword: "kw2",
					},
					{
						Code:    3,
						Keyword: "kw3",
					},
				},
				Affiliations: []struct {
					Code    uint16
					Keyword string `oscar:"len_prefix=uint16,nullterm"`
				}{
					{
						Code:    4,
						Keyword: "kw4",
					},
					{
						Code:    5,
						Keyword: "kw5",
					},
					{
						Code:    6,
						Keyword: "kw6",
					},
				},
			},
			mockParams: mockParams{
				icqUserUpdaterParams: icqUserUpdaterParams{
					setAffiliationsParams: setAffiliationsParams{
						{
							name: state.NewIdentScreenName("100003"),
							data: state.ICQAffiliations{
								PastAffiliations1Code:    1,
								PastAffiliations1Keyword: "kw1",
								PastAffiliations2Code:    2,
								PastAffiliations2Keyword: "kw2",
								PastAffiliations3Code:    3,
								PastAffiliations3Keyword: "kw3",
								Affiliations1Code:        4,
								Affiliations1Keyword:     "kw4",
								Affiliations2Code:        5,
								Affiliations2Keyword:     "kw5",
								Affiliations3Code:        6,
								Affiliations3Keyword:     "kw6",
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("100003"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMoreUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     100003,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													ReqSubType: wire.ICQDBQueryMetaReplySetAffiliations,
													Success:    wire.ICQStatusCodeOK,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userUpdater := newMockICQUserUpdater(t)
			for _, params := range tt.mockParams.setAffiliationsParams {
				userUpdater.EXPECT().
					SetAffiliations(params.name, params.data).
					Return(params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := ICQService{
				userUpdater:    userUpdater,
				messageRelayer: messageRelayer,
			}
			err := s.UpdateAffiliations(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_UpdateInterests(t *testing.T) {
	tests := []struct {
		name       string
		seq        uint16
		sess       *state.Session
		req        wire.ICQInterests
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "update interests - happy path",
			seq:  1,
			sess: newTestSession("100003", sessOptUIN(100003)),
			req: wire.ICQInterests{
				Interests: []struct {
					Code    uint16
					Keyword string `oscar:"len_prefix=uint16,nullterm"`
				}{
					{
						Code:    1,
						Keyword: "kw1",
					},
					{
						Code:    2,
						Keyword: "kw2",
					},
					{
						Code:    3,
						Keyword: "kw3",
					},
					{
						Code:    4,
						Keyword: "kw4",
					},
				},
			},
			mockParams: mockParams{
				icqUserUpdaterParams: icqUserUpdaterParams{
					setInterestsParams: setInterestsParams{
						{
							name: state.NewIdentScreenName("100003"),
							data: state.ICQInterests{
								Interest1Code:    1,
								Interest1Keyword: "kw1",
								Interest2Code:    2,
								Interest2Keyword: "kw2",
								Interest3Code:    3,
								Interest3Keyword: "kw3",
								Interest4Code:    4,
								Interest4Keyword: "kw4",
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("100003"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMoreUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     100003,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													ReqSubType: wire.ICQDBQueryMetaReplySetInterests,
													Success:    wire.ICQStatusCodeOK,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userUpdater := newMockICQUserUpdater(t)
			for _, params := range tt.mockParams.setInterestsParams {
				userUpdater.EXPECT().
					SetInterests(params.name, params.data).
					Return(params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := ICQService{
				userUpdater:    userUpdater,
				messageRelayer: messageRelayer,
			}
			err := s.UpdateInterests(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_UpdateUserNotes(t *testing.T) {
	tests := []struct {
		name       string
		seq        uint16
		sess       *state.Session
		req        wire.ICQNotes
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "update user notes - happy path",
			seq:  1,
			sess: newTestSession("100003", sessOptUIN(100003)),
			req: wire.ICQNotes{
				Notes: "here is my note",
			},
			mockParams: mockParams{
				icqUserUpdaterParams: icqUserUpdaterParams{
					setUserNotesParams: setUserNotesParams{
						{
							name: state.NewIdentScreenName("100003"),
							data: state.ICQUserNotes{
								Notes: "here is my note",
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("100003"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMoreUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     100003,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													ReqSubType: wire.ICQDBQueryMetaReplySetNotes,
													Success:    wire.ICQStatusCodeOK,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userUpdater := newMockICQUserUpdater(t)
			for _, params := range tt.mockParams.setUserNotesParams {
				userUpdater.EXPECT().
					SetUserNotes(params.name, params.data).
					Return(params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := ICQService{
				userUpdater:    userUpdater,
				messageRelayer: messageRelayer,
			}
			err := s.UpdateUserNotes(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_UpdateBasicInfo(t *testing.T) {
	tests := []struct {
		name       string
		seq        uint16
		sess       *state.Session
		req        wire.ICQUserInfoBasic
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "update basic info - happy path",
			seq:  1,
			sess: newTestSession("100003", sessOptUIN(100003)),
			req: wire.ICQUserInfoBasic{
				CellPhone:    "123-456-7890",
				CountryCode:  1,
				EmailAddress: "test@example.com",
				FirstName:    "John",
				GMTOffset:    5,
				HomeAddress:  "123 Main St",
				HomeCity:     "Anytown",
				HomeFax:      "098-765-4321",
				HomePhone:    "111-222-3333",
				HomeState:    "CA",
				LastName:     "Doe",
				Nickname:     "Johnny",
				PublishEmail: wire.ICQUserFlagPublishEmailYes,
				ZipCode:      "12345",
			},
			mockParams: mockParams{
				icqUserUpdaterParams: icqUserUpdaterParams{
					setBasicInfoParams: setBasicInfoParams{
						{
							name: state.NewIdentScreenName("100003"),
							data: state.ICQUserInfoBasic{
								CellPhone:    "123-456-7890",
								CountryCode:  1,
								EmailAddress: "test@example.com",
								FirstName:    "John",
								GMTOffset:    5,
								HomeAddress:  "123 Main St",
								HomeCity:     "Anytown",
								HomeFax:      "098-765-4321",
								HomePhone:    "111-222-3333",
								HomeState:    "CA",
								LastName:     "Doe",
								Nickname:     "Johnny",
								PublishEmail: true,
								ZipCode:      "12345",
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("100003"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMoreUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     100003,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													ReqSubType: wire.ICQDBQueryMetaReplySetBasicInfo,
													Success:    wire.ICQStatusCodeOK,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userUpdater := newMockICQUserUpdater(t)
			for _, params := range tt.mockParams.setBasicInfoParams {
				userUpdater.EXPECT().
					SetBasicInfo(params.name, params.data).
					Return(params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := ICQService{
				userUpdater:    userUpdater,
				messageRelayer: messageRelayer,
			}
			err := s.UpdateBasicInfo(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_UpdateWorkInfo(t *testing.T) {
	tests := []struct {
		name       string
		seq        uint16
		sess       *state.Session
		req        wire.ICQWorkInfo
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "update work info - happy path",
			seq:  1,
			sess: newTestSession("100003", sessOptUIN(100003)),
			req: wire.ICQWorkInfo{
				Company:         "TechCorp Inc.",
				Department:      "Engineering",
				OccupationCode:  1023,
				Position:        "Staff Software Engineer",
				WorkAddress:     "456 Technology Blvd",
				WorkCity:        "Innovate City",
				WorkCountryCode: 1,
				WorkFax:         "987-654-3210",
				WorkPhone:       "222-333-4444",
				WorkState:       "CA",
				WorkWebPage:     "http://www.techcorp.com",
				WorkZIP:         "67890",
			},
			mockParams: mockParams{
				icqUserUpdaterParams: icqUserUpdaterParams{
					setWorkInfoParams: setWorkInfoParams{
						{
							name: state.NewIdentScreenName("100003"),
							data: state.ICQWorkInfo{
								Company:         "TechCorp Inc.",
								Department:      "Engineering",
								OccupationCode:  1023,
								Position:        "Staff Software Engineer",
								WorkAddress:     "456 Technology Blvd",
								WorkCity:        "Innovate City",
								WorkCountryCode: 1,
								WorkFax:         "987-654-3210",
								WorkPhone:       "222-333-4444",
								WorkState:       "CA",
								WorkWebPage:     "http://www.techcorp.com",
								WorkZIP:         "67890",
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("100003"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMoreUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     100003,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													ReqSubType: wire.ICQDBQueryMetaReplySetWorkInfo,
													Success:    wire.ICQStatusCodeOK,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userUpdater := newMockICQUserUpdater(t)
			for _, params := range tt.mockParams.setWorkInfoParams {
				userUpdater.EXPECT().
					SetWorkInfo(params.name, params.data).
					Return(params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := ICQService{
				userUpdater:    userUpdater,
				messageRelayer: messageRelayer,
			}
			err := s.UpdateWorkInfo(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_UpdateMoreInfo(t *testing.T) {
	tests := []struct {
		name       string
		seq        uint16
		sess       *state.Session
		req        wire.SomeMoreUserInfo
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "update more info - happy path",
			seq:  1,
			sess: newTestSession("100003", sessOptUIN(100003)),
			req: wire.SomeMoreUserInfo{
				Age:          0,
				BirthDay:     7,
				BirthMonth:   8,
				BirthYear:    1994,
				Gender:       1,
				HomePageAddr: "http://www.johndoe.com",
				Lang1:        1,
				Lang2:        2,
				Lang3:        3,
			},
			mockParams: mockParams{
				icqUserUpdaterParams: icqUserUpdaterParams{
					setMoreInfoParams: setMoreInfoParams{
						{
							name: state.NewIdentScreenName("100003"),
							data: state.ICQMoreInfo{
								BirthDay:     7,
								BirthMonth:   8,
								BirthYear:    1994,
								Gender:       1,
								HomePageAddr: "http://www.johndoe.com",
								Lang1:        1,
								Lang2:        2,
								Lang3:        3,
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("100003"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMoreUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     100003,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													ReqSubType: wire.ICQDBQueryMetaReplySetMoreInfo,
													Success:    wire.ICQStatusCodeOK,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userUpdater := newMockICQUserUpdater(t)
			for _, params := range tt.mockParams.setMoreInfoParams {
				userUpdater.EXPECT().
					SetMoreInfo(params.name, params.data).
					Return(params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := ICQService{
				userUpdater:    userUpdater,
				messageRelayer: messageRelayer,
			}
			err := s.UpdateMoreInfo(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}
