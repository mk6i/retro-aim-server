package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminHandler_ConfirmRequest(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminAcctConfirmRequest,
		},
		Body: wire.SNAC_0x07_0x06_AdminConfirmRequest{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminAcctConfirmReply,
		},
		Body: wire.SNAC_0x07_0x07_AdminConfirmReply{
			Status: wire.AdminAcctConfirmStatusEmailSent,
		},
	}

	svc := newMockAdminService(t)
	svc.EXPECT().
		ConfirmRequest(nil, mock.Anything, input.Frame).
		Return(output, nil)

	h := NewAdminHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.ConfirmRequest(nil, nil, input.Frame, buf, responseWriter))
}

func TestAdminHandler_InfoQuery_RegistrationStatus(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminInfoQuery,
		},
		Body: wire.SNAC_0x07_0x02_AdminInfoQuery{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.AdminTLVRegistrationStatus, uint16(0x00))},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminInfoReply,
		},
		Body: wire.SNAC_0x07_0x03_AdminInfoReply{
			Permissions: wire.AdminInfoPermissionsReadWrite,
			TLVBlock: wire.TLVBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.AdminTLVRegistrationStatus, wire.AdminInfoRegStatusFullDisclosure),
				},
			},
		},
	}

	svc := newMockAdminService(t)
	svc.EXPECT().
		ConfirmRequest(mock.Anything, mock.Anything, input.Frame).
		Return(output, nil)

	h := NewAdminHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.ConfirmRequest(nil, nil, input.Frame, buf, responseWriter))
}

func TestAdminHandler_InfoChangeRequest_ScreenNameFormatted(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminInfoQuery,
		},
		Body: wire.SNAC_0x07_0x04_AdminInfoChangeRequest{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.AdminTLVScreenNameFormatted, "Chatting Chuck")},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminInfoChangeReply,
		},
		Body: wire.SNAC_0x07_0x05_AdminChangeReply{
			Permissions: wire.AdminInfoPermissionsReadWrite,
			TLVBlock: wire.TLVBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.AdminTLVScreenNameFormatted, "Chatting Chuck"),
				},
			},
		},
	}

	svc := newMockAdminService(t)
	svc.EXPECT().
		InfoChangeRequest(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewAdminHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.InfoChangeRequest(nil, nil, input.Frame, buf, responseWriter))
}
