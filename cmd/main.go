package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/mkaminski/goaim/server"
	"io"
	"log"
	"net"
)

func main() {

	var cfg server.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	fm, err := server.NewFeedbagStore(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}

	go server.StartManagementAPI(fm)

	sm := server.NewSessionManager()
	cr := server.NewChatRegistry()

	go listenBOS(cfg, sm, fm, cr)
	go listenChat(cfg, fm, cr)

	listener, err := net.Listen("tcp", server.Address("", cfg.OSCARPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("OSCAR server listening on %s\n", server.Address(cfg.OSCARHost, cfg.OSCARPort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleAuthConnection(cfg, sm, fm, conn)
	}
}

func listenBOS(cfg server.Config, sm *server.InMemorySessionManager, fm *server.FeedbagStore, cr *server.ChatRegistry) {
	listener, err := net.Listen("tcp", server.Address("", cfg.BOSPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("BOS server listening on %s\n", server.Address(cfg.OSCARHost, cfg.BOSPort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleBOSConnection(cfg, sm, fm, cr, conn)
	}
}

func listenChat(cfg server.Config, fm *server.FeedbagStore, cr *server.ChatRegistry) {
	listener, err := net.Listen("tcp", server.Address("", cfg.ChatPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("Chat server listening on %s\n", server.Address(cfg.OSCARHost, cfg.ChatPort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleChatConnection(cfg, fm, cr, conn)
	}
}

func handleAuthConnection(cfg server.Config, sm *server.InMemorySessionManager, fm *server.FeedbagStore, conn net.Conn) {
	defer conn.Close()
	seq := uint32(100)
	_, err := server.SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}

	err = server.ReceiveAndSendAuthChallenge(cfg, fm, conn, conn, &seq, uuid.New)
	if err != nil {
		log.Println(err)
		return
	}

	err = server.ReceiveAndSendBUCPLoginRequest(cfg, sm, fm, conn, conn, &seq, uuid.New)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(cfg server.Config, sm *server.InMemorySessionManager, fm *server.FeedbagStore, cr *server.ChatRegistry, conn net.Conn) {
	sess, seq, err := server.VerifyLogin(sm, conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}

	defer sess.Close()
	go func() {
		<-sess.Closed()
		server.Signout(sess, sm, fm)
		conn.Close()
	}()

	onClientReady := func(sess *server.Session, sm server.SessionManager, r io.Reader, w io.Writer, sequence *uint32) error {
		if err := server.NotifyArrival(sess, sm, fm); err != nil {
			return err
		}
		return server.GetAllOnlineBuddies(w, sess, sm, fm, sequence)
	}

	foodGroups := []uint16{0x0001, 0x0002, 0x0003, 0x0004, 0x0009, 0x0013, 0x000D}
	if err := server.ReadBos(cfg, onClientReady, sess, seq, sm, fm, cr, conn, foodGroups); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}

func handleChatConnection(cfg server.Config, fm *server.FeedbagStore, cr *server.ChatRegistry, conn net.Conn) {
	cookie, seq, err := server.VerifyChatLogin(conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}

	room, err := cr.Retrieve(string(cookie.Cookie))
	if err != nil {
		fmt.Printf("unable to find chat room: %s\n", err.Error())
		return
	}

	chatSess, found := room.Retrieve(cookie.SessID)
	if !found {
		fmt.Printf("unable to find user for session: %s\n", cookie.SessID)
		return
	}

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		server.AlertUserLeft(chatSess, room)
		room.Remove(chatSess)
		cr.MaybeRemoveRoom(room.Cookie)
		conn.Close()
	}()

	foodGroups := []uint16{0x0001, 0x0002, 0x0003, 0x0004, 0x0009, 0x0013, 0x000D, 0x000E}

	onClientReady := func(sess *server.Session, sm server.SessionManager, r io.Reader, w io.Writer, sequence *uint32) error {
		if err := server.SendChatRoomInfoUpdate(room, w, sequence); err != nil {
			return err
		}
		server.AlertUserJoined(sess, sm)
		return server.SetOnlineChatUsers(sm, w, sequence)
	}

	if err := server.ReadBos(cfg, onClientReady, chatSess, seq, room.SessionManager, fm, cr, conn, foodGroups); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}
