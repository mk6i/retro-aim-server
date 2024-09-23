package foodgroup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewODirService creates a new instance of ODirService.
func NewODirService(logger *slog.Logger, profileManager ProfileManager) ODirService {
	return ODirService{
		logger:         logger,
		profileManager: profileManager,
	}
}

// ODirService provides functionality for the ODir food group, which
// provides functionality for searching the user directory.
type ODirService struct {
	logger         *slog.Logger
	profileManager ProfileManager
}

// InfoQuery searches the user directory based on the query type: name/address,
// email, or interest. It dispatches the request to the appropriate search
// method and returns the search results or an error. The search type is
// determined by the presence of certain TLVs:
//
//   - wire.ODirTLVEmailAddress: Search by email.
//   - wire.ODirTLVInterest: Search by interest keyword.
//   - wire.ODirTLVFirstName or wire.ODirTLVLastName: Search by name and address.
//     First name or last name must be required to search by name and address.
//
// AIM 5.x sends wire.ODirTLVSearchType to specify the search type. This TLV is
// ignored in order to be backwards compatible with older versions that do not
// send it. It doesn't appear to make a difference, since AIM 5.x sends the
// same TLV types for each search type.
func (s ODirService) InfoQuery(_ context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0F_0x02_InfoQuery) (wire.SNACMessage, error) {
	snac := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ODir,
			SubGroup:  wire.ODirInfoReply,
			RequestID: inFrame.RequestID,
		},
	}

	switch {
	case inBody.HasTag(wire.ODirTLVEmailAddress):
		var err error
		snac.Body, err = s.searchByEmail(inBody)
		if err != nil {
			return wire.SNACMessage{}, err
		}
	case inBody.HasTag(wire.ODirTLVInterest):
		var err error
		snac.Body, err = s.searchByInterest(inBody)
		if err != nil {
			return wire.SNACMessage{}, err
		}
	case inBody.HasTag(wire.ODirTLVFirstName), inBody.HasTag(wire.ODirTLVLastName):
		var err error
		snac.Body, err = s.searchByNameAndAddr(inBody)
		if err != nil {
			return wire.SNACMessage{}, err
		}
	default:
		snac.Body = wire.SNAC_0x0F_0x03_InfoReply{
			Status: wire.ODirSearchResponseNameMissing,
		}
	}

	return snac, nil
}

// searchByNameAndAddr performs a directory search using the user's first or
// last name and address.
func (s ODirService) searchByNameAndAddr(inBody wire.SNAC_0x0F_0x02_InfoQuery) (wire.SNAC_0x0F_0x03_InfoReply, error) {
	foundUsers, err := s.profileManager.FindByAIMNameAndAddr(newAIMNameAndAddrFromTLVList(inBody.TLVList))
	if err != nil {
		return wire.SNAC_0x0F_0x03_InfoReply{}, fmt.Errorf("FindByAIMNameAndAddr: %w", err)
	}
	return s.searchResponse(foundUsers)
}

// searchByInterest performs a directory search using the user's specified
// interest keyword.
func (s ODirService) searchByInterest(inBody wire.SNAC_0x0F_0x02_InfoQuery) (wire.SNAC_0x0F_0x03_InfoReply, error) {
	var foundUsers []state.User

	interest, _ := inBody.String(wire.ODirTLVInterest)

	foundUsers, err := s.profileManager.FindByAIMKeyword(interest)
	if err != nil {
		return wire.SNAC_0x0F_0x03_InfoReply{}, fmt.Errorf("FindByAIMKeyword: %w", err)
	}

	return s.searchResponse(foundUsers)
}

// searchByEmail performs a directory search using the user's email address.
func (s ODirService) searchByEmail(inBody wire.SNAC_0x0F_0x02_InfoQuery) (wire.SNAC_0x0F_0x03_InfoReply, error) {
	email, _ := inBody.String(wire.ODirTLVEmailAddress)

	result, err := s.profileManager.FindByAIMEmail(email)
	if err != nil {
		if errors.Is(err, state.ErrNoUser) {
			return s.searchResponse(nil)
		}
		return wire.SNAC_0x0F_0x03_InfoReply{}, fmt.Errorf("FindByAIMEmail: %w", err)
	}

	return s.searchResponse([]state.User{result})
}

// searchResponse constructs the SNAC reply based on the users found during the
// search.
func (s ODirService) searchResponse(foundUsers []state.User) (wire.SNAC_0x0F_0x03_InfoReply, error) {
	body := wire.SNAC_0x0F_0x03_InfoReply{
		Status: wire.ODirSearchResponseOK,
	}

	for _, res := range foundUsers {
		body.Results.List = append(body.Results.List, wire.TLVBlock{
			TLVList: wire.TLVList{
				wire.NewTLVBE(wire.ODirTLVFirstName, res.AIMDirectoryInfo.FirstName),
				wire.NewTLVBE(wire.ODirTLVLastName, res.AIMDirectoryInfo.LastName),
				wire.NewTLVBE(wire.ODirTLVState, res.AIMDirectoryInfo.State),
				wire.NewTLVBE(wire.ODirTLVCity, res.AIMDirectoryInfo.City),
				wire.NewTLVBE(wire.ODirTLVCountry, res.AIMDirectoryInfo.Country),
				wire.NewTLVBE(wire.ODirTLVScreenName, res.DisplayScreenName.String()),
			},
		})
	}

	return body, nil
}

// KeywordListQuery returns a list of keywords that can be searched in the user
// directory.
func (s ODirService) KeywordListQuery(_ context.Context, inFrame wire.SNACFrame) (wire.SNACMessage, error) {
	interests, err := s.profileManager.InterestList()
	if err != nil {
		return wire.SNACMessage{}, fmt.Errorf("InterestList: %w", err)
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ODir,
			SubGroup:  wire.ODirKeywordListReply,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x0F_0x04_KeywordListReply{
			Status:    0x01,
			Interests: interests,
		},
	}, nil
}

// newAIMNameAndAddrFromTLVList constructs an AIMNameAndAddr structure from the
// TLV list containing user directory fields like first name, last name, etc.
func newAIMNameAndAddrFromTLVList(tlvList wire.TLVList) state.AIMNameAndAddr {
	a := state.AIMNameAndAddr{}

	if firstName, hasFirstName := tlvList.String(wire.ODirTLVFirstName); hasFirstName {
		a.FirstName = firstName
	}
	if lastName, hasLastName := tlvList.String(wire.ODirTLVLastName); hasLastName {
		a.LastName = lastName
	}
	if middleName, hasMiddleName := tlvList.String(wire.ODirTLVMiddleName); hasMiddleName {
		a.MiddleName = middleName
	}
	if maidenName, hasMaidenName := tlvList.String(wire.ODirTLVMaidenName); hasMaidenName {
		a.MaidenName = maidenName
	}
	if country, hasCountry := tlvList.String(wire.ODirTLVCountry); hasCountry {
		a.Country = country
	}
	if st, hasState := tlvList.String(wire.ODirTLVState); hasState {
		a.State = st
	}
	if city, hasCity := tlvList.String(wire.ODirTLVCity); hasCity {
		a.City = city
	}
	if nickName, hasNickName := tlvList.String(wire.ODirTLVNickName); hasNickName {
		a.NickName = nickName
	}
	if zipCode, hasZIPCode := tlvList.String(wire.ODirTLVZIP); hasZIPCode {
		a.ZIPCode = zipCode
	}
	if address, hasAddress := tlvList.String(wire.ODirTLVAddress); hasAddress {
		a.Address = address
	}

	return a
}
