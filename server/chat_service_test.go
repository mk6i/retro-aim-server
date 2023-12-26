package server

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatService_handleNewConnection(t *testing.T) {
	sess := state.NewSession()
	sess.SetID("login-cookie-1234")
	chatCookie := "chat-cookie"

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// < receive FLAPSignonFrame
		flap := oscar.FLAPFrame{}
		assert.NoError(t, oscar.Unmarshal(&flap, serverReader))
		buf, err := flap.SNACBuffer(serverReader)
		assert.NoError(t, err)
		flapSignonFrame := oscar.FLAPSignonFrame{}
		assert.NoError(t, oscar.Unmarshal(&flapSignonFrame, buf))

		// > send FLAPSignonFrame
		flapSignonFrame = oscar.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		cookie := ChatCookie{
			Cookie: []byte(chatCookie),
			SessID: sess.ID(),
		}
		flapSignonFrame.Append(oscar.NewTLV(oscar.OServiceTLVTagsLoginCookie, cookie))
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

		// < receive SNAC_0x01_0x03_OServiceHostOnline
		flap = oscar.FLAPFrame{}
		assert.NoError(t, oscar.Unmarshal(&flap, serverReader))
		buf, err = flap.SNACBuffer(serverReader)
		assert.NoError(t, err)
		frame := oscar.SNACFrame{}
		assert.NoError(t, oscar.Unmarshal(&frame, buf))
		body := oscar.SNAC_0x01_0x03_OServiceHostOnline{}
		assert.NoError(t, oscar.Unmarshal(&body, buf))

		// send the first request that should get relayed to BOSRouter.Route
		var seq uint32
		frame = oscar.SNACFrame{
			FoodGroup: oscar.Chat,
			SubGroup:  oscar.ChatNavNavInfo,
		}
		assert.NoError(t, sendSNAC(frame, struct{}{}, &seq, serverWriter))

		assert.NoError(t, serverWriter.Close())
	}()

	authHandler := newMockAuthHandler(t)
	authHandler.EXPECT().
		RetrieveChatSession(chatCookie, sess.ID()).
		Return(sess, nil)
	authHandler.EXPECT().
		SignoutChat(mock.Anything, sess, chatCookie).
		Run(func(ctx context.Context, sess *state.Session, chatID string) {
			wg.Done()
		}).
		Return(nil)

	chatHandler := newMockOServiceChatHandler(t)
	chatHandler.EXPECT().
		WriteOServiceHostOnline().
		Return(oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.OService,
				SubGroup:  oscar.OServiceHostOnline,
			},
			Body: oscar.SNAC_0x01_0x03_OServiceHostOnline{},
		})

	bosRouter := newMockChatServiceRouter(t)
	bosRouter.EXPECT().
		Route(mock.Anything, sess, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	rt := ChatService{
		AuthHandler:        authHandler,
		OServiceChatRouter: NewOServiceRouterForChat(slog.Default(), nil, chatHandler),
		ChatServiceRouter:  bosRouter,
	}
	rwc := pipeRWC{
		PipeReader: clientReader,
		PipeWriter: clientWriter,
	}
	rt.handleNewConnection(context.Background(), rwc)

	wg.Wait() // wait for server to drain the connection
}
