package foodgroup

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewICQService creates an instance of ICQService.
func NewICQService(
	messageRelayer MessageRelayer,
	finder ICQUserFinder,
	userUpdater ICQUserUpdater,
	logger *slog.Logger,
	sessionRetriever SessionRetriever,
) ICQService {
	return ICQService{
		messageRelayer:   messageRelayer,
		userFinder:       finder,
		userUpdater:      userUpdater,
		logger:           logger,
		sessionRetriever: sessionRetriever,
		timeNow:          time.Now,
	}
}

// ICQService provides functionality for the ICQ food group.
type ICQService struct {
	userFinder       ICQUserFinder
	logger           *slog.Logger
	messageRelayer   MessageRelayer
	sessionRetriever SessionRetriever
	userUpdater      ICQUserUpdater
	timeNow          func() time.Time
}

func (s ICQService) FindByDetails(ctx context.Context, sess *state.Session, req wire.ICQFindByDetails, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: wire.ICQDBQueryMetaReply,
			Seq:     seq,
		},
		Success:    wire.ICQStatusCodeOK,
		ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
	}

	res, err := s.userFinder.FindByDetails(req.FirstName, req.LastName, req.NickName)

	if err != nil {
		s.logger.Error("FindByDetails failed", "err", err.Error())
		resp.Success = wire.ICQStatusCodeErr
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}
	if len(res) == 0 {
		resp.Success = wire.ICQStatusCodeFail
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}

	for i := 0; i < len(res); i++ {
		if i == len(res)-1 {
			resp.LastResult()
		} else {
			resp.ReqSubType = wire.ICQDBQueryMetaReplyUserFound
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

func (s ICQService) FindByEmail(ctx context.Context, sess *state.Session, req wire.ICQFindByEmail, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: wire.ICQDBQueryMetaReply,
			Seq:     seq,
		},
		ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
		Success:    wire.ICQStatusCodeOK,
	}
	resp.LastResult()

	res, err := s.userFinder.FindByEmail(req.Email)

	switch {
	case errors.Is(err, state.ErrNoUser):
		resp.Success = wire.ICQStatusCodeFail
	case err != nil:
		s.logger.Error("FindByEmail failed", "err", err.Error())
		resp.Success = wire.ICQStatusCodeErr
	default:
		resp.Success = wire.ICQStatusCodeOK
		resp.Details = s.createResult(res)
	}

	return s.reply(ctx, sess, wire.ICQMessage{
		Message: resp,
	})
}

func (s ICQService) FindByInterests(ctx context.Context, sess *state.Session, req wire.ICQFindByWhitePages, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: wire.ICQDBQueryMetaReply,
			Seq:     seq,
		},
		Success:    wire.ICQStatusCodeOK,
		ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
	}

	interests := strings.Split(req.InterestsKeyword, ",")
	res, err := s.userFinder.FindByInterests(req.InterestsCode, interests)

	if err != nil {
		s.logger.Error("FindByInterests failed", "err", err.Error())
		resp.Success = wire.ICQStatusCodeErr
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}
	if len(res) == 0 {
		resp.Success = wire.ICQStatusCodeFail
		return s.reply(ctx, sess, wire.ICQMessage{
			Message: resp,
		})
	}

	for i := 0; i < len(res); i++ {
		if i == len(res)-1 {
			resp.LastResult()
		} else {
			resp.ReqSubType = wire.ICQDBQueryMetaReplyUserFound
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

func (s ICQService) FindByUIN(ctx context.Context, sess *state.Session, req wire.ICQFindByUIN, seq uint16) error {
	resp := wire.ICQUserSearchResult{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: wire.ICQDBQueryMetaReply,
			Seq:     seq,
		},
		ReqSubType: wire.ICQDBQueryMetaReplyLastUserFound,
		Success:    wire.ICQStatusCodeOK,
	}
	resp.LastResult()

	res, err := s.userFinder.FindByUIN(req.UIN)

	switch {
	case errors.Is(err, state.ErrNoUser):
		resp.Success = wire.ICQStatusCodeFail
	case err != nil:
		s.logger.Error("FindByUIN failed", "err", err.Error())
		resp.Success = wire.ICQStatusCodeErr
	default:
		resp.Success = wire.ICQStatusCodeOK
		resp.Details = s.createResult(res)
	}

	return s.reply(ctx, sess, wire.ICQMessage{
		Message: resp,
	})
}

func (s ICQService) FullUserInfo(ctx context.Context, sess *state.Session, req wire.ICQFindByUIN, seq uint16) error {

	user, err := s.userFinder.FindByUIN(req.UIN)
	if err != nil {
		return err
	}

	if err := s.userInfo(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.moreUserInfo(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.extraEmails(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.homepageCat(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.workInfo(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.notes(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.interests(ctx, sess, user, seq); err != nil {
		return err
	}

	if err := s.affiliations(ctx, sess, user, seq); err != nil {
		return err
	}
	return nil
}

func (s ICQService) MessagesEOF(ctx context.Context, sess *state.Session, seq uint16) error {
	s.logger.Debug("returning offline messages is not yet supported")
	msg := wire.ICQMessage{
		Message: wire.ICQMessagesEOF{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryOfflineMsgReplyLast,
				Seq:     seq,
			},
			DroppedMessages: 0,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) SetAffiliations(ctx context.Context, sess *state.Session, req wire.ICQAffiliations, seq uint16) error {
	u := state.ICQAffiliations{
		PastAffiliations1Code:    req.PastAffiliations[0].Code,
		PastAffiliations1Keyword: req.PastAffiliations[0].Keyword,
		PastAffiliations2Code:    req.PastAffiliations[1].Code,
		PastAffiliations2Keyword: req.PastAffiliations[1].Keyword,
		PastAffiliations3Code:    req.PastAffiliations[2].Code,
		PastAffiliations3Keyword: req.PastAffiliations[2].Keyword,
		Affiliations1Code:        req.Affiliations[0].Code,
		Affiliations1Keyword:     req.Affiliations[0].Keyword,
		Affiliations2Code:        req.Affiliations[1].Code,
		Affiliations2Keyword:     req.Affiliations[1].Keyword,
		Affiliations3Code:        req.Affiliations[2].Code,
		Affiliations3Keyword:     req.Affiliations[2].Keyword,
	}

	if err := s.userUpdater.SetAffiliations(sess.IdentScreenName(), u); err != nil {
		return err
	}

	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetAffiliations)
}

func (s ICQService) SetBasicInfo(ctx context.Context, sess *state.Session, req wire.ICQUserInfoBasic, seq uint16) error {
	u := state.ICQUserInfoBasic{
		CellPhone:    req.CellPhone,
		CountryCode:  req.CountryCode,
		EmailAddress: req.EmailAddress,
		FirstName:    req.FirstName,
		GMTOffset:    req.GMTOffset,
		HomeAddress:  req.HomeAddress,
		HomeCity:     req.HomeCity,
		HomeFax:      req.HomeFax,
		HomePhone:    req.HomePhone,
		HomeState:    req.HomeState,
		LastName:     req.LastName,
		Nickname:     req.Nickname,
		PublishEmail: req.PublishEmail == wire.ICQUserFlagPublishEmailYes,
		ZipCode:      req.ZipCode,
	}

	if err := s.userUpdater.SetBasicInfo(sess.IdentScreenName(), u); err != nil {
		return err
	}

	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetBasicInfo)
}

func (s ICQService) SetEmails(ctx context.Context, sess *state.Session, req wire.ICQEmailUserInfo, seq uint16) error {
	if len(req.Emails) > 0 {
		s.logger.Debug("adding additional emails is not yet supported")
	}
	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetEmails)
}

func (s ICQService) SetInterests(ctx context.Context, sess *state.Session, req wire.ICQInterests, seq uint16) error {
	// code defensively, check that there are 4 interests
	u := state.ICQInterests{
		Interest1Code:    req.Interests[0].Code,
		Interest1Keyword: req.Interests[0].Keyword,
		Interest2Code:    req.Interests[1].Code,
		Interest2Keyword: req.Interests[1].Keyword,
		Interest3Code:    req.Interests[2].Code,
		Interest3Keyword: req.Interests[2].Keyword,
		Interest4Code:    req.Interests[3].Code,
		Interest4Keyword: req.Interests[3].Keyword,
	}

	if err := s.userUpdater.SetInterests(sess.IdentScreenName(), u); err != nil {
		return err
	}

	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetInterests)
}

func (s ICQService) SetMoreInfo(ctx context.Context, sess *state.Session, req wire.SomeMoreUserInfo, seq uint16) error {
	u := state.ICQMoreInfo{
		Gender:       req.Gender,
		HomePageAddr: req.HomePageAddr,
		BirthYear:    req.BirthYear,
		BirthMonth:   req.BirthMonth,
		BirthDay:     req.BirthDay,
		Lang1:        req.Lang1,
		Lang2:        req.Lang2,
		Lang3:        req.Lang3,
	}

	if err := s.userUpdater.SetMoreInfo(sess.IdentScreenName(), u); err != nil {
		return err
	}

	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetMoreInfo)
}

func (s ICQService) SetPermissions(ctx context.Context, sess *state.Session, req wire.ICQInfoSetPerms, seq uint16) error {
	s.logger.Debug("setting permissions is not yet supported")
	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetPermissions)
}

func (s ICQService) SetUserNotes(ctx context.Context, sess *state.Session, req wire.ICQNotes, seq uint16) error {
	u := state.ICQUserNotes{
		Notes: req.Notes,
	}

	if err := s.userUpdater.SetUserNotes(sess.IdentScreenName(), u); err != nil {
		return err
	}

	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetNotes)
}

func (s ICQService) SetWorkInfo(ctx context.Context, sess *state.Session, req wire.ICQWorkInfo, seq uint16) error {
	icqWorkInfo := state.ICQWorkInfo{
		Company:         req.Company,
		Department:      req.Department,
		OccupationCode:  req.OccupationCode,
		Position:        req.Position,
		WorkAddress:     req.WorkAddress,
		WorkCity:        req.WorkCity,
		WorkCountryCode: req.WorkCountryCode,
		WorkFax:         req.WorkFax,
		WorkPhone:       req.WorkPhone,
		WorkState:       req.WorkState,
		WorkWebPage:     req.WorkWebPage,
		WorkZIP:         req.WorkZIP,
	}

	if err := s.userUpdater.SetWorkInfo(sess.IdentScreenName(), icqWorkInfo); err != nil {
		return err
	}

	return s.reqAck(ctx, sess, seq, wire.ICQDBQueryMetaReplySetWorkInfo)
}

func (s ICQService) affiliations(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMetaAffiliationsUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: wire.ICQDBQueryMetaReplyAffiliations,
			Success:    wire.ICQStatusCodeOK,
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

func (s ICQService) createResult(res state.User) wire.ICQUserSearchRecord {
	uin, _ := strconv.Atoi(res.IdentScreenName.String())

	searchRecord := wire.ICQUserSearchRecord{
		UIN:       uint32(uin),
		Nickname:  res.Nickname,
		FirstName: res.FirstName,
		LastName:  res.LastName,
		Email:     res.EmailAddress,
		Gender:    uint8(res.Gender),
		Age:       res.Age(s.timeNow),
	}
	if res.AuthReq {
		searchRecord.Authorization = 1
	}

	userSess := s.sessionRetriever.RetrieveSession(res.IdentScreenName)
	if userSess != nil {
		searchRecord.OnlineStatus = 1
	}
	return searchRecord
}

func (s ICQService) extraEmails(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQInfoEmailMore{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: wire.ICQDBQueryMetaReplyExtEmailInfo,
			Success:    wire.ICQStatusCodeOK,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) homepageCat(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQHomepageCat{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: wire.ICQDBQueryMetaReplyHomePageCat,
			Success:    wire.ICQStatusCodeOK,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) interests(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQUserInterests{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: wire.ICQDBQueryMetaReplyInterests,
			Success:    wire.ICQStatusCodeOK,
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

func (s ICQService) moreUserInfo(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMoreUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: wire.ICQDBQueryMetaReplyMoreInfo,
			Success:    wire.ICQStatusCodeOK,
			SomeMoreUserInfo: wire.SomeMoreUserInfo{
				Age:          uint8(user.Age(s.timeNow)),
				Gender:       user.Gender,
				HomePageAddr: user.HomePageAddr,
				BirthYear:    user.BirthYear,
				BirthMonth:   user.BirthMonth,
				BirthDay:     user.BirthDay,
				Lang1:        user.Lang1,
				Lang2:        user.Lang2,
				Lang3:        user.Lang3,
			},
			City:        user.HomeCity,
			State:       user.HomeState,
			CountryCode: user.CountryCode,
			TimeZone:    user.GMTOffset,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) notes(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQUserNotes{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: wire.ICQDBQueryMetaReplyNotes,
			Success:    wire.ICQStatusCodeOK,
			ICQNotes: wire.ICQNotes{
				Notes: user.Notes,
			},
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) reply(ctx context.Context, sess *state.Session, message wire.ICQMessage) error {
	msg := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.ICQTLVTagsMetadata, message),
				},
			},
		},
	}

	s.messageRelayer.RelayToScreenName(ctx, sess.IdentScreenName(), msg)
	return nil
}

func (s ICQService) reqAck(ctx context.Context, sess *state.Session, seq uint16, subType uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMoreUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: subType,
			Success:    wire.ICQStatusCodeOK,
		},
	}

	return s.reply(ctx, sess, msg)
}

func (s ICQService) userInfo(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	userInfo := wire.ICQUserInfo{
		ICQMetadata: wire.ICQMetadata{
			UIN:     sess.UIN(),
			ReqType: wire.ICQDBQueryMetaReply,
			Seq:     seq,
		},
		ReqSubType:  wire.ICQDBQueryMetaReplyBasicInfo,
		Success:     wire.ICQStatusCodeOK,
		Nickname:    user.Nickname,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.EmailAddress,
		HomeCity:    user.HomeCity,
		HomeState:   user.HomeState,
		HomePhone:   user.HomePhone,
		HomeFax:     user.HomeFax,
		HomeAddress: user.HomeAddress,
		CellPhone:   user.CellPhone,
		ZipCode:     user.ZipCode,
		CountryCode: user.CountryCode,
		GMTOffset:   user.GMTOffset,
		AuthFlag:    0, // todo figure these out
		WebAware:    1, // todo figure these out
		DCPerms:     0, // todo figure these out
	}

	if user.PublishEmail {
		userInfo.PublishEmail = wire.ICQUserFlagPublishEmailYes
	} else {
		userInfo.PublishEmail = wire.ICQUserFlagPublishEmailNo
	}

	msg := wire.ICQMessage{
		Message: userInfo,
	}
	return s.reply(ctx, sess, msg)

}

func (s ICQService) workInfo(ctx context.Context, sess *state.Session, user state.User, seq uint16) error {
	msg := wire.ICQMessage{
		Message: wire.ICQMetaWorkUserInfo{
			ICQMetadata: wire.ICQMetadata{
				UIN:     sess.UIN(),
				ReqType: wire.ICQDBQueryMetaReply,
				Seq:     seq,
			},
			ReqSubType: wire.ICQDBQueryMetaReplyWorkInfo,
			Success:    wire.ICQStatusCodeOK,
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
