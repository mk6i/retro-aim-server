package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// OServiceService provides functionality for the OService food group, which
// provides an assortment of services useful across multiple food groups.
type OServiceService struct {
	buddyBroadcaster buddyBroadcaster
	cfg              config.Config
	logger           *slog.Logger
	foodGroups       []uint16
	rateLimitClasses wire.RateLimitClasses
	snacRateLimits   wire.SNACRateLimits
	timeNow          func() time.Time
}

// ClientVersions informs the server what food group versions the client
// supports and returns to the client what food group versions it supports.
// This method simply regurgitates versions supplied by the client in inBody
// back to the client in a OServiceHostVersions SNAC. The server doesn't
// attempt to accommodate any particular food group version. The server
// implicitly accommodates any food group version for Windows AIM clients 5.x.
// It returns SNAC wire.OServiceHostVersions containing the server's supported
// food group versions.
// todo this documentation
func (s OServiceService) ClientVersions(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x17_OServiceClientVersions) wire.SNACMessage {
	var versions [wire.MDir + 1]uint16

	if len(bodyIn.Versions)%2 != 0 {
		s.logger.ErrorContext(ctx, "got uneven food group length")
		return wire.SNACMessage{}
	}

	for i := 0; i < len(bodyIn.Versions); i += 2 {
		fg := bodyIn.Versions[i]
		if fg < wire.OService || fg > wire.MDir {
			s.logger.ErrorContext(ctx, "invalid food group ID", "id", fg)
			continue
		}
		ver := bodyIn.Versions[i+1]
		if ver < 1 {
			s.logger.ErrorContext(ctx, "invalid food group version", "version", ver)
			continue
		}
		versions[fg] = ver
	}

	sess.SetFoodGroupVersions(versions)

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostVersions,
			RequestID: frame.RequestID,
		},
		Body: wire.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: bodyIn.Versions,
		},
	}
}

// RateParamsQuery returns SNAC rate limits. It returns SNAC
// wire.OServiceRateParamsReply containing rate limits for all food groups
// supported by this server.
//
// The purpose of this method is to convey per-SNAC server-side rate limits to
// the client. The response consists of two main parts: rate classes and rate
// groups. Rate classes define limits based on specific parameters, while rate
// groups associate these limits with relevant SNAC types.
//
// The current implementation does not enforce server-side rate limiting.
// Instead, the provided values inform the client about the recommended
// client-side rate limits.
//
// The rate limit values were taken from the example SNAC dump documented here:
// https://web.archive.org/web/20221207225518/https://wiki.nina.chat/wiki/Protocols/OSCAR/SNAC/OSERVICE_RATE_PARAMS_REPLY
//
// AIM clients silently fail when they expect a rate limit rule that does not
// exist in this response. When support for a new food group is added to the
// server, update this function accordingly.
func (s OServiceService) RateParamsQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame) wire.SNACMessage {
	// not contain LastTime and CurrentStatus fields.
	var limits = wire.SNAC_0x01_0x07_OServiceRateParamsReply{
		RateClasses: []wire.RateParamsSNAC{},
		RateGroups: []struct {
			ID    uint16
			Pairs []struct {
				FoodGroup uint16
				SubGroup  uint16
			} `oscar:"count_prefix=uint16"`
		}{
			{
				ID: 1,
				Pairs: []struct {
					FoodGroup uint16
					SubGroup  uint16
				}{},
			},
			{
				ID: 2,
				Pairs: []struct {
					FoodGroup uint16
					SubGroup  uint16
				}{},
			},
			{
				ID: 3,
				Pairs: []struct {
					FoodGroup uint16
					SubGroup  uint16
				}{},
			},
			{
				ID: 4,
				Pairs: []struct {
					FoodGroup uint16
					SubGroup  uint16
				}{},
			},
			{
				ID: 5,
				Pairs: []struct {
					FoodGroup uint16
					SubGroup  uint16
				}{},
			},
		},
	}

	for _, class := range s.rateLimitClasses.All() {
		str := wire.RateParamsSNAC{
			ID:              uint16(class.ID),
			WindowSize:      uint32(class.WindowSize),
			ClearLevel:      uint32(class.ClearLevel),
			AlertLevel:      uint32(class.AlertLevel),
			LimitLevel:      uint32(class.LimitLevel),
			DisconnectLevel: uint32(class.DisconnectLevel),
			CurrentLevel:    uint32(class.MaxLevel),
			MaxLevel:        uint32(class.MaxLevel),
		}
		if sess.FoodGroupVersions()[wire.OService] > 1 {
			str.V2Params = &struct {
				LastTime      uint32
				DroppingSNACs uint8
			}{
				LastTime: uint32(s.timeNow().Add(-time.Second).Unix()),
			}
		}
		limits.RateClasses = append(limits.RateClasses, str)
	}

	for snacClass := range s.snacRateLimits.All() {
		classID := int(snacClass.RateLimitClass) - 1
		limits.RateGroups[classID].Pairs = append(limits.RateGroups[classID].Pairs,
			struct {
				FoodGroup uint16
				SubGroup  uint16
			}{FoodGroup: snacClass.FoodGroup, SubGroup: snacClass.SubGroup})
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamsReply,
			RequestID: inFrame.RequestID,
		},
		Body: limits,
	}
}

// UserInfoQuery returns SNAC wire.OServiceUserInfoUpdate containing
// the user's info.
func (s OServiceService) UserInfoQuery(_ context.Context, sess *state.Session, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}
}

// SetUserInfoFields sets the user's visibility status to visible or invisible.
// The visibility status is set according to the inFrame TLV entry under key
// wire.OServiceUserInfoStatus. If the value is 0x0000, set invisible. If set
// to 0x0100, set invisible. Else, return an error for any other value.
// It returns SNAC wire.OServiceUserInfoUpdate containing the user's info.
func (s OServiceService) SetUserInfoFields(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (wire.SNACMessage, error) {
	if status, hasStatus := inBody.Uint32BE(wire.OServiceUserInfoStatus); hasStatus {
		sess.SetUserStatusBitmask(status)
		if sess.Invisible() {
			if err := s.buddyBroadcaster.BroadcastBuddyDeparted(ctx, sess); err != nil {
				return wire.SNACMessage{}, err
			}
		} else {
			if err := s.buddyBroadcaster.BroadcastBuddyArrived(ctx, sess); err != nil {
				return wire.SNACMessage{}, err
			}

		}
	}
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}, nil
}

// IdleNotification sets the user idle time.
// Set session idle time to the value of bodyIn.IdleTime. Return a user arrival
// message to all users who have this user on their buddy list.
func (s OServiceService) IdleNotification(ctx context.Context, sess *state.Session, bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification) error {
	if bodyIn.IdleTime == 0 {
		sess.UnsetIdle()
	} else {
		sess.SetIdle(time.Duration(bodyIn.IdleTime) * time.Second)
	}
	return s.buddyBroadcaster.BroadcastBuddyArrived(ctx, sess)
}

// SetPrivacyFlags sets client privacy settings. Currently, there's no action
// to take when these flags are set. This method simply logs the flags set by
// the client.
func (s OServiceService) SetPrivacyFlags(ctx context.Context, bodyIn wire.SNAC_0x01_0x14_OServiceSetPrivacyFlags) {
	attrs := slog.Group("request",
		slog.String("food_group", wire.FoodGroupName(wire.OService)),
		slog.String("sub_group", wire.SubGroupName(wire.OService, wire.OServiceSetPrivacyFlags)))

	if bodyIn.MemberFlag() {
		s.logger.LogAttrs(ctx, slog.LevelDebug, "client set member privacy flag, but we're not going to do anything", attrs)
	}
	if bodyIn.IdleFlag() {
		s.logger.LogAttrs(ctx, slog.LevelDebug, "client set idle privacy flag, but we're not going to do anything", attrs)
	}
}

// RateParamsSubAdd subscribes to rate parameter changes. AOL's OSCAR spec says
// that notifications will be queued after calling this method. I don't see the
// point of doing that since all clients appear to call RateParamsQuery at
// sign-on for all rate classes.
func (s OServiceService) RateParamsSubAdd(ctx context.Context, sess *state.Session, snac wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
	ids := make([]wire.RateLimitClassID, 0, len(snac.ClassIDs))

	for _, id := range snac.ClassIDs {
		if id < 1 || id > 5 {
			s.logger.DebugContext(ctx, "snac class ID out of range")
			continue
		}
		ids = append(ids, wire.RateLimitClassID(id))
	}

	if len(ids) == 0 {
		return
	}

	s.logger.DebugContext(ctx, "subscribing to rate limit updates", "classes", ids)
	sess.SubscribeRateLimits(ids)
}

func (s OServiceService) ServiceRequest(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x04_OServiceServiceRequest) (wire.SNACMessage, error) {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceErr,
			RequestID: frame.RequestID,
		},
		Body: wire.SNACError{
			Code: wire.ErrorCodeNotSupportedByHost,
		},
	}, nil
}

// ClientOnline informs the server that the client is ready.
func (s OServiceService) ClientOnline(ctx context.Context, bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	s.logger.DebugContext(ctx, "client is online", "group_versions", bodyIn.GroupVersions)
	return nil
}

// HostOnline initiates the Alert protocol sequence.
// It returns SNAC wire.OServiceHostOnline containing the list of food groups
// supported by the Alert service.
// Alert is provided by BOS in addition to the standalone Alert service.
// AIM 4.x always creates a secondary TCP connection for Alert, whereas 5.x
// can use the existing BOS connection for Alert services.
func (s OServiceService) HostOnline() wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
			RequestID: wire.ReqIDFromServer,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: s.foodGroups,
		},
	}
}

// RateLimitUpdates produces update messages reflecting any recent changes in
// rate limit class params or rate limit states for the current session.
// Changes are reported relative to the previous invocation for this session.
// Only newly observed transitions or updated rate parameters will be included.
func (s OServiceService) RateLimitUpdates(ctx context.Context, sess *state.Session, now time.Time) []wire.SNACMessage {
	msgs := make([]wire.SNACMessage, 0, 5)
	classDelta, stateDelta := sess.ObserveRateChanges(now)

	for _, curRate := range classDelta {
		s.logger.DebugContext(ctx, "rate limit class changed", "class", curRate.ID)
		msgs = append(msgs, buildRateLimitUpdate(1, curRate, sess, now))
	}

	for _, curRate := range stateDelta {
		s.logger.DebugContext(ctx, "rate limit state changed",
			"class", curRate.ID,
			"state", curRate.CurrentStatus)
		var code uint16
		switch curRate.CurrentStatus {
		case wire.RateLimitStatusLimited:
			code = 3
		case wire.RateLimitStatusAlert:
			code = 2
		case wire.RateLimitStatusClear:
			code = 4
		case wire.RateLimitStatusDisconnect:
			s.logger.DebugContext(ctx, "rate limit status disconnected, no point in returning status update")
			continue
		}

		msgs = append(msgs, buildRateLimitUpdate(code, curRate, sess, now))
	}

	return msgs
}

// buildRateLimitUpdate constructs a SNAC message notifying the client of a rate limit
// threshold update or a change in rate limiting status for a specific class.
//
// The message format varies depending on the client's supported protocol version.
// If OService version 2 or higher is supported, additional metadata such as
// time since last status change and whether SNACs are currently being dropped
// will be included.
func buildRateLimitUpdate(code uint16, curRate state.RateClassState, sess *state.Session, now time.Time) wire.SNACMessage {
	var droppingSNACs uint8
	if curRate.CurrentStatus == wire.RateLimitStatusLimited {
		droppingSNACs = 1
	}

	rate := wire.RateParamsSNAC{
		ID:              uint16(curRate.ID),
		WindowSize:      uint32(curRate.WindowSize),
		ClearLevel:      uint32(curRate.ClearLevel),
		AlertLevel:      uint32(curRate.AlertLevel),
		LimitLevel:      uint32(curRate.LimitLevel),
		DisconnectLevel: uint32(curRate.DisconnectLevel),
		CurrentLevel:    uint32(curRate.CurrentLevel),
		MaxLevel:        uint32(curRate.MaxLevel),
	}

	if sess.FoodGroupVersions()[wire.OService] > 1 {
		rate.V2Params = &struct {
			LastTime      uint32
			DroppingSNACs uint8
		}{
			LastTime:      uint32(max(0, now.Unix()-curRate.LastTime.Unix())),
			DroppingSNACs: droppingSNACs,
		}
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamChange,
			RequestID: wire.ReqIDFromServer,
		},
		Body: wire.SNAC_0x01_0x0A_OServiceRateParamsChange{
			Code: code,
			Rate: rate,
		},
	}
}

// NewOServiceServiceForBOS creates a new instance of OServiceServiceForBOS.
func NewOServiceServiceForBOS(
	cfg config.Config,
	messageRelayer MessageRelayer,
	logger *slog.Logger,
	cookieIssuer CookieBaker,
	chatRoomManager ChatRoomRegistry,
	relationshipFetcher RelationshipFetcher,
	sessionRetriever SessionRetriever,
	buddyIconManager BuddyIconManager,
	rateLimitClasses wire.RateLimitClasses,
	snacRateLimits wire.SNACRateLimits,
) *OServiceServiceForBOS {
	return &OServiceServiceForBOS{
		chatRoomManager: chatRoomManager,
		cookieIssuer:    cookieIssuer,
		messageRelayer:  messageRelayer,
		OServiceService: OServiceService{
			buddyBroadcaster: newBuddyNotifier(buddyIconManager, relationshipFetcher, messageRelayer, sessionRetriever),
			cfg:              cfg,
			logger:           logger,
			foodGroups: []uint16{
				wire.Alert,
				wire.BART,
				wire.Buddy,
				wire.Feedbag,
				wire.ICBM,
				wire.ICQ,
				wire.Locate,
				wire.OService,
				wire.PermitDeny,
				wire.UserLookup,
				wire.Invite,
				wire.Popup,
				wire.Stats,
			},
			rateLimitClasses: rateLimitClasses,
			snacRateLimits:   snacRateLimits,
			timeNow:          time.Now,
		},
	}
}

// OServiceServiceForBOS provides functionality for the OService food group
// running on the BOS server.
type OServiceServiceForBOS struct {
	OServiceService
	chatRoomManager ChatRoomRegistry
	cookieIssuer    CookieBaker
	messageRelayer  MessageRelayer
}

// chatLoginCookie represents credentials used to authenticate a user chat
// session.
type chatLoginCookie struct {
	ChatCookie string                  `oscar:"len_prefix=uint8"`
	ScreenName state.DisplayScreenName `oscar:"len_prefix=uint8"`
}

// ServiceRequest handles service discovery, providing a host name and metadata
// for connecting to the food group service specified in inFrame.
func (s OServiceServiceForBOS) ServiceRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x01_0x04_OServiceServiceRequest) (wire.SNACMessage, error) {
	fnIssueCookie := func(val any) ([]byte, error) {
		buf := &bytes.Buffer{}
		if err := wire.MarshalBE(val, buf); err != nil {
			return nil, err
		}
		return s.cookieIssuer.Issue(buf.Bytes())
	}

	switch inBody.FoodGroup {
	case wire.Admin:
		cookie, err := fnIssueCookie(bosCookie{
			ScreenName: sess.DisplayScreenName(),
		})
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.AdminPort)),
						wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.Admin),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.Alert:
		cookie, err := fnIssueCookie(bosCookie{
			ScreenName: sess.DisplayScreenName(),
		})
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.AlertPort)),
						wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.Alert),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.BART:
		cookie, err := fnIssueCookie(bosCookie{
			ScreenName: sess.DisplayScreenName(),
		})
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.BARTPort)),
						wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.BART),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.ChatNav:
		cookie, err := fnIssueCookie(bosCookie{
			ScreenName: sess.DisplayScreenName(),
		})
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.ChatNavPort)),
						wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.ChatNav),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.Chat:
		roomMeta, ok := inBody.Bytes(0x01)
		if !ok {
			return wire.SNACMessage{}, errors.New("missing room info")
		}

		roomSNAC := wire.SNAC_0x01_0x04_TLVRoomInfo{}
		if err := wire.UnmarshalBE(&roomSNAC, bytes.NewBuffer(roomMeta)); err != nil {
			return wire.SNACMessage{}, err
		}

		room, err := s.chatRoomManager.ChatRoomByCookie(ctx, roomSNAC.Cookie)
		if err != nil {
			return wire.SNACMessage{}, fmt.Errorf("unable to retrieve room info: %w", err)
		}

		cookie, err := fnIssueCookie(chatLoginCookie{
			ChatCookie: room.Cookie(),
			ScreenName: sess.DisplayScreenName(),
		})
		if err != nil {
			return wire.SNACMessage{}, err
		}

		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.ChatPort)),
						wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.Chat),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	case wire.ODir:
		cookie, err := fnIssueCookie(bosCookie{
			ScreenName: sess.DisplayScreenName(),
		})
		if err != nil {
			return wire.SNACMessage{}, err
		}
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceServiceResponse,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.OServiceTLVTagsReconnectHere, net.JoinHostPort(s.cfg.OSCARHost, s.cfg.ODirPort)),
						wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
						wire.NewTLVBE(wire.OServiceTLVTagsGroupID, wire.ODir),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLCertName, ""),
						wire.NewTLVBE(wire.OServiceTLVTagsSSLState, uint8(0x00)),
					},
				},
			},
		}, nil
	default:
		s.logger.InfoContext(ctx, "client service request for unsupported service", "food_group", wire.FoodGroupName(inBody.FoodGroup))
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceErr,
				RequestID: inFrame.RequestID,
			},
			Body: wire.SNACError{
				Code: wire.ErrorCodeServiceUnavailable,
			},
		}, nil
	}
}

// ClientOnline runs when the current user is ready to join.
// It announces current user's arrival to users who have the current user on
// their buddy list.
func (s OServiceServiceForBOS) ClientOnline(ctx context.Context, _ wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	sess.SetSignonComplete()

	if err := s.buddyBroadcaster.BroadcastVisibility(ctx, sess, nil, false); err != nil {
		return fmt.Errorf("unable to send buddy arrival notification: %w", err)
	}

	return nil
}

// NewOServiceServiceForChat creates a new instance of NewOServiceServiceForChat.
func NewOServiceServiceForChat(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	chatRoomManager ChatRoomRegistry,
	chatMessageRelayer ChatMessageRelayer,
	relationshipFetcher RelationshipFetcher,
	sessionRetriever SessionRetriever,
	buddyIconManager BuddyIconManager,
	rateLimitClasses wire.RateLimitClasses,
	snacRateLimits wire.SNACRateLimits,
) *OServiceServiceForChat {
	return &OServiceServiceForChat{
		OServiceService: OServiceService{
			buddyBroadcaster: newBuddyNotifier(buddyIconManager, relationshipFetcher, messageRelayer, sessionRetriever),
			cfg:              cfg,
			logger:           logger,
			foodGroups: []uint16{
				wire.OService,
				wire.Chat,
			},
			rateLimitClasses: rateLimitClasses,
			snacRateLimits:   snacRateLimits,
			timeNow:          time.Now,
		},
		chatRoomManager:    chatRoomManager,
		chatMessageRelayer: chatMessageRelayer,
	}
}

// OServiceServiceForChat provides functionality for the OService food group
// running on the Chat server.
type OServiceServiceForChat struct {
	OServiceService
	chatRoomManager    ChatRoomRegistry
	chatMessageRelayer ChatMessageRelayer
}

// ClientOnline runs when the current user is ready to join the chat.
// Trigger the following actions:
//   - Send current user the chat room metadata
//   - Announce current user's arrival to other chat room participants
//   - Send current user the chat room participant list
func (s OServiceServiceForChat) ClientOnline(ctx context.Context, _ wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error {
	room, err := s.chatRoomManager.ChatRoomByCookie(ctx, sess.ChatRoomCookie())
	if err != nil {
		return fmt.Errorf("error getting chat room: %w", err)
	}

	// Do not change the order of the following 3 methods. macOS client v4.0.9
	// requires this exact sequence, otherwise the chat session prematurely
	// closes seconds after users join a chat room.
	setOnlineChatUsers(ctx, sess, s.chatMessageRelayer)
	sendChatRoomInfoUpdate(ctx, sess, s.chatMessageRelayer, room)
	alertUserJoined(ctx, sess, s.chatMessageRelayer)

	return nil
}

// NewOServiceServiceForChatNav creates a new instance of OServiceService for
// ChatNav.
func NewOServiceServiceForChatNav(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	relationshipFetcher RelationshipFetcher,
	sessionRetriever SessionRetriever,
	buddyIconManager BuddyIconManager,
	rateLimitClasses wire.RateLimitClasses,
	snacRateLimits wire.SNACRateLimits,
) *OServiceService {
	return &OServiceService{
		buddyBroadcaster: newBuddyNotifier(buddyIconManager, relationshipFetcher, messageRelayer, sessionRetriever),
		cfg:              cfg,
		logger:           logger,
		foodGroups: []uint16{
			wire.ChatNav,
			wire.OService,
		},
		rateLimitClasses: rateLimitClasses,
		snacRateLimits:   snacRateLimits,
		timeNow:          time.Now,
	}
}

// NewOServiceServiceForAlert creates a new instance of OServiceService for the Alert
// server.
func NewOServiceServiceForAlert(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	relationshipFetcher RelationshipFetcher,
	sessionRetriever SessionRetriever,
	buddyIconManager BuddyIconManager,
	rateLimitClasses wire.RateLimitClasses,
	snacRateLimits wire.SNACRateLimits,
) *OServiceService {
	return &OServiceService{
		buddyBroadcaster: newBuddyNotifier(buddyIconManager, relationshipFetcher, messageRelayer, sessionRetriever),
		cfg:              cfg,
		logger:           logger,
		foodGroups: []uint16{
			wire.Alert,
			wire.OService,
		},
		rateLimitClasses: rateLimitClasses,
		snacRateLimits:   snacRateLimits,
		timeNow:          time.Now,
	}
}

// NewOServiceServiceForODir creates a new instance of OServiceService for the
// ODir server.
func NewOServiceServiceForODir(
	cfg config.Config,
	logger *slog.Logger,
	rateLimitClasses wire.RateLimitClasses,
	snacRateLimits wire.SNACRateLimits,
) *OServiceService {
	return &OServiceService{
		cfg:    cfg,
		logger: logger,
		foodGroups: []uint16{
			wire.ODir,
			wire.OService,
		},
		rateLimitClasses: rateLimitClasses,
		snacRateLimits:   snacRateLimits,
		timeNow:          time.Now,
	}
}

// NewOServiceServiceForAdmin creates a new instance of OServiceService for Admin server.
func NewOServiceServiceForAdmin(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	relationshipFetcher RelationshipFetcher,
	sessionRetriever SessionRetriever,
	buddyIconManager BuddyIconManager,
	rateLimitClasses wire.RateLimitClasses,
	snacRateLimits wire.SNACRateLimits,
) *OServiceService {
	return &OServiceService{
		buddyBroadcaster: newBuddyNotifier(buddyIconManager, relationshipFetcher, messageRelayer, sessionRetriever),
		cfg:              cfg,
		logger:           logger,
		foodGroups: []uint16{
			wire.OService,
			wire.Admin,
		},
		rateLimitClasses: rateLimitClasses,
		snacRateLimits:   snacRateLimits,
		timeNow:          time.Now,
	}
}

// NewOServiceServiceForBART creates a new instance of OServiceService for the
// BART server.
func NewOServiceServiceForBART(
	cfg config.Config,
	logger *slog.Logger,
	messageRelayer MessageRelayer,
	relationshipFetcher RelationshipFetcher,
	sessionRetriever SessionRetriever,
	buddyIconManager BuddyIconManager,
	rateLimitClasses wire.RateLimitClasses,
	snacRateLimits wire.SNACRateLimits,
) *OServiceService {
	return &OServiceService{
		buddyBroadcaster: newBuddyNotifier(buddyIconManager, relationshipFetcher, messageRelayer, sessionRetriever),
		cfg:              cfg,
		logger:           logger,
		foodGroups: []uint16{
			wire.BART,
			wire.OService,
		},
		rateLimitClasses: rateLimitClasses,
		snacRateLimits:   snacRateLimits,
		timeNow:          time.Now,
	}
}
