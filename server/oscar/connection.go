package oscar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type incomingMessage struct {
	flap    wire.FLAPFrame
	payload *bytes.Buffer
}

type flapClient struct {
	sequence uint32
	w        io.Writer
	r        io.Reader
}

func (f *flapClient) SignonHandshake() (wire.FLAPSignonFrame, error) {
	// send FLAPFrameSignon to client
	flap := wire.FLAPFrame{
		StartMarker:   42,
		FrameType:     wire.FLAPFrameSignon,
		Sequence:      uint16(f.sequence),
		PayloadLength: 4, // size of FLAPSignonFrame
	}
	if err := wire.Marshal(flap, f.w); err != nil {
		return wire.FLAPSignonFrame{}, err
	}
	flapSignonFrameOut := wire.FLAPSignonFrame{
		FLAPVersion: 1,
	}
	if err := wire.Marshal(flapSignonFrameOut, f.w); err != nil {
		return wire.FLAPSignonFrame{}, err
	}

	// receive FLAPFrameSignon from client
	flap = wire.FLAPFrame{}
	if err := wire.Unmarshal(&flap, f.r); err != nil {
		return wire.FLAPSignonFrame{}, err
	}
	buf, err := flap.ReadBody(f.r)
	if err != nil {
		return wire.FLAPSignonFrame{}, err
	}
	flapSignonFrameIn := wire.FLAPSignonFrame{}
	if err := wire.Unmarshal(&flapSignonFrameIn, buf); err != nil {
		return wire.FLAPSignonFrame{}, err
	}

	f.sequence++

	return flapSignonFrameIn, nil
}

// SendSignoffFrame sends a sign-off FLAP frame with attached TLVs as the last
// request sent in the FLAP auth flow. This is unrelated to the Disconnect()
// method, which sends a sign-off frame to terminate a BOS connection.
// todo: combine this method with Disconnect()
func (f *flapClient) SendSignoffFrame(tlvs wire.TLVRestBlock) error {
	tlvBuf := &bytes.Buffer{}
	if err := wire.Marshal(tlvs, tlvBuf); err != nil {
		return err
	}

	flap := wire.FLAPFrame{
		StartMarker:   42,
		FrameType:     wire.FLAPFrameSignoff,
		Sequence:      uint16(f.sequence),
		PayloadLength: uint16(tlvBuf.Len()),
	}

	if err := wire.Marshal(flap, f.w); err != nil {
		return err
	}

	expectLen := tlvBuf.Len()
	c, err := f.w.Write(tlvBuf.Bytes())
	if err != nil {
		return err
	}
	if c != expectLen {
		panic("did not write the expected # of bytes")
	}

	f.sequence++
	return nil
}

func (f *flapClient) SendSNAC(frame wire.SNACFrame, body any) error {
	snacBuf := &bytes.Buffer{}
	if err := wire.Marshal(frame, snacBuf); err != nil {
		return err
	}
	if err := wire.Marshal(body, snacBuf); err != nil {
		return err
	}

	flap := wire.FLAPFrame{
		StartMarker:   42,
		FrameType:     wire.FLAPFrameData,
		Sequence:      uint16(f.sequence),
		PayloadLength: uint16(snacBuf.Len()),
	}

	if err := wire.Marshal(flap, f.w); err != nil {
		return err
	}

	expectLen := snacBuf.Len()
	c, err := f.w.Write(snacBuf.Bytes())
	if err != nil {
		return err
	}
	if c != expectLen {
		panic("did not write the expected # of bytes")
	}

	f.sequence++
	return nil
}

func (f *flapClient) ReceiveSNAC(frame *wire.SNACFrame, body any) error {
	flap := wire.FLAPFrame{}
	if err := wire.Unmarshal(&flap, f.r); err != nil {
		return err
	}
	buf, err := flap.ReadBody(f.r)
	if err != nil {
		return err
	}
	if err := wire.Unmarshal(frame, buf); err != nil {
		return err
	}
	return wire.Unmarshal(body, buf)
}

func (f *flapClient) Disconnect() error {
	// gracefully disconnect so that the client does not try to
	// reconnect when the connection closes.
	flap := wire.FLAPFrame{
		StartMarker:   42,
		FrameType:     wire.FLAPFrameSignoff,
		Sequence:      uint16(f.sequence),
		PayloadLength: uint16(0),
	}
	return wire.Marshal(flap, f.w)
}

func sendInvalidSNACErr(frameIn wire.SNACFrame, rw ResponseWriter) error {
	frameOut := wire.SNACFrame{
		FoodGroup: frameIn.FoodGroup,
		SubGroup:  0x01, // error subgroup for all SNACs
		RequestID: frameIn.RequestID,
	}
	bodyOut := wire.SNACError{
		Code: wire.ErrorCodeInvalidSnac,
	}
	return rw.SendSNAC(frameOut, bodyOut)
}

func consumeFLAPFrames(r io.Reader, msgCh chan incomingMessage, errCh chan error) {
	defer close(msgCh)
	defer close(errCh)

	for {
		in := incomingMessage{}
		if err := wire.Unmarshal(&in.flap, r); err != nil {
			errCh <- err
			return
		}

		if in.flap.PayloadLength > 0 {
			buf := make([]byte, in.flap.PayloadLength)
			if _, err := io.ReadFull(r, buf); err != nil {
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
func dispatchIncomingMessages(ctx context.Context, sess *state.Session, flapc *flapClient, r io.Reader, logger *slog.Logger, router Handler, config config.Config) error {
	// buffered so that the go routine has room to exit
	msgCh := make(chan incomingMessage, 1)
	readErrCh := make(chan error, 1)
	go consumeFLAPFrames(r, msgCh, readErrCh)

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
			case wire.FLAPFrameData:
				inFrame := wire.SNACFrame{}
				if err := wire.Unmarshal(&inFrame, m.payload); err != nil {
					return err
				}
				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := router.Handle(ctx, sess, inFrame, m.payload, flapc); err != nil {
					middleware.LogRequestError(ctx, logger, inFrame, err)
					if errors.Is(err, ErrRouteNotFound) || errors.Is(err, wire.ErrUnsupportedFoodGroup) {
						if err1 := sendInvalidSNACErr(inFrame, flapc); err1 != nil {
							return errors.Join(err1, err)
						}
						if config.FailFast {
							panic(err.Error())
						}
						break
					}
					return err
				}
			case wire.FLAPFrameSignon:
				return fmt.Errorf("shouldn't get FLAPFrameSignon. flap: %v", m.flap)
			case wire.FLAPFrameError:
				return fmt.Errorf("got FLAPFrameError. flap: %v", m.flap)
			case wire.FLAPFrameSignoff:
				logger.InfoContext(ctx, "got FLAPFrameSignoff", "flap", m.flap)
				return nil
			case wire.FLAPFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				return fmt.Errorf("got unknown FLAP frame type. flap: %v", m.flap)
			}
		case m := <-sess.ReceiveMessage():
			// forward a notification sent from another client to this client
			if err := flapc.SendSNAC(m.Frame, m.Body); err != nil {
				middleware.LogRequestError(ctx, logger, m.Frame, err)
				return err
			}
			middleware.LogRequest(ctx, logger, m.Frame, m.Body)
		case <-sess.Closed():
			// gracefully disconnect so that the client does not try to
			// reconnect when the connection closes.
			if err := flapc.Disconnect(); err != nil {
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
