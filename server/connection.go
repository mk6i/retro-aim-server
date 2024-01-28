package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

var (
	ErrUnsupportedSubGroup = errors.New("unimplemented subgroup, your client version may be unsupported")
)

// Router is the interface for methods that route food group requests.
type Router interface {
	Route(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error
}

// snacSender is the function that packages a SNAC frame and body into a FLAP
// message and writes it to the output writer.
type snacSender func(frame oscar.SNACFrame, body any, sequence *uint32, w io.Writer) error

type incomingMessage struct {
	flap    oscar.FLAPFrame
	payload *bytes.Buffer
}

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
	buf, err := flap.ReadBody(r)
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

// dispatchIncomingMessages receives incoming messages and sends them to the
// appropriate message handler. Messages from the client are sent to the
// router. Messages relayed from the user session are forwarded to the client.
// This function ensures that the same sequence number is incremented for both
// types of messages. The function terminates upon receiving a connection error
// or when the session closes.
//
// todo: this method has too many params and should be folded into a new type
func dispatchIncomingMessages(ctx context.Context, sess *state.Session, seq uint32, rw io.ReadWriter, logger *slog.Logger, router Router, sendSNAC snacSender, config config.Config) error {
	// buffered so that the go routine has room to exit
	msgCh := make(chan incomingMessage, 1)
	readErrCh := make(chan error, 1)
	go consumeFLAPFrames(rw, msgCh, readErrCh)

	defer func() {
		logger.InfoContext(ctx, "user disconnected")
	}()

	for {
		select {
		case m, ok := <-msgCh:
			if !ok {
				return nil
			}
			switch m.flap.FrameType {
			case oscar.FLAPFrameData:
				inFrame := oscar.SNACFrame{}
				if err := oscar.Unmarshal(&inFrame, m.payload); err != nil {
					return err
				}
				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := router.Route(ctx, sess, inFrame, m.payload, rw, &seq); err != nil {
					logRequestError(ctx, logger, inFrame, err)
					if errors.Is(err, ErrUnsupportedSubGroup) {
						if err1 := sendInvalidSNACErr(inFrame, rw, &seq); err1 != nil {
							return errors.Join(err1, err)
						}
						if config.FailFast {
							panic(err.Error())
						}
						break
					}
					return err
				}
			case oscar.FLAPFrameSignon:
				return fmt.Errorf("shouldn't get FLAPFrameSignon. flap: %v", m.flap)
			case oscar.FLAPFrameError:
				return fmt.Errorf("got FLAPFrameError. flap: %v", m.flap)
			case oscar.FLAPFrameSignoff:
				logger.InfoContext(ctx, "got FLAPFrameSignoff", "flap", m.flap)
				return nil
			case oscar.FLAPFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				return fmt.Errorf("got unknown FLAP frame type. flap: %v", m.flap)
			}
		case m := <-sess.ReceiveMessage():
			// forward a notification sent from another client to this client
			if err := sendSNAC(m.Frame, m.Body, &seq, rw); err != nil {
				logRequestError(ctx, logger, m.Frame, err)
				return err
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
				return fmt.Errorf("unable to gracefully disconnect user. %w", err)
			}
			return nil
		case err := <-readErrCh:
			if !errors.Is(io.EOF, err) {
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return nil
		}
	}
}
