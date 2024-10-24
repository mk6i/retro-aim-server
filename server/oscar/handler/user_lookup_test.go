package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserLookupHandler_FindByEmail(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.UserLookup,
			SubGroup:  wire.UserLookupFindByEmail,
		},
		Body: wire.SNAC_0x0A_0x02_UserLookupFindByEmail{
			Email: []byte("haha@aol.com"),
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.UserLookup,
			SubGroup:  wire.UserLookupFindReply,
		},
		Body: wire.SNAC_0x0A_0x03_UserLookupFindReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(0x01, uint16(0x02)),
				},
			},
		},
	}

	svc := newMockUserLookupService(t)
	svc.EXPECT().
		FindByEmail(mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewUserLookupHandler(slog.Default(), svc)

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.FindByEmail(nil, nil, input.Frame, buf, ss))
}
