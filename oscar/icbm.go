package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type snac04_05 struct {
	snacFrame
	maxSlots             uint16
	ICBMFlags            uint32
	maxIncomingICBMLen   uint16
	maxSourceEvil        uint16
	maxDestinationEvil   uint16
	minInterICBMInterval uint32
}

func (s *snac04_05) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.maxSlots); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.ICBMFlags); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.maxIncomingICBMLen); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.maxSourceEvil); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.maxDestinationEvil); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.minInterICBMInterval); err != nil {
		return err
	}
	return nil
}

func SendAndReceiveICBMParameterReply(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveICBMParameterReply read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snacFrame{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("sendAndReceiveICBMParameterReply read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snac04_05{
		snacFrame: snacFrame{
			foodGroup: 0x04,
			subGroup:  0x05,
		},
		maxSlots:             100,
		ICBMFlags:            0x00000001,
		maxIncomingICBMLen:   8000,
		maxSourceEvil:        999,
		maxDestinationEvil:   999,
		minInterICBMInterval: 100,
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("sendAndReceiveICBMParameterReply write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveICBMParameterReply write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}
