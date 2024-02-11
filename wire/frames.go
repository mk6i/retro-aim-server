package wire

import (
	"bytes"
	"io"
)

type SNACError struct {
	Code uint16
	TLVRestBlock
}

const (
	FLAPFrameSignon    uint8 = 0x01
	FLAPFrameData      uint8 = 0x02
	FLAPFrameError     uint8 = 0x03
	FLAPFrameSignoff   uint8 = 0x04
	FLAPFrameKeepAlive uint8 = 0x05
)

type FLAPFrame struct {
	StartMarker   uint8
	FrameType     uint8
	Sequence      uint16
	PayloadLength uint16
}

func (f FLAPFrame) ReadBody(r io.Reader) (*bytes.Buffer, error) {
	b := make([]byte, f.PayloadLength)
	if _, err := r.Read(b); err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

type SNACFrame struct {
	FoodGroup uint16
	SubGroup  uint16
	Flags     uint16
	RequestID uint32
}

type FLAPSignonFrame struct {
	FLAPVersion uint32
	TLVRestBlock
}

type SNACMessage struct {
	Frame SNACFrame
	Body  any
}
