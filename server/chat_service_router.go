package server

import (
	"context"
	"errors"
	"io"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type ChatServiceRooterRouter struct {
	ChatRouter
	Config
	OServiceChatRouter
}

func (rt ChatServiceRooterRouter) Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32, chatID string) error {
	inFrame := oscar.SNACFrame{}
	if err := oscar.Unmarshal(&inFrame, r); err != nil {
		return err
	}

	err := func() error {
		switch inFrame.FoodGroup {
		case oscar.OService:
			return rt.RouteOService(ctx, sess, chatID, inFrame, r, w, sequence)
		case oscar.Chat:
			return rt.RouteChat(ctx, sess, chatID, inFrame, r, w, sequence)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, inFrame, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(inFrame, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.Config.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}
