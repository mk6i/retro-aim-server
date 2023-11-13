package server

import (
	"bufio"
	"bytes"
	"context"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
)

func TestHandleChatConnection_Notification(t *testing.T) {

	ctx := context.Background()
	cfg := Config{}
	cr := NewChatRegistry()
	logger := NewLogger(cfg)

	room := ChatRoom{
		Name:           "test chat room!",
		SessionManager: NewSessionManager(logger),
	}
	bobSess := room.NewSessionWithSN("bob-sess-id", "bob")
	cr.Register(room)

	msgIn := []oscar.XMessage{
		{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.CHAT,
				SubGroup:  oscar.ChatUsersJoined,
			},
			SnacOut: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []oscar.TLVUserInfo{
					bobSess.TLVUserInfo(),
				},
			},
		},
		{
			SnacFrame: oscar.SnacFrame{
				FoodGroup: oscar.CHAT,
				SubGroup:  oscar.ChatUsersLeft,
			},
			SnacOut: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
				Users: []oscar.TLVUserInfo{},
			},
		},
	}

	routeSig := func(ctx context.Context, buf io.Reader, w io.Writer, u *uint32) error {
		return nil
	}

	wg := sync.WaitGroup{}
	wg.Add(len(msgIn))

	var msgOut []oscar.XMessage
	alertHandler := func(ctx context.Context, msg oscar.XMessage, w io.Writer, u *uint32) error {
		msgOut = append(msgOut, msg)
		wg.Done()
		return nil
	}

	go func() {
		wg.Wait()
		bobSess.Close()
	}()

	pr, _ := io.Pipe()
	rw := bufio.NewReadWriter(bufio.NewReader(pr), bufio.NewWriter(&bytes.Buffer{}))

	for _, msg := range msgIn {
		room.SendToScreenName(ctx, "bob", msg)
	}

	dispatchIncomingMessages(ctx, bobSess, uint32(0), rw, logger, routeSig, alertHandler)

	assert.Equal(t, msgIn, msgOut)
}

func TestHandleChatConnection_ClientRequestFLAP(t *testing.T) {

	ctx := context.Background()
	cfg := Config{}
	cr := NewChatRegistry()
	logger := NewLogger(cfg)

	room := ChatRoom{
		Name:           "test chat room!",
		SessionManager: NewSessionManager(logger),
	}
	bobSess := room.NewSessionWithSN("bob-sess-id", "bob")
	cr.Register(room)

	payloads := [][]byte{
		{'a', 'b', 'c', 'd'},
		{'e', 'f', 'g', 'h'},
	}

	pr, pw := io.Pipe()
	_, pw2 := io.Pipe()
	go func() {
		for _, buf := range payloads {
			flap := oscar.FlapFrame{
				StartMarker:   42,
				FrameType:     oscar.FlapFrameData,
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
	alertHandler := func(ctx context.Context, msg oscar.XMessage, w io.Writer, u *uint32) error {
		return nil
	}

	rw := bufio.NewReadWriter(bufio.NewReader(pr), bufio.NewWriter(pw2))

	go func() {
		wg.Wait()
		pw.Close()
	}()

	dispatchIncomingMessages(ctx, bobSess, uint32(0), rw, logger, routeSig, alertHandler)

	assert.Equal(t, payloads, msgOut)
}

func TestHandleChatConnection_SessionClosed(t *testing.T) {

	ctx := context.Background()
	cfg := Config{}
	cr := NewChatRegistry()
	logger := NewLogger(cfg)

	room := ChatRoom{
		Name:           "test chat room!",
		SessionManager: NewSessionManager(logger),
	}
	sess := room.NewSessionWithSN("bob-sess-id", "bob")
	cr.Register(room)

	routeSig := func(ctx context.Context, buf io.Reader, w io.Writer, u *uint32) error {
		t.Fatal("not expecting any output")
		return nil
	}
	alertHandler := func(ctx context.Context, msg oscar.XMessage, w io.Writer, u *uint32) error {
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

	flap := oscar.FlapFrame{}
	assert.NoError(t, oscar.Unmarshal(&flap, pr2))
	assert.Equal(t, oscar.FlapFrameSignoff, flap.FrameType)
}
