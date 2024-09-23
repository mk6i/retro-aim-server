package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/wire"
)

func TestODirHandler_InfoQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ODir,
			SubGroup:  wire.ODirInfoQuery,
		},
		Body: wire.SNAC_0x0F_0x02_InfoQuery{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(1, uint16(2)),
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ODir,
			SubGroup:  wire.ODirInfoReply,
		},
		Body: wire.SNAC_0x0F_0x03_InfoReply{
			Status: 5, // OK has results/not found
		},
	}

	svc := newMockODirService(t)
	svc.EXPECT().
		InfoQuery(mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewODirHandler(slog.Default(), svc)

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.InfoQuery(nil, nil, input.Frame, buf, ss))
}

func TestODirHandler_KeywordListQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ODir,
			SubGroup:  wire.ODirKeywordListQuery,
		},
		Body: wire.SNAC_0x0F_0x02_InfoQuery{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(1, uint16(2)),
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ODir,
			SubGroup:  wire.ODirKeywordListReply,
		},
		Body: wire.SNAC_0x0F_0x04_KeywordListReply{
			Status: 0x01,
		},
	}

	svc := newMockODirService(t)
	svc.EXPECT().
		KeywordListQuery(mock.Anything, input.Frame).
		Return(output, nil)

	h := NewODirHandler(slog.Default(), svc)

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.KeywordListQuery(nil, nil, input.Frame, buf, ss))
}
