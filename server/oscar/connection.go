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

// dispatchIncomingMessages receives incoming messages and sends them to the
// appropriate message handler. Messages from the client are sent to the
// router. Messages relayed from the user session are forwarded to the client.
// This function ensures that the same sequence number is incremented for both
// types of messages. The function terminates upon receiving a connection error
// or when the session closes.
//
// todo: this method has too many params and should be folded into a new type
func dispatchIncomingMessages(ctx context.Context, sess *state.Session, flapc *wire.FlapClient, r io.Reader, logger *slog.Logger, router Handler, config config.Config) error {
	defer func() {
		logger.InfoContext(ctx, "user disconnected")
	}()

	// buffered so that the go routine has room to exit
	msgCh := make(chan wire.FLAPFrame, 1)
	errCh := make(chan error, 1)

	// consume flap frames
	go func() {
		defer close(msgCh)
		defer close(errCh)

		for {
			frame := wire.FLAPFrame{}
			if err := wire.UnmarshalBE(&frame, r); err != nil {
				errCh <- err
				return
			}
			msgCh <- frame
		}
	}()

	for {
		select {
		case flap, ok := <-msgCh:
			if !ok {
				return nil
			}
			switch flap.FrameType {
			case wire.FLAPFrameData:
				flapBuf := bytes.NewBuffer(flap.Payload)

				inFrame := wire.SNACFrame{}
				if err := wire.UnmarshalBE(&inFrame, flapBuf); err != nil {
					return err
				}
				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := router.Handle(ctx, sess, inFrame, flapBuf, flapc); err != nil {
					middleware.LogRequestError(ctx, logger, inFrame, err)
					if errors.Is(err, ErrRouteNotFound) {
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
				return fmt.Errorf("shouldn't get FLAPFrameSignon. flap: %v", flap)
			case wire.FLAPFrameError:
				return fmt.Errorf("got FLAPFrameError. flap: %v", flap)
			case wire.FLAPFrameSignoff:
				logger.InfoContext(ctx, "got FLAPFrameSignoff", "flap", flap)
				return nil
			case wire.FLAPFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				return fmt.Errorf("got unknown FLAP frame type. flap: %v", flap)
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
		case err := <-errCh:
			if !errors.Is(io.EOF, err) {
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return nil
		}
	}
}
