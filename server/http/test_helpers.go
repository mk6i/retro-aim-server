package http

import (
	"net/mail"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type mockParams struct {
	accountManagerParams
	bartRetrieverParams
	chatRoomRetrieverParams
	chatSessionRetrieverParams
	directoryManagerParams
	feedBagRetrieverParams
	profileRetrieverParams
	sessionRetrieverParams
	userManagerParams
}

// accountManagerParams is a helper struct that contains mock parameters for
// accountManager methods
type accountManagerParams struct {
	emailAddressByNameParams
	regStatusByNameParams
	confirmStatusByNameParams
	updateSuspendedStatusParams
}

// emailAddressByNameParams is the list of parameters passed at the mock
// accountManager.EmailAddressByName call site
type emailAddressByNameParams []struct {
	screenName state.IdentScreenName
	result     *mail.Address
	err        error
}

// regStatusByNameParams is the list of parameters passed at the mock
// accountManager.RegStatusByName call site
type regStatusByNameParams []struct {
	screenName state.IdentScreenName
	result     uint16
	err        error
}

// confirmStatusByNameParams is the list of parameters passed at the mock
// accountManager.ConfirmStatusByName call site
type confirmStatusByNameParams []struct {
	screenName state.IdentScreenName
	result     bool
	err        error
}

// updateSuspendedStatus is the list of parameters passed at the mock
// accountManager.updateSuspendedStatus call site
type updateSuspendedStatusParams []struct {
	suspendedStatus uint16
	screenName      state.IdentScreenName
	err             error
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

type directoryManagerParams struct {
	categoriesParams
	createCategoryParams
	createKeywordParams
	deleteCategoryParams
	deleteKeywordParams
	keywordsByCategoryParams
}

// categoriesParams is the list of parameters passed at the mock
// DirectoryManager.Categories call site
type categoriesParams []struct {
	result []state.Category
	err    error
}

// createCategoryParams is the list of parameters passed at the mock
// DirectoryManager.CreateCategory call site
type createCategoryParams []struct {
	name   string
	result state.Category
	err    error
}

// createKeywordParams is the list of parameters passed at the mock
// DirectoryManager.CreateKeyword call site
type createKeywordParams []struct {
	name       string
	categoryID uint8
	result     state.Keyword
	err        error
}

// deleteCategoryParams is the list of parameters passed at the mock
// DirectoryManager.DeleteCategory call site
type deleteCategoryParams []struct {
	categoryID uint8
	err        error
}

// deleteKeywordParams is the list of parameters passed at the mock
// DirectoryManager.DeleteKeyword call site
type deleteKeywordParams []struct {
	id  uint8
	err error
}

// keywordsByCategoryParams is the list of parameters passed at the mock
// DirectoryManager.KeywordsByCategory call site
type keywordsByCategoryParams []struct {
	categoryID uint8
	result     []state.Keyword
	err        error
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
