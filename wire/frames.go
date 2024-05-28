package wire

import (
	"bytes"
	"fmt"
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
	if f.PayloadLength > 0 {
		if _, err := io.ReadFull(r, b); err != nil {
			return nil, err
		}
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

// NewFlapClient creates a new FLAP client instance. startSeq is the initial
// sequence value, which is typically 0. r receives FLAP messages, w writes
// FLAP messages.
func NewFlapClient(startSeq uint32, r io.Reader, w io.Writer) *FlapClient {
	return &FlapClient{
		sequence: startSeq,
		r:        r,
		w:        w,
	}
}

// FlapClient sends and receive FLAP frames to and from the server. It ensures
// that the message sequence numbers are properly incremented after sending
// each successive message. It is not safe to use with multiple goroutines
// without synchronization.
type FlapClient struct {
	sequence uint32
	r        io.Reader
	w        io.Writer
}

// SendSignonFrame sends a signon FLAP frame containing a list of TLVs to
// authenticate or initiate a session.
func (f *FlapClient) SendSignonFrame(tlvs []TLV) error {
	signonFrame := FLAPSignonFrame{
		FLAPVersion: 1,
	}
	if len(tlvs) > 0 {
		signonFrame.AppendList(tlvs)
	}
	buf := &bytes.Buffer{}
	if err := Marshal(signonFrame, buf); err != nil {
		return err
	}

	flap := FLAPFrame{
		StartMarker:   42,
		FrameType:     FLAPFrameSignon,
		Sequence:      uint16(f.sequence),
		PayloadLength: uint16(buf.Len()),
	}
	if err := Marshal(flap, f.w); err != nil {
		return err
	}

	if _, err := f.w.Write(buf.Bytes()); err != nil {
		return err
	}

	f.sequence++

	return nil
}

// ReceiveSignonFrame receives a signon FLAP response message.
func (f *FlapClient) ReceiveSignonFrame() (FLAPSignonFrame, error) {
	flap := FLAPFrame{}
	if err := Unmarshal(&flap, f.r); err != nil {
		return FLAPSignonFrame{}, err
	}

	buf, err := flap.ReadBody(f.r)
	if err != nil {
		return FLAPSignonFrame{}, err
	}

	signonFrame := FLAPSignonFrame{}
	if err := Unmarshal(&signonFrame, buf); err != nil {
		return FLAPSignonFrame{}, err
	}

	return signonFrame, nil
}

// ReceiveFLAP receives a FLAP frame and body. It only returns a body if the
// FLAP frame is a data frame.
func (f *FlapClient) ReceiveFLAP() (FLAPFrame, *bytes.Buffer, error) {
	flap := FLAPFrame{}
	if err := Unmarshal(&flap, f.r); err != nil {
		return flap, nil, fmt.Errorf("unable to unmarshal FLAP frame: %w", err)
	}

	if flap.FrameType != FLAPFrameData {
		return flap, nil, nil
	}

	buf, err := flap.ReadBody(f.r)
	if err != nil {
		err = fmt.Errorf("unable to read FLAP body: %w", err)
	}
	return flap, buf, err
}

// SendSignoffFrame sends a sign-off FLAP frame with attached TLVs as the last
// request sent in the FLAP auth flow. This is unrelated to the Disconnect()
// method, which sends a sign-off frame to terminate a BOS connection.
// todo: combine this method with Disconnect()
func (f *FlapClient) SendSignoffFrame(tlvs TLVRestBlock) error {
	tlvBuf := &bytes.Buffer{}
	if err := Marshal(tlvs, tlvBuf); err != nil {
		return err
	}

	flap := FLAPFrame{
		StartMarker:   42,
		FrameType:     FLAPFrameSignoff,
		Sequence:      uint16(f.sequence),
		PayloadLength: uint16(tlvBuf.Len()),
	}

	if err := Marshal(flap, f.w); err != nil {
		return err
	}

	expectLen := tlvBuf.Len()
	c, err := f.w.Write(tlvBuf.Bytes())
	if err != nil {
		return err
	}
	if c != expectLen {
		panic("did not write the expected # of bytes")
	}

	f.sequence++
	return nil
}

// SendSNAC sends a SNAC message wrapped in a FLAP frame.
func (f *FlapClient) SendSNAC(frame SNACFrame, body any) error {
	snacBuf := &bytes.Buffer{}
	if err := Marshal(frame, snacBuf); err != nil {
		return err
	}
	if err := Marshal(body, snacBuf); err != nil {
		return err
	}

	flap := FLAPFrame{
		StartMarker:   42,
		FrameType:     FLAPFrameData,
		Sequence:      uint16(f.sequence),
		PayloadLength: uint16(snacBuf.Len()),
	}
	if err := Marshal(flap, f.w); err != nil {
		return err
	}

	if _, err := f.w.Write(snacBuf.Bytes()); err != nil {
		return err
	}

	f.sequence++
	return nil
}

// ReceiveSNAC receives a SNAC message wrapped in a FLAP frame.
func (f *FlapClient) ReceiveSNAC(frame *SNACFrame, body any) error {
	flap := FLAPFrame{}
	if err := Unmarshal(&flap, f.r); err != nil {
		return err
	}
	buf, err := flap.ReadBody(f.r)
	if err != nil {
		return err
	}
	if err := Unmarshal(frame, buf); err != nil {
		return err
	}
	return Unmarshal(body, buf)
}

// Disconnect sends a signoff FLAP frame.
func (f *FlapClient) Disconnect() error {
	// gracefully disconnect so that the client does not try to
	// reconnect when the connection closes.
	flap := FLAPFrame{
		StartMarker:   42,
		FrameType:     FLAPFrameSignoff,
		Sequence:      uint16(f.sequence),
		PayloadLength: uint16(0),
	}
	return Marshal(flap, f.w)
}
