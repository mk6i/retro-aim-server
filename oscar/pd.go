package oscar

import (
	"bytes"
	"fmt"
	"io"
)

func SendAndReceivePDRightsQuery(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceivePDRightsQuery read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snacFrame{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("sendAndReceivePDRightsQuery read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snacFrameTLV{
		snacFrame: snacFrame{
			foodGroup: 0x09,
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
				tType: 0x03,
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

	fmt.Printf("sendAndReceivePDRightsQuery write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceivePDRightsQuery write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}
