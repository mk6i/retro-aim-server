package main

import (
	"bytes"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"log"
	"net"
	"os"
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

	if err := oscar.WriteFlapSignonFrame(conn); err != nil {
		log.Println(err)
		return
	}

	_, err := oscar.ReadFlapSignonFrame(conn)
	if err != nil {
		log.Println(err)
		return
	}

	err = oscar.ReceiveAndSendAuthChallenge(conn, 101)
	if err != nil {
		log.Println(err)
		return
	}

	err = oscar.ReceiveAndSendBUCPLoginRequest(conn, 102)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(conn net.Conn) {
	//defer conn.Close()

	if err := oscar.WriteFlapSignonFrame(conn); err != nil {
		log.Println(err)
		return
	}

	bufLen, err := oscar.ReadFlapSignonFrame(conn)
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

	if err := oscar.PrintTLV(bytes.NewBuffer(buf)); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("writeOServiceHostOnline...")
	if err := oscar.WriteOServiceHostOnline(conn, 101); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveAndSendHostVersions...")
	if err := oscar.ReceiveAndSendHostVersions(conn, 102); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveAndSendServiceRateParams...")
	if err := oscar.ReceiveAndSendServiceRateParams(conn, 103); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveAndSendServiceRequestSelfInfo...")
	if err := oscar.ReceiveAndSendServiceRequestSelfInfo(conn, 104); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("sendAndReceiveFeedbagRightsQuery...")
	if err := oscar.SendAndReceiveFeedbagRightsQuery(conn, 105); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("receiveAndSendFeedbagQuery...")
	if err := oscar.ReceiveAndSendFeedbagQuery(conn, 106); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//fmt.Println("sendAndReceiveLocateRights...")
	//if err := sendAndReceiveLocateRights(conn, 107); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}
	//
	//fmt.Println("sendAndReceiveBuddyRights...")
	//if err := sendAndReceiveBuddyRights(conn, 108); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}
	//
	//fmt.Println("sendAndReceiveICBMParameterReply...")
	//if err := sendAndReceiveICBMParameterReply(conn, 109); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}
	//
	//fmt.Println("sendAndReceivePDRightsQuery...")
	//if err := sendAndReceivePDRightsQuery(conn, 110); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}
	//
	//fmt.Println("sendAndReceiveNextChatRights...")
	//if err := sendAndReceiveNextChatRights(conn, 111); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}
	//
	//fmt.Println("sendAndReceiveNext...")
	//if err := sendAndReceiveNext(conn, 112); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}
}

//func SendAndReceiveNext(rw io.ReadWriter, sequence uint16) error {
//	// receive
//	flap := &flapFrame{}
//	if err := flap.read(rw); err != nil {
//		return err
//	}
//
//	fmt.Printf("sendAndReceiveNext read FLAP: %+v\n", flap)
//
//	b := make([]byte, flap.payloadLength)
//	if _, err := rw.Read(b); err != nil {
//		return err
//	}
//
//	snac := &snacFrame{}
//	if err := snac.read(bytes.NewBuffer(b)); err != nil {
//		return err
//	}
//	fmt.Printf("sendAndReceiveNext read SNAC: %+v\n", snac)
//
//	return nil
//}

//func SendAndReceiveNext(rw io.ReadWriter, sequence uint16) error {
//	// receive
//	flap := &flapFrame{}
//	if err := flap.read(rw); err != nil {
//		return err
//	}
//
//	fmt.Printf("sendAndReceiveNext read FLAP: %+v\n", flap)
//
//	b := make([]byte, flap.payloadLength)
//	if _, err := rw.Read(b); err != nil {
//		return err
//	}
//
//	snac := &snacFrame{}
//	if err := snac.read(bytes.NewBuffer(b)); err != nil {
//		return err
//	}
//	fmt.Printf("sendAndReceiveNext read SNAC: %+v\n", snac)
//
//	return nil
//}
