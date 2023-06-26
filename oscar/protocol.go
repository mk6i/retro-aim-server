package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type routeHandler func(frame *flapFrame, rw io.ReadWriter) error

var familyRoute = map[uint16]routeHandler{
	0x01: routeOService,
}

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
