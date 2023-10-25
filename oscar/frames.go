package oscar

import (
	"bytes"
	"io"
)

type SnacError struct {
	Code uint16
	TLVRestBlock
}

type FlapFrame struct {
	StartMarker   uint8
	FrameType     uint8
	Sequence      uint16
	PayloadLength uint16
}

func (f FlapFrame) SNACBuffer(r io.Reader) (*bytes.Buffer, error) {
	b := make([]byte, f.PayloadLength)
	if _, err := r.Read(b); err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

type SnacFrame struct {
	FoodGroup uint16
	SubGroup  uint16
	Flags     uint16
	RequestID uint32
}

type FlapSignonFrame struct {
	FlapVersion uint32
	TLVRestBlock
}
