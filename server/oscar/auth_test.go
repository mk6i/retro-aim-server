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
		assert.NoError(t, wire.Unmarshal(&flap, serverReader))
		buf, err := flap.ReadBody(serverReader)
		assert.NoError(t, err)
		flapSignonFrame := wire.FLAPSignonFrame{}
		assert.NoError(t, wire.Unmarshal(&flapSignonFrame, buf))

		// > send FLAPSignonFrame
		flapSignonFrame = wire.FLAPSignonFrame{
			FLAPVersion: 1,
		}
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
		BUCPLogin(mock.Anything, mock.Anything, mock.Anything).
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
