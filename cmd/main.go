package main

import (
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"log"
	"net"
	"os"
)

func main() {

	go listenBOS()

	//todo implement CHATNAV and ALERT

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

	err := oscar.SendAndReceiveSignonFrame(conn, 100)
	if err != nil {
		log.Println(err)
		return
	}

	err = oscar.ReceiveAndSendAuthChallenge(conn, conn, 101)
	if err != nil {
		log.Println(err)
		return
	}

	err = oscar.ReceiveAndSendBUCPLoginRequest(conn, conn, 102)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(conn net.Conn) {
	//defer conn.Close()
	fmt.Println("SendAndReceiveSignonFrame...")
	if err := oscar.SendAndReceiveSignonFrame(conn, 100); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("writeOServiceHostOnline...")
	if err := oscar.WriteOServiceHostOnline(conn, 101); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := oscar.ReadBos(conn, 102); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
