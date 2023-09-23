package oscar

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const (
	BUCPErr                      uint16 = 0x0001
	BUCPLoginRequest                    = 0x0002
	BUCPRegisterRequest                 = 0x0004
	BUCPChallengeRequest                = 0x0006
	BUCPAsasnRequest                    = 0x0008
	BUCPSecuridRequest                  = 0x000A
	BUCPRegistrationImageRequest        = 0x000C
)

func routeBUCP(snac snacFrame) error {
	switch snac.subGroup {
	case BUCPErr:
		panic("not implemented")
	case BUCPLoginRequest:
		panic("not implemented")
	case BUCPRegisterRequest:
		panic("not implemented")
	case BUCPChallengeRequest:
		panic("not implemented")
	case BUCPAsasnRequest:
		panic("not implemented")
	case BUCPSecuridRequest:
		panic("not implemented")
	case BUCPRegistrationImageRequest:
		panic("not implemented")
	}

	return nil
}

func ReceiveAndSendAuthChallenge(s *Session, r io.Reader, w io.Writer, sequence *uint32) error {
	flap := flapFrame{}
	if err := flap.read(r); err != nil {
		return err
	}

	b := make([]byte, flap.payloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	snac := snacFrame{}
	if err := snac.read(buf); err != nil {
		return err
	}

	snacPayloadIn := SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := Unmarshal(&snacPayloadIn, buf); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := snacFrame{
		foodGroup: 0x17,
		subGroup:  0x07,
	}
	snacPayloadOut := SNAC_0x17_0x07_BUCPChallengeResponse{
		AuthKey: s.ID,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendBUCPLoginRequest(cfg Config, sess *Session, fm *FeedbagStore, r io.Reader, w io.Writer, sequence *uint32) error {
	flap := flapFrame{}
	if err := flap.read(r); err != nil {
		return err
	}

	b := make([]byte, flap.payloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	snac := snacFrame{}
	if err := snac.read(buf); err != nil {
		return err
	}

	snacPayloadIn := SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := Unmarshal(&snacPayloadIn, buf); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendBUCPLoginRequest read SNAC: %+v\n", snacPayloadIn)

	var found bool
	sess.ScreenName, found = snacPayloadIn.getString(TLV_SCREEN_NAME)
	if !found {
		return errors.New("unable to find screen name")
	}

	if err := fm.UpsertUser(sess.ScreenName); err != nil {
		return err
	}

	snacFrameOut := snacFrame{
		foodGroup: 0x17,
		subGroup:  0x03,
	}

	snacPayloadOut := SNAC_0x17_0x02_BUCPLoginRequest{
		TLVRestBlock: TLVRestBlock{
			TLVList: TLVList{
				{
					tType: TLV_SCREEN_NAME,
					val:   sess.ScreenName,
				},
				{
					tType: 0x08,
					val:   uint16(0x00),
				},
				{
					tType: 0x04,
					val:   "",
				},
				{
					tType: 0x05,
					val:   Address(cfg.OSCARHost, cfg.BOSPort),
				},
				{
					tType: 0x06,
					val:   []byte(sess.ID),
				},
				{
					tType: 0x11,
					val:   "mike@localhost",
				},
				{
					tType: 0x54,
					val:   "http://localhost",
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
