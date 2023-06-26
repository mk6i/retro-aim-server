package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"reflect"
)

type snac17_06 struct {
	snacFrame
	TLVs []*TLV
}

func (s *snac17_06) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}

	lookup := map[uint16]reflect.Kind{
		0x01: reflect.String,
		0x4B: reflect.String,
		0x5A: reflect.String,
	}

	for {
		// todo, don't like this extra alloc when we're EOF
		tlv := &TLV{}
		if err := tlv.read(r, lookup); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		s.TLVs = append(s.TLVs, tlv)
	}

	return nil
}

type snac17_07 struct {
	snacFrame
	authKey string
}

func (s *snac17_07) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.authKey))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(s.authKey)); err != nil {
		return err
	}
	return nil
}

func ReceiveAndSendAuthChallenge(rw net.Conn, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac17_06{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snac17_07{
		snacFrame: snacFrame{
			foodGroup: 0x17,
			subGroup:  0x07,
		},
		authKey: "theauthkey",
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("ReceiveAndSendAuthChallenge write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type snac17_02 struct {
	snacFrameTLV
}

func (s *snac17_02) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}

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

	for {
		// todo, don't like this extra alloc when we're EOF
		tlv := &TLV{}
		if err := tlv.read(r, lookup); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		s.TLVs = append(s.TLVs, tlv)
	}

	return nil
}

func ReceiveAndSendBUCPLoginRequest(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendBUCPLoginRequest read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac17_02{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendBUCPLoginRequest read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snacFrameTLV{
		snacFrame: snacFrame{
			foodGroup: 0x17,
			subGroup:  0x03,
		},
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
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("ReceiveAndSendAuthChallenge write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}
