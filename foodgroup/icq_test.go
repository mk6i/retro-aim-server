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
		wantSNAC   wire.SNACMessage
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "success",
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
