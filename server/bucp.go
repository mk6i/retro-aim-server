package server

import (
	"bytes"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
)

const (
	BUCPErr                      uint16 = 0x0001
	BUCPLoginRequest                    = 0x0002
	BUCPLoginResponse                   = 0x0003
	BUCPRegisterRequest                 = 0x0004
	BUCPChallengeRequest                = 0x0006
	BUCPChallengeResponse               = 0x0007
	BUCPAsasnRequest                    = 0x0008
	BUCPSecuridRequest                  = 0x000A
	BUCPRegistrationImageRequest        = 0x000C
)

func routeBUCP() error {
	return ErrUnimplementedSNAC
}

func ReceiveAndSendAuthChallenge(cfg Config, fm *FeedbagStore, r io.Reader, w io.Writer, sequence *uint32, newUUID func() uuid.UUID) error {
	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, r); err != nil {
		return err
	}
	b := make([]byte, flap.PayloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}
	snac := oscar.SnacFrame{}
	buf := bytes.NewBuffer(b)
	if err := oscar.Unmarshal(&snac, buf); err != nil {
		return err
	}

	snacPayloadIn := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn, buf); err != nil {
		return err
	}
	screenName, exists := snacPayloadIn.GetString(oscar.TLVScreenName)
	if !exists {
		return errors.New("screen name doesn't exist in tlv")
	}

	var authKey string

	u, err := fm.GetUser(screenName)
	switch {
	case err != nil:
		return err
	case u != nil:
		// user lookup succeeded
		authKey = u.AuthKey
	case cfg.DisableAuth:
		// can't find user, generate stub auth key
		authKey = newUUID().String()
	default:
		// can't find user, return login error
		snacFrameOut := oscar.SnacFrame{
			FoodGroup: BUCP,
			SubGroup:  BUCPLoginResponse,
		}
		snacPayloadOut := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
		snacPayloadOut.AddTLV(oscar.TLV{
			TType: oscar.TLVErrorSubcode,
			Val:   uint16(0x01),
		})
		return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPChallengeResponse,
	}
	snacPayloadOut := oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
		AuthKey: authKey,
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendBUCPLoginRequest(cfg Config, sm SessionManager, fm *FeedbagStore, r io.Reader, w io.Writer, sequence *uint32, newUUID func() uuid.UUID) error {
	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, r); err != nil {
		return err
	}
	snac := oscar.SnacFrame{}
	b := make([]byte, flap.PayloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}
	buf := bytes.NewBuffer(b)
	if err := oscar.Unmarshal(&snac, buf); err != nil {
		return err
	}

	snacPayloadIn := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn, buf); err != nil {
		return err
	}
	screenName, found := snacPayloadIn.GetString(oscar.TLVScreenName)
	if !found {
		return errors.New("screen name doesn't exist in tlv")
	}
	md5Hash, found := snacPayloadIn.GetSlice(oscar.TLVPasswordHash)
	if !found {
		return errors.New("password hash doesn't exist in tlv")
	}

	loginOK := false

	u, err := fm.GetUser(screenName)
	switch {
	case err != nil:
		return err
	case u != nil && bytes.Equal(u.PassHash, md5Hash):
		// password check succeeded
		loginOK = true
	case cfg.DisableAuth:
		// login failed but let them in anyway
		newUser, err := NewStubUser(screenName)
		if err != nil {
			return err
		}
		if err := fm.UpsertUser(newUser); err != nil {
			return err
		}
		loginOK = true
	}

	snacPayloadOut := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
	snacPayloadOut.AddTLV(oscar.TLV{
		TType: oscar.TLVScreenName,
		Val:   screenName,
	})

	if loginOK {
		sess := sm.NewSessionWithSN(newUUID().String(), screenName)
		snacPayloadOut.AddTLVList([]oscar.TLV{
			{
				TType: oscar.TLVReconnectHere,
				Val:   Address(cfg.OSCARHost, cfg.BOSPort),
			},
			{
				TType: oscar.TLVAuthorizationCookie,
				Val:   sess.ID,
			},
		})
	} else {
		snacPayloadOut.AddTLVList([]oscar.TLV{
			{
				TType: oscar.TLVErrorSubcode,
				Val:   uint16(0x01),
			},
		})
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPLoginResponse,
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
