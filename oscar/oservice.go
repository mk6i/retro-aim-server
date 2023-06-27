package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	OServiceErr               uint16 = 0x0001
	OServiceClientOnline             = 0x0002
	OServiceHostOnline               = 0x0003
	OServiceServiceRequest           = 0x0004
	OServiceServiceResponse          = 0x0005
	OServiceRateParamsQuery          = 0x0006
	OServiceRateParamsReply          = 0x0007
	OServiceRateParamsSubAdd         = 0x0008
	OServiceRateDelParamSub          = 0x0009
	OServiceRateParamChange          = 0x000A
	OServicePauseReq                 = 0x000B
	OServicePauseAck                 = 0x000C
	OServiceResume                   = 0x000D
	OServiceUserInfoQuery            = 0x000E
	OServiceUserInfoUpdate           = 0x000F
	OServiceEvilNotification         = 0x0010
	OServiceIdleNotification         = 0x0011
	OServiceMigrateGroups            = 0x0012
	OServiceMotd                     = 0x0013
	OServiceSetPrivacyFlags          = 0x0014
	OServiceWellKnownUrls            = 0x0015
	OServiceNoop                     = 0x0016
	OServiceClientVersions           = 0x0017
	OServiceHostVersions             = 0x0018
	OServiceMaxConfigQuery           = 0x0019
	OServiceMaxConfigReply           = 0x001A
	OServiceStoreConfig              = 0x001B
	OServiceConfigQuery              = 0x001C
	OServiceConfigReply              = 0x001D
	OServiceSetUserinfoFields        = 0x001E
	OServiceProbeReq                 = 0x001F
	OServiceProbeAck                 = 0x0020
	OServiceBartReply                = 0x0021
	OServiceBartQuery2               = 0x0022
	OServiceBartReply2               = 0x0023
)

func routeOService(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	switch snac.subGroup {
	case OServiceErr:
		panic("not implemented")
	case OServiceClientOnline:
		panic("not implemented")
	case OServiceHostOnline:
		panic("not implemented")
	case OServiceServiceRequest:
		panic("not implemented")
	case OServiceServiceResponse:
		panic("not implemented")
	case OServiceRateParamsQuery:
		panic("not implemented")
	case OServiceRateParamsReply:
		panic("not implemented")
	case OServiceRateParamsSubAdd:
		panic("not implemented")
	case OServiceRateDelParamSub:
		panic("not implemented")
	case OServiceRateParamChange:
		panic("not implemented")
	case OServicePauseReq:
		panic("not implemented")
	case OServicePauseAck:
		panic("not implemented")
	case OServiceResume:
		panic("not implemented")
	case OServiceUserInfoQuery:
		panic("not implemented")
	case OServiceUserInfoUpdate:
		panic("not implemented")
	case OServiceEvilNotification:
		panic("not implemented")
	case OServiceIdleNotification:
		panic("not implemented")
	case OServiceMigrateGroups:
		panic("not implemented")
	case OServiceMotd:
		panic("not implemented")
	case OServiceSetPrivacyFlags:
		panic("not implemented")
	case OServiceWellKnownUrls:
		panic("not implemented")
	case OServiceNoop:
		panic("not implemented")
	case OServiceClientVersions:
		return ReceiveAndSendHostVersions(flap, snac, r, w, sequence)
	case OServiceMaxConfigQuery:
		panic("not implemented")
	case OServiceMaxConfigReply:
		panic("not implemented")
	case OServiceStoreConfig:
		panic("not implemented")
	case OServiceConfigQuery:
		panic("not implemented")
	case OServiceConfigReply:
		panic("not implemented")
	case OServiceSetUserinfoFields:
		panic("not implemented")
	case OServiceProbeReq:
		panic("not implemented")
	case OServiceProbeAck:
		panic("not implemented")
	case OServiceBartReply:
		panic("not implemented")
	case OServiceBartQuery2:
		panic("not implemented")
	case OServiceBartReply2:
		panic("not implemented")
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

type snacVersions struct {
	versions map[uint16]uint16
}

func (s *snacVersions) read(r io.Reader) error {
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

func (s *snacVersions) write(w io.Writer) error {
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

func ReceiveAndSendHostVersions(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	fmt.Printf("receiveAndSendHostVersions read SNAC frame: %+v\n", snac)

	snacPayload := &snacVersions{
		versions: make(map[uint16]uint16),
	}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read SNAC: %+v\n", snacPayload)

	snacFrameOut := snacFrame{
		foodGroup: OSERVICE,
		subGroup:  OServiceHostVersions,
	}

	return writeOutSNAC(flap, snacFrameOut, snacPayload, sequence, w)
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
