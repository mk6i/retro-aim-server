package oscar

import (
	"bytes"
	"fmt"
	"io"
)

type snac02_03 struct {
	snacFrame
	TLVs []*TLV
}

func (s *snac02_03) write(w io.Writer) error {
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

func SendAndReceiveLocateRights(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveLocateRights read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snacFrame{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("sendAndReceiveLocateRights read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snac02_03{
		snacFrame: snacFrame{
			foodGroup: 0x02,
			subGroup:  0x03,
		},
		TLVs: []*TLV{
			{
				tType: 0x01,
				val:   uint16(0),
			},
			{
				tType: 0x02,
				val:   uint16(0),
			},
			{
				tType: 0x03,
				val:   uint16(0),
			},
			{
				tType: 0x04,
				val:   uint16(0),
			},
			{
				tType: 0x05,
				val:   uint16(0),
			},
		},
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("sendAndReceiveLocateRights write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveLocateRights write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}
