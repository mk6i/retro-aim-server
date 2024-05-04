package oscar

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestAlertServer_handleNewConnection(t *testing.T) {
	sess := state.NewSession()
	sess.SetID("login-cookie-1234")

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	go func() {
		// < receive FLAPSignonFrame
		flap := wire.FLAPFrame{}
		assert.NoError(t, wire.Unmarshal(&flap, serverReader))
		buf, err := flap.ReadBody(serverReader)
		assert.NoError(t, err)
		flapSignonFrame := wire.FLAPSignonFrame{}
		assert.NoError(t, wire.Unmarshal(&flapSignonFrame, buf))

		// > send FLAPSignonFrame
		flapSignonFrame = wire.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		flapSignonFrame.Append(wire.NewTLV(wire.OServiceTLVTagsLoginCookie, []byte(sess.ID())))
		buf = &bytes.Buffer{}
		assert.NoError(t, wire.Marshal(flapSignonFrame, buf))
		flap = wire.FLAPFrame{
			StartMarker:   42,
			FrameType:     wire.FLAPFrameSignon,
			PayloadLength: uint16(buf.Len()),
		}
		assert.NoError(t, wire.Marshal(flap, serverWriter))
		_, err = serverWriter.Write(buf.Bytes())
		assert.NoError(t, err)

		// < receive SNAC_0x01_0x03_OServiceHostOnline
		flap = wire.FLAPFrame{}
		assert.NoError(t, wire.Unmarshal(&flap, serverReader))
		buf, err = flap.ReadBody(serverReader)
		assert.NoError(t, err)
		frame := wire.SNACFrame{}
		assert.NoError(t, wire.Unmarshal(&frame, buf))
		body := wire.SNAC_0x01_0x03_OServiceHostOnline{}
		assert.NoError(t, wire.Unmarshal(&body, buf))

		// send the first request that should get relayed to BOSRouter.Handle
		flapc := flapClient{
			w: serverWriter,
		}
		frame = wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceClientOnline,
		}
		assert.NoError(t, flapc.SendSNAC(frame, struct{}{}))
		assert.NoError(t, serverWriter.Close())
	}()

	authService := newMockAuthService(t)
	authService.EXPECT().
		RetrieveBOSSession(sess.ID()).
		Return(sess, nil)
	authService.EXPECT().
		Signout(mock.Anything, sess).
		Return(nil)

	onlineNotifier := newMockOnlineNotifier(t)
	onlineNotifier.EXPECT().
		HostOnline().
		Return(wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceHostOnline,
			},
			Body: wire.SNAC_0x01_0x03_OServiceHostOnline{},
		})

	router := newMockHandler(t)
	router.EXPECT().
		Handle(mock.Anything, sess, mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) {
			assert.Equal(t, wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceClientOnline,
			}, inFrame)
		}).Return(nil)

	rt := AlertServer{
		AuthService:    authService,
		Handler:        router,
		Logger:         slog.Default(),
		OnlineNotifier: onlineNotifier,
	}
	rwc := pipeRWC{
		PipeReader: clientReader,
		PipeWriter: clientWriter,
	}
	assert.NoError(t, rt.handleNewConnection(context.Background(), rwc))
}
