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

	fm, err := oscar.NewFeedbagStore(testFile)
	if err != nil {
		log.Fatal(err)
	}

	sm := oscar.NewSessionManager()
	go listenBOS(sm, fm)

	chatSm := oscar.NewSessionManager()
	go listenChat(sm, chatSm, fm)

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
		go handleBOSConnection(sm, fm, conn)
	}
}

func listenChat(sm *oscar.SessionManager, chatSm *oscar.SessionManager, fm *oscar.FeedbagStore) {
	listener, err := net.Listen("tcp", ":5192")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 5192")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleChatConnection(sm, chatSm, fm, conn)
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
	defer conn.Close()

	sess, seq, err := oscar.VerifyLogin(sm, conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}

	defer oscar.Signout(sess, sm, fm)

	if err := oscar.ReadBos(sess, seq, sm, fm, conn); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}

func handleChatConnection(sm *oscar.SessionManager, chatSm *oscar.SessionManager, fm *oscar.FeedbagStore, conn net.Conn) {
	defer conn.Close()

	sess, seq, err := oscar.VerifyLogin(sm, conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}
	chatSess, err := chatSm.NewSessionWithSN(sess.ScreenName)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}
	defer chatSess.Close()
	defer chatSm.Remove(chatSess)

	if err := oscar.ReadBos(chatSess, seq, chatSm, fm, conn); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}
