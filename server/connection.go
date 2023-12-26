package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

var (
	ErrUnsupportedSubGroup = errors.New("unimplemented subgroup, your client version may be unsupported")
)

type (
	incomingMessage struct {
		flap    oscar.FLAPFrame
		payload *bytes.Buffer
	}
	alertHandler     func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, u *uint32) error
	clientReqHandler func(ctx context.Context, r io.Reader, w io.Writer, u *uint32) error
)

func sendSNAC(frame oscar.SNACFrame, body any, sequence *uint32, w io.Writer) error {
	snacBuf := &bytes.Buffer{}
	if err := oscar.Marshal(frame, snacBuf); err != nil {
		return err
	}
	if err := oscar.Marshal(body, snacBuf); err != nil {
		return err
	}

	flap := oscar.FLAPFrame{
		StartMarker:   42,
		FrameType:     oscar.FLAPFrameData,
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

func receiveSNAC(frame *oscar.SNACFrame, body any, r io.Reader) error {
	flap := oscar.FLAPFrame{}
	if err := oscar.Unmarshal(&flap, r); err != nil {
		return err
	}
	buf, err := flap.SNACBuffer(r)
	if err != nil {
		return err
	}
	if err := oscar.Unmarshal(frame, buf); err != nil {
		return err
	}
	return oscar.Unmarshal(body, buf)
}

func sendInvalidSNACErr(frameIn oscar.SNACFrame, w io.Writer, sequence *uint32) error {
	frameOut := oscar.SNACFrame{
		FoodGroup: frameIn.FoodGroup,
		SubGroup:  0x01, // error subgroup for all SNACs
		RequestID: frameIn.RequestID,
	}
	bodyOut := oscar.SNACError{
		Code: oscar.ErrorCodeInvalidSnac,
	}
	return sendSNAC(frameOut, bodyOut, sequence, w)
}

func consumeFLAPFrames(r io.Reader, msgCh chan incomingMessage, errCh chan error) {
	defer close(msgCh)
	defer close(errCh)

	for {
		in := incomingMessage{}
		if err := oscar.Unmarshal(&in.flap, r); err != nil {
			errCh <- err
			return
		}

		if in.flap.FrameType == oscar.FLAPFrameData {
			buf := make([]byte, in.flap.PayloadLength)
			if _, err := r.Read(buf); err != nil {
				errCh <- err
				return
			}
			in.payload = bytes.NewBuffer(buf)
		}

		msgCh <- in
	}
}

func dispatchIncomingMessages(ctx context.Context, sess *state.Session, seq uint32, rw io.ReadWriter, logger *slog.Logger, fn clientReqHandler, alertHandler alertHandler) {
	// buffered so that the go routine has room to exit
	msgCh := make(chan incomingMessage, 1)
	readErrCh := make(chan error, 1)
	go consumeFLAPFrames(rw, msgCh, readErrCh)

	defer func() {
		logger.InfoContext(ctx, "user disconnected")
	}()

	for {
		select {
		case m := <-msgCh:
			switch m.flap.FrameType {
			case oscar.FLAPFrameData:
				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := fn(ctx, m.payload, rw, &seq); err != nil {
					return
				}
			case oscar.FLAPFrameSignon:
				logger.ErrorContext(ctx, "shouldn't get FLAPFrameSignon", "flap", m.flap)
			case oscar.FLAPFrameError:
				logger.ErrorContext(ctx, "got FLAPFrameError", "flap", m.flap)
				return
			case oscar.FLAPFrameSignoff:
				logger.InfoContext(ctx, "got FLAPFrameSignoff", "flap", m.flap)
				return
			case oscar.FLAPFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				logger.ErrorContext(ctx, "got unknown FLAP frame type", "flap", m.flap)
				return
			}
		case m := <-sess.ReceiveMessage():
			// forward a notification sent from another client to this client
			if err := alertHandler(ctx, m, rw, &seq); err != nil {
				logRequestError(ctx, logger, m.Frame, err)
				return
			}
			logRequest(ctx, logger, m.Frame, m.Body)
		case <-sess.Closed():
			// gracefully disconnect so that the client does not try to
			// reconnect when the connection closes.
			flap := oscar.FLAPFrame{
				StartMarker:   42,
				FrameType:     oscar.FLAPFrameSignoff,
				Sequence:      uint16(seq),
				PayloadLength: uint16(0),
			}
			if err := oscar.Marshal(flap, rw); err != nil {
				logger.ErrorContext(ctx, "unable to gracefully disconnect user", "err", err)
			}
			return
		case err := <-readErrCh:
			if !errors.Is(io.EOF, err) {
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return
		}
	}
}
