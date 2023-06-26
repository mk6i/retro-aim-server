package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

var OServiceRoute = map[uint16]routeHandler{
	0x01: routeOService,
}

func routeOService(frame *flapFrame, rw io.ReadWriter) error {

	return nil
}

func WriteFlapSignonFrame(conn net.Conn) error {

	startMarker := uint8(42)
	if err := binary.Write(conn, binary.BigEndian, startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	frameType := uint8(1)
	if err := binary.Write(conn, binary.BigEndian, frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	sequenceNumber := uint16(100)
	if err := binary.Write(conn, binary.BigEndian, sequenceNumber); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	payloadLength := uint16(4)
	if err := binary.Write(conn, binary.BigEndian, payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	flapVersion := uint32(1)
	if err := binary.Write(conn, binary.BigEndian, flapVersion); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return nil
}

func ReadFlapSignonFrame(conn net.Conn) (uint16, error) {

	var startMarker uint8
	if err := binary.Read(conn, binary.BigEndian, &startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("start marker: %d\n", startMarker)

	var frameType uint8
	if err := binary.Read(conn, binary.BigEndian, &frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("frame type: %d\n", frameType)

	var sequenceNumber uint16
	if err := binary.Read(conn, binary.BigEndian, &sequenceNumber); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("sequence number: %d\n", sequenceNumber)

	var payloadLength uint16
	if err := binary.Read(conn, binary.BigEndian, &payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("payload length: %d\n", payloadLength)

	var flapVersion uint32
	if err := binary.Read(conn, binary.BigEndian, &flapVersion); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("flap version: %d\n", flapVersion)

	return payloadLength, nil
}

func WriteOServiceHostOnline(conn net.Conn, sequence uint16) error {

	snac := &snac01_03{
		snacFrame: snacFrame{
			foodGroup: 0x01,
			subGroup:  0x03,
		},
		foodGroups: []uint16{
			0x0001, 0x0002, 0x0003, 0x0004, 0x0009, 0x0013,
		},
	}

	fmt.Printf("writeOServiceHostOnline SNAC: %+v\n", snac)

	snacBuf := &bytes.Buffer{}
	if err := snac.write(snacBuf); err != nil {
		return err
	}

	flap := &flapFrame{
		startMarker:   42,
		frameType:     2,
		sequence:      sequence,
		payloadLength: uint16(snacBuf.Len()),
	}

	fmt.Printf("writeOServiceHostOnline FLAP: %+v\n", flap)

	if err := flap.write(conn); err != nil {
		return err
	}

	_, err := conn.Write(snacBuf.Bytes())
	return err
}

type snac01_17_18 struct {
	snacFrame
	versions map[uint16]uint16
}

func (s *snac01_17_18) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}
	for {
		var family uint16
		if err := binary.Read(r, binary.BigEndian, &family); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		var version uint16
		if err := binary.Read(r, binary.BigEndian, &version); err != nil {
			return err
		}
		s.versions[family] = version
	}
	return nil
}

func (s *snac01_17_18) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	for family, version := range s.versions {
		if err := binary.Write(w, binary.BigEndian, family); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, version); err != nil {
			return err
		}
	}
	return nil
}

type snac01_03 struct {
	snacFrame
	foodGroups []uint16
}

func (s *snac01_03) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.foodGroups); err != nil {
		return err
	}
	return nil
}

func ReceiveAndSendHostVersions(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac01_17_18{
		versions: make(map[uint16]uint16),
	}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read SNAC: %+v\n", snac)

	// respond
	snac.snacFrame.subGroup = 0x18

	snacBuf := &bytes.Buffer{}
	if err := snac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("receiveAndSendHostVersions write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions write SNAC: %+v\n", snac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type rateClass struct {
	ID              uint16
	windowSize      uint32
	clearLevel      uint32
	alertLevel      uint32
	limitLevel      uint32
	disconnectLevel uint32
	currentLevel    uint32
	maxLevel        uint32
	lastTime        uint32 // protocol v2 only
	currentState    byte   // protocol v2 only
}

type rateGroup struct {
	ID    uint16
	pairs []struct {
		foodGroup uint16
		subGroup  uint16
	}
}

type snac01_07 struct {
	snacFrame
	rateClasses []rateClass
	rateGroups  []rateGroup
}

func (s *snac01_07) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.rateClasses))); err != nil {
		return err
	}
	for _, rateClass := range s.rateClasses {
		if err := binary.Write(w, binary.BigEndian, rateClass.ID); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.windowSize); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.clearLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.alertLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.limitLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.disconnectLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.currentLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.maxLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.lastTime); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.currentState); err != nil {
			return err
		}
	}
	for _, rateGroup := range s.rateGroups {
		if err := binary.Write(w, binary.BigEndian, rateGroup.ID); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint16(len(rateGroup.pairs))); err != nil {
			return err
		}
		for _, pair := range rateGroup.pairs {
			if err := binary.Write(w, binary.BigEndian, pair.foodGroup); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, pair.subGroup); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReceiveAndSendServiceRateParams(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRateParams read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	incomingSnac := &snacFrame{}
	if err := incomingSnac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRateParams read SNAC: %+v\n", incomingSnac)

	// respond
	snac := &snac01_07{
		snacFrame: snacFrame{
			foodGroup: 0x01,
			subGroup:  0x07,
		},
		rateClasses: []rateClass{
			//{
			//	ID:              1,
			//	windowSize:      10,
			//	clearLevel:      10,
			//	alertLevel:      10,
			//	limitLevel:      10,
			//	disconnectLevel: 10,
			//	currentLevel:    10,
			//	maxLevel:        10,
			//	lastTime:        10,
			//	currentState:    10,
			//},
		},
		rateGroups: []rateGroup{
			//{
			//	ID: 1,
			//	pairs: []struct {
			//		foodGroup uint16
			//		subGroup  uint16
			//	}{
			//		{
			//			foodGroup: 1,
			//			subGroup:  1,
			//		},
			//	},
			//},
		},
	}

	snacBuf := &bytes.Buffer{}
	if err := snac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("receiveAndSendServiceRateParams write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRateParams write SNAC: %+v\n", snac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type snac01_08 struct {
	snacFrame
	subs []uint16
}

func (s *snac01_08) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}
	for {
		var rateClass uint16
		if err := binary.Read(r, binary.BigEndian, &rateClass); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		s.subs = append(s.subs, rateClass)
	}
	return nil
}

type snac01_0F struct {
	snacFrame
	screenName   string
	warningLevel uint16
	TLVs         []*TLV
}

func (s *snac01_0F) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint8(len(s.screenName))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(s.screenName)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.warningLevel); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.TLVs))); err != nil {
		return err
	}
	for _, t := range s.TLVs {
		if err := t.write(w); err != nil {
			return err
		}
	}
	return nil
}

func ReceiveAndSendServiceRequestSelfInfo(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRequestSelfInfo read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snacFrame{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("receiveAndSendServiceRequestSelfInfo read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snac01_0F{
		snacFrame: snacFrame{
			foodGroup: 0x01,
			subGroup:  0x0F,
		},
		screenName:   "screenname",
		warningLevel: 0,
		TLVs: []*TLV{
			{
				tType: 0x01,
				val:   uint32(0x0010),
			},
			{
				tType: 0x02,
				val:   uint32(time.Now().Unix()),
			},
			{
				tType: 0x03,
				val:   uint32(1687314861),
			},
			{
				tType: 0x04,
				val:   uint32(0),
			},
			{
				tType: 0x05,
				val:   uint32(1687314841),
			},
			{
				tType: 0x0D,
				val:   make([]byte, 0),
			},
			{
				tType: 0x0F,
				val:   uint32(0),
			},
		},
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("receiveAndSendServiceRequestSelfInfo write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRequestSelfInfo write SNAC: %+v\n", snac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}
