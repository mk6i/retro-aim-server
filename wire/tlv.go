package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// TLV represents dynamically typed data in the OSCAR protocol. Each message
// consists of a tag (or key) and a blob value. TLVs are typically grouped
// together in arrays.
type TLV struct {
	Tag   uint16
	Value []byte `oscar:"len_prefix=uint16"`
}

// NewTLV creates a new instance of TLV.
func NewTLV(tag uint16, val any) TLV {
	t := TLV{
		Tag: tag,
	}
	if _, ok := val.([]byte); ok {
		t.Value = val.([]byte)
	} else {
		buf := &bytes.Buffer{}
		if err := Marshal(val, buf); err != nil {
			panic(fmt.Sprintf("unable to create TLV: %s", err.Error()))
		}
		t.Value = buf.Bytes()
	}
	return t
}

// TLVRestBlock is a type of TLV array that does not have any length
// information encoded in the blob. This typically means that a given offset in
// the SNAC payload, the TLV occupies the "rest" of the payload.
type TLVRestBlock struct {
	TLVList
}

// TLVBlock is a type of TLV array that has the TLV element count encoded as a
// 2-byte value at the beginning of the encoded blob.
type TLVBlock struct {
	TLVList `oscar:"count_prefix=uint16"`
}

// TLVLBlock is a type of TLV array that has the TLV blob byte-length encoded
// as a 2-byte value at the beginning of the encoded blob.
type TLVLBlock struct {
	TLVList `oscar:"len_prefix=uint16"`
}

// TLVList is a list of TLV elements. It provides methods to append and access
// TLVs in the array. It provides methods that decode the data blob into the
// appropriate type at runtime. The caller assumes the TLV data type at runtime
// based on the protocol specification. These methods are not safe for
// read-write access by  multiple goroutines.
type TLVList []TLV

// Append adds a TLV to the end of the TLV list.
func (s *TLVList) Append(tlv TLV) {
	*s = append(*s, tlv)
}

// AppendList adds a TLV list to the end of the TLV list.
func (s *TLVList) AppendList(tlvs []TLV) {
	*s = append(*s, tlvs...)
}

// String retrieves the string value of a TLV with a tag value from the TLV
// list. It returns false if the tag does not exist in the list.
func (s *TLVList) String(tag uint16) (string, bool) {
	for _, tlv := range *s {
		if tag == tlv.Tag {
			return string(tlv.Value), true
		}
	}
	return "", false
}

// Slice retrieves the slice value of a TLV with a tag value from the TLV
// list. It returns false if the tag does not exist in the list.
func (s *TLVList) Slice(tag uint16) ([]byte, bool) {
	for _, tlv := range *s {
		if tag == tlv.Tag {
			return tlv.Value, true
		}
	}
	return nil, false
}

// Uint16 retrieves the uint16 value of a TLV with a tag value from the TLV
// list. It returns false if the tag does not exist in the list. It may panic
// if the TLV value is not uint16.
func (s *TLVList) Uint16(tag uint16) (uint16, bool) {
	for _, tlv := range *s {
		if tag == tlv.Tag {
			return binary.BigEndian.Uint16(tlv.Value), true
		}
	}
	return 0, false
}

// Uint32 retrieves the uint32 value of a TLV with a tag value from the TLV
// list. It returns false if the tag does not exist in the list. It may panic
// if the TLV value is not uint32.
func (s *TLVList) Uint32(tag uint16) (uint32, bool) {
	for _, tlv := range *s {
		if tag == tlv.Tag {
			return binary.BigEndian.Uint32(tlv.Value), true
		}
	}
	return 0, false
}
