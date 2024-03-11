package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/wire"
)

func TestBARTHandler_DownloadQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.BART,
			SubGroup:  wire.BARTDownloadQuery,
		},
		Body: wire.SNAC_0x10_0x04_BARTDownloadQuery{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.BART,
			SubGroup:  wire.BARTDownloadReply,
		},
		Body: wire.SNAC_0x10_0x05_BARTDownloadReply{
			ScreenName: "the-screen-name",
		},
	}

	svc := newMockBARTService(t)
	svc.EXPECT().
		RetrieveItem(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewBARTHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.DownloadQuery(nil, nil, input.Frame, buf, responseWriter))
}

func TestBARTHandler_UploadQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.BART,
			SubGroup:  wire.BARTUploadQuery,
		},
		Body: wire.SNAC_0x10_0x02_BARTUploadQuery{
			Type: 1,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.BART,
			SubGroup:  wire.BARTUploadReply,
		},
		Body: wire.SNAC_0x10_0x03_BARTUploadReply{
			Code: wire.BARTReplyCodesSuccess,
		},
	}

	svc := newMockBARTService(t)
	svc.EXPECT().
		UpsertItem(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewBARTHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.UploadQuery(nil, nil, input.Frame, buf, responseWriter))
}
