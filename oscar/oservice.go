package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync/atomic"
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

func routeOService(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case OServiceErr:
		panic("not implemented")
	case OServiceClientOnline:
		return ReceiveClientOnline(flap, snac, r, w, sequence)
	case OServiceHostOnline:
		panic("not implemented")
	case OServiceServiceRequest:
		return ReceiveAndSendServiceRequest(flap, snac, r, w, sequence)
	case OServiceRateParamsQuery:
		return ReceiveAndSendServiceRateParams(flap, snac, r, w, sequence)
	case OServiceRateParamsSubAdd:
		return ReceiveRateParamsSubAdd(flap, snac, r, w, sequence)
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
		return ReceiveAndSendServiceRequestSelfInfo(flap, snac, r, w, sequence)
	case OServiceUserInfoUpdate:
		panic("not implemented")
	case OServiceEvilNotification:
		panic("not implemented")
	case OServiceIdleNotification:
		return ReceiveIdleNotification(flap, snac, r, w, sequence)
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
		return ReceiveSetUserInfoFields(flap, snac, r, w, sequence)
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

func WriteOServiceHostOnline(conn net.Conn, sequence *uint32) error {
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
		sequence:      uint16(*sequence),
		payloadLength: uint16(snacBuf.Len()),
	}
	atomic.AddUint32(sequence, 1)
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

func ReceiveAndSendHostVersions(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
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

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayload, sequence, w)
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

type snacOServiceRateParamsReply struct {
	rateClasses []rateClass
	rateGroups  []rateGroup
}

func (s *snacOServiceRateParamsReply) write(w io.Writer) error {
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

func ReceiveAndSendServiceRateParams(flap *flapFrame, snac *snacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveAndSendServiceRateParams read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: OSERVICE,
		subGroup:  OServiceRateParamsReply,
	}

	snacPayloadOut := &snacOServiceRateParamsReply{
		rateClasses: []rateClass{
			{
				ID:              0x0001,
				windowSize:      0x00000050,
				clearLevel:      0x000009C4,
				alertLevel:      0x000007D0,
				limitLevel:      0x000005DC,
				disconnectLevel: 0x00000320,
				currentLevel:    0x00000D69,
				maxLevel:        0x00001770,
				lastTime:        0x00000000,
				currentState:    0x00,
			},
		},
		rateGroups: []rateGroup{
			{
				ID: 1,
				pairs: []struct {
					foodGroup uint16
					subGroup  uint16
				}{},
			},
		},
	}

	for i := uint16(0); i < 24; i++ { // for each food group
		for j := uint16(0); j < 32; j++ { // for each subgroup
			snacPayloadOut.rateGroups[0].pairs = append(snacPayloadOut.rateGroups[0].pairs,
				struct {
					foodGroup uint16
					subGroup  uint16
				}{
					foodGroup: i,
					subGroup:  j,
				})
		}
	}
	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacOServiceUserInfoUpdate struct {
	TLVPayload
	screenName   string
	warningLevel uint16
}

func (s *snacOServiceUserInfoUpdate) write(w io.Writer) error {
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
	return s.TLVPayload.write(w)
}

func ReceiveAndSendServiceRequestSelfInfo(flap *flapFrame, snac *snacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveAndSendServiceRequestSelfInfo read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: OSERVICE,
		subGroup:  OServiceUserInfoUpdate,
	}
	snacPayloadOut := &snacOServiceUserInfoUpdate{
		screenName:   "screenname",
		warningLevel: 0,
		TLVPayload: TLVPayload{
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
		},
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveRateParamsSubAdd(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveRateParamsSubAdd read SNAC frame: %+v\n", snac)

	snacPayload := &TLVPayload{}
	lookup := map[uint16]reflect.Kind{}
	if err := snacPayload.read(r, lookup); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read SNAC: %+v\n", snacPayload)

	return nil
}

type clientVersion struct {
	foodGroup   uint16
	version     uint16
	toolID      uint16
	toolVersion uint16
}

func (c *clientVersion) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &c.foodGroup); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &c.version); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &c.toolID); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &c.toolVersion)
}

func ReceiveClientOnline(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveClientOnline read SNAC frame: %+v\n", snac)

	b := make([]byte, flap.payloadLength-10)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)

	for buf.Len() > 0 {
		item := &clientVersion{}
		if err := item.read(buf); err != nil {
			return err
		}
		fmt.Printf("ReceiveClientOnline read SNAC client version: %+v\n", item)
	}

	return nil
}

func ReceiveSetUserInfoFields(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveSetUserInfoFields read SNAC frame: %+v\n", snac)

	b := make([]byte, flap.payloadLength-10)
	if _, err := r.Read(b); err != nil {
		return err
	}

	snacPayload := &TLVPayload{}
	return snacPayload.read(bytes.NewBuffer(b), map[uint16]reflect.Kind{
		0x06: reflect.Uint32,
		0x1D: reflect.Slice,
	})
}

type snacIdleNotification struct {
	idleTime uint32
}

func (s *snacIdleNotification) read(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, &s.idleTime)
}

func ReceiveIdleNotification(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveIdleNotification read SNAC frame: %+v\n", snac)

	snacPayload := &snacIdleNotification{}
	return snacPayload.read(r)
}

type snacServiceRequest struct {
	foodGroup uint16
	TLVPayload
}

func (s *snacServiceRequest) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.foodGroup); err != nil {
		return err
	}
	return s.TLVPayload.read(r, map[uint16]reflect.Kind{})
}

const (
	OserviceTlvTagsReconnectHere uint16 = 0x05
	OserviceTlvTagsLoginCookie          = 0x06
	OserviceTlvTagsGroupId              = 0x0D
	OserviceTlvTagsSslCertname          = 0x8D
	OserviceTlvTagsSslState             = 0x8E
)

// ServiceHosts config should live outside this package
var ServiceHosts = map[uint16]string{
	STATS: "192.168.64.1:5192",
	ALERT: "192.168.64.1:5193",
	ODIR:  "192.168.64.1:5194",
}

func ReceiveAndSendServiceRequest(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveAndSendServiceRequest read SNAC frame: %+v\n", snac)

	snacPayload := &snacServiceRequest{}
	if err := snacPayload.read(r); err != nil {
		return nil
	}

	fmt.Printf("receiveAndSendServiceRequest read SNAC body: %+v\n", snacPayload)

	host, ok := ServiceHosts[snacPayload.foodGroup]
	if !ok {
		return fmt.Errorf("unable to find hostname for %s", host)
	}

	snacFrameOut := snacFrame{
		foodGroup: OSERVICE,
		subGroup:  OServiceServiceResponse,
	}
	snacPayloadOut := &TLVPayload{
		TLVs: []*TLV{
			{
				tType: OserviceTlvTagsReconnectHere,
				val:   host,
			},
			{
				tType: OserviceTlvTagsLoginCookie,
				val:   "some-cookie",
			},
			{
				tType: OserviceTlvTagsGroupId,
				val:   snacPayload.foodGroup,
			},
			{
				tType: OserviceTlvTagsSslCertname,
				val:   "",
			},
			{
				tType: OserviceTlvTagsSslState,
				val:   uint8(0x00),
			},
		},
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}
