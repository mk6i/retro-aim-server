package server

import (
	"context"
	"io"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

type ChatServiceRouter struct {
	ChatRouter         Router
	OServiceChatRouter Router
}

func (rt ChatServiceRouter) Route(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.FoodGroup {
	case oscar.OService:
		return rt.OServiceChatRouter.Route(ctx, sess, inFrame, r, w, sequence)
	case oscar.Chat:
		return rt.ChatRouter.Route(ctx, sess, inFrame, r, w, sequence)
	default:
		return ErrUnsupportedSubGroup
	}
}
