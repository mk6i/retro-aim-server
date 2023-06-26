package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"reflect"
)

type snac17_06 struct {
	snacFrame
	TLVs []*TLV
}

type snac17_07 struct {
	snacFrame
	authKey string
}

func (s *snac17_07) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.authKey))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(s.authKey)); err != nil {
		return err
	}
	return nil
}

func (s *snac17_06) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}

	lookup := map[uint16]reflect.Kind{
		0x01: reflect.String,
		0x4B: reflect.String,
		0x5A: reflect.String,
	}

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

func ReceiveAndSendAuthChallenge(rw net.Conn, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac17_06{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snac17_07{
		snacFrame: snacFrame{
			foodGroup: 0x17,
			subGroup:  0x07,
		},
		authKey: "theauthkey",
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("ReceiveAndSendAuthChallenge write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type snac17_02 struct {
	snacFrameTLV
}

func (s *snac17_02) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}

	lookup := map[uint16]reflect.Kind{
		0x01: reflect.String, // screen name
		0x03: reflect.String, // client ID string
		0x25: reflect.Slice,  // password md5 hash
		0x16: reflect.Uint16, // client ID
		0x17: reflect.Uint16, // client major version
		0x18: reflect.Uint16, // client minor version
		0x19: reflect.Uint16, // client lesser version
		0x1A: reflect.Uint16, // client build number
		0x14: reflect.Uint32, // distribution number
		0x0F: reflect.String, // client language
		0x0E: reflect.String, // client country
		0x4A: reflect.Slice,  // SSI use flag
		0x06: reflect.String, // SSI use flag
		0x4C: reflect.Slice,  // use old md5?
	}

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

func ReceiveAndSendBUCPLoginRequest(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendBUCPLoginRequest read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac17_02{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendBUCPLoginRequest read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snacFrameTLV{
		snacFrame: snacFrame{
			foodGroup: 0x17,
			subGroup:  0x03,
		},
		TLVs: []*TLV{
			{
				tType: 0x01,
				val:   "myscreenname",
			},
			{
				tType: 0x08,
				val:   uint16(0x00),
			},
			{
				tType: 0x04,
				val:   "",
			},
			{
				tType: 0x05,
				val:   "192.168.64.1:5191",
			},
			{
				tType: 0x06,
				val:   []byte("thecookie"),
			},
			{
				tType: 0x11,
				val:   "mike@localhost",
			},
			{
				tType: 0x54,
				val:   "http://localhost",
			},
		},
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("ReceiveAndSendAuthChallenge write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendAuthChallenge write SNAC: %+v\n", writeSnac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

func PrintTLV(r io.Reader) error {

	for {
		var tlvID uint16
		if err := binary.Read(r, binary.BigEndian, &tlvID); err != nil {
			return err
		}

		var tlvValLen uint16
		if err := binary.Read(r, binary.BigEndian, &tlvValLen); err != nil {
			return err
		}

		fmt.Printf("(%d) ", tlvID)
		switch tlvID {
		case 0x01: // screen name
			val, err := readString(r, tlvValLen)
			if err != nil {
				return err
			}
			fmt.Printf("screen name: %s", val)
		case 0x03: // client id string
			val, err := readString(r, tlvValLen)
			if err != nil {
				return err
			}
			fmt.Printf("client id string: %s", val)
		case 0x25: // password md5 hash
			val, err := readBytes(r, tlvValLen)
			if err != nil {
				return err
			}
			fmt.Printf("password md5 hash: %b\n", val)
		case 0x16: // client id
			val, err := readUint16(r)
			if err != nil {
				return err
			}
			fmt.Printf("client id (len=%d): %d", tlvValLen, val)
		case 0x17: // client major version
			val, err := readUint16(r)
			if err != nil {
				return err
			}
			fmt.Printf("client major version (len=%d): %d", tlvValLen, val)
		case 0x18: // client minor version
			val, err := readUint16(r)
			if err != nil {
				return err
			}
			fmt.Printf("client minor version (len=%d): %d", tlvValLen, val)
		case 0x19: // client lesser version
			val, err := readUint16(r)
			if err != nil {
				return err
			}
			fmt.Printf("client lesser version (len=%d): %d", tlvValLen, val)
		case 0x1A: // client build number
			val, err := readUint16(r)
			if err != nil {
				return err
			}
			fmt.Printf("client build number (len=%d): %d", tlvValLen, val)
		case 0x14: // distribution number
			val, err := readUint32(r)
			if err != nil {
				return err
			}
			fmt.Printf("distribution number (len=%d): %d", tlvValLen, val)
		case 0x0F: // client language (2 symbols)
			val, err := readString(r, tlvValLen)
			if err != nil {
				return err
			}
			fmt.Printf("client language (2 symbols) (len=%d): %s", tlvValLen, val)
		case 0x0E: // client country (2 symbols)
			val, err := readString(r, tlvValLen)
			if err != nil {
				return err
			}
			fmt.Printf("client country (2 symbols) (len=%d): %s", tlvValLen, val)
		case 0x4A: // SSI use flag
			val, err := readBytes(r, tlvValLen)
			if err != nil {
				return err
			}
			// buddy list thing?
			fmt.Printf("SSI use flag (len=%d): %d", tlvValLen, val[0])
			return nil
		case 0x004c:
			fmt.Printf("Use old MD5?\n")
		case 0x06:
			val, err := readString(r, tlvValLen)
			if err != nil {
				return err
			}
			fmt.Printf("login cookie (len=%d): %s\n", tlvValLen, val)
			for {
				buf := make([]byte, 1)
				_, err := r.Read(buf)
				if err != nil {
					return nil
				}
				fmt.Println(buf)
			}
		default:
			fmt.Printf("unknown TLV: %d (len=%d)", tlvID, tlvValLen)
			_, err := r.Read(make([]byte, tlvValLen))
			if err != nil {
				return err
			}
		}

		fmt.Println()
	}
}

func readString(r io.Reader, len uint16) (string, error) {
	buf := make([]byte, len)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readBytes(r io.Reader, len uint16) ([]byte, error) {
	buf := make([]byte, len)
	if _, err := r.Read(buf); err != nil {
		return buf, err
	}
	return buf, nil
}

func readUint16(r io.Reader) (uint16, error) {
	var val uint16
	binary.Read(r, binary.BigEndian, &val)
	return val, nil
}

func readUint32(r io.Reader) (uint32, error) {
	var val uint32
	binary.Read(r, binary.BigEndian, &val)
	return val, nil
}
