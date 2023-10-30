package main

import (
	"errors"
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

	router := server.NewRouter()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleBOSConnection(cfg, sm, fm, cr, conn, router)
	}
}

func listenChat(cfg server.Config, fm *server.FeedbagStore, cr *server.ChatRegistry) {
	listener, err := net.Listen("tcp", server.Address("", cfg.ChatPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("Chat server listening on %s\n", server.Address(cfg.OSCARHost, cfg.ChatPort))

	router := server.NewRouterForChat()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleChatConnection(cfg, fm, cr, conn, router)
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

func handleBOSConnection(cfg server.Config, sm *server.InMemorySessionManager, fm *server.FeedbagStore, cr *server.ChatRegistry, conn net.Conn, router server.Router) {
	sess, seq, err := server.VerifyLogin(sm, conn)
	if err != nil {
		fmt.Printf("user disconnected with error: %s\n", err.Error())
		return
	}

	defer sess.Close()
	defer conn.Close()

	go func() {
		<-sess.Closed()
		server.Signout(sess, sm, fm)
	}()

	if err := server.ReadBos(cfg, sess, seq, sm, fm, cr, conn, server.ChatRoom{}, router); err != nil {
		switch {
		case errors.Is(io.EOF, err):
			fallthrough
		case errors.Is(server.ErrSignedOff, err):
			fmt.Println("user signed off")
		default:
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		}
	}
}

func handleChatConnection(cfg server.Config, fm *server.FeedbagStore, cr *server.ChatRegistry, conn net.Conn, router server.Router) {
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

	if err := server.ReadBos(cfg, chatSess, seq, room.SessionManager, fm, cr, conn, room, router); err != nil {
		if err != io.EOF {
			fmt.Printf("user disconnected with error: %s\n", err.Error())
		} else {
			fmt.Println("user disconnected")
		}
	}
}
