package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	configPath := flag.String("config", "/etc/lar/config.yaml", "path to config file")
	flag.Parse()

	var err error
	cfg, err := config.LoadGateway(*configPath)
	if err != nil {
		log.Fatalf("Can not load config: %v\n", err)
	}

	s := server.NewServer(cfg)
	log.Fatal(s.ListenAndServer())
}
