package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFeedbagHandler_DeleteItem(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagDeleteItem,
		},
		Body: wire.SNAC_0x13_0x0A_FeedbagDeleteItem{
			Items: []wire.FeedbagItem{
				{
					Name: "my-item",
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagStatus,
		},
		Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
			Results: []uint16{1234},
		},
	}

	svc := newMockFeedbagService(t)
	svc.EXPECT().
		DeleteItem(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewFeedbagHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.DeleteItem(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_EndCluster(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagEndCluster,
		},
		Body: struct{}{},
	}

	svc := newMockFeedbagService(t)
	h := NewFeedbagHandler(slog.Default(), svc)
	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.EndCluster(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_InsertItem(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagInsertItem,
		},
		Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
			Items: []wire.FeedbagItem{
				{
					Name: "my-item",
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagStatus,
		},
		Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
			Results: []uint16{1234},
		},
	}

	svc := newMockFeedbagService(t)
	svc.EXPECT().
		InsertItem(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewFeedbagHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.InsertItem(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_Query(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagQuery,
		},
		Body: wire.SNAC_0x13_0x02_FeedbagRightsQuery{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					{
						Tag:   0x01,
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagReply,
		},
		Body: wire.SNAC_0x13_0x06_FeedbagReply{
			Version: 4,
		},
	}

	svc := newMockFeedbagService(t)
	svc.EXPECT().
		Query(mock.Anything, mock.Anything, input.Frame).
		Return(output, nil)

	h := NewFeedbagHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.Query(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_QueryIfModified(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagQueryIfModified,
		},
		Body: wire.SNAC_0x13_0x05_FeedbagQueryIfModified{
			LastUpdate: 1234,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagReply,
		},
		Body: wire.SNAC_0x13_0x06_FeedbagReply{
			LastUpdate: 1234,
		},
	}

	svc := newMockFeedbagService(t)
	svc.EXPECT().
		QueryIfModified(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewFeedbagHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.QueryIfModified(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_RightsQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagRightsQuery,
		},
		Body: wire.SNAC_0x13_0x02_FeedbagRightsQuery{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					{
						Tag:   0x01,
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagRightsReply,
		},
		Body: wire.SNAC_0x13_0x03_FeedbagRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					{
						Tag:   0x01,
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
	}

	svc := newMockFeedbagService(t)
	svc.EXPECT().
		RightsQuery(mock.Anything, input.Frame).
		Return(output)

	h := NewFeedbagHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.RightsQuery(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_StartCluster(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagStartCluster,
		},
		Body: wire.SNAC_0x13_0x11_FeedbagStartCluster{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					{
						Tag:   0x01,
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
	}

	svc := newMockFeedbagService(t)
	svc.EXPECT().
		StartCluster(mock.Anything, input.Frame, input.Body)

	h := NewFeedbagHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.StartCluster(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_UpdateItem(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagUpdateItem,
		},
		Body: wire.SNAC_0x13_0x09_FeedbagUpdateItem{
			Items: []wire.FeedbagItem{
				{
					Name: "my-item",
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagStatus,
		},
		Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
			Results: []uint16{1234},
		},
	}

	svc := newMockFeedbagService(t)
	svc.EXPECT().
		UpdateItem(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewFeedbagHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.UpdateItem(nil, nil, input.Frame, buf, responseWriter))
}

func TestFeedbagHandler_Use(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagUse,
		},
		Body: struct{}{},
	}

	svc := newMockFeedbagService(t)
	h := NewFeedbagHandler(slog.Default(), svc)
	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.Use(nil, nil, input.Frame, buf, responseWriter))
}
