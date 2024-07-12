package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLocateHandler_GetDirInfo(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateGetDirInfo,
		},
		Body: wire.SNAC_0x02_0x0B_LocateGetDirInfo{
			WatcherScreenNames: "screen-name",
		},
	}

	svc := newMockLocateService(t)
	h := NewLocateHandler(svc, slog.Default())
	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.GetDirInfo(nil, nil, input.Frame, buf, responseWriter))
}

func TestLocateHandler_RightsQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateRightsQuery,
		},
		Body: struct{}{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateRightsReply,
		},
		Body: wire.SNAC_0x02_0x03_LocateRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, uint16(1000)),
				},
			},
		},
	}

	svc := newMockLocateService(t)
	svc.EXPECT().
		RightsQuery(mock.Anything, input.Frame).
		Return(output)

	h := NewLocateHandler(svc, slog.Default())

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.RightsQuery(nil, nil, input.Frame, buf, responseWriter))
}

func TestLocateHandler_SetDirInfo(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateSetDirInfo,
		},
		Body: wire.SNAC_0x02_0x09_LocateSetDirInfo{
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
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateSetDirReply,
		},
		Body: wire.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}

	svc := newMockLocateService(t)
	svc.EXPECT().
		SetDirInfo(mock.Anything, input.Frame).
		Return(output)

	h := NewLocateHandler(svc, slog.Default())

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.SetDirInfo(nil, nil, input.Frame, buf, responseWriter))
}

func TestLocateHandler_SetInfo(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateSetInfo,
		},
		Body: wire.SNAC_0x02_0x04_LocateSetInfo{
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

	svc := newMockLocateService(t)
	svc.EXPECT().
		SetInfo(mock.Anything, mock.Anything, input.Body).
		Return(nil)

	h := NewLocateHandler(svc, slog.Default())
	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.SetInfo(nil, nil, input.Frame, buf, responseWriter))
}

func TestLocateHandler_SetKeywordInfo(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateSetKeywordInfo,
		},
		Body: wire.SNAC_0x02_0x0F_LocateSetKeywordInfo{
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
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateSetKeywordReply,
		},
		Body: wire.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}

	svc := newMockLocateService(t)
	svc.EXPECT().
		SetKeywordInfo(mock.Anything, input.Frame).
		Return(output)

	h := NewLocateHandler(svc, slog.Default())

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.SetKeywordInfo(nil, nil, input.Frame, buf, responseWriter))
}

func TestLocateHandler_UserInfoQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateUserInfoQuery,
		},
		Body: wire.SNAC_0x02_0x05_LocateUserInfoQuery{
			Type: 1,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateUserInfoReply,
		},
		Body: wire.SNAC_0x02_0x06_LocateUserInfoReply{
			TLVUserInfo: wire.TLVUserInfo{
				ScreenName: "screen-name",
			},
			LocateInfo: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					{
						Tag:   0x01,
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
	}

	svc := newMockLocateService(t)
	svc.EXPECT().
		UserInfoQuery(mock.Anything, mock.Anything, input.Frame, input.Body).
		Return(output, nil)

	h := NewLocateHandler(svc, slog.Default())

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.UserInfoQuery(nil, nil, input.Frame, buf, responseWriter))
}

func TestLocateHandler_UserInfoQuery2(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateUserInfoQuery2,
		},
		Body: wire.SNAC_0x02_0x15_LocateUserInfoQuery2{
			Type2: 1,
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Locate,
			SubGroup:  wire.LocateUserInfoReply,
		},
		Body: wire.SNAC_0x02_0x06_LocateUserInfoReply{
			TLVUserInfo: wire.TLVUserInfo{
				ScreenName: "screen-name",
			},
			LocateInfo: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					{
						Tag:   0x01,
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
	}

	svc := newMockLocateService(t)
	svc.EXPECT().
		UserInfoQuery(mock.Anything, mock.Anything, input.Frame, wire.SNAC_0x02_0x05_LocateUserInfoQuery{Type: 1}).
		Return(output, nil)

	h := NewLocateHandler(svc, slog.Default())

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.UserInfoQuery2(nil, nil, input.Frame, buf, responseWriter))
}
