package server

import (
	"bytes"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOServiceRouter_RouteOService_ForBOS(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input oscar.SNACMessage
		// output is the response payload
		output oscar.SNACMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive OServiceClientOnline, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceClientOnline,
				},
				Body: oscar.SNAC_0x01_0x02_OServiceClientOnline{
					GroupVersions: []struct {
						FoodGroup   uint16
						Version     uint16
						ToolID      uint16
						ToolVersion uint16
					}{
						{
							FoodGroup: 10,
						},
					},
				},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive OServiceServiceRequest, return OServiceServiceResponse",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceServiceRequest,
				},
				Body: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: 10,
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceServiceResponse,
				},
				Body: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, uint16(1000)),
						},
					},
				},
			},
		},
		{
			name: "receive OServiceRateParamsQuery, return OServiceRateParamsReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceRateParamsQuery,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceRateParamsReply,
				},
				Body: oscar.SNAC_0x01_0x07_OServiceRateParamsReply{
					RateGroups: []struct {
						ID    uint16
						Pairs []struct {
							FoodGroup uint16
							SubGroup  uint16
						} `count_prefix:"uint16"`
					}{
						{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "receive OServiceRateParamsSubAdd, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceRateParamsSubAdd,
				},
				Body: oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, []byte{1, 2, 3, 4}),
						},
					},
				},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive OServiceUserInfoQuery, return OServiceUserInfoUpdate",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoQuery,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoUpdate,
				},
				Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
				},
			},
		},
		{
			name: "receive OServiceIdleNotification, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceIdleNotification,
				},
				Body: oscar.SNAC_0x01_0x11_OServiceIdleNotification{
					IdleTime: 10,
				},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive OServiceClientVersions, return OServiceHostVersions",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceClientVersions,
				},
				Body: oscar.SNAC_0x01_0x17_OServiceClientVersions{
					Versions: []uint16{
						10,
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceHostVersions,
				},
				Body: oscar.SNAC_0x01_0x18_OServiceHostVersions{
					Versions: []uint16{
						10,
					},
				},
			},
		},
		{
			name: "receive OServiceSetUserInfoFields, return OServiceUserInfoUpdate",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceSetUserInfoFields,
				},
				Body: oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, []byte{1, 2, 3, 4}),
						},
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoUpdate,
				},
				Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
				},
			},
		},
		{
			name: "receive OServicePauseReq, expect ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServicePauseReq,
				},
				Body: struct{}{}, // empty SNAC
			},
			output:    oscar.SNACMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockOServiceHandler(t)
			svc.EXPECT().
				RateParamsQueryHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				UserInfoQueryHandler(mock.Anything, mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				ClientVersionsHandler(mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetUserInfoFieldsHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				RateParamsSubAddHandler(mock.Anything, tc.input.Body).
				Maybe()
			svc.EXPECT().
				IdleNotificationHandler(mock.Anything, mock.Anything, tc.input.Body).
				Return(tc.handlerErr).
				Maybe()

			svcBOS := newMockOServiceBOSHandler(t)
			svcBOS.EXPECT().
				ServiceRequestHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svcBOS.EXPECT().
				ClientOnlineHandler(mock.Anything, tc.input.Body, mock.Anything).
				Return(tc.handlerErr).
				Maybe()

			router := OServiceBOSRouter{
				OServiceRouter: OServiceRouter{
					OServiceHandler: svc,
					RouteLogger: RouteLogger{
						Logger: NewLogger(Config{}),
					},
				},
				OServiceBOSHandler: svcBOS,
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteOService(nil, nil, tc.input.Frame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == (oscar.SNACMessage{}) {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
			}

			// verify the FLAP frame
			flap := oscar.FLAPFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

			flapBuf, err := flap.SNACBuffer(bufOut)
			assert.NoError(t, err)

			// verify the SNAC frame
			snacFrame := oscar.SNACFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, flapBuf))
			assert.Equal(t, tc.output.Frame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.Body, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), flapBuf.Bytes())
		})
	}
}

func TestOServiceRouter_RouteOService_ForChat(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input oscar.SNACMessage
		// output is the response payload
		output oscar.SNACMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive OServiceClientOnline, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceClientOnline,
				},
				Body: oscar.SNAC_0x01_0x02_OServiceClientOnline{
					GroupVersions: []struct {
						FoodGroup   uint16
						Version     uint16
						ToolID      uint16
						ToolVersion uint16
					}{
						{
							FoodGroup: 10,
						},
					},
				},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive OServiceServiceRequest, return OServiceErr",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceServiceRequest,
				},
				Body: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: 10,
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceErr,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeInvalidSnac,
				},
			},
		},
		{
			name: "receive OServiceRateParamsQuery, return OServiceRateParamsReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceRateParamsQuery,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceRateParamsReply,
				},
				Body: oscar.SNAC_0x01_0x07_OServiceRateParamsReply{
					RateGroups: []struct {
						ID    uint16
						Pairs []struct {
							FoodGroup uint16
							SubGroup  uint16
						} `count_prefix:"uint16"`
					}{
						{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "receive OServiceRateParamsSubAdd, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceRateParamsSubAdd,
				},
				Body: oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, []byte{1, 2, 3, 4}),
						},
					},
				},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive OServiceUserInfoQuery, return OServiceUserInfoUpdate",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoQuery,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoUpdate,
				},
				Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
				},
			},
		},
		{
			name: "receive OServiceIdleNotification, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceIdleNotification,
				},
				Body: oscar.SNAC_0x01_0x11_OServiceIdleNotification{
					IdleTime: 10,
				},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive OServiceClientVersions, return OServiceHostVersions",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceClientVersions,
				},
				Body: oscar.SNAC_0x01_0x17_OServiceClientVersions{
					Versions: []uint16{
						10,
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceHostVersions,
				},
				Body: oscar.SNAC_0x01_0x18_OServiceHostVersions{
					Versions: []uint16{
						10,
					},
				},
			},
		},
		{
			name: "receive OServiceSetUserInfoFields, return OServiceUserInfoUpdate",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceSetUserInfoFields,
				},
				Body: oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, []byte{1, 2, 3, 4}),
						},
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoUpdate,
				},
				Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
				},
			},
		},
		{
			name: "receive OServicePauseReq, expect ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServicePauseReq,
				},
				Body: struct{}{}, // empty SNAC
			},
			output:    oscar.SNACMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockOServiceHandler(t)
			svc.EXPECT().
				RateParamsQueryHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				UserInfoQueryHandler(mock.Anything, mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				ClientVersionsHandler(mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetUserInfoFieldsHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				RateParamsSubAddHandler(mock.Anything, tc.input.Body).
				Maybe()
			svc.EXPECT().
				IdleNotificationHandler(mock.Anything, mock.Anything, tc.input.Body).
				Return(tc.handlerErr).
				Maybe()

			svcBOS := newMockOServiceChatHandler(t)
			svcBOS.EXPECT().
				ClientOnlineHandler(mock.Anything, tc.input.Body, mock.Anything, mock.Anything).
				Return(tc.handlerErr).
				Maybe()

			router := OServiceChatRouter{
				OServiceRouter: OServiceRouter{
					OServiceHandler: svc,
					RouteLogger: RouteLogger{
						Logger: NewLogger(Config{}),
					},
				},
				OServiceChatHandler: svcBOS,
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteOService(nil, nil, "", tc.input.Frame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == (oscar.SNACMessage{}) {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
			}

			// verify the FLAP frame
			flap := oscar.FLAPFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

			flapBuf, err := flap.SNACBuffer(bufOut)
			assert.NoError(t, err)

			// verify the SNAC frame
			snacFrame := oscar.SNACFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, flapBuf))
			assert.Equal(t, tc.output.Frame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.Body, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), flapBuf.Bytes())
		})
	}
}
