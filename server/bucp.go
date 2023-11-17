package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

func routeBUCP(context.Context) error {
	return ErrUnsupportedSubGroup
}

func ReceiveAndSendAuthChallenge(cfg Config, fm *SQLiteFeedbagStore, r io.Reader, w io.Writer, sequence *uint32, newUUID func() uuid.UUID) error {
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
			FoodGroup: oscar.BUCP,
			SubGroup:  BUCPLoginResponse,
		}
		snacPayloadOut := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
		snacPayloadOut.AddTLV(oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)))
		return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: oscar.BUCP,
		SubGroup:  BUCPChallengeResponse,
	}
	snacPayloadOut := oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
		AuthKey: authKey,
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendBUCPLoginRequest(cfg Config, sm SessionManager, fm *SQLiteFeedbagStore, r io.Reader, w io.Writer, sequence *uint32, newUUID func() uuid.UUID) error {
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
	snacPayloadOut.AddTLV(oscar.NewTLV(oscar.TLVScreenName, screenName))

	if loginOK {
		sess := sm.NewSessionWithSN(newUUID().String(), screenName)
		snacPayloadOut.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.TLVReconnectHere, Address(cfg.OSCARHost, cfg.BOSPort)),
			oscar.NewTLV(oscar.TLVAuthorizationCookie, sess.ID()),
		})
	} else {
		snacPayloadOut.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
		})
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: oscar.BUCP,
		SubGroup:  BUCPLoginResponse,
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func SendAndReceiveSignonFrame(rw io.ReadWriter, sequence *uint32) (oscar.FlapSignonFrame, error) {
	flapFrameOut := oscar.FlapFrame{
		StartMarker:   42,
		FrameType:     oscar.FlapFrameSignon,
		Sequence:      uint16(*sequence),
		PayloadLength: 4, // size of FlapSignonFrame
	}
	if err := oscar.Marshal(flapFrameOut, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	flapSignonFrameOut := oscar.FlapSignonFrame{
		FlapVersion: 1,
	}
	if err := oscar.Marshal(flapSignonFrameOut, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}

	// receive
	flapFrameIn := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flapFrameIn, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	b := make([]byte, flapFrameIn.PayloadLength)
	if _, err := rw.Read(b); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	flapSignonFrameIn := oscar.FlapSignonFrame{}
	if err := oscar.Unmarshal(&flapSignonFrameIn, bytes.NewBuffer(b)); err != nil {
		return oscar.FlapSignonFrame{}, err
	}

	*sequence++

	return flapSignonFrameIn, nil
}

func VerifyLogin(sm SessionManager, rw io.ReadWriter) (*Session, uint32, error) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(rw, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	ID, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session id from payload")
	}

	sess, ok := sm.Retrieve(string(ID))
	if !ok {
		return nil, 0, fmt.Errorf("unable to find session by id %s", ID)
	}

	return sess, seq, nil
}

func VerifyChatLogin(rw io.ReadWriter) (*ChatCookie, uint32, error) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(rw, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	buf, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session id from payload")
	}

	cookie := ChatCookie{}
	err = oscar.Unmarshal(&cookie, bytes.NewBuffer(buf))

	return &cookie, seq, err
}
