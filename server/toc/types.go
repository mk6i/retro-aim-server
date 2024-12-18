package toc

import (
	"context"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type LocateService interface {
	//DirInfo(ctx context.Context, frame wire.SNACFrame, body wire.SNAC_0x02_0x0B_LocateGetDirInfo) (wire.SNACMessage, error)
	//RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	//SetDirInfo(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x09_LocateSetDirInfo) (wire.SNACMessage, error)
	SetInfo(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x02_0x04_LocateSetInfo) error
	//SetKeywordInfo(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, body wire.SNAC_0x02_0x0F_LocateSetKeywordInfo) (wire.SNACMessage, error)
	UserInfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery) (wire.SNACMessage, error)
}
