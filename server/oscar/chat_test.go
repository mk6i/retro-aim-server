package oscar

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatService_handleNewConnection(t *testing.T) {
	sess := state.NewSession()

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
		flapSignonFrame.Append(wire.NewTLV(wire.OServiceTLVTagsLoginCookie, []byte(`the-chat-login-cookie`)))
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
		flapc := wire.NewFlapClient(0, nil, serverWriter)
		frame = wire.SNACFrame{
			FoodGroup: wire.Chat,
			SubGroup:  wire.ChatNavNavInfo,
		}
		assert.NoError(t, flapc.SendSNAC(frame, struct{}{}))
		assert.NoError(t, serverWriter.Close())
	}()

	authService := newMockAuthService(t)
	authService.EXPECT().
		RegisterChatSession([]byte(`user-screen-name`)).
		Return(sess, nil)
	authService.EXPECT().
		SignoutChat(mock.Anything, sess).
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

	cookieCracker := newMockCookieCracker(t)
	cookieCracker.EXPECT().
		Crack([]byte(`the-chat-login-cookie`)).
		Return([]byte(`user-screen-name`), nil)

	bosRouter := newMockHandler(t)
	bosRouter.EXPECT().
		Handle(mock.Anything, sess, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	rt := ChatServer{
		AuthService:    authService,
		CookieCracker:  cookieCracker,
		Handler:        bosRouter,
		Logger:         slog.Default(),
		OnlineNotifier: onlineNotifier,
	}
	rwc := pipeRWC{
		PipeReader: clientReader,
		PipeWriter: clientWriter,
	}
	assert.NoError(t, rt.handleNewConnection(context.Background(), rwc))
}
