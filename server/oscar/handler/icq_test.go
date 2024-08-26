package handler

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestICQHandler_DBQuery(t *testing.T) {
	type ICQMetaRequest struct {
		wire.ICQMetadata
		ReqSubType  uint16
		MetaRequest any
	}
	type reqParams struct {
		ctx     context.Context
		sess    *state.Session
		inFrame wire.SNACFrame
		inBody  wire.SNAC_0x15_0x02_BQuery
		rw      oscar.ResponseWriter
		seq     uint16
		wantErr error
	}
	type mockParam struct {
		req     any
		wantErr error
	}
	type allMockParams struct {
		deleteMsgReq    *mockParam
		findByDetails   *mockParam
		findByEmail     *mockParam
		findByInterests *mockParam
		findByUIN       *mockParam
		fullUserInfo    *mockParam
		offlineMsgReq   *mockParam
		setAffiliations *mockParam
		setBasicInfo    *mockParam
		setEmails       *mockParam
		setInterests    *mockParam
		setMoreInfo     *mockParam
		setPermissions  *mockParam
		setUserNotes    *mockParam
		setWorkInfo     *mockParam
		shortUserInfo   *mockParam
		xmlReqData      *mockParam
	}
	tests := []struct {
		name          string
		reqParams     reqParams
		allMockParams allMockParams
	}{
		{
			name: "MetaReqFullInfo - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqFullInfo,
									MetaRequest: wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{
										UIN: 123456789,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				fullUserInfo: &mockParam{
					req: wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{
						UIN: 123456789,
					},
				},
			},
		},
		{
			name: "MetaReqShortInfo - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqShortInfo,
									MetaRequest: wire.ICQ_0x07D0_0x04BA_DBQueryMetaReqShortInfo{
										UIN: 123456789,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				shortUserInfo: &mockParam{
					req: wire.ICQ_0x07D0_0x04BA_DBQueryMetaReqShortInfo{
						UIN: 123456789,
					},
				},
			},
		},
		{
			name: "MetaReqFullInfo2 - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqFullInfo2,
									MetaRequest: wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{
										UIN: 123456789,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				fullUserInfo: &mockParam{
					req: wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{
						UIN: 123456789,
					},
				},
			},
		},
		{
			name: "MetaReqXMLReq - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqXMLReq,
									MetaRequest: wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq{
										XMLRequest: "<xml></xml>",
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				xmlReqData: &mockParam{
					req: wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq{
						XMLRequest: "<xml></xml>",
					},
				},
			},
		},
		{
			name: "MetaReqSetPermissions - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetPermissions,
									MetaRequest: wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions{
										Authorization: 1,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setPermissions: &mockParam{
					req: wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions{
						Authorization: 1,
					},
				},
			},
		},
		{
			name: "MetaReqSearchByUIN - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSearchByUIN,
									MetaRequest: wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{
										UIN: 123456789,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				findByUIN: &mockParam{
					req: wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{
						UIN: 123456789,
					},
				},
			},
		},
		{
			name: "MetaReqSearchByEmail - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSearchByEmail,
									MetaRequest: wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail{
										Email: "test@aol.com",
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				findByEmail: &mockParam{
					req: wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail{
						Email: "test@aol.com",
					},
				},
			},
		},
		{
			name: "MetaReqSearchByDetails - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSearchByDetails,
									MetaRequest: wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails{
										FirstName: "john",
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				findByDetails: &mockParam{
					req: wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails{
						FirstName: "john",
					},
				},
			},
		},
		{
			name: "MetaReqSearchWhitePages - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSearchWhitePages,
									MetaRequest: wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages{
										InterestsCode: 1,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				findByInterests: &mockParam{
					req: wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages{
						InterestsCode: 1,
					},
				},
			},
		},
		{
			name: "MetaReqSetBasicInfo - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetBasicInfo,
									MetaRequest: wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo{
										FirstName: "john",
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setBasicInfo: &mockParam{
					req: wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo{
						FirstName: "john",
					},
				},
			},
		},
		{
			name: "MetaReqSetWorkInfo - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetWorkInfo,
									MetaRequest: wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo{
										ZIP: "11111",
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setWorkInfo: &mockParam{
					req: wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo{
						ZIP: "11111",
					},
				},
			},
		},
		{
			name: "MetaReqSetMoreInfo - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetMoreInfo,
									MetaRequest: wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo{
										Age: 100,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setMoreInfo: &mockParam{
					req: wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo{
						Age: 100,
					},
				},
			},
		},
		{
			name: "MetaReqSetNotes - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetNotes,
									MetaRequest: wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes{
										Notes: "my note",
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setUserNotes: &mockParam{
					req: wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes{
						Notes: "my note",
					},
				},
			},
		},
		{
			name: "MetaReqSetEmails - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetEmails,
									MetaRequest: wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails{
										Emails: []struct {
											Publish uint8
											Email   string `oscar:"len_prefix=uint16,nullterm"`
										}{
											{
												Email: "test@aol.com",
											},
										},
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setEmails: &mockParam{
					req: wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails{
						Emails: []struct {
							Publish uint8
							Email   string `oscar:"len_prefix=uint16,nullterm"`
						}{
							{
								Email: "test@aol.com",
							},
						},
					},
				},
			},
		},
		{
			name: "MetaReqSetInterests - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetInterests,
									MetaRequest: wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests{
										Interests: []struct {
											Code    uint16
											Keyword string `oscar:"len_prefix=uint16,nullterm"`
										}{
											{
												Keyword: "an_interest",
											},
										},
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setInterests: &mockParam{
					req: wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests{
						Interests: []struct {
							Code    uint16
							Keyword string `oscar:"len_prefix=uint16,nullterm"`
						}{
							{
								Keyword: "an_interest",
							},
						},
					},
				},
			},
		},
		{
			name: "MetaReqSetAffiliations - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType: wire.ICQDBQueryMetaReqSetAffiliations,
									MetaRequest: wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations{
										PastAffiliations: []struct {
											Code    uint16
											Keyword string `oscar:"len_prefix=uint16,nullterm"`
										}{
											{
												Keyword: "a_past_affiliation",
											},
										},
										Affiliations: []struct {
											Code    uint16
											Keyword string `oscar:"len_prefix=uint16,nullterm"`
										}{
											{
												Keyword: "an_affiliation",
											},
										},
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				setAffiliations: &mockParam{
					req: wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations{
						PastAffiliations: []struct {
							Code    uint16
							Keyword string `oscar:"len_prefix=uint16,nullterm"`
						}{
							{
								Keyword: "a_past_affiliation",
							},
						},
						Affiliations: []struct {
							Code    uint16
							Keyword string `oscar:"len_prefix=uint16,nullterm"`
						}{
							{
								Keyword: "an_affiliation",
							},
						},
					},
				},
			},
		},
		{
			name: "MetaReqStat - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType:  wire.ICQDBQueryMetaReqStat0a8c,
									MetaRequest: struct{}{},
								},
							}),
						},
					},
				},
				seq: 1,
			},
		},
		{
			name: "unknown metadata request subtype",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryMetaReq,
										Seq:     1,
									},
									ReqSubType:  0xCA8E,
									MetaRequest: struct{}{},
								},
							}),
						},
					},
				},
				seq:     1,
				wantErr: errUnknownICQMetaReqSubType,
			},
		},
		{
			name: "OfflineMsgReq - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: wire.ICQMetadataWithSubType{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryOfflineMsgReq,
										Seq:     1,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				offlineMsgReq: &mockParam{},
			},
		},
		{
			name: "DeleteMsgReq - happy path",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: wire.ICQMetadataWithSubType{
									ICQMetadata: wire.ICQMetadata{
										ReqType: wire.ICQDBQueryDeleteMsgReq,
										Seq:     1,
									},
								},
							}),
						},
					},
				},
				seq: 1,
			},
			allMockParams: allMockParams{
				deleteMsgReq: &mockParam{},
			},
		},
		{
			name: "unknown request type",
			reqParams: reqParams{
				sess: &state.Session{},
				inBody: wire.SNAC_0x15_0x02_BQuery{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ICQTLVTagsMetadata, wire.ICQMessageReplyEnvelope{
								Message: ICQMetaRequest{
									ICQMetadata: wire.ICQMetadata{
										ReqType: 0x13B4,
										Seq:     1,
									},
									ReqSubType:  0xCA8E,
									MetaRequest: struct{}{},
								},
							}),
						},
					},
				},
				seq:     1,
				wantErr: errUnknownICQMetaReqType,
			},
		}, // todo: add to a separate test
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			icqService := newMockICQService(t)
			switch {
			case tt.allMockParams.fullUserInfo != nil:
				icqService.EXPECT().
					FullUserInfo(mock.Anything, tt.reqParams.sess, tt.allMockParams.fullUserInfo.req, tt.reqParams.seq).
					Return(tt.allMockParams.fullUserInfo.wantErr)
			case tt.allMockParams.shortUserInfo != nil:
				icqService.EXPECT().
					ShortUserInfo(mock.Anything, tt.reqParams.sess, tt.allMockParams.shortUserInfo.req, tt.reqParams.seq).
					Return(tt.allMockParams.shortUserInfo.wantErr)
			case tt.allMockParams.xmlReqData != nil:
				icqService.EXPECT().
					XMLReqData(mock.Anything, tt.reqParams.sess, tt.allMockParams.xmlReqData.req, tt.reqParams.seq).
					Return(tt.allMockParams.xmlReqData.wantErr)
			case tt.allMockParams.setPermissions != nil:
				icqService.EXPECT().
					SetPermissions(mock.Anything, tt.reqParams.sess, tt.allMockParams.setPermissions.req, tt.reqParams.seq).
					Return(tt.allMockParams.setPermissions.wantErr)
			case tt.allMockParams.findByUIN != nil:
				icqService.EXPECT().
					FindByUIN(mock.Anything, tt.reqParams.sess, tt.allMockParams.findByUIN.req, tt.reqParams.seq).
					Return(tt.allMockParams.findByUIN.wantErr)
			case tt.allMockParams.findByEmail != nil:
				icqService.EXPECT().
					FindByEmail(mock.Anything, tt.reqParams.sess, tt.allMockParams.findByEmail.req, tt.reqParams.seq).
					Return(tt.allMockParams.findByEmail.wantErr)
			case tt.allMockParams.findByDetails != nil:
				icqService.EXPECT().
					FindByDetails(mock.Anything, tt.reqParams.sess, tt.allMockParams.findByDetails.req, tt.reqParams.seq).
					Return(tt.allMockParams.findByDetails.wantErr)
			case tt.allMockParams.findByInterests != nil:
				icqService.EXPECT().
					FindByInterests(mock.Anything, tt.reqParams.sess, tt.allMockParams.findByInterests.req, tt.reqParams.seq).
					Return(tt.allMockParams.findByInterests.wantErr)
			case tt.allMockParams.setBasicInfo != nil:
				icqService.EXPECT().
					SetBasicInfo(mock.Anything, tt.reqParams.sess, tt.allMockParams.setBasicInfo.req, tt.reqParams.seq).
					Return(tt.allMockParams.setBasicInfo.wantErr)
			case tt.allMockParams.setWorkInfo != nil:
				icqService.EXPECT().
					SetWorkInfo(mock.Anything, tt.reqParams.sess, tt.allMockParams.setWorkInfo.req, tt.reqParams.seq).
					Return(tt.allMockParams.setWorkInfo.wantErr)
			case tt.allMockParams.setMoreInfo != nil:
				icqService.EXPECT().
					SetMoreInfo(mock.Anything, tt.reqParams.sess, tt.allMockParams.setMoreInfo.req, tt.reqParams.seq).
					Return(tt.allMockParams.setMoreInfo.wantErr)
			case tt.allMockParams.setUserNotes != nil:
				icqService.EXPECT().
					SetUserNotes(mock.Anything, tt.reqParams.sess, tt.allMockParams.setUserNotes.req, tt.reqParams.seq).
					Return(tt.allMockParams.setUserNotes.wantErr)
			case tt.allMockParams.setEmails != nil:
				icqService.EXPECT().
					SetEmails(mock.Anything, tt.reqParams.sess, tt.allMockParams.setEmails.req, tt.reqParams.seq).
					Return(tt.allMockParams.setEmails.wantErr)
			case tt.allMockParams.setInterests != nil:
				icqService.EXPECT().
					SetInterests(mock.Anything, tt.reqParams.sess, tt.allMockParams.setInterests.req, tt.reqParams.seq).
					Return(tt.allMockParams.setInterests.wantErr)
			case tt.allMockParams.setAffiliations != nil:
				icqService.EXPECT().
					SetAffiliations(mock.Anything, tt.reqParams.sess, tt.allMockParams.setAffiliations.req, tt.reqParams.seq).
					Return(tt.allMockParams.setAffiliations.wantErr)
			case tt.allMockParams.offlineMsgReq != nil:
				icqService.EXPECT().
					OfflineMsgReq(mock.Anything, tt.reqParams.sess, tt.reqParams.seq).
					Return(tt.allMockParams.offlineMsgReq.wantErr)
			case tt.allMockParams.deleteMsgReq != nil:
				icqService.EXPECT().
					DeleteMsgReq(mock.Anything, tt.reqParams.sess, tt.reqParams.seq).
					Return(tt.allMockParams.deleteMsgReq.wantErr)
			}

			rt := NewICQHandler(slog.Default(), icqService)

			buf := &bytes.Buffer{}
			assert.NoError(t, wire.MarshalBE(tt.reqParams.inBody, buf))

			err := rt.DBQuery(nil, tt.reqParams.sess, wire.SNACFrame{}, buf, nil)
			assert.ErrorIs(t, err, tt.reqParams.wantErr)
		})
	}
}
