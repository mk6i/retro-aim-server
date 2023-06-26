package oscar

import (
	"bytes"
	"fmt"
	"io"
)

func SendAndReceiveNextChatRights(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveNextChatRights read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snacFrame{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("sendAndReceiveNextChatRights read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snacFrameTLV{
		snacFrame: snacFrame{
			foodGroup: 0x0D,
			subGroup:  0x09,
		},
		TLVs: []*TLV{},
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("sendAndReceiveNextChatRights write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveNextChatRights write SNAC: %+v\n", writeSnac)

	return nil
}
