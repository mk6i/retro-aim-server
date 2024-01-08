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

	routeSig := func(ctx context.Context, buf io.Reader, w io.Writer, u *uint32) error {
		return nil
	}

	wg := sync.WaitGroup{}
	wg.Add(len(msgIn))

	var msgOut []oscar.SNACMessage
	alertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, u *uint32) error {
		msgOut = append(msgOut, msg)
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

	dispatchIncomingMessages(ctx, sess, uint32(0), rw, logger, routeSig, alertHandler)

	assert.Equal(t, msgIn, msgOut)
}

func TestHandleChatConnection_ClientRequestFLAP(t *testing.T) {

	ctx := context.Background()
	cfg := config.Config{}
	logger := NewLogger(cfg)

	sessionManager := state.NewInMemorySessionManager(logger)
	sess := sessionManager.AddSession("bob-sess-id", "bob")

	payloads := [][]byte{
		{'a', 'b', 'c', 'd'},
		{'e', 'f', 'g', 'h'},
	}

	pr, pw := io.Pipe()
	_, pw2 := io.Pipe()
	go func() {
		for _, buf := range payloads {
			flap := oscar.FLAPFrame{
				StartMarker:   42,
				FrameType:     oscar.FLAPFrameData,
				PayloadLength: uint16(len(buf)),
			}
			assert.NoError(t, oscar.Marshal(flap, pw))
			assert.NoError(t, oscar.Marshal(buf, pw))
		}
	}()

	var msgOut [][]byte
	wg := sync.WaitGroup{}
	wg.Add(len(payloads))

	routeSig := func(ctx context.Context, buf io.Reader, w io.Writer, u *uint32) error {
		var err error
		b, err := io.ReadAll(buf)
		msgOut = append(msgOut, b)
		wg.Done()
		return err
	}
	alertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, u *uint32) error {
		return nil
	}

	rw := bufio.NewReadWriter(bufio.NewReader(pr), bufio.NewWriter(pw2))

	go func() {
		wg.Wait()
		pw.Close()
	}()

	dispatchIncomingMessages(ctx, sess, uint32(0), rw, logger, routeSig, alertHandler)

	assert.Equal(t, payloads, msgOut)
}

func TestHandleChatConnection_SessionClosed(t *testing.T) {

	ctx := context.Background()
	cfg := config.Config{}
	logger := NewLogger(cfg)

	sessionManager := state.NewInMemorySessionManager(logger)
	sess := sessionManager.AddSession("bob-sess-id", "bob")

	routeSig := func(ctx context.Context, buf io.Reader, w io.Writer, u *uint32) error {
		t.Fatal("not expecting any output")
		return nil
	}
	alertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, u *uint32) error {
		t.Fatal("not expecting any alerts")
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

	go dispatchIncomingMessages(ctx, sess, 0, in, logger, routeSig, alertHandler)

	flap := oscar.FLAPFrame{}
	assert.NoError(t, oscar.Unmarshal(&flap, pr2))
	assert.Equal(t, oscar.FLAPFrameSignoff, flap.FrameType)
}
