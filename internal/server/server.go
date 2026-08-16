package server

import (
	"fmt"
	"net/http"
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
	return &s
}

func (s *Server) ListenAndServer(port uint16) error {
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        loggingMiddleware(s.mux),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	return server.ListenAndServe()
}
