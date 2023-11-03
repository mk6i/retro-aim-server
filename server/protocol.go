package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log/slog"
)

const (
	ErrorCodeInvalidSnac          uint16 = 0x01
	ErrorCodeRateToHost           uint16 = 0x02
	ErrorCodeRateToClient         uint16 = 0x03
	ErrorCodeNotLoggedOn          uint16 = 0x04
	ErrorCodeServiceUnavailable   uint16 = 0x05
	ErrorCodeServiceNotDefined    uint16 = 0x06
	ErrorCodeObsoleteSnac         uint16 = 0x07
	ErrorCodeNotSupportedByHost   uint16 = 0x08
	ErrorCodeNotSupportedByClient uint16 = 0x09
	ErrorCodeRefusedByClient      uint16 = 0x0A
	ErrorCodeReplyTooBig          uint16 = 0x0B
	ErrorCodeResponsesLost        uint16 = 0x0C
	ErrorCodeRequestDenied        uint16 = 0x0D
	ErrorCodeBustedSnacPayload    uint16 = 0x0E
	ErrorCodeInsufficientRights   uint16 = 0x0F
	ErrorCodeInLocalPermitDeny    uint16 = 0x10
	ErrorCodeTooEvilSender        uint16 = 0x11
	ErrorCodeTooEvilReceiver      uint16 = 0x12
	ErrorCodeUserTempUnavail      uint16 = 0x13
	ErrorCodeNoMatch              uint16 = 0x14
	ErrorCodeListOverflow         uint16 = 0x15
	ErrorCodeRequestAmbigous      uint16 = 0x16
	ErrorCodeQueueFull            uint16 = 0x17
	ErrorCodeNotWhileOnAol        uint16 = 0x18
	ErrorCodeQueryFail            uint16 = 0x19
	ErrorCodeTimeout              uint16 = 0x1A
	ErrorCodeErrorText            uint16 = 0x1B
	ErrorCodeGeneralFailure       uint16 = 0x1C
	ErrorCodeProgress             uint16 = 0x1D
	ErrorCodeInFreeArea           uint16 = 0x1E
	ErrorCodeRestrictedByPc       uint16 = 0x1F
	ErrorCodeRemoteRestrictedByPc uint16 = 0x20
)

const (
	ErrorTagsFailUrl        = 0x04
	ErrorTagsErrorSubcode   = 0x08
	ErrorTagsErrorText      = 0x1B
	ErrorTagsErrorInfoClsid = 0x29
	ErrorTagsErrorInfoData  = 0x2A
)

var (
	CapChat, _ = uuid.MustParse("748F2420-6287-11D1-8222-444553540000").MarshalBinary()
)

var (
	ErrUnsupportedFoodGroup = errors.New("unimplemented food group, your client version may be unsupported")
	ErrUnsupportedSubGroup  = errors.New("unimplemented subgroup, your client version may be unsupported")
)

type Config struct {
	BOSPort     int    `envconfig:"BOS_PORT" default:"5191"`
	ChatPort    int    `envconfig:"CHAT_PORT" default:"5192"`
	DBPath      string `envconfig:"DB_PATH" required:"true"`
	DisableAuth bool   `envconfig:"DISABLE_AUTH" default:"false"`
	FailFast    bool   `envconfig:"FAIL_FAST" default:"false"`
	OSCARHost   string `envconfig:"OSCAR_HOST" required:"true"`
	OSCARPort   int    `envconfig:"OSCAR_PORT" default:"5190"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

func Address(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func SendAndReceiveSignonFrame(rw io.ReadWriter, sequence *uint32) (oscar.FlapSignonFrame, error) {
	flapFrameOut := oscar.FlapFrame{
		StartMarker:   42,
		FrameType:     oscar.FlapFrameSignon,
		Sequence:      uint16(*sequence),
		PayloadLength: 4, // size of FlapSignonFrame
	}
	if err := oscar.Marshal(flapFrameOut, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	flapSignonFrameOut := oscar.FlapSignonFrame{
		FlapVersion: 1,
	}
	if err := oscar.Marshal(flapSignonFrameOut, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}

	// receive
	flapFrameIn := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flapFrameIn, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	b := make([]byte, flapFrameIn.PayloadLength)
	if _, err := rw.Read(b); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	flapSignonFrameIn := oscar.FlapSignonFrame{}
	if err := oscar.Unmarshal(&flapSignonFrameIn, bytes.NewBuffer(b)); err != nil {
		return oscar.FlapSignonFrame{}, err
	}

	*sequence++

	return flapSignonFrameIn, nil
}

func VerifyLogin(sm SessionManager, rw io.ReadWriter) (*Session, uint32, error) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(rw, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	ID, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session ID from payload")
	}

	sess, ok := sm.Retrieve(string(ID))
	if !ok {
		return nil, 0, fmt.Errorf("unable to find session by ID %s", ID)
	}

	return sess, seq, nil
}

func VerifyChatLogin(rw io.ReadWriter) (*ChatCookie, uint32, error) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(rw, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	buf, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session ID from payload")
	}

	cookie := ChatCookie{}
	err = oscar.Unmarshal(&cookie, bytes.NewBuffer(buf))

	return &cookie, seq, err
}

type IncomingMessage struct {
	flap oscar.FlapFrame
	snac oscar.SnacFrame
	buf  io.Reader
}

type XMessage struct {
	snacFrame oscar.SnacFrame
	snacOut   any
}

func sendInvalidSNACErr(snac oscar.SnacFrame, w io.Writer, sequence *uint32) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: snac.FoodGroup,
		SubGroup:  0x01, // error subgroup for all SNACs
	}
	snacPayloadOut := oscar.SnacError{
		Code: ErrorCodeInvalidSnac,
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func readIncomingRequests(ctx context.Context, logger *slog.Logger, rw io.Reader, msgCh chan IncomingMessage, errCh chan error) {
	defer close(msgCh)
	defer close(errCh)

	for {
		flap := oscar.FlapFrame{}
		if err := oscar.Unmarshal(&flap, rw); err != nil {
			errCh <- err
			return
		}

		switch flap.FrameType {
		case oscar.FlapFrameSignon:
			errCh <- errors.New("shouldn't get FlapFrameSignon")
			return
		case oscar.FlapFrameData:
			b := make([]byte, flap.PayloadLength)
			if _, err := rw.Read(b); err != nil {
				errCh <- err
				return
			}

			snac := oscar.SnacFrame{}
			buf := bytes.NewBuffer(b)
			if err := oscar.Unmarshal(&snac, buf); err != nil {
				errCh <- err
				return
			}

			msgCh <- IncomingMessage{
				flap: flap,
				snac: snac,
				buf:  buf,
			}
		case oscar.FlapFrameError:
			errCh <- fmt.Errorf("got FlapFrameError: %v", flap)
			return
		case oscar.FlapFrameSignoff:
			errCh <- ErrSignedOff
			return
		case oscar.FlapFrameKeepAlive:
			logger.DebugContext(ctx, "keepalive heartbeat")
		default:
			errCh <- fmt.Errorf("unknown frame type: %v", flap)
			return
		}
	}
}

func Signout(ctx context.Context, logger *slog.Logger, sess *Session, sm SessionManager, fm *FeedbagStore) {
	if err := BroadcastDeparture(ctx, sess, sm, fm); err != nil {
		logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
	}
	sm.Remove(sess)
}

func ReadBos(ctx context.Context, cfg Config, sess *Session, seq uint32, sm SessionManager, fm *FeedbagStore, cr *ChatRegistry, rwc io.ReadWriter, room ChatRoom, router Router, logger *slog.Logger) {
	if err := router.WriteOServiceHostOnline(rwc, &seq); err != nil {
		logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
	}

	// buffered so that the go routine has room to exit
	msgCh := make(chan IncomingMessage, 1)
	errCh := make(chan error, 1)
	go readIncomingRequests(ctx, logger, rwc, msgCh, errCh)

	rl := RouteLogger{
		Logger: logger,
	}

	for {
		select {
		case m := <-msgCh:
			if err := router.routeIncomingRequests(ctx, cfg, sm, sess, fm, cr, rwc, &seq, m.snac, m.buf, room); err != nil {
				if errors.Is(err, ErrUnsupportedSubGroup) || errors.Is(err, ErrUnsupportedFoodGroup) {
					if err1 := sendInvalidSNACErr(m.snac, rwc, &seq); err1 != nil {
						err = errors.Join(err1, err)
					}
					if cfg.FailFast {
						panic(err.Error())
					}
				}
				logRequestError(ctx, logger, m.snac, err)
				return
			}
		case m := <-sess.RecvMessage():
			if err := writeOutSNAC(oscar.SnacFrame{}, m.snacFrame, m.snacOut, &seq, rwc); err != nil {
				logRequestError(ctx, logger, m.snacFrame, err)
				return
			}
			rl.logRequest(ctx, m.snacFrame, m.snacOut)
		case <-sess.Closed():
			if err := gracefulDisconnect(seq, rwc); err != nil {
				logger.ErrorContext(ctx, "unable to gracefully disconnect user", "err", err)
			}
			return
		case err := <-errCh:
			switch {
			case errors.Is(io.EOF, err):
				fallthrough
			case errors.Is(ErrSignedOff, err):
				logger.InfoContext(ctx, "client signed off")
			default:
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return
		}
	}
}

func logRequestError(ctx context.Context, logger *slog.Logger, inFrame oscar.SnacFrame, err error) {
	logger.LogAttrs(ctx, slog.LevelError, "client disconnected with error",
		slog.Group("request",
			slog.String("food_group", oscar.FoodGroupStr(inFrame.FoodGroup)),
			slog.String("sub_group", oscar.SubGroupStr(inFrame.FoodGroup, inFrame.SubGroup)),
		),
		slog.String("err", err.Error()),
	)
}

func gracefulDisconnect(seq uint32, rwc io.ReadWriter) error {
	return oscar.Marshal(oscar.FlapFrame{
		StartMarker: 42,
		FrameType:   oscar.FlapFrameSignoff,
		Sequence:    uint16(seq),
	}, rwc)
}

func NewRouter(logger *slog.Logger) Router {
	return Router{
		AlertRouter:    NewAlertRouter(logger),
		BuddyRouter:    NewBuddyRouter(logger),
		ChatNavRouter:  NewChatNavRouter(logger),
		ChatRouter:     NewChatRouter(logger),
		FeedbagRouter:  NewFeedbagRouter(logger),
		ICBMRouter:     NewICBMRouter(logger),
		LocateRouter:   NewLocateRouter(logger),
		OServiceRouter: NewOServiceRouter(logger),
	}
}

func NewRouterForChat(logger *slog.Logger) Router {
	r := NewRouter(logger)
	r.OServiceRouter = NewOServiceRouterForChat(logger)
	return r
}

type Router struct {
	AlertRouter
	BuddyRouter
	ChatNavRouter
	ChatRouter
	FeedbagRouter
	ICBMRouter
	LocateRouter
	OServiceRouter
}

func (rt *Router) routeIncomingRequests(ctx context.Context, cfg Config, sm SessionManager, sess *Session, fm *FeedbagStore, cr *ChatRegistry, rw io.ReadWriter, sequence *uint32, snac oscar.SnacFrame, buf io.Reader, room ChatRoom) error {
	switch snac.FoodGroup {
	case oscar.OSERVICE:
		return rt.RouteOService(ctx, cfg, cr, sm, fm, sess, room, snac, buf, rw, sequence)
	case oscar.LOCATE:
		return rt.RouteLocate(ctx, sess, sm, fm, snac, buf, rw, sequence)
	case oscar.BUDDY:
		return rt.RouteBuddy(ctx, snac, buf, rw, sequence)
	case oscar.ICBM:
		return rt.RouteICBM(ctx, sm, fm, sess, snac, buf, rw, sequence)
	case oscar.CHAT_NAV:
		return rt.RouteChatNav(ctx, sess, cr, snac, buf, rw, sequence)
	case oscar.FEEDBAG:
		return rt.RouteFeedbag(ctx, sm, sess, fm, snac, buf, rw, sequence)
	case oscar.BUCP:
		return routeBUCP(ctx)
	case oscar.CHAT:
		return rt.RouteChat(ctx, sess, sm, snac, buf, rw, sequence)
	case oscar.ALERT:
		return rt.RouteAlert(ctx, snac)
	default:
		return ErrUnsupportedFoodGroup
	}
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
