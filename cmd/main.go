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
	cr := oscar.NewChatRegistry()

	go listenBOS(sm, fm, cr)
	go listenChat(fm, cr)

	listener, err := net.Listen("tcp", ":5190")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 5190")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

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

func listenBOS(sm *oscar.SessionManager, fm *oscar.FeedbagStore, cr *oscar.ChatRegistry) {
	listener, err := net.Listen("tcp", ":5191")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 5191")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleBOSConnection(sm, fm, cr, conn)
	}
}

func listenChat(fm *oscar.FeedbagStore, cr *oscar.ChatRegistry) {
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
		go handleChatConnection(fm, cr, conn)
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

func handleBOSConnection(sm *oscar.SessionManager, fm *oscar.FeedbagStore, cr *oscar.ChatRegistry, conn net.Conn) {
	defer conn.Close()

	sess, seq, err := oscar.VerifyLogin(sm, conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}

	defer oscar.Signout(sess, sm, fm)

	onClientReady := func(sess *oscar.Session, sm *oscar.SessionManager, r io.Reader, w io.Writer, sequence *uint32) error {
		err := oscar.NotifyArrival(sess, sm, fm)
		if err != nil {
			return err
		}

		return oscar.GetOnlineBuddies(w, sess, sm, fm, sequence)
	}

	foodGroups := []uint16{0x0001, 0x0002, 0x0003, 0x0004, 0x0009, 0x0013, 0x000D}
	if err := oscar.ReadBos(onClientReady, sess, seq, sm, fm, cr, conn, foodGroups); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}

func handleChatConnection(fm *oscar.FeedbagStore, cr *oscar.ChatRegistry, conn net.Conn) {
	defer conn.Close()

	cookie, seq, err := oscar.VerifyChatLogin(conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}

	room, err := cr.Retrieve(string(cookie.Cookie))

	chatSess, found := room.SessionManager.Retrieve(cookie.SessID)
	if !found {
		fmt.Printf("unable to find user for session: %s\n", cookie.SessID)
		return
	}
	defer exitChat(chatSess, room, cr)

	foodGroups := []uint16{0x0001, 0x0002, 0x0003, 0x0004, 0x0009, 0x0013, 0x000D, 0x000E}

	onClientReady := func(sess *oscar.Session, sm *oscar.SessionManager, r io.Reader, w io.Writer, sequence *uint32) error {
		if err := oscar.SendChatRoomInfoUpdate(room, w, sequence); err != nil {
			return err
		}
		oscar.AlertUserJoined(sess, sm)
		return oscar.SetOnlineChatUsers(sm, w, sequence)
	}

	if err := oscar.ReadBos(onClientReady, chatSess, seq, room.SessionManager, fm, cr, conn, foodGroups); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}

func exitChat(chatSess *oscar.Session, room oscar.ChatRoom, cr *oscar.ChatRegistry) {
	oscar.AlertUserLeft(chatSess, room.SessionManager)
	chatSess.Close()
	room.SessionManager.Remove(chatSess)
	cr.MaybeRemoveRoom(room.ID)
}
