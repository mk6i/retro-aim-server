package main

import (
	"fmt"
	"os"
	"sync"

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

	fm, err := server.NewFeedbagStore(cfg.DBPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to create feedbag store: %s", err.Error())
		os.Exit(1)
	}

	logger := server.NewLogger(cfg)
	sm := server.NewSessionManager(logger)
	cr := server.NewChatRegistry()

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		server.StartManagementAPI(fm, logger)
		wg.Done()
	}()
	go func() {
		server.ListenBOS(cfg, sm, fm, cr, logger.With("svc", "BOS"))
		wg.Done()
	}()
	go func() {
		server.ListenChat(cfg, fm, cr, logger.With("svc", "CHAT"))
		wg.Done()
	}()
	go func() {
		server.ListenBUCPLogin(cfg, err, logger, sm, fm)
		wg.Done()
	}()

	wg.Wait()
}
