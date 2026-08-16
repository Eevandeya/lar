package server

import (
	"net/http"
	"time"

	"github.com/eevandeya/lar/internal/config"
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
	return &s
}

func (s *Server) ListenAndServer() error {
	server := &http.Server{
		Addr:           ":8080",
		Handler:        s.mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return server.ListenAndServe()
}
