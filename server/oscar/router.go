package oscar

import (
	"context"
	"errors"
	"io"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// ErrRouteNotFound is an error that indicates a failure to find a matching
// route for an OSCAR protocol request.
var ErrRouteNotFound = errors.New("route not found")

// ResponseWriter is the interface for sending a SNAC response to the client
// from the server handlers.
type ResponseWriter interface {
	SendSNAC(frame wire.SNACFrame, body any) error
}

// Router defines a structure for routing OSCAR protocol requests to
// appropriate handlers based on group:subGroup identifiers.
type Router struct {
	AdminHandler
	AlertHandler
	BARTHandler
	BuddyHandler
	ChatHandler
	ChatNavHandler
	FeedbagHandler
	ICBMHandler
	ICQHandler
	LocateHandler
	ODirHandler
	OServiceHandler
	PermitDenyHandler
	StatsHandler
	UserLookupHandler
}

// Handle directs an incoming OSCAR request to the appropriate handler based on
// its group and subGroup identifiers found in the SNAC frame. It returns an
// ErrRouteNotFound error if no matching handler is found for the group:subGroup
// pair in the request.
func (rt Router) Handle(ctx context.Context, server uint16, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter, connectHere string) error {
	switch {
	case inFrame.FoodGroup == wire.Admin && inFrame.SubGroup == wire.AdminAcctConfirmRequest:
		return rt.AdminHandler.ConfirmRequest(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Admin && inFrame.SubGroup == wire.AdminInfoChangeRequest:
		return rt.AdminHandler.InfoChangeRequest(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Admin && inFrame.SubGroup == wire.AdminInfoQuery:
		return rt.AdminHandler.InfoQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Alert && inFrame.SubGroup == wire.AlertNotifyCapabilities:
		return rt.AlertHandler.NotifyCapabilities(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Alert && inFrame.SubGroup == wire.AlertNotifyDisplayCapabilities:
		return rt.AlertHandler.NotifyDisplayCapabilities(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.BART && inFrame.SubGroup == wire.BARTDownloadQuery:
		return rt.BARTHandler.DownloadQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.BART && inFrame.SubGroup == wire.BARTUploadQuery:
		return rt.BARTHandler.UploadQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Buddy && inFrame.SubGroup == wire.BuddyAddBuddies:
		return rt.BuddyHandler.AddBuddies(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Buddy && inFrame.SubGroup == wire.BuddyDelBuddies:
		return rt.BuddyHandler.DelBuddies(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Buddy && inFrame.SubGroup == wire.BuddyRightsQuery:
		return rt.BuddyHandler.RightsQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Chat && inFrame.SubGroup == wire.ChatChannelMsgToHost:
		return rt.ChatHandler.ChannelMsgToHost(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ChatNav && inFrame.SubGroup == wire.ChatNavCreateRoom:
		return rt.ChatNavHandler.CreateRoom(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ChatNav && inFrame.SubGroup == wire.ChatNavRequestChatRights:
		return rt.ChatNavHandler.RequestChatRights(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ChatNav && inFrame.SubGroup == wire.ChatNavRequestExchangeInfo:
		return rt.ChatNavHandler.RequestExchangeInfo(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ChatNav && inFrame.SubGroup == wire.ChatNavRequestRoomInfo:
		return rt.ChatNavHandler.RequestRoomInfo(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagDeleteItem:
		return rt.FeedbagHandler.DeleteItem(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagEndCluster:
		return rt.FeedbagHandler.EndCluster(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagInsertItem:
		return rt.FeedbagHandler.InsertItem(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagQuery:
		return rt.FeedbagHandler.Query(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagQueryIfModified:
		return rt.FeedbagHandler.QueryIfModified(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagRespondAuthorizeToHost:
		return rt.FeedbagHandler.RespondAuthorizeToHost(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagRightsQuery:
		return rt.FeedbagHandler.RightsQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagStartCluster:
		return rt.FeedbagHandler.StartCluster(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagUpdateItem:
		return rt.FeedbagHandler.UpdateItem(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Feedbag && inFrame.SubGroup == wire.FeedbagUse:
		return rt.FeedbagHandler.Use(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ICQ && inFrame.SubGroup == wire.ICQDBQuery:
		return rt.ICQHandler.DBQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ICBM && inFrame.SubGroup == wire.ICBMAddParameters:
		return rt.ICBMHandler.AddParameters(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ICBM && inFrame.SubGroup == wire.ICBMChannelMsgToHost:
		return rt.ICBMHandler.ChannelMsgToHost(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ICBM && inFrame.SubGroup == wire.ICBMClientErr:
		return rt.ICBMHandler.ClientErr(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ICBM && inFrame.SubGroup == wire.ICBMClientEvent:
		return rt.ICBMHandler.ClientEvent(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ICBM && inFrame.SubGroup == wire.ICBMEvilRequest:
		return rt.ICBMHandler.EvilRequest(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ICBM && inFrame.SubGroup == wire.ICBMParameterQuery:
		return rt.ICBMHandler.ParameterQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Locate && inFrame.SubGroup == wire.LocateGetDirInfo:
		return rt.LocateHandler.GetDirInfo(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Locate && inFrame.SubGroup == wire.LocateRightsQuery:
		return rt.LocateHandler.RightsQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Locate && inFrame.SubGroup == wire.LocateSetDirInfo:
		return rt.LocateHandler.SetDirInfo(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Locate && inFrame.SubGroup == wire.LocateSetInfo:
		return rt.LocateHandler.SetInfo(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Locate && inFrame.SubGroup == wire.LocateSetKeywordInfo:
		return rt.LocateHandler.SetKeywordInfo(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Locate && inFrame.SubGroup == wire.LocateUserInfoQuery:
		return rt.LocateHandler.UserInfoQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Locate && inFrame.SubGroup == wire.LocateUserInfoQuery2:
		return rt.LocateHandler.UserInfoQuery2(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ODir && inFrame.SubGroup == wire.ODirInfoQuery:
		return rt.ODirHandler.InfoQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.ODir && inFrame.SubGroup == wire.ODirKeywordListQuery:
		return rt.ODirHandler.KeywordListQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceClientOnline:
		return rt.OServiceHandler.ClientOnline(ctx, server, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceClientVersions:
		return rt.OServiceHandler.ClientVersions(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceIdleNotification:
		return rt.OServiceHandler.IdleNotification(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceNoop:
		return rt.OServiceHandler.Noop(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceRateParamsQuery:
		return rt.OServiceHandler.RateParamsQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceRateParamsSubAdd:
		return rt.OServiceHandler.RateParamsSubAdd(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceServiceRequest:
		return rt.OServiceHandler.ServiceRequest(ctx, server, sess, inFrame, r, rw, connectHere)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceSetPrivacyFlags:
		return rt.OServiceHandler.SetPrivacyFlags(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceSetUserInfoFields:
		return rt.OServiceHandler.SetUserInfoFields(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.OService && inFrame.SubGroup == wire.OServiceUserInfoQuery:
		return rt.OServiceHandler.UserInfoQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.PermitDeny && inFrame.SubGroup == wire.PermitDenyAddDenyListEntries:
		return rt.PermitDenyHandler.AddDenyListEntries(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.PermitDeny && inFrame.SubGroup == wire.PermitDenyAddPermListEntries:
		return rt.PermitDenyHandler.AddPermListEntries(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.PermitDeny && inFrame.SubGroup == wire.PermitDenyDelDenyListEntries:
		return rt.PermitDenyHandler.DelDenyListEntries(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.PermitDeny && inFrame.SubGroup == wire.PermitDenyDelPermListEntries:
		return rt.PermitDenyHandler.DelPermListEntries(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.PermitDeny && inFrame.SubGroup == wire.PermitDenyRightsQuery:
		return rt.PermitDenyHandler.RightsQuery(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.PermitDeny && inFrame.SubGroup == wire.PermitDenySetGroupPermitMask:
		return rt.PermitDenyHandler.SetGroupPermitMask(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.Stats && inFrame.SubGroup == wire.StatsReportEvents:
		return rt.StatsHandler.ReportEvents(ctx, sess, inFrame, r, rw)
	case inFrame.FoodGroup == wire.UserLookup && inFrame.SubGroup == wire.UserLookupFindByEmail:
		return rt.UserLookupHandler.FindByEmail(ctx, sess, inFrame, r, rw)

	}

	return ErrRouteNotFound
}
