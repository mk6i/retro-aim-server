package oscar

import (
	"bytes"
	"encoding/binary"
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

type SnacFrame struct {
	FoodGroup uint16
	SubGroup  uint16
	Flags     uint16
	RequestID uint32
}

type FlapSignonFrame struct {
	FlapFrame
	FlapVersion uint32
	TLVRestBlock
}

func (f FlapSignonFrame) Write(w io.Writer) error {
	if err := Marshal(f.FlapFrame, w); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, f.FlapVersion)
}

func (f *FlapSignonFrame) Read(r io.Reader) error {
	if err := Unmarshal(&f.FlapFrame, r); err != nil {
		return err
	}

	// todo: combine b+buf?
	b := make([]byte, f.PayloadLength)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.BigEndian, &f.FlapVersion); err != nil {
		return err
	}

	return f.TLVRestBlock.Read(buf)
}
