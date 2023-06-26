package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"
)

type snac13_02 struct {
	snacFrame
	TLVs []*TLV
}

func (s *snac13_02) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}

	lookup := map[uint16]reflect.Kind{0x0B: reflect.Uint16}

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

func SendAndReceiveFeedbagRightsQuery(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveFeedbagRightsQuery read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac13_02{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("sendAndReceiveFeedbagRightsQuery read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snacFrameTLV{
		snacFrame: snacFrame{
			foodGroup: 0x13,
			subGroup:  0x03,
		},
		TLVs: []*TLV{
			{
				tType: 0x03,
				val:   uint16(200),
			},
			{
				tType: 0x04,
				val:   uint16(200),
			},
			{
				tType: 0x05,
				val:   uint16(200),
			},
			{
				tType: 0x06,
				val:   uint16(200),
			},
			{
				tType: 0x07,
				val:   uint16(200),
			},
			{
				tType: 0x08,
				val:   uint16(200),
			},
			{
				tType: 0x09,
				val:   uint16(200),
			},
			{
				tType: 0x0A,
				val:   uint16(200),
			},
			{
				tType: 0x0C,
				val:   uint16(200),
			},
			{
				tType: 0x0D,
				val:   uint16(200),
			},
			{
				tType: 0x0E,
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

	fmt.Printf("sendAndReceiveFeedbagRightsQuery write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveFeedbagRightsQuery write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type feedbagItem struct {
	name    string
	groupID uint16
	itemID  uint16
	classID uint16
	tlvs    []*TLV
}

func (f *feedbagItem) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.name))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(f.name)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.groupID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.itemID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.classID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.tlvs))); err != nil {
		return err
	}
	for _, tlv := range f.tlvs {
		if err := tlv.write(w); err != nil {
			return err
		}
	}
	return nil
}

type snac13_06 struct {
	snacFrame
	version    uint8
	items      []*feedbagItem
	lastUpdate uint32
}

func (s *snac13_06) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.version); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.items))); err != nil {
		return err
	}
	for _, t := range s.items {
		if err := t.write(w); err != nil {
			return err
		}
	}
	return binary.Write(w, binary.BigEndian, s.lastUpdate)
}

func ReceiveAndSendFeedbagQuery(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendFeedbagQuery read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snacFrame{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("receiveAndSendFeedbagQuery read SNAC: %+v\n", snac)

	// send
	writeSnac := &snac13_06{
		snacFrame: snacFrame{
			foodGroup: 0x13,
			subGroup:  0x06,
		},
		version: 0,
		items: []*feedbagItem{
			{
				groupID: 0,
				itemID:  0,
				classID: 0,
				name:    "",
				tlvs: []*TLV{
					{
						tType: 0x00C8,
						val:   []uint16{321, 10},
					},
				},
			},
			{
				groupID: 0,
				itemID:  1805,
				classID: 3,
				name:    "spimmer123",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0,
				itemID:  4046,
				classID: 0x14,
				name:    "5",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0,
				itemID:  12108,
				classID: 4,
				name:    "",
				tlvs: []*TLV{
					{
						tType: 202,
						val:   uint8(0x04),
					},
					{
						tType: 203,
						val:   uint32(0xffffffff),
					},
					{
						tType: 204,
						val:   uint32(1),
					},
				},
			},
			{
				groupID: 0x0A,
				itemID:  0,
				classID: 1,
				name:    "Friends",
				tlvs: []*TLV{
					{
						tType: 200,
						val:   []uint16{110, 147},
					},
				},
			},
			{
				groupID: 0x0A,
				itemID:  110,
				classID: 0,
				name:    "ChattingChuck",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0x0A,
				itemID:  147,
				classID: 0,
				name:    "example@example.com",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0,
				itemID:  0,
				classID: 1,
				name:    "Empty Group",
				tlvs: []*TLV{
					{
						tType: 200,
						val:   []uint16{},
					},
				},
			},
		},
		lastUpdate: uint32(time.Now().Unix()),
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("receiveAndSendFeedbagQuery write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendFeedbagQuery write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())

	return err
}
