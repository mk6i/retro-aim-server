package server

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBUCPAuthService_handleNewConnection(t *testing.T) {
	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	go func() {
		// < receive FLAPSignonFrame
		flap := oscar.FLAPFrame{}
		assert.NoError(t, oscar.Unmarshal(&flap, serverReader))
		buf, err := flap.ReadBody(serverReader)
		assert.NoError(t, err)
		flapSignonFrame := oscar.FLAPSignonFrame{}
		assert.NoError(t, oscar.Unmarshal(&flapSignonFrame, buf))

		// > send FLAPSignonFrame
		flapSignonFrame = oscar.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		buf = &bytes.Buffer{}
		assert.NoError(t, oscar.Marshal(flapSignonFrame, buf))
		flap = oscar.FLAPFrame{
			StartMarker:   42,
			FrameType:     oscar.FLAPFrameSignon,
			PayloadLength: uint16(buf.Len()),
		}
		assert.NoError(t, oscar.Marshal(flap, serverWriter))
		_, err = serverWriter.Write(buf.Bytes())
		assert.NoError(t, err)

		// > send SNAC_0x17_0x06_BUCPChallengeRequest
		var seq uint32
		frame := oscar.SNACFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPChallengeRequest,
		}
		bodyIn := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
		assert.NoError(t, sendSNAC(frame, bodyIn, &seq, serverWriter))

		// < receive SNAC_0x17_0x07_BUCPChallengeResponse
		frame = oscar.SNACFrame{}
		assert.NoError(t, receiveSNAC(&frame, &oscar.SNAC_0x17_0x07_BUCPChallengeResponse{}, serverReader))
		assert.Equal(t, oscar.SNACFrame{FoodGroup: oscar.BUCP, SubGroup: oscar.BUCPChallengeResponse}, frame)

		// > send SNAC_0x17_0x02_BUCPLoginRequest
		frame = oscar.SNACFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPLoginRequest,
		}
		assert.NoError(t, sendSNAC(frame, oscar.SNAC_0x17_0x02_BUCPLoginRequest{}, &seq, serverWriter))

		// < receive SNAC_0x17_0x03_BUCPLoginResponse
		frame = oscar.SNACFrame{}
		assert.NoError(t, receiveSNAC(&frame, &oscar.SNAC_0x17_0x03_BUCPLoginResponse{}, serverReader))
		assert.Equal(t, oscar.SNACFrame{FoodGroup: oscar.BUCP, SubGroup: oscar.BUCPLoginResponse}, frame)

		assert.NoError(t, serverWriter.Close())
	}()

	authHandler := newMockAuthHandler(t)
	authHandler.EXPECT().
		BUCPChallengeRequestHandler(mock.Anything, mock.Anything).
		Return(oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  oscar.BUCPChallengeResponse,
			},
			Body: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{},
		}, nil)
	authHandler.EXPECT().
		BUCPLoginRequestHandler(mock.Anything, mock.Anything, mock.Anything).
		Return(oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.BUCP,
				SubGroup:  oscar.BUCPLoginResponse,
			},
			Body: oscar.SNAC_0x17_0x03_BUCPLoginResponse{},
		}, nil)

	rt := BUCPAuthService{
		AuthHandler: authHandler,
		Logger:      slog.Default(),
	}
	rwc := pipeRWC{
		PipeReader: clientReader,
		PipeWriter: clientWriter,
	}
	assert.NoError(t, rt.handleNewConnection(rwc))
}
