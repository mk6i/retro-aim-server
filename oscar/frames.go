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

func (f FlapFrame) Write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, f.StartMarker); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.FrameType); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.Sequence); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, f.PayloadLength)
}

func (f *FlapFrame) Read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &f.StartMarker); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.FrameType); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.Sequence); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &f.PayloadLength)
}

type SnacFrame struct {
	FoodGroup uint16
	SubGroup  uint16
	Flags     uint16
	RequestID uint32
}

func (s SnacFrame) Write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.FoodGroup); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.SubGroup); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.Flags); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.RequestID); err != nil {
		return err
	}
	return nil
}

func (s *SnacFrame) Read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.FoodGroup); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.SubGroup); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.Flags); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &s.RequestID)
}

type FlapSignonFrame struct {
	FlapFrame
	FlapVersion uint32
	TLVRestBlock
}

func (f FlapSignonFrame) Write(w io.Writer) error {
	if err := f.FlapFrame.Write(w); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, f.FlapVersion)
}

func (f *FlapSignonFrame) Read(r io.Reader) error {
	if err := f.FlapFrame.Read(r); err != nil {
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
