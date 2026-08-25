package server

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/eevandeya/lar/internal/config"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	maxHeaderBytes = 1 << 20
)

type Server struct {
	cfg *config.GatewayConfig
	mux *http.ServeMux
}

func NewServer(cfg *config.GatewayConfig) *Server {
	s := Server{
		cfg: cfg,
		mux: http.NewServeMux(),
	}
	s.mux.HandleFunc("GET /status", s.statusHandler)
	s.mux.HandleFunc("POST /wake", s.wakeHandler)
	s.mux.HandleFunc("POST /shutdown", s.shutdownHandler)
	return &s
}

func (s *Server) ListenAndServe(useHTTP bool) error {
	handler := loggingMiddleware(slog.Default(), s.authMiddleware(s.mux))

	server := &http.Server{
		Addr: net.JoinHostPort(
			net.IP(s.cfg.Server.Host).String(),
			strconv.FormatUint(uint64(s.cfg.Server.Port), 10)),
		Handler:        handler,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	if useHTTP {
		slog.Warn("Using unencrypted HTTP." +
			"Traffic is not encrypted and may be intercepted.")
		return server.ListenAndServe()
	}

	if s.cfg.Server.TLSConfig != nil {
		return server.ListenAndServeTLS(
			s.cfg.Server.TLSConfig.CertificateFilePath,
			s.cfg.Server.TLSConfig.KeyFilePath)
	}

	return errors.New("no tls config found")
}
