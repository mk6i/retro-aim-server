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

// pipeRWC provides a mock for ReadWriteCloser that uses pipes instead of TCP
// connections
type pipeRWC struct {
	*io.PipeReader
	*io.PipeWriter
}

func (m pipeRWC) Close() error {
	if err := m.PipeReader.Close(); err != nil {
		return err
	}
	return m.PipeWriter.Close()
}

func TestBOSService_handleNewConnection(t *testing.T) {
	sess := state.NewSession()

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	go func() {
		// < receive FLAPSignonFrame
		flap := wire.FLAPFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flap, serverReader))
		flapSignonFrame := wire.FLAPSignonFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flapSignonFrame, bytes.NewBuffer(flap.Payload)))

		// > send FLAPSignonFrame
		flapSignonFrame = wire.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		flapSignonFrame.Append(wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, []byte("the-cookie")))
		buf := &bytes.Buffer{}
		assert.NoError(t, wire.MarshalBE(flapSignonFrame, buf))
		flap = wire.FLAPFrame{
			StartMarker: 42,
			FrameType:   wire.FLAPFrameSignon,
			Payload:     buf.Bytes(),
		}
		assert.NoError(t, wire.MarshalBE(flap, serverWriter))

		flapc := wire.NewFlapClient(0, serverReader, serverWriter)

		// < receive SNAC_0x01_0x03_OServiceHostOnline
		frame := wire.SNACFrame{}
		body := wire.SNAC_0x01_0x03_OServiceHostOnline{}
		assert.NoError(t, flapc.ReceiveSNAC(&frame, &body))

		// send the first request that should get relayed to BOSRouter.Handle
		frame = wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceClientOnline,
		}
		assert.NoError(t, flapc.SendSNAC(frame, struct{}{}))
		assert.NoError(t, serverWriter.Close())
	}()

	authService := newMockAuthService(t)
	authService.EXPECT().
		RegisterBOSSession(mock.Anything, []byte("the-cookie")).
		Return(sess, nil)
	authService.EXPECT().
		Signout(mock.Anything, sess)

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

	rt := BOSServer{
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
