package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"time"
)

func main() {

	go listenBOS()

	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5190")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 5190")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Handle connection in a separate goroutine
		go handleAuthConnection(conn)
	}
}

func listenBOS() {
	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5191")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 5191")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleBOSConnection(conn)
	}
}

func handleAuthConnection(conn net.Conn) {
	defer conn.Close()

	if err := writeFlapSignonFrame(conn); err != nil {
		log.Println(err)
		return
	}

	_, err := ReadFlapSignonFrame(conn)
	if err != nil {
		log.Println(err)
		return
	}

	sequenceNumber, err := ReadAuthChallengeRequest(conn)
	if err != nil {
		log.Println(err)
		return
	}

	err = WriteAuthChallengeResponse(conn, sequenceNumber+1)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = ReadBUCPLoginRequest(conn)
	if err != nil {
		log.Println(err)
		return
	}

	err = WriteBUCPLoginResponse(conn, sequenceNumber+1)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(conn net.Conn) {
	//defer conn.Close()

	if err := writeFlapSignonFrame(conn); err != nil {
		log.Println(err)
		return
	}

	bufLen, err := ReadFlapSignonFrame(conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Reading BOS login request...")

	buf := make([]byte, bufLen)

	_, err = conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}

	if err := printTLV(bytes.NewBuffer(buf)); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("writeOServiceHostOnline...")
	if err := writeOServiceHostOnline(conn, 101); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveAndSendHostVersions...")
	if err := receiveAndSendHostVersions(conn, 102); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveAndSendServiceRateParams...")
	if err := receiveAndSendServiceRateParams(conn, 103); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveAndSendServiceRequestSelfInfo...")
	if err := receiveAndSendServiceRequestSelfInfo(conn, 104); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveFeedbagRightsQuery...")
	if err := receiveFeedbagRightsQuery(conn); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func writeFlapSignonFrame(conn net.Conn) error {

	startMarker := uint8(42)
	if err := binary.Write(conn, binary.BigEndian, startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	frameType := uint8(1)
	if err := binary.Write(conn, binary.BigEndian, frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	sequenceNumber := uint16(100)
	if err := binary.Write(conn, binary.BigEndian, sequenceNumber); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	payloadLength := uint16(4)
	if err := binary.Write(conn, binary.BigEndian, payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	flapVersion := uint32(1)
	if err := binary.Write(conn, binary.BigEndian, flapVersion); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return nil
}

func ReadFlapSignonFrame(conn net.Conn) (uint16, error) {

	var startMarker uint8
	if err := binary.Read(conn, binary.BigEndian, &startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("start marker: %d\n", startMarker)

	var frameType uint8
	if err := binary.Read(conn, binary.BigEndian, &frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("frame type: %d\n", frameType)

	var sequenceNumber uint16
	if err := binary.Read(conn, binary.BigEndian, &sequenceNumber); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("sequence number: %d\n", sequenceNumber)

	var payloadLength uint16
	if err := binary.Read(conn, binary.BigEndian, &payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("payload length: %d\n", payloadLength)

	var flapVersion uint32
	if err := binary.Read(conn, binary.BigEndian, &flapVersion); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("flap version: %d\n", flapVersion)

	return payloadLength, nil
}

func ReadAuthChallengeRequest(conn net.Conn) (uint16, error) {

	fmt.Println("Reading snac...")

	var startMarker uint8
	if err := binary.Read(conn, binary.BigEndian, &startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("start marker: %d\n", startMarker)

	var frameType uint8
	if err := binary.Read(conn, binary.BigEndian, &frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("frame type: %d\n", frameType)

	var sequenceNumber uint16
	if err := binary.Read(conn, binary.BigEndian, &sequenceNumber); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("sequence number: %d\n", sequenceNumber)

	var payloadLength uint16
	if err := binary.Read(conn, binary.BigEndian, &payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("payload length: %d\n", payloadLength)

	remainder := make([]byte, payloadLength)
	if _, err := conn.Read(remainder); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	snacBuf := bytes.NewBuffer(remainder)

	fmt.Println("Reading Snac header...")

	var foodGroup uint16
	if err := binary.Read(snacBuf, binary.BigEndian, &foodGroup); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("food group: %d\n", foodGroup)
	// 23 = 0x17 = https://wiki.nina.chat/wiki/Protocols/OSCAR#Foodgroups BUCP (0x0017)
	var foodGroupType uint16
	if err := binary.Read(snacBuf, binary.BigEndian, &foodGroupType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("foodGroupType: %d\n", foodGroupType)
	// 6 = 0x0006 = BUCP__CHALLENGE_REQUEST
	var flags uint16
	if err := binary.Read(snacBuf, binary.BigEndian, &flags); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("flags: %d\n", flags)

	var requestID uint32
	if err := binary.Read(snacBuf, binary.BigEndian, &requestID); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("requestID: %d\n", requestID)

	var tag uint16
	if err := binary.Read(snacBuf, binary.BigEndian, &tag); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("tag: %d\n", tag)

	var screenNameLen uint16
	if err := binary.Read(snacBuf, binary.BigEndian, &screenNameLen); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("screenNameLen: %d\n", screenNameLen)

	screenNameBuf := make([]byte, screenNameLen)
	if _, err := snacBuf.Read(screenNameBuf); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("screen name: %s\n", screenNameBuf)

	return sequenceNumber, nil
}

func WriteAuthChallengeResponse(conn net.Conn, sequenceNumber uint16) error {
	fmt.Println("Writing auth challenge response...")

	startMarker := uint8(42)
	if err := binary.Write(conn, binary.BigEndian, startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	frameType := uint8(2)
	if err := binary.Write(conn, binary.BigEndian, frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	seq := uint16(101)
	if err := binary.Write(conn, binary.BigEndian, seq); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	b := make([]byte, 0)
	snacBuf := bytes.NewBuffer(b)

	{
		foodGroup := uint16(0x17)
		if err := binary.Write(snacBuf, binary.BigEndian, foodGroup); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		foodGroupType := uint16(7)
		if err := binary.Write(snacBuf, binary.BigEndian, foodGroupType); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		flags := uint16(0x00)
		if err := binary.Write(snacBuf, binary.BigEndian, flags); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		requestID := uint32(0x00)
		if err := binary.Write(snacBuf, binary.BigEndian, requestID); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		authKey := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5}
		authKeyLen := uint32(len(authKey))
		if err := binary.Write(snacBuf, binary.BigEndian, authKeyLen); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if _, err := snacBuf.Write(authKey); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	payloadLength := uint16(snacBuf.Len())
	if err := binary.Write(conn, binary.BigEndian, payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if _, err := conn.Write(snacBuf.Bytes()); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return nil
}

func ReadBUCPLoginRequest(conn net.Conn) (uint16, error) {

	fmt.Println("Reading BUCP login request...")

	var startMarker uint8
	if err := binary.Read(conn, binary.BigEndian, &startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("start marker: %d\n", startMarker)

	var frameType uint8
	if err := binary.Read(conn, binary.BigEndian, &frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("frame type: %d\n", frameType)

	var sequenceNumber uint16
	if err := binary.Read(conn, binary.BigEndian, &sequenceNumber); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("sequence number: %d\n", sequenceNumber)

	var payloadLength uint16
	if err := binary.Read(conn, binary.BigEndian, &payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("payload length: %d\n", payloadLength)

	fmt.Println("Reading Snac header...")

	var foodGroup uint16
	if err := binary.Read(conn, binary.BigEndian, &foodGroup); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("food group: %d\n", foodGroup)
	// 23 = 0x17 = https://wiki.nina.chat/wiki/Protocols/OSCAR#Foodgroups BUCP (0x0017)
	var foodGroupType uint16
	if err := binary.Read(conn, binary.BigEndian, &foodGroupType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("foodGroupType: %d\n", foodGroupType)
	// 6 = 0x0006 = BUCP__CHALLENGE_REQUEST
	var flags uint16
	if err := binary.Read(conn, binary.BigEndian, &flags); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("flags: %d\n", flags)

	var requestID uint32
	if err := binary.Read(conn, binary.BigEndian, &requestID); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("requestID: %d\n", requestID)

	err := printBUCPTLV(conn)
	if err != nil && err != io.EOF {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return sequenceNumber, nil
}

func WriteBUCPLoginResponse(conn net.Conn, sequenceNumber uint16) error {
	fmt.Println("Writing bucp login response...")

	startMarker := uint8(42)
	if err := binary.Write(conn, binary.BigEndian, startMarker); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	frameType := uint8(2)
	if err := binary.Write(conn, binary.BigEndian, frameType); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	seq := uint16(102)
	if err := binary.Write(conn, binary.BigEndian, seq); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	b := make([]byte, 0)
	snacBuf := bytes.NewBuffer(b)

	{
		foodGroup := uint16(0x17)
		if err := binary.Write(snacBuf, binary.BigEndian, foodGroup); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		foodGroupType := uint16(0x03)
		if err := binary.Write(snacBuf, binary.BigEndian, foodGroupType); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		flags := uint16(0x00)
		if err := binary.Write(snacBuf, binary.BigEndian, flags); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		requestID := uint32(0x00)
		if err := binary.Write(snacBuf, binary.BigEndian, requestID); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if err := writeTLV(snacBuf, 0x01, "myscreenname"); err != nil {
			return err
		}

		if err := writeTLV(snacBuf, 0x08, uint16(0x00)); err != nil {
			return err
		}

		if err := writeTLV(snacBuf, 0x04, ""); err != nil {
			return err
		}

		if err := writeTLV(snacBuf, 0x05, "192.168.64.1:5191"); err != nil {
			return err
		}

		if err := writeTLV(snacBuf, 0x06, []byte("thecookie")); err != nil {
			return err
		}

		if err := writeTLV(snacBuf, 0x11, "mike@localhost"); err != nil {
			return err
		}

		if err := writeTLV(snacBuf, 0x54, "http://localhost"); err != nil {
			return err
		}
	}

	payloadLength := uint16(snacBuf.Len())
	if err := binary.Write(conn, binary.BigEndian, payloadLength); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if _, err := conn.Write(snacBuf.Bytes()); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return nil
}

func writeTLV(w io.Writer, tlvID uint16, val any) error {
	if err := binary.Write(w, binary.BigEndian, tlvID); err != nil {
		return err
	}
	switch val := val.(type) {
	case uint16:
		tlvValLen := uint16(2)
		if err := binary.Write(w, binary.BigEndian, tlvValLen); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, val); err != nil {
			return err
		}
	case uint32:
		tlvValLen := uint16(4)
		if err := binary.Write(w, binary.BigEndian, tlvValLen); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, val); err != nil {
			return err
		}
	case string:
		tlvValLen := uint16(len(val))
		if err := binary.Write(w, binary.BigEndian, tlvValLen); err != nil {
			return err
		}
		_, err := w.Write([]byte(val))
		if err != nil {
			return err
		}
	case []byte:
		tlvValLen := uint16(len(val))
		if err := binary.Write(w, binary.BigEndian, tlvValLen); err != nil {
			return err
		}
		_, err := w.Write(val)
		if err != nil {
			return err
		}
	}
	return nil
}

func printBUCPTLV(r io.Reader) error {

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
			fmt.Printf("login cookie (len=%d): %s", tlvValLen, val)
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

func printTLV(r io.Reader) error {

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

type snac01_03 struct {
	snacFrame
	foodGroups []uint16
}

func (s *snac01_03) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.foodGroups); err != nil {
		return err
	}
	return nil
}

func writeOServiceHostOnline(conn net.Conn, sequence uint16) error {

	snac := &snac01_03{
		snacFrame: snacFrame{
			foodGroup: 0x01,
			subGroup:  0x03,
		},
		foodGroups: []uint16{
			0x0001, 0x0002, 0x0003, 0x0004, 0x0005, 0x0006, 0x0007, 0x0008, 0x0009,
			0x000A, 0x000B, 0x000C, 0x000D, 0x000E, 0x000F, 0x0010, 0x0013, 0x0015,
			0x0017, 0x0018, 0x0022, 0x0024, 0x0025, 0x044A,
		},
	}

	fmt.Printf("writeOServiceHostOnline SNAC: %+v\n", snac)

	snacBuf := &bytes.Buffer{}
	if err := snac.write(snacBuf); err != nil {
		return err
	}

	flap := &flapFrame{
		startMarker:   42,
		frameType:     2,
		sequence:      sequence,
		payloadLength: uint16(snacBuf.Len()),
	}

	fmt.Printf("writeOServiceHostOnline FLAP: %+v\n", flap)

	if err := flap.write(conn); err != nil {
		return err
	}

	_, err := conn.Write(snacBuf.Bytes())
	return err
}

type snac01_17_18 struct {
	snacFrame
	versions map[uint16]uint16
}

func (s *snac01_17_18) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}
	for {
		var family uint16
		if err := binary.Read(r, binary.BigEndian, &family); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		var version uint16
		if err := binary.Read(r, binary.BigEndian, &version); err != nil {
			return err
		}
		s.versions[family] = version
	}
	return nil
}

func (s *snac01_17_18) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	for family, version := range s.versions {
		if err := binary.Write(w, binary.BigEndian, family); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, version); err != nil {
			return err
		}
	}
	return nil
}

func receiveAndSendHostVersions(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac01_17_18{
		versions: make(map[uint16]uint16),
	}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read SNAC: %+v\n", snac)

	// respond
	snac.snacFrame.subGroup = 0x18

	snacBuf := &bytes.Buffer{}
	if err := snac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("receiveAndSendHostVersions write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions write SNAC: %+v\n", snac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type rateClass struct {
	ID              uint16
	windowSize      uint32
	clearLevel      uint32
	alertLevel      uint32
	limitLevel      uint32
	disconnectLevel uint32
	currentLevel    uint32
	maxLevel        uint32
	lastTime        uint32 // protocol v2 only
	currentState    byte   // protocol v2 only
}

type rateGroup struct {
	ID    uint16
	pairs []struct {
		foodGroup uint16
		subGroup  uint16
	}
}

type snac01_07 struct {
	snacFrame
	rateClasses []rateClass
	rateGroups  []rateGroup
}

func (s *snac01_07) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.rateClasses))); err != nil {
		return err
	}
	for _, rateClass := range s.rateClasses {
		if err := binary.Write(w, binary.BigEndian, rateClass.ID); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.windowSize); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.clearLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.alertLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.limitLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.disconnectLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.currentLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.maxLevel); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.lastTime); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, rateClass.currentState); err != nil {
			return err
		}
	}
	for _, rateGroup := range s.rateGroups {
		if err := binary.Write(w, binary.BigEndian, rateGroup.ID); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint16(len(rateGroup.pairs))); err != nil {
			return err
		}
		for _, pair := range rateGroup.pairs {
			if err := binary.Write(w, binary.BigEndian, pair.foodGroup); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, pair.subGroup); err != nil {
				return err
			}
		}
	}

	return nil
}

func receiveAndSendServiceRateParams(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRateParams read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	incomingSnac := &snacFrame{}
	if err := incomingSnac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRateParams read SNAC: %+v\n", incomingSnac)

	// respond
	snac := &snac01_07{
		snacFrame: snacFrame{
			foodGroup: 0x01,
			subGroup:  0x07,
		},
		rateClasses: []rateClass{
			//{
			//	ID:              1,
			//	windowSize:      10,
			//	clearLevel:      10,
			//	alertLevel:      10,
			//	limitLevel:      10,
			//	disconnectLevel: 10,
			//	currentLevel:    10,
			//	maxLevel:        10,
			//	lastTime:        10,
			//	currentState:    10,
			//},
		},
		rateGroups: []rateGroup{
			//{
			//	ID: 1,
			//	pairs: []struct {
			//		foodGroup uint16
			//		subGroup  uint16
			//	}{
			//		{
			//			foodGroup: 1,
			//			subGroup:  1,
			//		},
			//	},
			//},
		},
	}

	snacBuf := &bytes.Buffer{}
	if err := snac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("receiveAndSendServiceRateParams write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRateParams write SNAC: %+v\n", snac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type snac01_08 struct {
	snacFrame
	subs []uint16
}

func (s *snac01_08) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}
	for {
		var rateClass uint16
		if err := binary.Read(r, binary.BigEndian, &rateClass); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		s.subs = append(s.subs, rateClass)
	}
	return nil
}

type TLV struct {
	tType uint16
	val   any
}

func (t *TLV) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, t.tType); err != nil {
		return err
	}

	var valLen uint16

	switch t.val.(type) {
	case uint8:
		valLen = 1
	case uint16:
		valLen = 2
	case uint32:
		valLen = 4
	case []byte:
		valLen = uint16(len(t.val.([]byte)))
	}

	if err := binary.Write(w, binary.BigEndian, valLen); err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, t.val)
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
	case reflect.Uint16:
		var val uint16
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return err
		}
		t.val = val
	default:
		panic("unsupported data type")
	}

	return nil
}

type snac01_0F struct {
	snacFrame
	screenName   string
	warningLevel uint16
	TLVs         []*TLV
}

func (s *snac01_0F) write(w io.Writer) error {
	if err := s.snacFrame.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint8(len(s.screenName))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(s.screenName)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.warningLevel); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.TLVs))); err != nil {
		return err
	}
	for _, t := range s.TLVs {
		if err := t.write(w); err != nil {
			return err
		}
	}
	return nil
}

func receiveAndSendServiceRequestSelfInfo(rw io.ReadWriter, sequence uint16) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRequestSelfInfo read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snacFrame{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("receiveAndSendServiceRequestSelfInfo read SNAC: %+v\n", snac)

	// respond
	writeSnac := &snac01_0F{
		snacFrame: snacFrame{
			foodGroup: 0x01,
			subGroup:  0x0F,
		},
		screenName:   "screenname",
		warningLevel: 0,
		TLVs: []*TLV{
			{
				tType: 0x01,
				val:   uint32(0x0010),
			},
			{
				tType: 0x02,
				val:   uint32(time.Now().Unix()),
			},
			{
				tType: 0x03,
				val:   uint32(1687314861),
			},
			{
				tType: 0x04,
				val:   uint32(0),
			},
			{
				tType: 0x05,
				val:   uint32(1687314841),
			},
			{
				tType: 0x0D,
				val:   make([]byte, 0),
			},
			{
				tType: 0x0F,
				val:   uint32(0),
			},
		},
	}

	snacBuf := &bytes.Buffer{}
	if err := writeSnac.write(snacBuf); err != nil {
		return err
	}

	flap.sequence = sequence
	flap.payloadLength = uint16(snacBuf.Len())

	fmt.Printf("receiveAndSendServiceRequestSelfInfo write FLAP: %+v\n", flap)

	if err := flap.write(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRequestSelfInfo write SNAC: %+v\n", snac)

	_, err := rw.Write(snacBuf.Bytes())
	return err
}

type snac13_02 struct {
	snacFrame
	TLVs []*TLV
}

func (s *snac13_02) read(r io.Reader) error {
	if err := s.snacFrame.read(r); err != nil {
		return err
	}

	lookup := map[uint16]reflect.Kind{0x0B: reflect.Uint16}

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

func receiveFeedbagRightsQuery(rw io.ReadWriter) error {
	// receive
	flap := &flapFrame{}
	if err := flap.read(rw); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendServiceRequestSelfInfo read FLAP: %+v\n", flap)

	b := make([]byte, flap.payloadLength)
	if _, err := rw.Read(b); err != nil {
		return err
	}

	snac := &snac13_02{}
	if err := snac.read(bytes.NewBuffer(b)); err != nil {
		return err
	}
	fmt.Printf("receiveAndSendServiceRequestSelfInfo read SNAC: %+v\n", snac)

	return nil
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
