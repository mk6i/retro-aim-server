package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/mkaminski/goaim/server"
)

func main() {

	var cfg server.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to process app config: %s", err.Error())
		os.Exit(1)
	}

	logger := server.NewLogger(cfg)

	fm, err := server.NewFeedbagStore(cfg.DBPath)
	if err != nil {
		logger.Error("unable to create feedbag store", "err", err.Error())
		os.Exit(1)
	}

	go server.StartManagementAPI(fm, logger)

	sm := server.NewSessionManager(logger)
	cr := server.NewChatRegistry()

	go listenBOS(cfg, sm, fm, cr, logger.With("svc", "BOS"))
	go listenChat(cfg, fm, cr, logger.With("svc", "CHAT"))

	addr := server.Address("", cfg.OSCARPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("unable to bind OSCAR server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("starting OSCAR server", "addr", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleAuthConnection(cfg, sm, fm, conn)
	}
}

func listenBOS(cfg server.Config, sm *server.InMemorySessionManager, fm *server.FeedbagStore, cr *server.ChatRegistry, logger *slog.Logger) {
	addr := server.Address("", cfg.BOSPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("unable to bind BOS server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("starting service", "addr", addr)

	router := server.NewRouter(logger)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())
		logger.DebugContext(ctx, "accepted connection")
		go handleBOSConnection(ctx, cfg, sm, fm, cr, conn, router, logger)
	}
}

func listenChat(cfg server.Config, fm *server.FeedbagStore, cr *server.ChatRegistry, logger *slog.Logger) {
	addr := server.Address("", cfg.ChatPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("unable to bind chat server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("starting service", "addr", addr)

	router := server.NewRouterForChat(logger)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())
		logger.DebugContext(ctx, "accepted connection")
		go handleChatConnection(ctx, cfg, fm, cr, conn, router, logger)
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

func handleBOSConnection(ctx context.Context, cfg server.Config, sm *server.InMemorySessionManager, fm *server.FeedbagStore, cr *server.ChatRegistry, conn net.Conn, router server.Router, logger *slog.Logger) {
	sess, seq, err := server.VerifyLogin(sm, conn)
	if err != nil {
		logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	defer sess.Close()
	defer conn.Close()

	go func() {
		<-sess.Closed()
		server.Signout(ctx, logger, sess, sm, fm)
	}()

	ctx = context.WithValue(ctx, "screenName", sess.ScreenName)

	server.ReadBos(ctx, cfg, sess, seq, sm, fm, cr, conn, server.ChatRoom{}, router, logger)
}

func handleChatConnection(ctx context.Context, cfg server.Config, fm *server.FeedbagStore, cr *server.ChatRegistry, conn net.Conn, router server.Router, logger *slog.Logger) {
	cookie, seq, err := server.VerifyChatLogin(conn)
	if err != nil {
		logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	room, err := cr.Retrieve(string(cookie.Cookie))
	if err != nil {
		logger.ErrorContext(ctx, "unable to find chat room", "err", err.Error())
		return
	}

	chatSess, found := room.Retrieve(cookie.SessID)
	if !found {
		logger.ErrorContext(ctx, "unable to find user for session", "sessID", cookie.SessID)
		return
	}

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		server.AlertUserLeft(ctx, chatSess, room)
		room.Remove(chatSess)
		cr.MaybeRemoveRoom(room.Cookie)
		conn.Close()
	}()

	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName)

	server.ReadBos(ctx, cfg, chatSess, seq, room.SessionManager, fm, cr, conn, room, router, logger)
}
