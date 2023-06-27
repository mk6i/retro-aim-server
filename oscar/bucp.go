package oscar

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
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

func routeBUCP(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	switch snac.subGroup {
	case BUCPErr:
		panic("not implemented")
	case BUCPLoginRequest:
		return ReceiveAndSendBUCPLoginRequest(flap, snac, r, w, sequence)
	case BUCPRegisterRequest:
		panic("not implemented")
	case BUCPChallengeRequest:
		return ReceiveAndSendAuthChallenge(flap, snac, r, w, sequence)
	case BUCPAsasnRequest:
		panic("not implemented")
	case BUCPSecuridRequest:
		panic("not implemented")
	case BUCPRegistrationImageRequest:
		panic("not implemented")
	}

	return nil
}

type snacBUCPChallengeRequest struct {
	TLVPayload
}

func (s *snacBUCPChallengeRequest) read(r io.Reader) error {
	lookup := map[uint16]reflect.Kind{
		0x01: reflect.String,
		0x4B: reflect.String,
		0x5A: reflect.String,
	}
	return s.TLVPayload.read(r, lookup)
}

type snacBUCPChallengeResponse struct {
	authKey string
}

func (s *snacBUCPChallengeResponse) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.authKey))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(s.authKey)); err != nil {
		return err
	}
	return nil
}

func ReceiveAndSendAuthChallenge(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	snacPayload := &snacBUCPChallengeRequest{}
	if err := snac.read(r); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge read SNAC: %+v\n", snac)
	fmt.Printf("ReceiveAndSendAuthChallenge read SNAC payload: %+v\n", snacPayload)

	snacFrameOut := snacFrame{
		foodGroup: 0x17,
		subGroup:  0x07,
	}
	snacPayloadOut := &snacBUCPChallengeResponse{
		authKey: "theauthkey",
	}

	return writeOutSNAC(flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacBUCPLoginRequest struct {
	TLVPayload
}

func (s *snacBUCPLoginRequest) read(r io.Reader) error {
	lookup := map[uint16]reflect.Kind{
		0x01: reflect.String, // screen name
		0x03: reflect.String, // client ID string
		0x25: reflect.Slice,  // password md5 hash
		0x16: reflect.Uint16, // client ID
		0x17: reflect.Uint16, // client major version
		0x18: reflect.Uint16, // client minor version
		0x19: reflect.Uint16, // client lesser version
		0x1A: reflect.Uint16, // client build number
		0x14: reflect.Uint32, // distribution number
		0x0F: reflect.String, // client language
		0x0E: reflect.String, // client country
		0x4A: reflect.Slice,  // SSI use flag
		0x06: reflect.String, // SSI use flag
		0x4C: reflect.Slice,  // use old md5?
	}
	return s.TLVPayload.read(r, lookup)
}

func ReceiveAndSendBUCPLoginRequest(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	snacPayload := &snacBUCPLoginRequest{}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendBUCPLoginRequest read SNAC: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: 0x17,
		subGroup:  0x03,
	}
	snacPayloadOut := &snacBUCPLoginRequest{
		TLVPayload: TLVPayload{
			TLVs: []*TLV{
				{
					tType: 0x01,
					val:   "myscreenname",
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
					val:   "192.168.64.1:5191",
				},
				{
					tType: 0x06,
					val:   []byte("thecookie"),
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

	return writeOutSNAC(flap, snacFrameOut, snacPayloadOut, sequence, w)
}
