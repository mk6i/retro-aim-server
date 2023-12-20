package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	TLVScreenName          uint16 = 0x01
	TLVReconnectHere       uint16 = 0x05
	TLVAuthorizationCookie uint16 = 0x06
	TLVErrorSubcode        uint16 = 0x08
	TLVPasswordHash        uint16 = 0x25
)

type TLVRestBlock struct {
	TLVList
}

type TLVBlock struct {
	TLVList `count_prefix:"uint16"`
}

type TLVLBlock struct {
	TLVList `len_prefix:"uint16"`
}

type TLVList []TLV

func (s *TLVList) AddTLV(tlv TLV) {
	*s = append(*s, tlv)
}

func (s *TLVList) AddTLVList(tlvs []TLV) {
	*s = append(*s, tlvs...)
}

func (s TLVList) GetString(tType uint16) (string, bool) {
	for _, tlv := range s {
		if tType == tlv.TType {
			return string(tlv.Val), true
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
			return tlv.Val, true
		}
	}
	return nil, false
}

func (s TLVList) GetUint16(tType uint16) (uint16, bool) {
	for _, tlv := range s {
		if tType == tlv.TType {
			return binary.BigEndian.Uint16(tlv.Val), true
		}
	}
	return 0, false
}

func (s TLVList) GetUint32(tType uint16) (uint32, bool) {
	for _, tlv := range s {
		if tType == tlv.TType {
			return binary.BigEndian.Uint32(tlv.Val), true
		}
	}
	return 0, false
}

func NewTLV(ttype uint16, val any) TLV {
	t := TLV{
		TType: ttype,
	}
	if _, ok := val.([]byte); ok {
		t.Val = val.([]byte)
	} else {
		buf := &bytes.Buffer{}
		if err := Marshal(val, buf); err != nil {
			panic(fmt.Sprintf("unable to create TLV: %s", err.Error()))
		}
		t.Val = buf.Bytes()
	}
	return t
}

type TLV struct {
	TType uint16
	Val   []byte `len_prefix:"uint16"`
}
