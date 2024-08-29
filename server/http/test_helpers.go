package http

import (
	"context"
	"net/mail"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type mockParams struct {
	accountRetrieverParams
	bartRetrieverParams
	chatRoomCreatorParams
	chatRoomRetrieverParams
	chatSessionRetrieverParams
	feedBagRetrieverParams
	messageRelayerParams
	profileRetrieverParams
	sessionRetrieverParams
	userManagerParams
}

// accountRetrieverParams is a helper struct that contains mock parameters for
// accountRetriever methods
type accountRetrieverParams struct {
	emailAddressByNameParams
	regStatusByNameParams
	confirmStatusByNameParams
}

// emailAddressByNameParams is the list of parameters passed at the mock
// accountRetriever.EmailAddressByName call site
type emailAddressByNameParams []struct {
	screenName state.IdentScreenName
	result     *mail.Address
	err        error
}

// regStatusByNameParams is the list of parameters passed at the mock
// accountRetriever.RegStatusByName call site
type regStatusByNameParams []struct {
	screenName state.IdentScreenName
	result     uint16
	err        error
}

// confirmStatusByNameParams is the list of parameters passed at the mock
// accountRetriever.ConfirmStatusByName call site
type confirmStatusByNameParams []struct {
	screenName state.IdentScreenName
	result     bool
	err        error
}

// bartRetrieverParams is a helper struct that contains mock parameters for
// BARTRetriever methods
type bartRetrieverParams struct {
	bartRetrieveParams
}

// bartRetrieveParams is the list of parameters passed at the mock
// BARTRetriever.BARTRetrieveParams call site
type bartRetrieveParams []struct {
	itemHash []byte
	result   []byte
	err      error
}

// chatRoomCreatorParams is a helper struct that contains mock parameters for
// ChatRoomCreator methods
type chatRoomCreatorParams struct {
	createChatRoomParams
}

// createChatRoomParams is the list of parameters passed at the mock
// ChatRoomCreator.CreateChatRoom call site
type createChatRoomParams []struct {
	chatRoom *state.ChatRoom
	err      error
}

// chatRoomRetrieverParams is a helper struct that contains mock parameters for
// ChatRoomRetriever methods
type chatRoomRetrieverParams struct {
	allChatRoomsParams
}

// allChatRoomsParams is the list of parameters passed at the mock
// ChatRoomRetriever.AllChatRooms call site
type allChatRoomsParams []struct {
	exchange uint16
	result   []state.ChatRoom
	err      error
}

// chatRoomRetrieverParams is a helper struct that contains mock parameters for
// ChatRoomRetriever methods
type chatSessionRetrieverParams struct {
	chatSessionRetrieverAllSessionsParams
}

// chatSessionRetrieverAllSessionsParams is the list of parameters passed at the mock
// ChatSessionRetriever.AllSessions call site
type chatSessionRetrieverAllSessionsParams []struct {
	cookie string
	result []*state.Session
}

// feedBagRetrieverParams is a helper struct that contains mock parameters for
// FeedBagRetriever methods
type feedBagRetrieverParams struct {
	buddyIconRefByNameParams
}

// buddyIconRefByNameParams is the list of parameters passed at the mock
// FeedBagRetriever.BuddyIconRefByNameParams call site
type buddyIconRefByNameParams []struct {
	screenName state.IdentScreenName
	result     *wire.BARTID
	err        error
}

// messageRelayerParams is a helper struct that contains mock parameters for
// MessageRelayer methods
type messageRelayerParams struct {
	relayToScreenNameParams
}

// relayToScreenNameParams is the list of parameters passed at the mock
// MessageRelayer.RelayToScreenNameParams call site
type relayToScreenNameParams []struct {
	ctx        context.Context
	screenName state.IdentScreenName
	msg        wire.SNACMessage
}

// profileRetrieverParams is a helper struct that contains mock parameters for
// ProfileRetriever methods
type profileRetrieverParams struct {
	retrieveProfileParams
}

// retrieveProfileParams is the list of parameters passed at the mock
// ProfileRetriever.Profile call site
type retrieveProfileParams []struct {
	screenName state.IdentScreenName
	result     string
	err        error
}

// sessionRetrieverParams is a helper struct that contains mock parameters for
// SessionRetriever methods
type sessionRetrieverParams struct {
	sessionRetrieverAllSessionsParams
	retrieveSessionByNameParams
}

// sessionRetrieverAllSessionsParams is the list of parameters passed at the mock
// SessionRetriever.AllSessions call site
type sessionRetrieverAllSessionsParams []struct {
	result []*state.Session
}

// retrieveSessionParams is the list of parameters passed at the mock
// SessionRetriever.RetrieveSessionByName call site
type retrieveSessionByNameParams []struct {
	screenName state.IdentScreenName
	result     *state.Session
}

// userManagerParams is a helper struct that contains mock parameters for
// UserManager methods
type userManagerParams struct {
	allUsersParams
	deleteUserParams
	getUserParams
	insertUserParams
	setUserPasswordParams
}

// allUsersParams is the list of parameters passed at the mock
// UserManager.AllUsers call site
type allUsersParams []struct {
	result []state.User
	err    error
}

// deleteUserParams is the list of parameters passed at the mock
// UserManager.DeleteUser call site
type deleteUserParams []struct {
	screenName state.IdentScreenName
	err        error
}

// getUserParams is the list of parameters passed at the mock
// UserManager.User call site
type getUserParams []struct {
	screenName state.IdentScreenName
	result     *state.User
	err        error
}

// insertUserParams is the list of parameters passed at the mock
// UserManager.InsertUser call site
type insertUserParams []struct {
	u   state.User
	err error
}

// setUserPasswordParams is the list of parameters passed at the mock
// UserManager.SetUserPassword call site
type setUserPasswordParams []struct {
	screenName  state.IdentScreenName
	newPassword string
	err         error
}
