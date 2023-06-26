package oscar

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

type snac03_02 struct {
	snacFrame
	TLVs []*TLV
}

func (s *snac03_02) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}

	lookup := map[uint16]reflect.Kind{0x05: reflect.Uint16}

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

type snac03_03 struct {
	snacFrame
	TLVs []*TLV
}

func (s *snac03_03) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	for _, tlv := range s.TLVs {
		if err := tlv.write(w); err != nil {
			return err
		}
	}
	return nil
}

func SendAndReceiveBuddyRights(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveBuddyRights read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac03_02{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("sendAndReceiveBuddyRights read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snac03_03{
		snacFrame: snacFrame{
			foodGroup: 0x03,
			subGroup:  0x03,
		},
		TLVs: []*TLV{
			{
				tType: 0x01,
				val:   uint16(100),
			},
			{
				tType: 0x02,
				val:   uint16(100),
			},
			{
				tType: 0x04,
				val:   uint16(100),
			},
		},
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("sendAndReceiveBuddyRights write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveBuddyRights write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}
