package oscar

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBUCPAuthService_handleNewConnection(t *testing.T) {
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
		buf := &bytes.Buffer{}
		assert.NoError(t, wire.MarshalBE(flapSignonFrame, buf))
		flap = wire.FLAPFrame{
			StartMarker: 42,
			FrameType:   wire.FLAPFrameSignon,
			Payload:     buf.Bytes(),
		}
		assert.NoError(t, wire.MarshalBE(flap, serverWriter))

		// > send SNAC_0x17_0x06_BUCPChallengeRequest
		flapc := wire.NewFlapClient(0, serverReader, serverWriter)
		frame := wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPChallengeRequest,
		}
		bodyIn := wire.SNAC_0x17_0x06_BUCPChallengeRequest{}
		assert.NoError(t, flapc.SendSNAC(frame, bodyIn))

		// < receive SNAC_0x17_0x07_BUCPChallengeResponse
		frame = wire.SNACFrame{}
		assert.NoError(t, flapc.ReceiveSNAC(&frame, &wire.SNAC_0x17_0x07_BUCPChallengeResponse{}))
		assert.Equal(t, wire.SNACFrame{FoodGroup: wire.BUCP, SubGroup: wire.BUCPChallengeResponse}, frame)

		// > send keep alive frame (like BSFlite does mid-login)
		assert.NoError(t, flapc.SendKeepAliveFrame())

		// > send SNAC_0x17_0x02_BUCPLoginRequest
		frame = wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPLoginRequest,
		}
		assert.NoError(t, flapc.SendSNAC(frame, wire.SNAC_0x17_0x02_BUCPLoginRequest{}))

		// < receive SNAC_0x17_0x03_BUCPLoginResponse
		frame = wire.SNACFrame{}
		assert.NoError(t, flapc.ReceiveSNAC(&frame, &wire.SNAC_0x17_0x03_BUCPLoginResponse{}))
		assert.Equal(t, wire.SNACFrame{FoodGroup: wire.BUCP, SubGroup: wire.BUCPLoginResponse}, frame)

		assert.NoError(t, serverWriter.Close())
	}()

	authService := newMockAuthService(t)
	authService.EXPECT().
		BUCPChallenge(mock.Anything, mock.Anything).
		Return(wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.BUCP,
				SubGroup:  wire.BUCPChallengeResponse,
			},
			Body: wire.SNAC_0x17_0x07_BUCPChallengeResponse{},
		}, nil)
	authService.EXPECT().
		BUCPLogin(mock.Anything, mock.Anything).
		Return(wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.BUCP,
				SubGroup:  wire.BUCPLoginResponse,
			},
			Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{},
		}, nil)

	rt := AuthServer{
		AuthService: authService,
		Logger:      slog.Default(),
	}
	rwc := pipeRWC{
		PipeReader: clientReader,
		PipeWriter: clientWriter,
	}
	assert.NoError(t, rt.handleNewConnection(rwc))
}
