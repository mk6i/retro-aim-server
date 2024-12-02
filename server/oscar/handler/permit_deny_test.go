package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/wire"
)

func TestPermitDenyHandler_RightsQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyRightsQuery,
		},
		Body: struct{}{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyRightsReply,
		},
		Body: wire.SNAC_0x09_0x03_PermitDenyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(0x01, uint16(1000)),
				},
			},
		},
	}

	svc := newMockPermitDenyService(t)
	svc.EXPECT().
		RightsQuery(mock.Anything, input.Frame).
		Return(output)

	h := NewPermitDenyHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.RightsQuery(nil, nil, input.Frame, buf, responseWriter))
}

func TestPermitDenyHandler_AddDenyListEntries(t *testing.T) {
	sess := state.NewSession()
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyAddDenyListEntries,
		},
		Body: wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{
			Users: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: "friend1",
				},
				{
					ScreenName: "friend2",
				},
			},
		},
	}
	svc := newMockPermitDenyService(t)
	svc.EXPECT().
		AddDenyListEntries(mock.Anything, sess, input.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	h := NewPermitDenyHandler(slog.Default(), svc)
	assert.NoError(t, h.AddDenyListEntries(nil, sess, input.Frame, buf, nil))
}

func TestPermitDenyHandler_DelDenyListEntries(t *testing.T) {
	sess := state.NewSession()
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyDelDenyListEntries,
		},
		Body: wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries{
			Users: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: "friend1",
				},
				{
					ScreenName: "friend2",
				},
			},
		},
	}
	svc := newMockPermitDenyService(t)
	svc.EXPECT().
		DelDenyListEntries(mock.Anything, sess, input.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	h := NewPermitDenyHandler(slog.Default(), svc)
	assert.NoError(t, h.DelDenyListEntries(nil, sess, input.Frame, buf, nil))
}

func TestPermitDenyHandler_AddPermListEntries(t *testing.T) {
	sess := state.NewSession()
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyAddPermListEntries,
		},
		Body: wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{
			Users: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: "friend1",
				},
				{
					ScreenName: "friend2",
				},
			},
		},
	}
	svc := newMockPermitDenyService(t)
	svc.EXPECT().
		AddPermListEntries(mock.Anything, sess, input.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	h := NewPermitDenyHandler(slog.Default(), svc)
	assert.NoError(t, h.AddPermListEntries(nil, sess, input.Frame, buf, nil))
}

func TestPermitDenyHandler_DelPermListEntries(t *testing.T) {
	sess := state.NewSession()
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyDelPermListEntries,
		},
		Body: wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries{
			Users: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: "friend1",
				},
				{
					ScreenName: "friend2",
				},
			},
		},
	}
	svc := newMockPermitDenyService(t)
	svc.EXPECT().
		DelPermListEntries(mock.Anything, sess, input.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	h := NewPermitDenyHandler(slog.Default(), svc)
	assert.NoError(t, h.DelPermListEntries(nil, sess, input.Frame, buf, nil))
}

func TestPermitDenyHandler_SetGroupPermitMask(t *testing.T) {
	sess := state.NewSession()
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyDelPermListEntries,
		},
		Body: wire.SNAC_0x09_0x04_PermitDenySetGroupPermitMask{
			PermMask: 1234,
		},
	}
	svc := newMockPermitDenyService(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	h := NewPermitDenyHandler(slog.Default(), svc)
	assert.NoError(t, h.SetGroupPermitMask(nil, sess, input.Frame, buf, nil))
}
