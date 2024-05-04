package handler

import (
	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/wire"
)

// Handlers aggregates various specialized handlers required for processing
// different types of OSCAR protocol requests. Each field corresponds to a
// specific handler responsible for a distinct aspect of the OSCAR service,
// such as managing buddy lists, chat sessions, and user alerts.
type Handlers struct {
	AlertHandler
	BARTHandler
	BuddyHandler
	ChatHandler
	ChatNavHandler
	FeedbagHandler
	ICBMHandler
	LocateHandler
	OServiceAlertHandler
	OServiceBOSHandler
	OServiceChatHandler
	OServiceChatNavHandler
}

// NewBOSRouter initializes and configures a new Router instance for handling
// OSCAR protocol requests in the context of a Basic Oscar Service (BOS).
func NewBOSRouter(h Handlers) oscar.Router {
	router := oscar.NewRouter()

	router.Register(wire.Alert, wire.AlertNotifyCapabilities, h.AlertHandler.NotifyCapabilities)
	router.Register(wire.Alert, wire.AlertNotifyDisplayCapabilities, h.AlertHandler.NotifyDisplayCapabilities)

	router.Register(wire.BART, wire.BARTUploadQuery, h.BARTHandler.UploadQuery)
	router.Register(wire.BART, wire.BARTDownloadQuery, h.BARTHandler.DownloadQuery)

	router.Register(wire.Buddy, wire.BuddyAddBuddies, h.BuddyHandler.AddBuddies)
	router.Register(wire.Buddy, wire.BuddyDelBuddies, h.BuddyHandler.DelBuddies)
	router.Register(wire.Buddy, wire.BuddyRightsQuery, h.BuddyHandler.RightsQuery)

	router.Register(wire.ChatNav, wire.ChatNavCreateRoom, h.ChatNavHandler.CreateRoom)
	router.Register(wire.ChatNav, wire.ChatNavRequestChatRights, h.ChatNavHandler.RequestChatRights)
	router.Register(wire.ChatNav, wire.ChatNavRequestRoomInfo, h.ChatNavHandler.RequestRoomInfo)

	router.Register(wire.Feedbag, wire.FeedbagDeleteItem, h.FeedbagHandler.DeleteItem)
	router.Register(wire.Feedbag, wire.FeedbagEndCluster, h.FeedbagHandler.EndCluster)
	router.Register(wire.Feedbag, wire.FeedbagInsertItem, h.FeedbagHandler.InsertItem)
	router.Register(wire.Feedbag, wire.FeedbagQuery, h.FeedbagHandler.Query)
	router.Register(wire.Feedbag, wire.FeedbagQueryIfModified, h.FeedbagHandler.QueryIfModified)
	router.Register(wire.Feedbag, wire.FeedbagRightsQuery, h.FeedbagHandler.RightsQuery)
	router.Register(wire.Feedbag, wire.FeedbagStartCluster, h.FeedbagHandler.StartCluster)
	router.Register(wire.Feedbag, wire.FeedbagUpdateItem, h.FeedbagHandler.UpdateItem)
	router.Register(wire.Feedbag, wire.FeedbagUse, h.FeedbagHandler.Use)

	router.Register(wire.ICBM, wire.ICBMAddParameters, h.ICBMHandler.AddParameters)
	router.Register(wire.ICBM, wire.ICBMChannelMsgToHost, h.ICBMHandler.ChannelMsgToHost)
	router.Register(wire.ICBM, wire.ICBMClientErr, h.ICBMHandler.ClientErr)
	router.Register(wire.ICBM, wire.ICBMClientEvent, h.ICBMHandler.ClientEvent)
	router.Register(wire.ICBM, wire.ICBMEvilRequest, h.ICBMHandler.EvilRequest)
	router.Register(wire.ICBM, wire.ICBMParameterQuery, h.ICBMHandler.ParameterQuery)

	router.Register(wire.Locate, wire.LocateGetDirInfo, h.LocateHandler.GetDirInfo)
	router.Register(wire.Locate, wire.LocateRightsQuery, h.LocateHandler.RightsQuery)
	router.Register(wire.Locate, wire.LocateSetDirInfo, h.LocateHandler.SetDirInfo)
	router.Register(wire.Locate, wire.LocateSetInfo, h.LocateHandler.SetInfo)
	router.Register(wire.Locate, wire.LocateSetKeywordInfo, h.LocateHandler.SetKeywordInfo)
	router.Register(wire.Locate, wire.LocateUserInfoQuery, h.LocateHandler.UserInfoQuery)
	router.Register(wire.Locate, wire.LocateUserInfoQuery2, h.LocateHandler.UserInfoQuery2)

	router.Register(wire.OService, wire.OServiceClientOnline, h.OServiceBOSHandler.ClientOnline)
	router.Register(wire.OService, wire.OServiceClientVersions, h.OServiceBOSHandler.OServiceHandler.ClientVersions)
	router.Register(wire.OService, wire.OServiceIdleNotification, h.OServiceBOSHandler.OServiceHandler.IdleNotification)
	router.Register(wire.OService, wire.OServiceNoop, h.OServiceBOSHandler.OServiceHandler.Noop)
	router.Register(wire.OService, wire.OServiceRateParamsQuery, h.OServiceBOSHandler.OServiceHandler.RateParamsQuery)
	router.Register(wire.OService, wire.OServiceRateParamsSubAdd, h.OServiceBOSHandler.OServiceHandler.RateParamsSubAdd)
	router.Register(wire.OService, wire.OServiceServiceRequest, h.OServiceBOSHandler.ServiceRequest)
	router.Register(wire.OService, wire.OServiceSetUserInfoFields, h.OServiceBOSHandler.OServiceHandler.SetUserInfoFields)
	router.Register(wire.OService, wire.OServiceUserInfoQuery, h.OServiceBOSHandler.OServiceHandler.UserInfoQuery)

	return router
}

// NewChatRouter initializes and configures a new Router instance specifically
// for handling chat-related OSCAR protocol requests.
func NewChatRouter(h Handlers) oscar.Router {
	router := oscar.NewRouter()

	router.Register(wire.Chat, wire.ChatChannelMsgToHost, h.ChatHandler.ChannelMsgToHost)

	router.Register(wire.OService, wire.OServiceClientOnline, h.OServiceChatHandler.ClientOnline)
	router.Register(wire.OService, wire.OServiceClientVersions, h.OServiceChatHandler.OServiceHandler.ClientVersions)
	router.Register(wire.OService, wire.OServiceIdleNotification, h.OServiceChatHandler.OServiceHandler.IdleNotification)
	router.Register(wire.OService, wire.OServiceRateParamsQuery, h.OServiceChatHandler.OServiceHandler.RateParamsQuery)
	router.Register(wire.OService, wire.OServiceRateParamsSubAdd, h.OServiceChatHandler.OServiceHandler.RateParamsSubAdd)
	router.Register(wire.OService, wire.OServiceSetUserInfoFields, h.OServiceChatHandler.OServiceHandler.SetUserInfoFields)
	router.Register(wire.OService, wire.OServiceUserInfoQuery, h.OServiceChatHandler.OServiceHandler.UserInfoQuery)

	return router
}

// NewChatNavRouter initializes and configures a new Router instance for
// handling OSCAR protocol requests in the context of the ChatNav service.
func NewChatNavRouter(h Handlers) oscar.Router {
	router := oscar.NewRouter()

	router.Register(wire.ChatNav, wire.ChatNavCreateRoom, h.ChatNavHandler.CreateRoom)
	router.Register(wire.ChatNav, wire.ChatNavRequestChatRights, h.ChatNavHandler.RequestChatRights)
	router.Register(wire.ChatNav, wire.ChatNavRequestRoomInfo, h.ChatNavHandler.RequestRoomInfo)

	router.Register(wire.OService, wire.OServiceClientOnline, h.OServiceChatNavHandler.ClientOnline)
	router.Register(wire.OService, wire.OServiceClientVersions, h.OServiceChatNavHandler.OServiceHandler.ClientVersions)
	router.Register(wire.OService, wire.OServiceIdleNotification, h.OServiceChatNavHandler.OServiceHandler.IdleNotification)
	router.Register(wire.OService, wire.OServiceRateParamsQuery, h.OServiceChatNavHandler.OServiceHandler.RateParamsQuery)
	router.Register(wire.OService, wire.OServiceRateParamsSubAdd, h.OServiceChatNavHandler.OServiceHandler.RateParamsSubAdd)
	router.Register(wire.OService, wire.OServiceSetUserInfoFields, h.OServiceChatNavHandler.OServiceHandler.SetUserInfoFields)
	router.Register(wire.OService, wire.OServiceUserInfoQuery, h.OServiceChatNavHandler.OServiceHandler.UserInfoQuery)

	return router
}

func NewAlertRouter(h Handlers) oscar.Router {
	router := oscar.NewRouter()

	router.Register(wire.Alert, wire.AlertNotifyCapabilities, h.AlertHandler.NotifyCapabilities)
	router.Register(wire.Alert, wire.AlertNotifyDisplayCapabilities, h.AlertHandler.NotifyDisplayCapabilities)
	router.Register(wire.Alert, wire.AlertUserOnline, h.AlertHandler.NotifyDisplayCapabilities)

	router.Register(wire.OService, wire.OServiceClientOnline, h.OServiceAlertHandler.ClientOnline)
	router.Register(wire.OService, wire.OServiceClientVersions, h.OServiceAlertHandler.OServiceHandler.ClientVersions)
	router.Register(wire.OService, wire.OServiceRateParamsQuery, h.OServiceAlertHandler.OServiceHandler.RateParamsQuery)
	router.Register(wire.OService, wire.OServiceRateParamsSubAdd, h.OServiceAlertHandler.OServiceHandler.RateParamsSubAdd)

	return router
}
