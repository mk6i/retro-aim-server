package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type flapFrame struct {
	startMarker   uint8
	frameType     uint8
	sequence      uint16
	payloadLength uint16
}

func (f *flapFrame) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, f.startMarker); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.frameType); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.sequence); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, f.payloadLength)
}

func (f *flapFrame) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &f.startMarker); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.frameType); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.sequence); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &f.payloadLength)
}

type snacFrame struct {
	foodGroup uint16
	subGroup  uint16
	flags     uint16
	requestID uint32
}

func (s *snacFrame) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.foodGroup); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.subGroup); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.flags); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.requestID); err != nil {
		return err
	}
	return nil
}

func (s *snacFrame) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.foodGroup); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.subGroup); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.flags); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &s.requestID)
}

type snacWriter interface {
	write(w io.Writer) error
}

type TLVPayload struct {
	TLVs []*TLV
}

func (s *TLVPayload) read(r io.Reader, lookup map[uint16]reflect.Kind) error {
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

func (s *TLVPayload) write(w io.Writer) error {
	for _, tlv := range s.TLVs {
		if err := tlv.write(w); err != nil {
			return err
		}
	}
	return nil
}

type TLV struct {
	tType uint16
	val   any
}

type snacFrameTLV struct {
	snacFrame
	TLVs []*TLV
}

func (s *snacFrameTLV) write(w io.Writer) error {
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

func (t *TLV) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, t.tType); err != nil {
		return err
	}

	var valLen uint16
	val := t.val

	switch t.val.(type) {
	case uint8:
		valLen = 1
	case uint16:
		valLen = 2
	case uint32:
		valLen = 4
	case []uint16:
		valLen = uint16(len(t.val.([]uint16)))
	case []byte:
		valLen = uint16(len(t.val.([]byte)))
	case string:
		valLen = uint16(len(t.val.(string)))
		val = []byte(t.val.(string))
	}

	if err := binary.Write(w, binary.BigEndian, valLen); err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, val)
}

func (t *TLV) read(r io.Reader, typeLookup map[uint16]reflect.Kind) error {
	if err := binary.Read(r, binary.BigEndian, &t.tType); err != nil {
		return err
	}
	var tlvValLen uint16
	if err := binary.Read(r, binary.BigEndian, &tlvValLen); err != nil {
		return err
	}

	kind, ok := typeLookup[t.tType]
	if !ok {
		return fmt.Errorf("unknown data type for TLV %d", t.tType)
	}

	switch kind {
	case reflect.Uint8:
		var val uint16
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return err
		}
		t.val = val
	case reflect.Uint16:
		var val uint16
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return err
		}
		t.val = val
	case reflect.Uint32:
		var val uint32
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return err
		}
		t.val = val
	case reflect.String:
		buf := make([]byte, tlvValLen)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		t.val = string(buf)
	case reflect.Slice:
		buf := make([]byte, tlvValLen)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		t.val = string(buf)
	default:
		panic("unsupported data type")
	}

	return nil
}

type flapSignonFrame struct {
	flapFrame
	flapVersion uint32
	TLVs        []*TLV
}

func (f *flapSignonFrame) write(w io.Writer) error {
	if err := f.flapFrame.write(w); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, f.flapVersion)
}

func (f *flapSignonFrame) read(r io.Reader) error {
	if err := f.flapFrame.read(r); err != nil {
		return err
	}

	// todo: combine b+buf?
	b := make([]byte, f.payloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.BigEndian, &f.flapVersion); err != nil {
		return err
	}

	lookup := map[uint16]reflect.Kind{
		0x06: reflect.String,
		0x4A: reflect.Uint8,
	}

	for {
		// todo, don't like this extra alloc when we're EOF
		tlv := &TLV{}
		if err := tlv.read(buf, lookup); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		f.TLVs = append(f.TLVs, tlv)
	}

	return nil
}

func SendAndReceiveSignonFrame(rw io.ReadWriter, sequence uint16) error {
	// send
	flap := &flapSignonFrame{
		flapFrame: flapFrame{
			startMarker:   42,
			frameType:     1,
			sequence:      sequence,
			payloadLength: 4, // size of flapVersion
		},
		flapVersion: 1,
	}

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("SendAndReceiveSignonFrame read FLAP: %+v\n", flap)

	// receive
	flap = &flapSignonFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("SendAndReceiveSignonFrame write FLAP: %+v\n", flap)

	return nil
}

const (
	OSERVICE      uint16 = 0x0001
	LOCATE               = 0x0002
	BUDDY                = 0x0003
	ICBM                 = 0x0004
	ADVERT               = 0x0005
	INVITE               = 0x0006
	ADMIN                = 0x0007
	POPUP                = 0x0008
	PD                   = 0x0009
	USER_LOOKUP          = 0x000A
	STATS                = 0x000B
	TRANSLATE            = 0x000C
	CHAT_NAV             = 0x000D
	CHAT                 = 0x000E
	ODIR                 = 0x000F
	BART                 = 0x0010
	FEEDBAG              = 0x0013
	ICQ                  = 0x0015
	BUCP                 = 0x0017
	ALERT                = 0x0018
	PLUGIN               = 0x0022
	UNNAMED_FG_24        = 0x0024
	MDIR                 = 0x0025
	ARS                  = 0x044A
)

func ReadBos(rw io.ReadWriter, sequence uint16) error {
	for {
		// receive
		flap := &flapFrame{}
		if err := flap.read(rw); err != nil {
			return err
		}

		b := make([]byte, flap.payloadLength)
		if _, err := rw.Read(b); err != nil {
			return err
		}

		buf := bytes.NewBuffer(b)

		snac := &snacFrame{}
		if err := snac.read(buf); err != nil {
			return err
		}

		switch snac.foodGroup {
		case OSERVICE:
			if err := routeOService(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case LOCATE:
			if err := routeLocate(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case BUDDY:
			if err := routeBuddy(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case ICBM:
			if err := routeICBM(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case ADVERT:
		case INVITE:
		case ADMIN:
		case POPUP:
		case PD:
			if err := routePD(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case USER_LOOKUP:
		case STATS:
		case TRANSLATE:
		case CHAT_NAV:
			if err := routeChatNav(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case CHAT:
		case ODIR:
		case BART:
		case FEEDBAG:
			if err := routeFeedbag(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case ICQ:
		case BUCP:
			if err := routeBUCP(flap, snac, buf, rw, sequence); err != nil {
				return err
			}
			sequence++
		case ALERT:
		case PLUGIN:
		case UNNAMED_FG_24:
		case MDIR:
		case ARS:
		}

	}
}

func writeOutSNAC(flap *flapFrame, snacFrame snacFrame, snacOut snacWriter, sequence uint16, w io.Writer) error {
	snacBuf := &bytes.Buffer{}
	if err := snacFrame.write(snacBuf); err != nil {
		return err
	}
	if err := snacOut.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf(" write FLAP: %+v\n", flap)

	if err := flap.write(w); err != nil {
		return err
	}

	fmt.Printf(" write SNAC: %+v\n", snacOut)

	_, err := w.Write(snacBuf.Bytes())
	return err
}
