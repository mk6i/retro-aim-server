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
	DeleteMsgReq(ctx context.Context, sess *state.Session, seq uint16) error
	FindByDetails(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails, seq uint16) error
	FindByEmail(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail, seq uint16) error
	FindByInterests(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages, seq uint16) error
	FindByUIN(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, seq uint16) error
	FullUserInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN, seq uint16) error
	OfflineMsgReq(ctx context.Context, sess *state.Session, seq uint16) error
	SetAffiliations(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations, seq uint16) error
	SetBasicInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo, seq uint16) error
	SetEmails(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails, seq uint16) error
	SetInterests(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests, seq uint16) error
	SetMoreInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo, seq uint16) error
	SetPermissions(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions, seq uint16) error
	SetUserNotes(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes, seq uint16) error
	SetWorkInfo(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo, seq uint16) error
	XMLReqData(ctx context.Context, sess *state.Session, req wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq, seq uint16) error
}

var (
	errUnknownICQMetaReqType    = errors.New("unknown ICQ request type")
	errUnknownICQMetaReqSubType = errors.New("unknown ICQ metadata request subtype")
)

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

func (rt ICQHandler) DBQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0F_0x02_BQuery{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}

	md, ok := inBody.Slice(wire.ICQTLVTagsMetadata)
	if !ok {
		return errors.New("invalid ICQ frame")
	}

	icqChunk := wire.ICQMessageRequestEnvelope{}
	if err := wire.UnmarshalLE(&icqChunk, bytes.NewBuffer(md)); err != nil {
		return err
	}
	buf := bytes.NewBuffer(icqChunk.Body)
	icqMD := wire.ICQMetadataWithSubType{}
	if err := wire.UnmarshalLE(&icqMD, buf); err != nil {
		return err
	}

	switch icqMD.ReqType {
	case wire.ICQDBQueryOfflineMsgReq:
		return rt.ICQService.OfflineMsgReq(ctx, sess, icqMD.Seq)
	case wire.ICQDBQueryDeleteMsgReq:
		return rt.ICQService.DeleteMsgReq(ctx, sess, icqMD.Seq)
	case wire.ICQDBQueryMetaReq:
		if icqMD.Optional == nil {
			return errors.New("got req without subtype")
		}
		rt.Logger.Debug("ICQ client request",
			"query_name", wire.ICQDBQueryName(icqMD.ReqType),
			"query_type", wire.ICQDBQueryMetaName(icqMD.Optional.ReqSubType),
			"uin", sess.UIN())

		switch icqMD.Optional.ReqSubType {
		case wire.ICQDBQueryMetaReqFullInfo, wire.ICQDBQueryMetaReqFullInfo2:
			userInfo := wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{}
			if err := binary.Read(buf, binary.LittleEndian, &userInfo); err != nil {
				return nil
			}
			return rt.ICQService.FullUserInfo(ctx, sess, userInfo, icqMD.Seq)
		case wire.ICQDBQueryMetaReqXMLReq:
			req := wire.ICQ_0x07D0_0x0898_DBQueryMetaReqXMLReq{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.XMLReqData(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetPermissions:
			req := wire.ICQ_0x07D0_0x0424_DBQueryMetaReqSetPermissions{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetPermissions(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchByUIN:
			req := wire.ICQ_0x07D0_0x051F_DBQueryMetaReqSearchByUIN{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByUIN(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchByEmail:
			req := wire.ICQ_0x07D0_0x0529_DBQueryMetaReqSearchByEmail{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByEmail(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchByDetails:
			req := wire.ICQ_0x07D0_0x0515_DBQueryMetaReqSearchByDetails{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByDetails(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchWhitePages:
			req := wire.ICQ_0x07D0_0x0533_DBQueryMetaReqSearchWhitePages{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByInterests(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetBasicInfo:
			req := wire.ICQ_0x07D0_0x03EA_DBQueryMetaReqSetBasicInfo{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetBasicInfo(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetWorkInfo:
			req := wire.ICQ_0x07D0_0x03F3_DBQueryMetaReqSetWorkInfo{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetWorkInfo(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetMoreInfo:
			req := wire.ICQ_0x07D0_0x03FD_DBQueryMetaReqSetMoreInfo{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetMoreInfo(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetNotes:
			req := wire.ICQ_0x07D0_0x0406_DBQueryMetaReqSetNotes{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetUserNotes(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetEmails:
			req := wire.ICQ_0x07D0_0x040B_DBQueryMetaReqSetEmails{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetEmails(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetInterests:
			req := wire.ICQ_0x07D0_0x0410_DBQueryMetaReqSetInterests{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetInterests(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetAffiliations:
			req := wire.ICQ_0x07D0_0x041A_DBQueryMetaReqSetAffiliations{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.SetAffiliations(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqStat0a8c,
			wire.ICQDBQueryMetaReqStat0a96,
			wire.ICQDBQueryMetaReqStat0aaa,
			wire.ICQDBQueryMetaReqStat0ab4,
			wire.ICQDBQueryMetaReqStat0ab9,
			wire.ICQDBQueryMetaReqStat0abe,
			wire.ICQDBQueryMetaReqStat0ac8,
			wire.ICQDBQueryMetaReqStat0acd,
			wire.ICQDBQueryMetaReqStat0ad2,
			wire.ICQDBQueryMetaReqStat0ad7,
			wire.ICQDBQueryMetaReqStat0758:
			rt.Logger.Debug("got a request for stats, not doing anything right now")
		default:
			return fmt.Errorf("%w: %X", errUnknownICQMetaReqSubType, icqMD.Optional.ReqSubType)
		}
	default:
		return fmt.Errorf("%w: %X", errUnknownICQMetaReqType, icqMD.ReqType)
	}

	return nil
}
