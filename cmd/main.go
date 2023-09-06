package main

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {

	var cfg oscar.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	fm, err := oscar.NewFeedbagStore(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}

	sm := oscar.NewSessionManager()
	cr := oscar.NewChatRegistry()

	go listenBOS(cfg, sm, fm, cr)
	go listenChat(cfg, fm, cr)

	listener, err := net.Listen("tcp", oscar.Address("", cfg.OSCARPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("OSCAR server listening on %s\n", oscar.Address(cfg.OSCARHost, cfg.OSCARPort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleAuthConnection(cfg, sm, fm, conn)
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

func listenBOS(cfg oscar.Config, sm *oscar.SessionManager, fm *oscar.FeedbagStore, cr *oscar.ChatRegistry) {
	listener, err := net.Listen("tcp", oscar.Address("", cfg.BOSPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("BOS server listening on %s\n", oscar.Address(cfg.OSCARHost, cfg.BOSPort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleBOSConnection(cfg, sm, fm, cr, conn)
	}
}

func listenChat(cfg oscar.Config, fm *oscar.FeedbagStore, cr *oscar.ChatRegistry) {
	listener, err := net.Listen("tcp", oscar.Address("", cfg.ChatPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("Chat server listening on %s\n", oscar.Address(cfg.OSCARHost, cfg.ChatPort))

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleChatConnection(cfg, fm, cr, conn)
	}
}

func handleAuthConnection(cfg oscar.Config, sm *oscar.SessionManager, fm *oscar.FeedbagStore, conn net.Conn) {
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

	err = oscar.ReceiveAndSendBUCPLoginRequest(cfg, sess, fm, conn, conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(cfg oscar.Config, sm *oscar.SessionManager, fm *oscar.FeedbagStore, cr *oscar.ChatRegistry, conn net.Conn) {
	sess, seq, err := oscar.VerifyLogin(sm, conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}

	defer sess.Close()
	go func() {
		<-sess.Closed()
		oscar.Signout(sess, sm, fm)
		conn.Close()
	}()

	onClientReady := func(sess *oscar.Session, sm *oscar.SessionManager, r io.Reader, w io.Writer, sequence *uint32) error {
		err := oscar.NotifyArrival(sess, sm, fm)
		if err != nil {
			return err
		}

		return oscar.GetOnlineBuddies(w, sess, sm, fm, sequence)
	}

	foodGroups := []uint16{0x0001, 0x0002, 0x0003, 0x0004, 0x0009, 0x0013, 0x000D}
	if err := oscar.ReadBos(cfg, onClientReady, sess, seq, sm, fm, cr, conn, foodGroups); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}

func handleChatConnection(cfg oscar.Config, fm *oscar.FeedbagStore, cr *oscar.ChatRegistry, conn net.Conn) {
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

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		oscar.AlertUserLeft(chatSess, room.SessionManager)
		room.SessionManager.Remove(chatSess)
		cr.MaybeRemoveRoom(room.ID)
		conn.Close()
	}()

	foodGroups := []uint16{0x0001, 0x0002, 0x0003, 0x0004, 0x0009, 0x0013, 0x000D, 0x000E}

	onClientReady := func(sess *oscar.Session, sm *oscar.SessionManager, r io.Reader, w io.Writer, sequence *uint32) error {
		if err := oscar.SendChatRoomInfoUpdate(room, w, sequence); err != nil {
			return err
		}
		oscar.AlertUserJoined(sess, sm)
		return oscar.SetOnlineChatUsers(sm, w, sequence)
	}

	if err := oscar.ReadBos(cfg, onClientReady, chatSess, seq, room.SessionManager, fm, cr, conn, foodGroups); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}
