package server

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleChatConnection_Notification(t *testing.T) {

	ctx := context.Background()
	cfg := config.Config{}
	logger := NewLogger(cfg)

	sessionManager := state.NewInMemorySessionManager(logger)
	sess := sessionManager.AddSession("bob-sess-id", "bob")

	msgIn := []oscar.SNACMessage{
		{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.Chat,
				SubGroup:  oscar.ChatUsersJoined,
			},
			Body: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []oscar.TLVUserInfo{
					sess.TLVUserInfo(),
				},
			},
		},
		{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.Chat,
				SubGroup:  oscar.ChatUsersLeft,
			},
			Body: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []oscar.TLVUserInfo{},
			},
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(len(msgIn))

	var msgOut []oscar.SNACMessage
	alertHandler := func(frame oscar.SNACFrame, body any, sequence *uint32, w io.Writer) error {
		msgOut = append(msgOut, oscar.SNACMessage{
			Frame: frame,
			Body:  body,
		})
		wg.Done()
		return nil
	}

	go func() {
		wg.Wait()
		sess.Close()
	}()

	pr, _ := io.Pipe()
	rw := bufio.NewReadWriter(bufio.NewReader(pr), bufio.NewWriter(&bytes.Buffer{}))

	for _, msg := range msgIn {
		sessionManager.RelayToScreenName(ctx, "bob", msg)
	}

	assert.NoError(t, dispatchIncomingMessages(ctx, sess, uint32(0), rw, logger, nil, alertHandler, config.Config{}))

	assert.Equal(t, msgIn, msgOut)
}

func TestHandleChatConnection_ClientRequestFLAP(t *testing.T) {

	ctx := context.Background()
	cfg := config.Config{}
	logger := NewLogger(cfg)

	sessionManager := state.NewInMemorySessionManager(logger)
	sess := sessionManager.AddSession("bob-sess-id", "bob")

	payloads := []oscar.SNACFrame{
		{FoodGroup: oscar.ICBM, SubGroup: oscar.ICBMChannelMsgToClient},
		{FoodGroup: oscar.ChatNav, SubGroup: oscar.ChatNavNavInfo},
	}

	pr, pw := io.Pipe()
	_, pw2 := io.Pipe()
	go func() {
		for _, buf := range payloads {
			var seq uint32
			sendSNAC(buf, struct{}{}, &seq, pw)
		}
	}()

	var msgOut []oscar.SNACFrame
	wg := sync.WaitGroup{}
	wg.Add(len(payloads))

	router := newMockRouter(t)
	router.EXPECT().
		Route(mock.Anything, sess, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) {
			msgOut = append(msgOut, inFrame)
			wg.Done()
		}).Return(nil)

	alertHandler := func(frame oscar.SNACFrame, body any, sequence *uint32, w io.Writer) error {
		return nil
	}

	rw := bufio.NewReadWriter(bufio.NewReader(pr), bufio.NewWriter(pw2))

	go func() {
		wg.Wait()
		pw.Close()
	}()

	assert.NoError(t, dispatchIncomingMessages(ctx, sess, uint32(0), rw, logger, router, alertHandler, config.Config{}))

	assert.Equal(t, payloads, msgOut)
}

func TestHandleChatConnection_SessionClosed(t *testing.T) {

	ctx := context.Background()
	cfg := config.Config{}
	logger := NewLogger(cfg)

	sessionManager := state.NewInMemorySessionManager(logger)
	sess := sessionManager.AddSession("bob-sess-id", "bob")

	router := newMockRouter(t)
	router.EXPECT().
		Route(mock.Anything, sess, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) {
			t.Fatal("not expecting any output")
		}).
		Maybe().
		Return(nil)

	alertHandler := func(frame oscar.SNACFrame, body any, sequence *uint32, w io.Writer) error {
		t.Fatal("not expecting any output")
		return nil
	}

	pr1, _ := io.Pipe()
	pr2, pw2 := io.Pipe()

	in := struct {
		io.Reader
		io.Writer
	}{
		Reader: pr1,
		Writer: pw2,
	}
	sess.Close()

	go dispatchIncomingMessages(ctx, sess, 0, in, logger, router, alertHandler, config.Config{})

	flap := oscar.FLAPFrame{}
	assert.NoError(t, oscar.Unmarshal(&flap, pr2))
	assert.Equal(t, oscar.FLAPFrameSignoff, flap.FrameType)
}
