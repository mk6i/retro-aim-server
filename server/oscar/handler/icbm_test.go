package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestICBMHandler_AddParameters(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMAddParameters,
		},
		Body: wire.SNAC_0x04_0x02_ICBMAddParameters{
			Channel: 1,
		},
	}

	svc := newMockICBMService(t)
	h := NewICBMHandler(slog.Default(), svc)
	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.AddParameters(nil, nil, input.Frame, buf, responseWriter))
}

func TestICBMHandler_ChannelMsgToHost(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMChannelMsgToHost,
		},
		Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
			ScreenName: "recipient-screen-name",
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMHostAck,
		},
		Body: wire.SNAC_0x04_0x0C_ICBMHostAck{
			ChannelID: 4,
		},
	}

	svc := newMockICBMService(t)
	svc.EXPECT().
		ChannelMsgToHost(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(&output, nil)

	h := NewICBMHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ChannelMsgToHost(nil, nil, input.Frame, buf, responseWriter))
}

func TestICBMHandler_ClientErr(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMClientErr,
		},
		Body: wire.SNAC_0x04_0x0B_ICBMClientErr{
			Code: 4,
		},
	}

	svc := newMockICBMService(t)
	h := NewICBMHandler(slog.Default(), svc)
	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ClientErr(nil, nil, input.Frame, buf, responseWriter))
}

func TestICBMHandler_ClientEvent(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMClientEvent,
		},
		Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
			ScreenName: "recipient-screen-name",
		},
	}

	svc := newMockICBMService(t)
	svc.EXPECT().
		ClientEvent(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(nil)

	h := NewICBMHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ClientEvent(nil, nil, input.Frame, buf, responseWriter))
}

func TestICBMHandler_EvilRequest(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMEvilRequest,
		},
		Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
			ScreenName: "recipient-screen-name",
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMEvilReply,
		},
		Body: wire.SNAC_0x04_0x09_ICBMEvilReply{
			EvilDeltaApplied: 100,
		},
	}

	svc := newMockICBMService(t)
	svc.EXPECT().
		EvilRequest(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewICBMHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.EvilRequest(nil, nil, input.Frame, buf, responseWriter))
}

func TestICBMHandler_ParameterQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMParameterQuery,
		},
		Body: struct{}{}, // empty SNAC
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMParameterReply,
		},
		Body: wire.SNAC_0x04_0x05_ICBMParameterReply{
			MaxSlots: 100,
		},
	}

	svc := newMockICBMService(t)
	svc.EXPECT().
		ParameterQuery(mock.Anything, input.Frame).
		Return(output)

	h := NewICBMHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ParameterQuery(nil, nil, input.Frame, buf, responseWriter))
}
