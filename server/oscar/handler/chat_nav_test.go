package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatNavHandler_CreateRoom(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavCreateRoom,
		},
		Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
			Exchange: 1,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{},
	}

	svc := newMockChatNavService(t)
	svc.EXPECT().
		CreateRoom(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewChatNavHandler(svc, slog.Default())

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.CreateRoom(nil, nil, input.Frame, buf, ss))
}

func TestChatNavHandler_CreateRoom_ReadErr(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavCreateRoom,
		},
		Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
			Exchange: 1,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{},
	}

	svc := newMockChatNavService(t)
	svc.EXPECT().
		CreateRoom(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewChatNavHandler(svc, slog.Default())

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.CreateRoom(nil, nil, input.Frame, buf, ss))
}

func TestChatNavHandler_RequestChatRights(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavRequestChatRights,
		},
		Body: struct{}{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{},
	}

	svc := newMockChatNavService(t)
	svc.EXPECT().
		RequestChatRights(mock.Anything, input.Frame).
		Return(output)

	h := NewChatNavHandler(svc, slog.Default())

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.RequestChatRights(nil, nil, input.Frame, buf, ss))
}

func TestChatNavHandler_RequestRoomInfo(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavRequestRoomInfo,
		},
		Body: wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
			Exchange: 1,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{},
	}

	svc := newMockChatNavService(t)
	svc.EXPECT().
		RequestRoomInfo(mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewChatNavHandler(svc, slog.Default())

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.RequestRoomInfo(nil, nil, input.Frame, buf, ss))
}

func TestChatNavHandler_RequestExchangeInfo(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavRequestExchangeInfo,
		},
		Body: wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo{
			Exchange: 4,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{},
	}

	svc := newMockChatNavService(t)
	svc.EXPECT().
		ExchangeInfo(mock.Anything, input.Frame, input.Body).
		Return(output)

	h := NewChatNavHandler(svc, slog.Default())

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.RequestExchangeInfo(nil, nil, input.Frame, buf, ss))
}
