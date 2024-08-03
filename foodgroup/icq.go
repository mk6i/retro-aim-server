package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewICQService creates an instance of ICQService.
func NewICQService(
	messageRelayer MessageRelayer,
	icqFinder ICQFinder,
	userUpdater UserUpdater,
	logger *slog.Logger,
	sessionRetriever SessionRetriever,
) ICQService {
	return ICQService{
		messageRelayer:   messageRelayer,
		icqFinder:        icqFinder,
		userUpdater:      userUpdater,
		logger:           logger,
		sessionRetriever: sessionRetriever,
	}
}

// ICQService provides functionality for the ICQ (PD) food group.
// The PD food group manages settings for permit/deny (allow/block) for
// pre-feedbag (sever-side buddy list) AIM clients. Right now it's stubbed out
// to support pidgin. Eventually this food group will be fully implemented in
// order to support client blocking in AIM <= 3.0.
type ICQService struct {
	icqFinder        ICQFinder
	logger           *slog.Logger
	messageRelayer   MessageRelayer
	sessionRetriever SessionRetriever
	userUpdater      UserUpdater
}

type ReqUserInfo struct {
	SearchUIN uint32
}

func (s ICQService) GetICQFullUserInfo(ctx context.Context, sess *state.Session, userInfo ReqUserInfo, seq uint16) error {

	user, err := s.icqFinder.FindByUIN(userInfo.SearchUIN)
	if err != nil {
		return err
	}

	if err := s.getICQUserInfo(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.getICQMoreUserInfo(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.getICQInfoEmailMore(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.getICQHomepageCat(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.getICQMetaWorkUserInfo(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.getICQUserNotes(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.getICQUserInterests(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.getICQMetaAffiliationsUserInfo(ctx, sess, user, seq); err != nil {
		return err
	}
	return nil
}

func (s ICQService) getICQUserInfo(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType:   0x00C8,
			Success:      0x0A,
			Nickname:     user.Nickname,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Email:        user.EmailAddress,
			HomeCity:     user.HomeCity,
			HomeState:    user.HomeState,
			HomePhone:    user.HomePhone,
			HomeFax:      user.HomeFax,
			HomeAddress:  user.HomeAddress,
			CellPhone:    user.CellPhone,
			ZipCode:      user.ZipCode,
			CountryCode:  user.CountryCode,
			GMTOffset:    user.GMTOffset,
			AuthFlag:     0,
			WebAware:     1,
			DCPerms:      0,
			PublishEmail: 1,
		},
	}

	return s.reply(ctx, sess, msg)

}

func (s ICQService) getICQMoreUserInfo(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMoreUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: 0x00DC,
			Success:    0x0A,
			SomeMoreUserInfo: wire.SomeMoreUserInfo{
				Age:          uint8(user.Age()),
				Gender:       user.Gender,
				HomePageAddr: user.HomePageAddr,
				BirthYear:    user.BirthYear,
				BirthMonth:   user.BirthMonth,
				BirthDay:     user.BirthDay,
				Lang1:        user.Lang1,
				Lang2:        user.Lang2,
				Lang3:        user.Lang3,
			},
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) getICQInfoEmailMore(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQInfoEmailMore{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: 0x00EB,
			Success:    0x0A,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) getICQHomepageCat(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQHomepageCat{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: 0x010E,
			Success:    0x0A,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) getICQMetaWorkUserInfo(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMetaWorkUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: 0x00D2,
			Success:    0x0A,
			ICQWorkInfo: wire.ICQWorkInfo{
				WorkCity:        user.WorkCity,
				WorkState:       user.WorkState,
				WorkPhone:       user.WorkPhone,
				WorkFax:         user.WorkFax,
				WorkAddress:     user.WorkAddress,
				WorkZIP:         user.WorkZIP,
				WorkCountryCode: user.WorkCountryCode,
				Company:         user.Company,
				Department:      user.Department,
				Position:        user.Position,
				OccupationCode:  user.OccupationCode,
				WorkWebPage:     user.WorkWebPage,
			},
		},
	}
	return s.reply(ctx, sess, msg)
}

func (s ICQService) getICQUserNotes(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQUserNotes{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: 0x00E6,
			Success:    0x0A,
			ICQNotes: wire.ICQNotes{
				Notes: user.Notes,
			},
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) getICQUserInterests(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQUserInterests{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: 0x00F0,
			Success:    0x0A,
			Interests: []struct {
				Code    uint16
				Keyword string `oscar:"len_prefix=uint16,nullterm"`
			}{
				{
					Code:    user.Interest1Code,
					Keyword: user.Interest1Keyword,
				},
				{
					Code:    user.Interest2Code,
					Keyword: user.Interest2Keyword,
				},
				{
					Code:    user.Interest3Code,
					Keyword: user.Interest3Keyword,
				},
				{
					Code:    user.Interest4Code,
					Keyword: user.Interest4Keyword,
				},
			},
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) getICQMetaAffiliationsUserInfo(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMetaAffiliationsUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: 0x00FA,
			Success:    0x0A,
			ICQAffiliations: wire.ICQAffiliations{
				PastAffiliations: []struct {
					Code    uint16
					Keyword string `oscar:"len_prefix=uint16,nullterm"`
				}{
					{
						Code:    user.PastAffiliations1Code,
						Keyword: user.PastAffiliations1Keyword,
					},
					{
						Code:    user.PastAffiliations2Code,
						Keyword: user.PastAffiliations2Keyword,
					},
					{
						Code:    user.PastAffiliations3Code,
						Keyword: user.PastAffiliations3Keyword,
					},
				},
				Affiliations: []struct {
					Code    uint16
					Keyword string `oscar:"len_prefix=uint16,nullterm"`
				}{
					{
						Code:    user.Affiliations1Code,
						Keyword: user.Affiliations1Keyword,
					},
					{
						Code:    user.Affiliations2Code,
						Keyword: user.Affiliations2Keyword,
					},
					{
						Code:    user.Affiliations3Code,
						Keyword: user.Affiliations3Keyword,
					},
				},
			},
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) GetICQMessagesEOF(ctx context.Context, sess *state.Session, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMessagesEOF{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x0042,
				Seq:     seq,
			},
			DroppedMessages: 0,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) FindByUIN(ctx context.Context, sess *state.Session, req wire.ICQFindByUIN, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: 0x07DA,
			Seq:     seq,
		},
		ReqSubType: 0x01AE,
		Success:    0x0A,
	}
	resp.LastResult()

	res, err := s.icqFinder.FindByUIN(req.UIN)

	switch {
	case errors.Is(err, state.ErrNoUser):
		resp.Success = 0x32
	case err != nil:
		s.logger.Error("FindByUIN failed", "err", err.Error())
		resp.Success = 0x14
	default:
		resp.Success = 0x0a
		resp.Details = s.createResult(res)
	}

	return s.reply(ctx, sess, wire.ICQMessage{
		Message: resp,
	})
}

func (s ICQService) FindByEmail(ctx context.Context, sess *state.Session, req wire.ICQFindByEmail, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: 0x07DA,
			Seq:     seq,
		},
		ReqSubType: 0x01AE,
		Success:    0x0A,
	}
	resp.LastResult()

	res, err := s.icqFinder.FindByEmail(req.Email)

	switch {
	case errors.Is(err, state.ErrNoUser):
		resp.Success = 0x32
	case err != nil:
		s.logger.Error("FindByEmail failed", "err", err.Error())
		resp.Success = 0x14
	default:
		resp.Success = 0x0a
		resp.Details = s.createResult(res)
	}

	return s.reply(ctx, sess, wire.ICQMessage{
		Message: resp,
	})
}

func (s ICQService) FindByDetails(ctx context.Context, sess *state.Session, req wire.ICQFindByDetails, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: 0x07DA,
			Seq:     seq,
		},
		Success:    0x0A,
		ReqSubType: 0x01AE,
	}

	res, err := s.icqFinder.FindByDetails(req.FirstName, req.LastName, req.NickName)

	if err != nil {
		s.logger.Error("FindByDetails failed", "err", err.Error())
		resp.Success = 0x14
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}
	if len(res) == 0 {
		resp.Success = 0x32
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}

	for i := 0; i < len(res); i++ {
		if i == len(res)-1 {
			resp.LastResult()
		} else {
			resp.ReqSubType = 0x01A4
		}
		resp.Details = s.createResult(res[i])
		if err := s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s ICQService) FindByWhitepages(ctx context.Context, sess *state.Session, req wire.ICQFindByWhitePages, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: 0x07DA,
			Seq:     seq,
		},
		Success:    0x0A,
		ReqSubType: 0x01AE,
	}

	interests := strings.Split(req.InterestsKeyword, ",")
	res, err := s.icqFinder.FindByInterests(req.InterestsCode, interests)

	if err != nil {
		s.logger.Error("FindByWhitepages failed", "err", err.Error())
		resp.Success = 0x14
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}
	if len(res) == 0 {
		resp.Success = 0x32
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}

	for i := 0; i < len(res); i++ {
		if i == len(res)-1 {
			resp.LastResult()
		} else {
			resp.ReqSubType = 0x01A4
		}
		resp.Details = s.createResult(res[i])
		if err := s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s ICQService) createResult(res state.User) wire.ICQUserSearchRecord {
	searchRecord := res.ICQUserSearchRecord(time.Now())
	userSess := s.sessionRetriever.RetrieveSession(res.IdentScreenName)
	if userSess != nil {
		searchRecord.OnlineStatus = 1
	}
	return searchRecord
}

func (s ICQService) GetICQReqAck(ctx context.Context, sess *state.Session, seq uint16, subType uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMoreUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: 0x07DA,
				Seq:     seq,
			},
			ReqSubType: subType,
			Success:    0x0A,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) reply(ctx context.Context, sess *state.Session, userInfo wire.ICQMessage) error {
	buf := &bytes.Buffer{}
	if err := wire.MarshalLE(userInfo, buf); err != nil {
		return err
	}

	msg := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf.Bytes()),
				},
			},
		},
	}

	s.messageRelayer.RelayToScreenName(ctx, sess.IdentScreenName(), msg)
	return nil
}

func (s ICQService) UpdateBasicInfo(ctx context.Context, sess *state.Session, req wire.ICQUserInfoBasic, seq uint16) error {
	err := s.userUpdater.UpdateUser(sess.IdentScreenName(), func(u *state.User) {
		u.Nickname = req.Nickname
		u.FirstName = req.FirstName
		u.LastName = req.LastName
		u.EmailAddress = req.Email
		u.HomeCity = req.HomeCity
		u.HomeState = req.HomeState
		u.HomePhone = req.HomePhone
		u.HomeFax = req.HomeFax
		u.HomeAddress = req.HomeAddress
		u.CellPhone = req.CellPhone
		u.ZipCode = req.ZipCode
		u.CountryCode = req.CountryCode
		u.GMTOffset = req.GMTOffset
		if req.PublishEmail == 1 {
			u.PublishEmail = true
		}
	})

	if err != nil {
		return err
	}
	return s.GetICQReqAck(ctx, sess, seq, 0x0064)
}

func (s ICQService) UpdateWorkInfo(ctx context.Context, sess *state.Session, req wire.ICQWorkInfo, seq uint16) error {
	err := s.userUpdater.UpdateUser(sess.IdentScreenName(), func(u *state.User) {
		u.Company = req.Company
		u.Department = req.Department
		u.OccupationCode = req.OccupationCode
		u.Position = req.Position
		u.WorkAddress = req.WorkAddress
		u.WorkCity = req.WorkCity
		u.WorkCountryCode = req.WorkCountryCode
		u.WorkFax = req.WorkFax
		u.WorkPhone = req.WorkPhone
		u.WorkState = req.WorkState
		u.WorkWebPage = req.WorkWebPage
		u.WorkZIP = req.WorkZIP
	})

	if err != nil {
		return err
	}
	return s.GetICQReqAck(ctx, sess, seq, 0x006E)
}

func (s ICQService) UpdateMoreInfo(ctx context.Context, sess *state.Session, req wire.SomeMoreUserInfo, seq uint16) error {
	err := s.userUpdater.UpdateUser(sess.IdentScreenName(), func(u *state.User) {
		u.Gender = req.Gender
		u.HomePageAddr = req.HomePageAddr
		u.BirthYear = req.BirthYear
		u.BirthMonth = req.BirthMonth
		u.BirthDay = req.BirthDay
		u.Lang1 = req.Lang1
		u.Lang2 = req.Lang2
		u.Lang3 = req.Lang3
	})

	if err != nil {
		return err
	}
	return s.GetICQReqAck(ctx, sess, seq, 0x0078)
}

func (s ICQService) UpdateUserNotes(ctx context.Context, sess *state.Session, req wire.ICQNotes, seq uint16) error {
	err := s.userUpdater.UpdateUser(sess.IdentScreenName(), func(u *state.User) {
		u.Notes = req.Notes
	})

	if err != nil {
		return err
	}
	return s.GetICQReqAck(ctx, sess, seq, 0x0082)
}

func (s ICQService) UpdateInterests(ctx context.Context, sess *state.Session, req wire.ICQInterests, seq uint16) error {
	err := s.userUpdater.UpdateUser(sess.IdentScreenName(), func(u *state.User) {
		u.Interest1Code = req.Interests[0].Code
		u.Interest1Keyword = req.Interests[0].Keyword
		u.Interest2Code = req.Interests[1].Code
		u.Interest2Keyword = req.Interests[1].Keyword
		u.Interest3Code = req.Interests[2].Code
		u.Interest3Keyword = req.Interests[2].Keyword
		u.Interest4Code = req.Interests[3].Code
		u.Interest4Keyword = req.Interests[3].Keyword
	})

	if err != nil {
		return err
	}
	return s.GetICQReqAck(ctx, sess, seq, 0x008C)
}

func (s ICQService) UpdateAffiliations(ctx context.Context, sess *state.Session, req wire.ICQAffiliations, seq uint16) error {
	err := s.userUpdater.UpdateUser(sess.IdentScreenName(), func(u *state.User) {
		u.PastAffiliations1Code = req.PastAffiliations[0].Code
		u.PastAffiliations1Keyword = req.PastAffiliations[0].Keyword
		u.PastAffiliations2Code = req.PastAffiliations[1].Code
		u.PastAffiliations2Keyword = req.PastAffiliations[1].Keyword
		u.PastAffiliations3Code = req.PastAffiliations[2].Code
		u.PastAffiliations3Keyword = req.PastAffiliations[2].Keyword
		u.Affiliations1Code = req.Affiliations[0].Code
		u.Affiliations1Keyword = req.Affiliations[0].Keyword
		u.Affiliations2Code = req.Affiliations[1].Code
		u.Affiliations2Keyword = req.Affiliations[1].Keyword
		u.Affiliations3Code = req.Affiliations[2].Code
		u.Affiliations3Keyword = req.Affiliations[2].Keyword
	})

	if err != nil {
		return err
	}
	return s.GetICQReqAck(ctx, sess, seq, 0x0096)
}

func (s ICQService) UpdateEmails(ctx context.Context, sess *state.Session, req wire.ICQEmailUserInfo, seq uint16) error {
	if len(req.Emails) > 0 {
		s.logger.Debug("adding additional emails is not yet supported")
	}
	return s.GetICQReqAck(ctx, sess, seq, 0x0087)
}
