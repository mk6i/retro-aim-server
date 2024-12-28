package oscar

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

func TestHandleChatConnection_MessageRelay(t *testing.T) {
	sessionManager := state.NewInMemorySessionManager(slog.Default())
	// add a user to session that will receive relayed messages
	sess, _ := sessionManager.AddSession(nil, "bob")

	// start the server connection handler in the background
	serverReader, _ := io.Pipe()
	clientReader, serverWriter := io.Pipe()
	go func() {
		flapc := wire.NewFlapClient(0, nil, serverWriter)
		err := dispatchIncomingMessages(context.Background(), sess, flapc, serverReader, slog.Default(), nil)
		assert.NoError(t, err)
	}()

	inboundMsgs := []wire.SNACMessage{
		{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Chat,
				SubGroup:  wire.ChatUsersJoined,
			},
			Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []wire.TLVUserInfo{
					{
						ScreenName: "screenname1",
					},
				},
			},
		},
		{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Chat,
				SubGroup:  wire.ChatUsersLeft,
			},
			Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []wire.TLVUserInfo{
					{
						ScreenName: "screenname2",
					},
				},
			},
		},
	}

	// relay messages to user session
	for _, msg := range inboundMsgs {
		sessionManager.RelayToScreenName(context.Background(), state.NewIdentScreenName("bob"), msg)
	}

	// consume and verify the relayed messages
	for i := 0; i < len(inboundMsgs); i++ {
		flap := wire.FLAPFrame{}
		assert.NoError(t, wire.UnmarshalBE(&flap, clientReader))
		frame := wire.SNACFrame{}
		buf := bytes.NewBuffer(flap.Payload)
		assert.NoError(t, wire.UnmarshalBE(&frame, buf))
		assert.Equal(t, inboundMsgs[i].Frame, frame)
		body := wire.SNAC_0x0E_0x03_ChatUsersJoined{}
		assert.NoError(t, wire.UnmarshalBE(&body, buf))
		assert.Equal(t, inboundMsgs[i].Body, body)
	}

	// stop the session, which terminates the connection handler goroutine
	sess.Close()
	<-sess.Closed()

	// verify the connection handler sends client disconnection message before
	// terminating
	flap := wire.FLAPFrame{}
	assert.NoError(t, wire.UnmarshalBE(&flap, clientReader))
	assert.Equal(t, wire.FLAPFrameSignoff, flap.FrameType)
}

func TestHandleChatConnection_ClientRequest(t *testing.T) {
	sessionManager := state.NewInMemorySessionManager(slog.Default())
	// add session so that the function can terminate upon closure
	sess, _ := sessionManager.AddSession(nil, "bob")

	inboundMsgs := []wire.SNACMessage{
		{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Chat,
				SubGroup:  wire.ChatUsersJoined,
			},
			Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []wire.TLVUserInfo{
					{
						ScreenName: "screenname1",
					},
				},
			},
		},
		{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Chat,
				SubGroup:  wire.ChatUsersLeft,
			},
			Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []wire.TLVUserInfo{
					{
						ScreenName: "screenname2",
					},
				},
			},
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(inboundMsgs))

	// set up mock handlers to receive messages and verify their contents
	router := newMockHandler(t)
	for _, msg := range inboundMsgs {
		msg := msg
		router.EXPECT().
			Handle(mock.Anything, sess, msg.Frame, mock.Anything, mock.Anything).
			Run(func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) {
				defer wg.Done()
				body := wire.SNAC_0x0E_0x03_ChatUsersJoined{}
				assert.NoError(t, wire.UnmarshalBE(&body, r))
				assert.Equal(t, msg.Body, body)
			}).
			Return(nil)
	}

	// start the server connection handler in the background
	serverReader, clientWriter := io.Pipe()
	clientReader, serverWriter := io.Pipe()
	go func() {
		flapc := wire.NewFlapClient(0, nil, serverWriter)
		assert.NoError(t, dispatchIncomingMessages(context.Background(), sess, flapc, serverReader, slog.Default(), router))
	}()

	// send client messages
	flapc := wire.NewFlapClient(0, nil, clientWriter)
	for _, msg := range inboundMsgs {
		err := flapc.SendSNAC(msg.Frame, msg.Body)
		assert.NoError(t, err)
	}
	wg.Wait()

	// stop the session, which terminates the connection handler goroutine
	sess.Close()
	<-sess.Closed()

	// verify the connection handler sends client disconnection message before
	// terminating
	flap := wire.FLAPFrame{}
	assert.NoError(t, wire.UnmarshalBE(&flap, clientReader))
	assert.Equal(t, wire.FLAPFrameSignoff, flap.FrameType)
}
