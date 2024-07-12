package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestAlertHandler_NotifyCapabilities(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Alert,
			SubGroup:  wire.AlertNotifyCapabilities,
		},
		Body: wire.SNACFrame{},
	}

	h := NewAlertHandler(slog.Default())

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.NotifyCapabilities(nil, nil, input.Frame, buf, nil))
}

func TestAlertHandler_NotifyDisplayCapabilities(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Alert,
			SubGroup:  wire.AlertNotifyDisplayCapabilities,
		},
		Body: wire.SNACFrame{},
	}

	h := NewAlertHandler(slog.Default())

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.NotifyDisplayCapabilities(nil, nil, input.Frame, buf, nil))
}
