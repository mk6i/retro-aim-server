package oscar

import (
	"context"
	"io"
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestRouter_OK(t *testing.T) {
	r := NewRouter()

	var called bool
	r.Register(wire.OService, wire.OServiceServiceRequest, func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
		called = true
		return nil
	})

	frame := wire.SNACFrame{
		FoodGroup: wire.OService,
		SubGroup:  wire.OServiceServiceRequest,
	}
	assert.NoError(t, r.Handle(nil, nil, frame, nil, nil))
	assert.True(t, called)
}

func TestRouter_Err(t *testing.T) {
	r := NewRouter()

	r.Register(wire.OService, wire.OServiceServiceRequest, func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
		return io.EOF
	})

	frame := wire.SNACFrame{
		FoodGroup: wire.OService,
		SubGroup:  wire.OServiceServiceRequest,
	}
	assert.ErrorIs(t, r.Handle(nil, nil, frame, nil, nil), io.EOF)
}

func TestRouter_ErrRouteNotFound(t *testing.T) {
	r := NewRouter()

	r.Register(wire.OService, wire.OServiceServiceRequest, func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
		return nil
	})

	frame := wire.SNACFrame{
		FoodGroup: wire.ICBM,
		SubGroup:  wire.ICBMChannelMsgToClient,
	}
	assert.ErrorIs(t, r.Handle(nil, nil, frame, nil, nil), ErrRouteNotFound)
}
