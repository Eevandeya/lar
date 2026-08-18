package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/server"
	"github.com/eevandeya/lar/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "version for lar-gateway")
	configPath := flag.String("config", "/etc/lar/config.yaml", "path to config file")
	debug := flag.Bool("debug", false, "enable debug output")
	flag.Parse()

	var logLevel slog.Level
	if *debug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	if *showVersion {
		slog.Debug("--version flag detected, showing version", "version", version.Version)
		fmt.Printf("lar-gateway version %s\n", version.Version)
		return
	}

	var err error
	slog.Info("loading gateway config", "path", *configPath)
	cfg, err := config.LoadGateway(*configPath)
	if err != nil {
		slog.Debug("failed to load config", "err", err)
		fmt.Printf("Failed to load config file.\nUse flag --config to specify path to config.\n")
		return
	}

	s := server.NewServer(cfg)
	slog.Info("starting server", "port", cfg.Server.Port)
	if err = s.ListenAndServer(cfg.Server.Port); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
