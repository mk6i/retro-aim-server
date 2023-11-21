package server

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"io"
	"net"
)

type AuthHandler interface {
	ReceiveAndSendAuthChallenge(snacPayloadIn oscar.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (oscar.XMessage, error)
	ReceiveAndSendBUCPLoginRequest(snacPayloadIn oscar.SNAC_0x17_0x02_BUCPLoginRequest, newUUID func() uuid.UUID) (oscar.XMessage, error)
	RetrieveChatSession(ctx context.Context, chatID string, sessID string) (*state.Session, error)
	SendAndReceiveSignonFrame(rw io.ReadWriter, sequence *uint32) (oscar.FlapSignonFrame, error)
	Signout(ctx context.Context, sess *state.Session) error
	SignoutChat(ctx context.Context, sess *state.Session, chatID string)
	VerifyChatLogin(rw io.ReadWriter) (*ChatCookie, uint32, error)
	VerifyLogin(conn net.Conn) (*state.Session, uint32, error)
}

type BOSService struct {
	AlertRouter
	AuthHandler
	BuddyRouter
	ChatNavRouter
	FeedbagRouter
	ICBMRouter
	LocateRouter
	OServiceBOSRouter
	Cfg Config
	RouteLogger
}

type BOSServiceManager interface {
	Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32) error
	Signout(ctx context.Context, sess *state.Session) error
	VerifyLogin(conn net.Conn) (*state.Session, uint32, error)
	WriteOServiceHostOnline() oscar.XMessage
}

func (rt BOSService) Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32) error {
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, r); err != nil {
		return err
	}

	err := func() error {
		switch snac.FoodGroup {
		case oscar.OSERVICE:
			return rt.RouteOService(ctx, sess, snac, r, w, sequence)
		case oscar.LOCATE:
			return rt.RouteLocate(ctx, sess, snac, r, w, sequence)
		case oscar.BUDDY:
			return rt.RouteBuddy(ctx, snac, r, w, sequence)
		case oscar.ICBM:
			return rt.RouteICBM(ctx, sess, snac, r, w, sequence)
		case oscar.CHAT_NAV:
			return rt.RouteChatNav(ctx, sess, snac, r, w, sequence)
		case oscar.FEEDBAG:
			return rt.RouteFeedbag(ctx, sess, snac, r, w, sequence)
		case oscar.BUCP:
			return routeBUCP(ctx)
		case oscar.ALERT:
			return rt.RouteAlert(ctx, snac)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, snac, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(snac, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.Cfg.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}

type ChatService struct {
	AuthHandler
	ChatRouter
	Config Config
	OServiceChatRouter
	RouteLogger
}

type ChatServiceManager interface {
	RetrieveChatSession(ctx context.Context, chatID string, sessID string) (*state.Session, error)
	Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32, chatID string) error
	SignoutChat(ctx context.Context, sess *state.Session, chatID string)
	VerifyChatLogin(rw io.ReadWriter) (*ChatCookie, uint32, error)
	WriteOServiceHostOnline() oscar.XMessage
}

func (rt ChatService) Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32, chatID string) error {
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, r); err != nil {
		return err
	}

	err := func() error {
		switch snac.FoodGroup {
		case oscar.OSERVICE:
			return rt.RouteOService(ctx, sess, chatID, snac, r, w, sequence)
		case oscar.CHAT:
			return rt.RouteChat(ctx, sess, chatID, snac, r, w, sequence)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, snac, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(snac, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.Config.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}

func writeOutSNAC(originsnac oscar.SnacFrame, snacFrame oscar.SnacFrame, snacOut any, sequence *uint32, w io.Writer) error {
	if originsnac.RequestID != 0 {
		snacFrame.RequestID = originsnac.RequestID
	}

	snacBuf := &bytes.Buffer{}
	if err := oscar.Marshal(snacFrame, snacBuf); err != nil {
		return err
	}
	if err := oscar.Marshal(snacOut, snacBuf); err != nil {
		return err
	}

	flap := oscar.FlapFrame{
		StartMarker:   42,
		FrameType:     oscar.FlapFrameData,
		Sequence:      uint16(*sequence),
		PayloadLength: uint16(snacBuf.Len()),
	}

	if err := oscar.Marshal(flap, w); err != nil {
		return err
	}

	expectLen := snacBuf.Len()
	c, err := w.Write(snacBuf.Bytes())
	if err != nil {
		return err
	}
	if c != expectLen {
		panic("did not write the expected # of bytes")
	}

	*sequence++
	return nil
}

func sendInvalidSNACErr(snac oscar.SnacFrame, w io.Writer, sequence *uint32) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: snac.FoodGroup,
		SubGroup:  0x01, // error subgroup for all SNACs
	}
	snacPayloadOut := oscar.SnacError{
		Code: oscar.ErrorCodeInvalidSnac,
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
