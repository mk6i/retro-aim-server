package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOServiceBOSHandler_ClientOnline(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceClientOnline,
		},
		Body: wire.SNAC_0x01_0x02_OServiceClientOnline{
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
	}

	svc := newMockOServiceBOSService(t)
	svc.EXPECT().
		ClientOnline(mock.Anything, input.Body, mock.Anything).
		Return(nil)

	h := NewOServiceHandlerForBOS(slog.Default(), nil, svc)

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ClientOnline(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceBOSHandler_ServiceRequest(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceServiceRequest,
		},
		Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
			FoodGroup: wire.Chat,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceServiceResponse,
		},
		Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, uint16(1000)),
				},
			},
		},
	}

	svc := newMockOServiceBOSService(t)
	svc.EXPECT().
		ServiceRequest(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewOServiceHandlerForBOS(slog.Default(), nil, svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ServiceRequest(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceChatHandler_ClientOnline(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceClientOnline,
		},
		Body: wire.SNAC_0x01_0x02_OServiceClientOnline{
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
	}

	svc := newMockOServiceChatService(t)
	svc.EXPECT().
		ClientOnline(mock.Anything, mock.Anything).
		Return(nil)

	h := NewOServiceHandlerForChat(slog.Default(), nil, svc)

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ClientOnline(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceHandler_IdleNotification(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceIdleNotification,
		},
		Body: wire.SNAC_0x01_0x11_OServiceIdleNotification{
			IdleTime: 10,
		},
	}

	svc := newMockOServiceService(t)
	svc.EXPECT().
		IdleNotification(mock.Anything, mock.Anything, input.Body).
		Return(nil)

	h := OServiceHandler{
		OServiceService: svc,
		RouteLogger: middleware.RouteLogger{
			Logger: slog.Default(),
		},
	}

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.IdleNotification(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceHandler_ClientVersions(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceClientVersions,
		},
		Body: wire.SNAC_0x01_0x17_OServiceClientVersions{
			Versions: []uint16{
				10,
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostVersions,
		},
		Body: wire.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: []uint16{
				10,
			},
		},
	}

	svc := newMockOServiceService(t)
	svc.EXPECT().
		ClientVersions(mock.Anything, input.Frame, input.Body).
		Return(output)

	h := OServiceHandler{
		OServiceService: svc,
		RouteLogger: middleware.RouteLogger{
			Logger: slog.Default(),
		},
	}

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ClientVersions(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceHandler_RateParamsQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamsQuery,
		},
		Body: struct{}{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamsReply,
		},
		Body: wire.SNAC_0x01_0x07_OServiceRateParamsReply{
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
	}

	svc := newMockOServiceService(t)
	svc.EXPECT().
		RateParamsQuery(mock.Anything, input.Frame).
		Return(output)

	h := OServiceHandler{
		OServiceService: svc,
		RouteLogger: middleware.RouteLogger{
			Logger: slog.Default(),
		},
	}

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.RateParamsQuery(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceHandler_RateParamsSubAdd(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamsSubAdd,
		},
		Body: wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, []byte{1, 2, 3, 4}),
				},
			},
		},
	}

	svc := newMockOServiceService(t)
	svc.EXPECT().
		RateParamsSubAdd(mock.Anything, input.Body)

	h := OServiceHandler{
		OServiceService: svc,
		RouteLogger: middleware.RouteLogger{
			Logger: slog.Default(),
		},
	}

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.RateParamsSubAdd(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceHandler_SetUserInfoFields(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceSetUserInfoFields,
		},
		Body: wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, []byte{1, 2, 3, 4}),
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: wire.TLVUserInfo{
				ScreenName: "screen-name",
			},
		},
	}

	svc := newMockOServiceService(t)
	svc.EXPECT().
		SetUserInfoFields(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := OServiceHandler{
		OServiceService: svc,
		RouteLogger: middleware.RouteLogger{
			Logger: slog.Default(),
		},
	}

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.SetUserInfoFields(nil, nil, input.Frame, buf, responseWriter))
}

func TestOServiceHandler_UserInfoQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoQuery,
		},
		Body: struct{}{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: wire.TLVUserInfo{
				ScreenName: "screen-name",
			},
		},
	}

	svc := newMockOServiceService(t)
	svc.EXPECT().
		UserInfoQuery(mock.Anything, mock.Anything, input.Frame).
		Return(output)

	h := OServiceHandler{
		OServiceService: svc,
		RouteLogger: middleware.RouteLogger{
			Logger: slog.Default(),
		},
	}

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.UserInfoQuery(nil, nil, input.Frame, buf, responseWriter))
}
