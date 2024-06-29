package foodgroup

import (
	"context"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewAdminService creates an instance of AdminService.
func NewAdminService(
	sessionManager SessionManager,
	accountManager AccountManager,
	buddyUpdateBroadcaster buddyBroadcaster,
) *AdminService {
	return &AdminService{
		sessionManager:         sessionManager,
		accountManager:         accountManager,
		buddyUpdateBroadcaster: buddyUpdateBroadcaster,
	}
}

// AdminService provides functionality for the Admin food group.
// The Admin food group is used for client control of passwords, screen name formatting,
// email address, and account confirmation.
type AdminService struct {
	sessionManager         SessionManager
	accountManager         AccountManager
	buddyUpdateBroadcaster buddyBroadcaster
}

// ConfirmRequest returns the ScreenName account status. It returns SNAC
// wire.AdminConfirmReply. The values in the return SNAC are
// flag, URL, length
func (s AdminService) ConfirmRequest(_ context.Context, frame wire.SNACFrame) (wire.SNACMessage, error) {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminAcctConfirmReply,
			RequestID: frame.RequestID,
		},
		Body: wire.SNAC_0x07_0x07_AdminConfirmReply{
			Status: wire.AdminAcctConfirmStatusEmailSent, // todo: get from session/db
		},
	}, nil
}

// InfoQuery returns the requested information about the account
func (s AdminService) InfoQuery(_ context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x07_0x02_AdminInfoQuery) (wire.SNACMessage, error) {
	var getAdminInfoReply = func(tag uint16, val any) wire.SNACMessage {
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Admin,
				SubGroup:  wire.AdminInfoReply,
				RequestID: frame.RequestID,
			},
			Body: wire.SNAC_0x07_0x03_AdminInfoReply{
				Permissions: wire.AdminInfoPermissionsReadWrite, // todo: what does this actually control?
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(tag, val),
					},
				},
			},
		}
	}

	// wire.AdminTLVRegistrationStatus is used in the AIM Preferences > Privacy panel to control
	// Allow users who know my e-mail address to find...
	//	o Nothing about me - wire.AdminInfoRegStatusNoDisclosure
	//	o Only that I have an account - wire.AdminInfoRegStatusLimitDisclosure
	//	o My screen name - wire.AdminInfoRegStatusFullDisclosure
	if _, hasRegStatus := body.TLVRestBlock.Slice(wire.AdminTLVRegistrationStatus); hasRegStatus {
		return getAdminInfoReply(wire.AdminTLVRegistrationStatus, wire.AdminInfoRegStatusFullDisclosure), nil // todo: get from session/db

	} else if _, hasEmail := body.TLVRestBlock.Slice(wire.AdminTLVEmailAddress); hasEmail {
		return getAdminInfoReply(wire.AdminTLVEmailAddress, sess.IdentScreenName().String()+"@aol.com"), nil // todo: get from session/db

	} else if _, hasNickName := body.TLVRestBlock.Slice(wire.AdminTLVScreenNameFormatted); hasNickName {
		return getAdminInfoReply(wire.AdminTLVScreenNameFormatted, sess.DisplayScreenName().String()), nil

	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminErr,
			RequestID: frame.RequestID,
		},
		Body: wire.SNACError{
			Code: wire.ErrorCodeNotSupportedByHost,
		},
	}, nil
}

func (s AdminService) InfoChangeRequest(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x07_0x04_AdminInfoChangeRequest) (wire.SNACMessage, error) {
	var replyMessage = func(tag uint16, val any) wire.SNACMessage {
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.Admin,
				SubGroup:  wire.AdminInfoChangeReply,
				RequestID: frame.RequestID,
			},
			Body: wire.SNAC_0x07_0x05_AdminChangeReply{
				Permissions: wire.AdminInfoPermissionsReadWrite,
				TLVBlock: wire.TLVBlock{
					TLVList: wire.TLVList{
						wire.NewTLV(tag, val),
					},
				},
			},
		}
	}

	var validateProposedName = func(name state.DisplayScreenName) (ok bool, errorCode uint16) {
		if len(name) > 16 {
			// proposed name is too long
			// todo: 16 should be defined elsewhere
			return false, wire.AdminInfoErrorInvalidNickNameLength
		} else if name.IdentScreenName() != sess.IdentScreenName() {
			// proposed name does not match session name (e.g. malicious client)
			return false, wire.AdminInfoErrorInvalidNickName
		}
		return true, 0
	}

	if sn, hasScreenNameFormatted := body.TLVRestBlock.Slice(wire.AdminTLVScreenNameFormatted); hasScreenNameFormatted {
		proposedName := state.DisplayScreenName(sn)
		if ok, errorCode := validateProposedName(proposedName); !ok {
			return wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Admin,
					SubGroup:  wire.AdminErr,
					RequestID: frame.RequestID,
				},
				Body: wire.SNACError{
					Code: errorCode,
				},
			}, nil
		}
		if err := s.accountManager.UpdateDisplayScreenName(proposedName); err != nil {
			return wire.SNACMessage{}, err
		}
		sess.SetDisplayScreenName(proposedName)
		if err := s.buddyUpdateBroadcaster.BroadcastBuddyArrived(ctx, sess); err != nil {
			return wire.SNACMessage{}, err
		}
		return replyMessage(wire.AdminTLVScreenNameFormatted, proposedName.String()), nil
	}
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Admin,
			SubGroup:  wire.AdminErr,
			RequestID: frame.RequestID,
		},
		Body: wire.SNACError{
			Code: wire.ErrorCodeNotSupportedByHost,
		},
	}, nil
}
