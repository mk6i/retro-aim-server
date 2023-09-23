package server

import (
	"testing"
	"time"
)

func TestSession_SendMessage_SessSendOK(t *testing.T) {
	s := Session{
		msgCh:       make(chan XMessage, 1),
		stopCh:      make(chan struct{}),
		sendTimeout: sendTimeout,
	}
	if res := s.SendMessage(XMessage{}); res != SessSendOK {
		t.Fatalf("expected SessSendOK, got %+v", res)
	}
}

func TestSession_SendMessage_SessSendClosed(t *testing.T) {
	s := Session{
		msgCh:       make(chan XMessage, 1),
		stopCh:      make(chan struct{}),
		sendTimeout: sendTimeout,
	}
	s.Close()
	if res := s.SendMessage(XMessage{}); res != SessSendClosed {
		t.Fatalf("expected SessSendClosed, got %+v", res)
	}
}

func TestSession_SendMessage_SessSendTimeout(t *testing.T) {
	s := Session{
		msgCh:       make(chan XMessage),
		stopCh:      make(chan struct{}),
		sendTimeout: 0,
	}
	if res := s.SendMessage(XMessage{}); res != SessSendTimeout {
		t.Fatalf("expected SessSendTimeout got %+v", res)
	}
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
