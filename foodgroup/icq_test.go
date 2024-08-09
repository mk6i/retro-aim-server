package foodgroup

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

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

func TestICQService_FindByWhitePages(t *testing.T) {
	tests := []struct {
		name       string
		timeNow    func() time.Time
		seq        uint16
		sess       *state.Session
		req        wire.ICQFindByWhitePages
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "find by white pages - happy path",
			timeNow: func() time.Time {
				return time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
			},
			seq:  1,
			sess: newTestSession("11111111", sessOptUIN(11111111)),
			req: wire.ICQFindByWhitePages{
				InterestsCode:    10,
				InterestsKeyword: "knitting,crocheting,sewing",
			},
			mockParams: mockParams{
				icqUserFinderParams: icqUserFinderParams{
					findByInterestsParams: findByInterestsParams{
						{
							code:     10,
							keywords: []string{"knitting", "crocheting", "sewing"},
							result: []state.User{
								{
									AuthReq:         false,
									EmailAddress:    "janey@example.com",
									FirstName:       "Jane",
									Gender:          2,
									IdentScreenName: state.NewIdentScreenName("987654321"),
									LastName:        "Doe",
									Nickname:        "Janey",
									BirthDay:        31,
									BirthMonth:      7,
									BirthYear:       1995,
								},
								{
									AuthReq:         true,
									EmailAddress:    "alice@example.com",
									FirstName:       "Alice",
									Gender:          1,
									IdentScreenName: state.NewIdentScreenName("123456789"),
									LastName:        "Smith",
									Nickname:        "Ally123",
									BirthDay:        31,
									BirthMonth:      7,
									BirthYear:       1999,
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserSearchResult{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyUserFound,
													Details: wire.ICQUserSearchRecord{
														UIN:           987654321,
														Nickname:      "Janey",
														FirstName:     "Jane",
														LastName:      "Doe",
														Email:         "janey@example.com",
														Authorization: 0,
														OnlineStatus:  0,
														Gender:        2,
														Age:           25,
													},
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserSearchResult{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
													Details: wire.ICQUserSearchRecord{
														UIN:           123456789,
														Nickname:      "Ally123",
														FirstName:     "Alice",
														LastName:      "Smith",
														Email:         "alice@example.com",
														Authorization: 1,
														OnlineStatus:  1,
														Gender:        1,
														Age:           21,
													},
													LastMessageFooter: &struct {
														FoundUsersLeft uint32
													}{
														FoundUsersLeft: 0,
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
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("987654321"),
							result:     nil,
						},
						{
							screenName: state.NewIdentScreenName("123456789"),
							result:     &state.Session{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userFinder := newMockICQUserFinder(t)
			for _, params := range tt.mockParams.findByInterestsParams {
				userFinder.EXPECT().
					FindByInterests(params.code, params.keywords).
					Return(params.result, params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			sessionRetriever := newMockSessionRetriever(t)
			for _, params := range tt.mockParams.retrieveSessionParams {
				sessionRetriever.EXPECT().
					RetrieveSession(params.screenName).
					Return(params.result)
			}

			s := ICQService{
				messageRelayer:   messageRelayer,
				sessionRetriever: sessionRetriever,
				timeNow:          tt.timeNow,
				userFinder:       userFinder,
			}
			err := s.FindByWhitePages(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_FindByDetails(t *testing.T) {
	tests := []struct {
		name       string
		timeNow    func() time.Time
		seq        uint16
		sess       *state.Session
		req        wire.ICQFindByDetails
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "find by details - happy path",
			timeNow: func() time.Time {
				return time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
			},
			seq:  1,
			sess: newTestSession("11111111", sessOptUIN(11111111)),
			req: wire.ICQFindByDetails{
				FirstName: "John",
				LastName:  "Doe",
				NickName:  "Johnny",
			},
			mockParams: mockParams{
				icqUserFinderParams: icqUserFinderParams{
					findByDetailsParams: findByDetailsParams{
						{
							firstName: "John",
							lastName:  "Doe",
							nickName:  "Johnny",
							result: []state.User{
								{
									AuthReq:         false,
									EmailAddress:    "janey@example.com",
									FirstName:       "Jane",
									Gender:          2,
									IdentScreenName: state.NewIdentScreenName("987654321"),
									LastName:        "Doe",
									Nickname:        "Janey",
									BirthDay:        31,
									BirthMonth:      7,
									BirthYear:       1995,
								},
								{
									AuthReq:         true,
									EmailAddress:    "john@example.com",
									FirstName:       "John",
									Gender:          1,
									IdentScreenName: state.NewIdentScreenName("123456789"),
									LastName:        "Doe",
									Nickname:        "Johnny",
									BirthDay:        31,
									BirthMonth:      7,
									BirthYear:       1999,
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserSearchResult{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyUserFound,
													Details: wire.ICQUserSearchRecord{
														UIN:           987654321,
														Nickname:      "Janey",
														FirstName:     "Jane",
														LastName:      "Doe",
														Email:         "janey@example.com",
														Authorization: 0,
														OnlineStatus:  0,
														Gender:        2,
														Age:           25,
													},
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserSearchResult{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
													Details: wire.ICQUserSearchRecord{
														UIN:           123456789,
														Nickname:      "Johnny",
														FirstName:     "John",
														LastName:      "Doe",
														Email:         "john@example.com",
														Authorization: 1,
														OnlineStatus:  1,
														Gender:        1,
														Age:           21,
													},
													LastMessageFooter: &struct {
														FoundUsersLeft uint32
													}{
														FoundUsersLeft: 0,
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
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("987654321"),
							result:     nil,
						},
						{
							screenName: state.NewIdentScreenName("123456789"),
							result:     &state.Session{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userFinder := newMockICQUserFinder(t)
			for _, params := range tt.mockParams.findByDetailsParams {
				userFinder.EXPECT().
					FindByDetails(params.firstName, params.lastName, params.nickName).
					Return(params.result, params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			sessionRetriever := newMockSessionRetriever(t)
			for _, params := range tt.mockParams.retrieveSessionParams {
				sessionRetriever.EXPECT().
					RetrieveSession(params.screenName).
					Return(params.result)
			}

			s := ICQService{
				messageRelayer:   messageRelayer,
				sessionRetriever: sessionRetriever,
				timeNow:          tt.timeNow,
				userFinder:       userFinder,
			}
			err := s.FindByDetails(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_FindByEmail(t *testing.T) {
	tests := []struct {
		name       string
		timeNow    func() time.Time
		seq        uint16
		sess       *state.Session
		req        wire.ICQFindByEmail
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "find by details - happy path",
			timeNow: func() time.Time {
				return time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
			},
			seq:  1,
			sess: newTestSession("11111111", sessOptUIN(11111111)),
			req: wire.ICQFindByEmail{
				Email: "john@example.com",
			},
			mockParams: mockParams{
				icqUserFinderParams: icqUserFinderParams{
					findByEmailParams: findByEmailParams{
						{
							email: "john@example.com",
							result: state.User{
								AuthReq:         true,
								EmailAddress:    "john@example.com",
								FirstName:       "John",
								Gender:          1,
								IdentScreenName: state.NewIdentScreenName("123456789"),
								LastName:        "Doe",
								Nickname:        "Johnny",
								BirthDay:        31,
								BirthMonth:      7,
								BirthYear:       1999,
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserSearchResult{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
													Details: wire.ICQUserSearchRecord{
														UIN:           123456789,
														Nickname:      "Johnny",
														FirstName:     "John",
														LastName:      "Doe",
														Email:         "john@example.com",
														Authorization: 1,
														OnlineStatus:  1,
														Gender:        1,
														Age:           21,
													},
													LastMessageFooter: &struct {
														FoundUsersLeft uint32
													}{
														FoundUsersLeft: 0,
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
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("123456789"),
							result:     &state.Session{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userFinder := newMockICQUserFinder(t)
			for _, params := range tt.mockParams.findByEmailParams {
				userFinder.EXPECT().
					FindByEmail(params.email).
					Return(params.result, params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			sessionRetriever := newMockSessionRetriever(t)
			for _, params := range tt.mockParams.retrieveSessionParams {
				sessionRetriever.EXPECT().
					RetrieveSession(params.screenName).
					Return(params.result)
			}

			s := ICQService{
				messageRelayer:   messageRelayer,
				sessionRetriever: sessionRetriever,
				timeNow:          tt.timeNow,
				userFinder:       userFinder,
			}
			err := s.FindByEmail(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_FindByUIN(t *testing.T) {
	tests := []struct {
		name       string
		timeNow    func() time.Time
		seq        uint16
		sess       *state.Session
		req        wire.ICQFindByUIN
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "find by details - happy path",
			timeNow: func() time.Time {
				return time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
			},
			seq:  1,
			sess: newTestSession("11111111", sessOptUIN(11111111)),
			req: wire.ICQFindByUIN{
				UIN: 123456789,
			},
			mockParams: mockParams{
				icqUserFinderParams: icqUserFinderParams{
					findByUINParams: findByUINParams{
						{
							UIN: 123456789,
							result: state.User{
								AuthReq:         true,
								EmailAddress:    "john@example.com",
								FirstName:       "John",
								Gender:          1,
								IdentScreenName: state.NewIdentScreenName("123456789"),
								LastName:        "Doe",
								Nickname:        "Johnny",
								BirthDay:        31,
								BirthMonth:      7,
								BirthYear:       1999,
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserSearchResult{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
													Details: wire.ICQUserSearchRecord{
														UIN:           123456789,
														Nickname:      "Johnny",
														FirstName:     "John",
														LastName:      "Doe",
														Email:         "john@example.com",
														Authorization: 1,
														OnlineStatus:  1,
														Gender:        1,
														Age:           21,
													},
													LastMessageFooter: &struct {
														FoundUsersLeft uint32
													}{
														FoundUsersLeft: 0,
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
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("123456789"),
							result:     &state.Session{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userFinder := newMockICQUserFinder(t)
			for _, params := range tt.mockParams.findByUINParams {
				userFinder.EXPECT().
					FindByUIN(params.UIN).
					Return(params.result, params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			sessionRetriever := newMockSessionRetriever(t)
			for _, params := range tt.mockParams.retrieveSessionParams {
				sessionRetriever.EXPECT().
					RetrieveSession(params.screenName).
					Return(params.result)
			}

			s := ICQService{
				messageRelayer:   messageRelayer,
				sessionRetriever: sessionRetriever,
				timeNow:          tt.timeNow,
				userFinder:       userFinder,
			}
			err := s.FindByUIN(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_GetICQFullUserInfo(t *testing.T) {
	tests := []struct {
		name       string
		timeNow    func() time.Time
		seq        uint16
		sess       *state.Session
		req        wire.ICQFindByUIN
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "get full user info - happy path",
			timeNow: func() time.Time {
				return time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC)
			},
			seq:  1,
			sess: newTestSession("11111111", sessOptUIN(11111111)),
			req: wire.ICQFindByUIN{
				UIN: 123456789,
			},
			mockParams: mockParams{
				icqUserFinderParams: icqUserFinderParams{
					findByUINParams: findByUINParams{
						{
							UIN: 123456789,
							result: state.User{
								IdentScreenName:          state.NewIdentScreenName("123456789"),
								Nickname:                 "CoolUser123",
								FirstName:                "John",
								LastName:                 "Doe",
								EmailAddress:             "john.doe@example.com",
								AuthReq:                  true,
								Gender:                   1,
								HomeCity:                 "New York",
								HomeState:                "NY",
								HomePhone:                "123-456-7890",
								HomeFax:                  "123-456-7891",
								HomeAddress:              "123 Main St, New York, NY 10001",
								CellPhone:                "987-654-3210",
								ZipCode:                  "10001",
								CountryCode:              1,
								GMTOffset:                5,
								PublishEmail:             true,
								WorkCity:                 "Los Angeles",
								WorkState:                "CA",
								WorkPhone:                "234-567-8901",
								WorkFax:                  "234-567-8902",
								WorkAddress:              "456 Work St, Los Angeles, CA 90001",
								WorkZIP:                  "90001",
								WorkCountryCode:          1,
								Company:                  "TechCorp",
								Department:               "Engineering",
								Position:                 "Staff Software Engineer",
								OccupationCode:           1234,
								WorkWebPage:              "https://techcorp.com",
								HomePageAddr:             "https://johnsdomain.com",
								BirthYear:                1990,
								BirthMonth:               6,
								BirthDay:                 15,
								Lang1:                    1,
								Lang2:                    2,
								Lang3:                    3,
								Notes:                    "This is a test user.",
								Interest1Code:            5678,
								Interest1Keyword:         "Programming",
								Interest2Code:            6789,
								Interest2Keyword:         "Gaming",
								Interest3Code:            7890,
								Interest3Keyword:         "Music",
								Interest4Code:            8901,
								Interest4Keyword:         "Traveling",
								PastAffiliations1Code:    9012,
								PastAffiliations1Keyword: "College",
								PastAffiliations2Code:    1233,
								PastAffiliations2Keyword: "High School",
								PastAffiliations3Code:    3456,
								PastAffiliations3Keyword: "Previous Job",
								Affiliations1Code:        4567,
								Affiliations1Keyword:     "Professional Org",
								Affiliations2Code:        5678,
								Affiliations2Keyword:     "Alumni Group",
								Affiliations3Code:        6789,
								Affiliations3Keyword:     "Community Group",
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyBasicInfo,

													Nickname:     "CoolUser123",
													FirstName:    "John",
													LastName:     "Doe",
													Email:        "john.doe@example.com",
													HomeCity:     "New York",
													HomeState:    "NY",
													HomePhone:    "123-456-7890",
													HomeFax:      "123-456-7891",
													HomeAddress:  "123 Main St, New York, NY 10001",
													CellPhone:    "987-654-3210",
													ZipCode:      "10001",
													CountryCode:  1,
													GMTOffset:    5,
													AuthFlag:     0, // todo figure these out
													WebAware:     1, // todo figure these out
													DCPerms:      0, // todo figure these out
													PublishEmail: wire.ICQUserFlagPublishEmailYes,
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
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
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyMoreInfo,
													SomeMoreUserInfo: wire.SomeMoreUserInfo{
														Age:          30,
														Gender:       1,
														HomePageAddr: "https://johnsdomain.com",
														BirthYear:    1990,
														BirthMonth:   6,
														BirthDay:     15,
														Lang1:        1,
														Lang2:        2,
														Lang3:        3,
													},
													City:        "New York",
													State:       "NY",
													CountryCode: 1,
													TimeZone:    5,
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQInfoEmailMore{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyExtEmailInfo,
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQHomepageCat{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyHomePageCat,
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMetaWorkUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyWorkInfo,
													ICQWorkInfo: wire.ICQWorkInfo{
														WorkCity:        "Los Angeles",
														WorkState:       "CA",
														WorkPhone:       "234-567-8901",
														WorkFax:         "234-567-8902",
														WorkAddress:     "456 Work St, Los Angeles, CA 90001",
														WorkZIP:         "90001",
														WorkCountryCode: 1,
														Company:         "TechCorp",
														Department:      "Engineering",
														Position:        "Staff Software Engineer",
														OccupationCode:  1234,
														WorkWebPage:     "https://techcorp.com",
													},
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserNotes{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyNotes,
													ICQNotes: wire.ICQNotes{
														Notes: "This is a test user.",
													},
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQUserInterests{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyInterests,
													Interests: []struct {
														Code    uint16
														Keyword string `oscar:"len_prefix=uint16,nullterm"`
													}{
														{
															Code:    5678,
															Keyword: "Programming",
														},
														{
															Code:    6789,
															Keyword: "Gaming",
														},
														{
															Code:    7890,
															Keyword: "Music",
														},
														{
															Code:    8901,
															Keyword: "Traveling",
														},
													},
												},
											}),
										},
									},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMetaAffiliationsUserInfo{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryMetaReply,
														Seq:     1,
													},
													Success:    wire.ICQStatusCodeOK,
													ReqSubType: wire.ICQDBQueryMetaReplyAffiliations,
													ICQAffiliations: wire.ICQAffiliations{
														PastAffiliations: []struct {
															Code    uint16
															Keyword string `oscar:"len_prefix=uint16,nullterm"`
														}{
															{
																Code:    9012,
																Keyword: "College",
															},
															{
																Code:    1233,
																Keyword: "High School",
															},
															{
																Code:    3456,
																Keyword: "Previous Job",
															},
														},
														Affiliations: []struct {
															Code    uint16
															Keyword string `oscar:"len_prefix=uint16,nullterm"`
														}{
															{
																Code:    4567,
																Keyword: "Professional Org",
															},
															{
																Code:    5678,
																Keyword: "Alumni Group",
															},
															{
																Code:    6789,
																Keyword: "Community Group",
															},
														},
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
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("123456789"),
							result:     &state.Session{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userFinder := newMockICQUserFinder(t)
			for _, params := range tt.mockParams.findByUINParams {
				userFinder.EXPECT().
					FindByUIN(params.UIN).
					Return(params.result, params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := ICQService{
				messageRelayer: messageRelayer,
				timeNow:        tt.timeNow,
				userFinder:     userFinder,
			}
			err := s.GetICQFullUserInfo(nil, tt.sess, tt.req, tt.seq)
			assert.NoError(t, err)
		})
	}
}

func TestICQService_GetICQMessagesEOF(t *testing.T) {
	tests := []struct {
		name       string
		seq        uint16
		sess       *state.Session
		mockParams mockParams
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			seq:  1,
			sess: newTestSession("11111111", sessOptUIN(11111111)),
			mockParams: mockParams{
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("11111111"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.ICQ,
									SubGroup:  wire.ICQDBReply,
								},
								Body: wire.SNAC_0x0F_0x02_ICQDBReply{
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessage{
												Message: wire.ICQMessagesEOF{
													ICQMetadata: wire.ICQMetadata{
														UIN:     11111111,
														ReqType: wire.ICQDBQueryOfflineMsgReplyLast,
														Seq:     1,
													},
													DroppedMessages: 0,
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
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			s := NewICQService(messageRelayer, nil, nil, slog.Default(), nil)
			err := s.GetICQMessagesEOF(nil, tt.sess, tt.seq)
			assert.NoError(t, err)
		})
	}
}
