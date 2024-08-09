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
	FindByDetails(ctx context.Context, sess *state.Session, req wire.ICQFindByDetails, seq uint16) error
	FindByEmail(ctx context.Context, sess *state.Session, req wire.ICQFindByEmail, seq uint16) error
	FindByUIN(ctx context.Context, sess *state.Session, req wire.ICQFindByUIN, seq uint16) error
	FindByWhitePages(ctx context.Context, sess *state.Session, req wire.ICQFindByWhitePages, seq uint16) error
	GetICQFullUserInfo(ctx context.Context, sess *state.Session, req wire.ICQFindByUIN, seq uint16) error
	GetICQMessagesEOF(ctx context.Context, sess *state.Session, seq uint16) error
	UpdateBasicInfo(ctx context.Context, sess *state.Session, req wire.ICQUserInfoBasic, seq uint16) error
	UpdateWorkInfo(ctx context.Context, sess *state.Session, req wire.ICQWorkInfo, seq uint16) error
	UpdateMoreInfo(ctx context.Context, sess *state.Session, req wire.SomeMoreUserInfo, seq uint16) error
	UpdateUserNotes(ctx context.Context, sess *state.Session, req wire.ICQNotes, seq uint16) error
	UpdateInterests(ctx context.Context, sess *state.Session, req wire.ICQInterests, seq uint16) error
	UpdateAffiliations(ctx context.Context, sess *state.Session, req wire.ICQAffiliations, seq uint16) error
	UpdateEmails(ctx context.Context, sess *state.Session, req wire.ICQEmailUserInfo, seq uint16) error
	UpdatePermissions(ctx context.Context, sess *state.Session, req wire.ICQInfoSetPerms, seq uint16) error
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

func (rt ICQHandler) DBQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0F_0x02_ICQDBQuery{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}

	md, ok := inBody.Slice(wire.ICQTLVTagsMetadata)
	if !ok {
		return errors.New("invalid ICQ frame")
	}

	icqChunk := wire.ICQChunk{}
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
		return rt.ICQService.GetICQMessagesEOF(ctx, sess, icqMD.Seq)
	case wire.ICQDBQueryDeleteMsgReq:
		fmt.Println("hello")
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
			userInfo := wire.ICQFindByUIN{}
			if err := binary.Read(buf, binary.LittleEndian, &userInfo); err != nil {
				return nil
			}
			return rt.ICQService.GetICQFullUserInfo(ctx, sess, userInfo, icqMD.Seq)
		case wire.ICQDBQueryMetaReqXMLReq:
			req := wire.ICQXMLReqData{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetPermissions:
			req := wire.ICQInfoSetPerms{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdatePermissions(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchByUIN:
			req := wire.ICQFindByUIN{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByUIN(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchByEmail:
			req := wire.ICQFindByEmail{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByEmail(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchByDetails:
			req := wire.ICQFindByDetails{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByDetails(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSearchWhitePages:
			req := wire.ICQFindByWhitePages{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.FindByWhitePages(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetBasicInfo:
			req := wire.ICQUserInfoBasic{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdateBasicInfo(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetWorkInfo:
			req := wire.ICQWorkInfo{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdateWorkInfo(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetMoreInfo:
			req := wire.SomeMoreUserInfo{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdateMoreInfo(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetNotes:
			req := wire.ICQNotes{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdateUserNotes(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetEmails:
			req := wire.ICQEmailUserInfo{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdateEmails(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetInterests:
			req := wire.ICQInterests{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdateInterests(ctx, sess, req, icqMD.Seq); err != nil {
				return err
			}
		case wire.ICQDBQueryMetaReqSetAffiliations:
			req := wire.ICQAffiliations{}
			if err := wire.UnmarshalLE(&req, buf); err != nil {
				return err
			}
			if err := rt.ICQService.UpdateAffiliations(ctx, sess, req, icqMD.Seq); err != nil {
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
			return fmt.Errorf("unknown request subtype %X", icqMD.Optional.ReqSubType)
		}
	default:
		return fmt.Errorf("unknown request type %X", icqMD.ReqType)
	}

	return nil
}
