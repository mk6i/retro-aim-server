package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
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

func routeBUCP(snac oscar.SnacFrame) error {
	switch snac.SubGroup {
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
	flap := oscar.FlapFrame{}
	if err := flap.Read(r); err != nil {
		return err
	}

	b := make([]byte, flap.PayloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	snac := oscar.SnacFrame{}
	if err := snac.Read(buf); err != nil {
		return err
	}

	snacPayloadIn := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn, buf); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: 0x17,
		SubGroup:  0x07,
	}
	snacPayloadOut := oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
		AuthKey: s.ID,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendBUCPLoginRequest(cfg Config, sess *Session, fm *FeedbagStore, r io.Reader, w io.Writer, sequence *uint32) error {
	flap := oscar.FlapFrame{}
	if err := flap.Read(r); err != nil {
		return err
	}

	b := make([]byte, flap.PayloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	snac := oscar.SnacFrame{}
	if err := snac.Read(buf); err != nil {
		return err
	}

	snacPayloadIn := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn, buf); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendBUCPLoginRequest read SNAC: %+v\n", snacPayloadIn)

	var found bool
	sess.ScreenName, found = snacPayloadIn.GetString(0x01)
	if !found {
		return errors.New("unable to find screen name")
	}

	if err := fm.UpsertUser(sess.ScreenName); err != nil {
		return err
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: 0x17,
		SubGroup:  0x03,
	}

	snacPayloadOut := oscar.SNAC_0x17_0x02_BUCPLoginRequest{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x01,
					Val:   sess.ScreenName,
				},
				{
					TType: 0x08,
					Val:   uint16(0x00),
				},
				{
					TType: 0x04,
					Val:   "",
				},
				{
					TType: 0x05,
					Val:   Address(cfg.OSCARHost, cfg.BOSPort),
				},
				{
					TType: 0x06,
					Val:   []byte(sess.ID),
				},
				{
					TType: 0x11,
					Val:   "mike@localhost",
				},
				{
					TType: 0x54,
					Val:   "http://localhost",
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
