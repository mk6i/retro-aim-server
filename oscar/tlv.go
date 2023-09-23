package oscar

import (
	"bytes"
	"encoding/binary"
	"io"
)

type TLVRestBlock struct {
	TLVList
}

// read consumes the remainder of the read buffer
func (s *TLVRestBlock) Read(r io.Reader) error {
	for {
		tlv := TLV{}
		if err := tlv.Read(r); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		s.TLVList = append(s.TLVList, tlv)
	}
	return nil
}

type TLVBlock struct {
	TLVList
}

// read consumes up to n TLVs, as specified in payload
func (s *TLVBlock) Read(r io.Reader) error {
	var tlvCount uint16
	if err := binary.Read(r, binary.BigEndian, &tlvCount); err != nil {
		return err
	}
	if tlvCount == 0 {
		return nil
	}
	for i := uint16(0); i < tlvCount; i++ {
		tlv := TLV{}
		if err := tlv.Read(r); err != nil {
			return err
		}
		s.TLVList = append(s.TLVList, tlv)
	}
	return nil
}

func (s TLVBlock) WriteTLV(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.TLVList))); err != nil {
		return err
	}
	return s.TLVList.WriteTLV(w)
}

type TLVLBlock struct {
	TLVList
}

// read consumes up to n bytes, as specified in the payload
func (s *TLVLBlock) Read(r io.Reader) error {
	var tlvLen uint16
	if err := binary.Read(r, binary.BigEndian, &tlvLen); err != nil {
		return err
	}
	p := make([]byte, tlvLen)
	if _, err := r.Read(p); err != nil {
		return err
	}
	buf := bytes.NewBuffer(p)
	for buf.Len() > 0 {
		tlv := TLV{}
		if err := tlv.Read(buf); err != nil {
			return err
		}
		s.TLVList = append(s.TLVList, tlv)
	}
	return nil
}

func (s TLVLBlock) WriteTLV(w io.Writer) error {
	buf := &bytes.Buffer{}
	if err := s.TLVList.WriteTLV(buf); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(buf.Len())); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

type TLVList []TLV

func (s *TLVList) AddTLV(tlv TLV) {
	*s = append(*s, tlv)
}

func (s TLVList) WriteTLV(w io.Writer) error {
	for _, tlv := range s {
		if err := tlv.WriteTLV(w); err != nil {
			return err
		}
	}
	return nil
}

func (s TLVList) GetString(tType uint16) (string, bool) {
	for _, tlv := range s {
		if tType == tlv.TType {
			return string(tlv.Val.([]byte)), true
		}
	}
	return "", false
}

func (s TLVList) GetTLV(tType uint16) (TLV, bool) {
	for _, tlv := range s {
		if tType == tlv.TType {
			return tlv, true
		}
	}
	return TLV{}, false
}

func (s TLVList) GetSlice(tType uint16) ([]byte, bool) {
	for _, tlv := range s {
		if tType == tlv.TType {
			return tlv.Val.([]byte), true
		}
	}
	return nil, false
}

func (s TLVList) GetUint32(tType uint16) (uint32, bool) {
	for _, tlv := range s {
		if tType == tlv.TType {
			return binary.BigEndian.Uint32(tlv.Val.([]byte)), true
		}
	}
	return 0, false
}

type TLV struct {
	TType uint16
	Val   any
}

func (t TLV) WriteTLV(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, t.TType); err != nil {
		return err
	}

	var valLen uint16
	val := t.Val

	switch t.Val.(type) {
	case uint8:
		valLen = 1
	case uint16:
		valLen = 2
	case uint32:
		valLen = 4
	case []uint16:
		valLen = uint16(len(t.Val.([]uint16)) * 2)
	case []byte:
		valLen = uint16(len(t.Val.([]byte)))
	case string:
		valLen = uint16(len(t.Val.(string)))
		val = []byte(t.Val.(string))
	default:
		buf := &bytes.Buffer{}
		if err := Marshal(t.Val, buf); err != nil {
			return err
		}
		valLen = uint16(buf.Len())
		val = buf.Bytes()
	}

	if err := binary.Write(w, binary.BigEndian, valLen); err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, val)
}

func (t *TLV) Read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &t.TType); err != nil {
		return err
	}
	var tlvValLen uint16
	if err := binary.Read(r, binary.BigEndian, &tlvValLen); err != nil {
		return err
	}
	buf := make([]byte, tlvValLen)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	t.Val = buf
	return nil
}
