package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/server"
	"github.com/eevandeya/lar/internal/version"
	"gopkg.in/yaml.v3"
)

const (
	red   = "\033[31m"
	reset = "\033[0m"
)

func main() {
	showVersion := flag.Bool("version", false, "version for lar-gateway")
	useHTTP := flag.Bool("insecure", false, "allow the server to run without TLS using unencrypted HTTP.")
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
		errorPrefix := fmt.Sprintf("%sError:%s", red, reset)

		var addrErr *net.AddrError
		var missingValuesErr config.MissingConfigValuesErr
		var incompleteTLSErr config.IncompleteTLSConfigErr
		var typeErr *yaml.TypeError

		switch {
		case errors.As(err, &addrErr):
			_, _ = fmt.Fprintf(os.Stderr, "%s invalid config: invalid mac address: %s\n",
				errorPrefix, addrErr.Addr)

		case errors.As(err, &missingValuesErr):
			_, _ = fmt.Fprintf(os.Stderr, "%s invalid config: missing required values: %s \n",
				errorPrefix, strings.Join(missingValuesErr, ", "))

		case errors.As(err, &incompleteTLSErr):
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorPrefix, incompleteTLSErr.Error())

		case errors.As(err, &typeErr):
			_, _ = fmt.Fprintf(os.Stderr, "%s invalid config: %s\n", errorPrefix, typeErr.Error())

		case errors.Is(err, config.ErrUnsupportedIPVersion) || errors.Is(err, config.ErrInvalidIPAddress):
			// TODO: too little error message context here
			_, _ = fmt.Fprintf(os.Stderr, "%s invalid config: %s\n", errorPrefix, err.Error())

		default:
			_, _ = fmt.Fprintf(os.Stderr, "%s invalid config: %s\n", errorPrefix, err.Error())

		}
		return
	}

	if !*useHTTP && cfg.Server.TLSConfig == nil {
		fmt.Print(
			"TLS configuration is not set and --insecure is not enabled.\n\n" +
				"The server requires TLS configuration to run securely over HTTPS.\n" +
				"If you intentionally want to use unencrypted HTTP, start the server with\n--insecure.\n\n" +
				"WARNING: HTTP traffic is unencrypted and may expose sensitive data.\n" +
				"Use --insecure only if you understand and accept the security risks.\n")
		return
	}

	s := server.NewServerWithDependencies(cfg)
	slog.Info("starting server", "port", cfg.Server.Port)
	if err = s.ListenAndServe(*useHTTP); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
