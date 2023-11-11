package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSession_SendMessage_SessSendOK(t *testing.T) {
	s := Session{
		msgCh:  make(chan XMessage, 1),
		stopCh: make(chan struct{}),
	}
	if res := s.SendMessage(XMessage{}); res != SessSendOK {
		t.Fatalf("expected SessSendOK, got %+v", res)
	}
}

func TestSession_SendMessage_SessSendClosed(t *testing.T) {
	s := Session{
		msgCh:  make(chan XMessage, 1),
		stopCh: make(chan struct{}),
	}
	s.Close()
	if res := s.SendMessage(XMessage{}); res != SessSendClosed {
		t.Fatalf("expected SessSendClosed, got %+v", res)
	}
}

func TestSession_SendMessage_SessQueueFull(t *testing.T) {
	bufSize := 10
	s := Session{
		msgCh:  make(chan XMessage, bufSize),
		stopCh: make(chan struct{}),
	}
	for i := 0; i < bufSize; i++ {
		assert.Equal(t, SessSendOK, s.SendMessage(XMessage{}))
	}
	assert.Equal(t, SessQueueFull, s.SendMessage(XMessage{}))
}

func TestSession_Close_Twice(t *testing.T) {
	s := Session{
		stopCh: make(chan struct{}),
	}
	s.Close()
	s.Close() // make sure close is idempotent
	if !s.closed {
		t.Fatal("expected session to be closed")
	}
	select {
	case <-s.Closed():
	case <-time.After(1 * time.Second):
		t.Fatalf("channel is not closed")
	}
}
