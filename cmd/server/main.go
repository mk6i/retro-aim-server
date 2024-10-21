package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

var (
	// default build fields populated by GoReleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	cfgFile := flag.String("config", "settings.env", "Path to config file")
	showHelp := flag.Bool("help", false, "Display help")
	showVersion := flag.Bool("version", false, "Display build information")

	flag.Parse()

	switch {
	case *showVersion:
		fmt.Printf("%-10s %s\n", "version:", version)
		fmt.Printf("%-10s %s\n", "commit:", commit)
		fmt.Printf("%-10s %s\n", "date:", date)
		os.Exit(0)
	case *showHelp:
		flag.PrintDefaults()
		os.Exit(0)
	}

	// optionally populate environment variables with config file
	if err := godotenv.Load(*cfgFile); err != nil {
		fmt.Printf("Config file (%s) not found, defaulting to env vars for app config...\n", *cfgFile)
	} else {
		fmt.Printf("Successfully loaded config file (%s)\n", *cfgFile)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	type starter interface {
		Start(ctx context.Context) error
	}

	var g errgroup.Group
	start := func(fn starter) {
		g.Go(func() error { return fn.Start(ctx) })
	}

	deps, err := MakeCommonDeps()
	if err != nil {
		fmt.Printf("error initializing common deps: %v\n", err)
		os.Exit(1)
	}

	start(Admin(deps))
	start(Alert(deps))
	start(Auth(deps))
	start(BART(deps))
	start(BOS(deps))
	start(Chat(deps))
	start(ChatNav(deps))
	start(MgmtAPI(deps))
	start(ODir(deps))

	if err := g.Wait(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
