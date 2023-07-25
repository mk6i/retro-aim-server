package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

const (
	ICBMErr                uint16 = 0x0001
	ICBMAddParameters             = 0x0002
	ICBMDelParameters             = 0x0003
	ICBMParameterQuery            = 0x0004
	ICBMParameterReply            = 0x0005
	ICBMChannelMsgTohost          = 0x0006
	ICBMChannelMsgToclient        = 0x0007
	ICBMEvilRequest               = 0x0008
	ICBMEvilReply                 = 0x0009
	ICBMMissedCalls               = 0x000A
	ICBMClientErr                 = 0x000B
	ICBMHostAck                   = 0x000C
	ICBMSinStored                 = 0x000D
	ICBMSinListQuery              = 0x000E
	ICBMSinListReply              = 0x000F
	ICBMSinRetrieve               = 0x0010
	ICBMSinDelete                 = 0x0011
	ICBMNotifyRequest             = 0x0012
	ICBMNotifyReply               = 0x0013
	ICBMClientEvent               = 0x0014
	ICBMSinReply                  = 0x0017
)

func routeICBM(sm *SessionManager, fm *FeedbagStore, sess *Session, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case ICBMErr:
		panic("not implemented")
	case ICBMAddParameters:
		return ReceiveAddParameters(flap, snac, r, w, sequence)
	case ICBMDelParameters:
		panic("not implemented")
	case ICBMParameterQuery:
		return SendAndReceiveICBMParameterReply(flap, snac, r, w, sequence)
	case ICBMChannelMsgTohost:
		return SendAndReceiveChannelMsgTohost(sm, fm, sess, flap, snac, r, w, sequence)
	case ICBMChannelMsgToclient:
		panic("not implemented")
	case ICBMEvilRequest:
		panic("not implemented")
	case ICBMEvilReply:
		panic("not implemented")
	case ICBMMissedCalls:
		panic("not implemented")
	case ICBMClientErr:
		panic("not implemented")
	case ICBMHostAck:
		panic("not implemented")
	case ICBMSinStored:
		panic("not implemented")
	case ICBMSinListQuery:
		panic("not implemented")
	case ICBMSinListReply:
		panic("not implemented")
	case ICBMSinRetrieve:
		panic("not implemented")
	case ICBMSinDelete:
		panic("not implemented")
	case ICBMNotifyRequest:
		panic("not implemented")
	case ICBMNotifyReply:
		panic("not implemented")
	case ICBMClientEvent:
		panic("not implemented")
	case ICBMSinReply:
		panic("not implemented")
	}

	return nil
}

type snacParameterRequest struct {
	channel              uint16
	ICBMFlags            uint32
	maxIncomingICBMLen   uint16
	maxSourceEvil        uint16
	maxDestinationEvil   uint16
	minInterICBMInterval uint32
}

func (s *snacParameterRequest) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.channel); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.ICBMFlags); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.maxIncomingICBMLen); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.maxSourceEvil); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.maxDestinationEvil); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.minInterICBMInterval); err != nil {
		return err
	}
	return nil
}

type snacParameterResponse struct {
	maxSlots             uint16
	ICBMFlags            uint32
	maxIncomingICBMLen   uint16
	maxSourceEvil        uint16
	maxDestinationEvil   uint16
	minInterICBMInterval uint32
}

func (s *snacParameterResponse) write(w io.Writer) error {
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

func SendAndReceiveICBMParameterReply(flap *flapFrame, snac *snacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveICBMParameterReply read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: ICBM,
		subGroup:  ICBMParameterReply,
	}
	snacPayloadOut := &snacParameterResponse{
		maxSlots:             100,
		ICBMFlags:            3,
		maxIncomingICBMLen:   512,
		maxSourceEvil:        999,
		maxDestinationEvil:   999,
		minInterICBMInterval: 0,
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacMessageToHost struct {
	cookie     [8]byte
	channelID  uint16
	screenName string
	TLVPayload
}

func (s *snacMessageToHost) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.cookie); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.channelID); err != nil {
		return err
	}

	var l uint8
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	buf := make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	s.screenName = string(buf)

	return s.TLVPayload.read(r, map[uint16]reflect.Kind{
		0x02: reflect.Slice,
		0x03: reflect.Slice,
	})
}

func SendAndReceiveChannelMsgTohost(sm *SessionManager, fm *FeedbagStore, sess *Session, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveChannelMsgTohost read SNAC frame: %+v\n", snac)

	snacPayloadIn := &snacMessageToHost{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	SendIM(sm, snacPayloadIn.screenName, "hey there")

	//session := sm.RetrieveByScreenName(snacPayloadIn.screenName)
	//
	//messagePayload, found := snacPayloadIn.TLVPayload.getSlice(0x02)
	//if !found {
	//	return errors.New("unable to find message data tlv")
	//}
	//
	//mm := &XMessage{
	//	flap: &flapFrame{
	//		startMarker: 42,
	//		frameType:   2,
	//	},
	//	snacFrame: snacFrame{
	//		foodGroup: ICBM,
	//		subGroup:  ICBMChannelMsgToclient,
	//		requestID: 12425,
	//	},
	//	snacOut: &snacClientIM{
	//		cookie:       [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	//		channelID:    1,
	//		screenName:   snacPayloadIn.screenName,
	//		warningLevel: 0,
	//		TLVPayload: TLVPayload{
	//			TLVs: []*TLV{
	//				{
	//					tType: 0x02,
	//					val:   messagePayload,
	//				},
	//			},
	//		},
	//	},
	//}
	//session.MsgChan <- mm

	return nil
}

func ReceiveAddParameters(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveAddParameters read SNAC frame: %+v\n", snac)

	snacPayload := &snacParameterRequest{}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	fmt.Printf("ReceiveAddParameters read SNAC: %+v\n", snacPayload)
	return nil
}

type snacClientIM struct {
	cookie       [8]byte
	channelID    uint16
	screenName   string
	warningLevel uint16
	TLVPayload
}

func (f *snacClientIM) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, f.cookie); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.channelID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint8(len(f.screenName))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(f.screenName)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.warningLevel); err != nil {
		return err
	}
	// attrs.num
	if err := binary.Write(w, binary.BigEndian, uint16(0)); err != nil {
		return err
	}
	return f.TLVPayload.write(w)
}

type messageData struct {
	text string
}

func (m *messageData) write(w io.Writer) error {
	// required capabilities
	if err := binary.Write(w, binary.BigEndian, uint8(0x05)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint8(0x01)); err != nil {
		return err
	}
	l := []byte{0x01}
	if err := binary.Write(w, binary.BigEndian, uint16(len(l))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, l); err != nil {
		return err
	}

	// message text

	// identifier
	if err := binary.Write(w, binary.BigEndian, uint8(0x01)); err != nil {
		return err
	}
	// version
	if err := binary.Write(w, binary.BigEndian, uint8(0x01)); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	// charset num
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return err
	}
	// charset subset
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return err
	}
	// message text
	if err := binary.Write(buf, binary.BigEndian, []byte(m.text)); err != nil {
		return err
	}

	// TLV len
	if err := binary.Write(w, binary.BigEndian, uint16(buf.Len())); err != nil {
		return err
	}
	// TLV payload
	if err := binary.Write(w, binary.BigEndian, buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func SendIM(sm *SessionManager, screenName string, msg string) error {

	session := sm.RetrieveByScreenName(screenName)

	session.MsgChan <- &XMessage{
		flap: &flapFrame{
			startMarker: 42,
			frameType:   2,
		},
		snacFrame: snacFrame{
			foodGroup: ICBM,
			subGroup:  ICBMChannelMsgToclient,
			requestID: 12425,
		},
		snacOut: &snacClientIM{
			cookie:     [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			channelID:  1,
			screenName: screenName,
			TLVPayload: TLVPayload{
				TLVs: []*TLV{
					{
						tType: 0x02,
						val: &messageData{
							text: msg,
						},
					},
				},
			},
		},
	}

	return nil
}
