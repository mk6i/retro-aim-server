package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/wire"
)

func TestChatHandler_ChannelMsgToHost_WithReflectedResponse(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatChannelMsgToHost,
		},
		Body: wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{
			Channel: 4,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatChannelMsgToClient,
		},
		Body: wire.SNAC_0x0E_0x06_ChatChannelMsgToClient{
			Channel: 4,
		},
	}

	svc := newMockChatService(t)
	svc.EXPECT().
		ChannelMsgToHost(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(&output, nil)

	h := NewChatHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ChannelMsgToHost(nil, nil, input.Frame, buf, responseWriter))
}

func TestChatHandler_ChannelMsgToHost_WithoutReflectedResponse(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatChannelMsgToHost,
		},
		Body: wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{
			Channel: 4,
		},
	}
	// nil response from handler means the response is not reflected back to
	// the caller
	var output *wire.SNACMessage

	svc := newMockChatService(t)
	svc.EXPECT().
		ChannelMsgToHost(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewChatHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t) // omit mock handler call

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.ChannelMsgToHost(nil, nil, input.Frame, buf, responseWriter))
}
