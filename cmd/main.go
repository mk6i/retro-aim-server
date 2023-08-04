package main

import (
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log"
	"net"
	"net/http"
)

const testFile string = "/Users/mike/dev/goaim/aim.db"

func main() {

	sm := oscar.NewSessionManager()
	fm, err := oscar.NewFeedbagStore(testFile)
	if err != nil {
		log.Fatal(err)
	}

	go listenBOS(sm, fm)

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
		go handleAuthConnection(sm, fm, conn)
	}
}

func webServer(ch chan string) {
	http.HandleFunc("/send-im", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ch <- string(body)
	})

	if err := http.ListenAndServe(":3333", nil); err != nil {
		panic(err.Error())
	}
}

func listenBOS(sm *oscar.SessionManager, fm *oscar.FeedbagStore) {
	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5191")
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan string, 1)
	go webServer(ch)

	defer listener.Close()

	fmt.Println("Server is listening on port 5191")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleBOSConnection(sm, fm, conn)
	}
}

func handleAuthConnection(sm *oscar.SessionManager, fm *oscar.FeedbagStore, conn net.Conn) {
	defer conn.Close()
	seq := uint32(100)
	_, err := oscar.SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}

	sess, err := sm.NewSession()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = oscar.ReceiveAndSendAuthChallenge(sess, conn, conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}

	err = oscar.ReceiveAndSendBUCPLoginRequest(sess, fm, conn, conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(sm *oscar.SessionManager, fm *oscar.FeedbagStore, conn net.Conn) {
	if err := oscar.ReadBos(sm, fm, conn); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Println(err.Error())
		}
	}
}
