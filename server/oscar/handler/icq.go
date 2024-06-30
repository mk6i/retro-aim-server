package handler

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type ICQService interface {
	DBQuery(_ context.Context, frame wire.SNACFrame, body wire.SNAC_0x0F_0x02_ICQDBQuery) wire.SNACMessage
}

func NewICQHandler(logger *slog.Logger, ICQService ICQService) ICQHandler {
	return ICQHandler{
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
		ICQService: ICQService,
	}
}

type ICQHandler struct {
	ICQService
	middleware.RouteLogger
}

func (rt ICQHandler) DBQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0F_0x02_ICQDBQuery{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	md, ok := inBody.Slice(0x01)
	if !ok {
		return errors.New("invalid ICQ frame")
	}

	icqMD := wire.ICQMetadata{}
	buf := bytes.NewBuffer(md)

	if err := binary.Read(buf, binary.LittleEndian, &icqMD.ChunkSize); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &icqMD.UIN); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &icqMD.ReqType); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &icqMD.Seq); err != nil {
		return err
	}

	switch icqMD.ReqType {
	case wire.ICQReqTypeOfflineMsg:
		fmt.Println("hello")
	case wire.ICQReqTypeDeleteMsg:
		fmt.Println("hello")
	case wire.ICQReqTypeInfo:
		if err := binary.Read(buf, binary.LittleEndian, &icqMD.ReqSubType); err != nil {
			return err
		}
		switch icqMD.ReqSubType {
		case 0x04D0:

			userInfo := ReqUserInfo{}
			if err := binary.Read(buf, binary.LittleEndian, &userInfo); err != nil {
				return err
			}
			requestInfo(userInfo)
			fmt.Println("hello")
		}
		fmt.Println("hello")
	}

	outSNAC := rt.ICQService.DBQuery(ctx, inFrame, inBody)
	rt.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

type ReqUserInfo struct {
	SearchUIN uint32
}

func requestInfo(userInfo ReqUserInfo) {

}
