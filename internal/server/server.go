package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/eevandeya/lar/internal/arp"
	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/ssh"
	"github.com/eevandeya/lar/internal/wol"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	maxHeaderBytes = 1 << 20
)

type Dependencies struct {
	InterfaceByName func(name string) (*net.Interface, error)
	ArpProbe        func(macAddr net.HardwareAddr, machineIP net.IP, ifi *net.Interface) (bool, error)
	Wake            func(macAddr net.HardwareAddr, ip net.IP) error
	Shutdown        func(user string, keyPath string, machineIP net.IP, sshPort uint16) error
}

type Server struct {
	cfg  *config.GatewayConfig
	mux  *http.ServeMux
	deps Dependencies
	http *http.Server
}

func NewServerWithDependencies(cfg *config.GatewayConfig) *Server {
	return NewServer(cfg, Dependencies{
		InterfaceByName: net.InterfaceByName,
		ArpProbe:        arp.Probe,
		Wake:            wol.Wake,
		Shutdown:        ssh.Shutdown,
	})
}

func NewServer(cfg *config.GatewayConfig, deps Dependencies) *Server {
	s := Server{
		cfg:  cfg,
		mux:  http.NewServeMux(),
		deps: deps,
	}

	handler := loggingMiddleware(slog.Default(), s.authMiddleware(s.mux))

	httpServer := &http.Server{
		Addr: net.JoinHostPort(
			net.IP(s.cfg.Server.Host).String(),
			strconv.FormatUint(uint64(s.cfg.Server.Port), 10)),
		Handler:        handler,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	s.http = httpServer

	s.mux.HandleFunc("GET /status", s.statusHandler)
	s.mux.HandleFunc("POST /wake", s.wakeHandler)
	s.mux.HandleFunc("POST /shutdown", s.shutdownHandler)

	return &s
}

func (s *Server) Serve(listener net.Listener) error {
	return s.http.Serve(listener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) ListenAndServe(useHTTP bool) error {
	if useHTTP {
		slog.Warn("Using unencrypted HTTP." +
			"Traffic is not encrypted and may be intercepted.")
		return s.http.ListenAndServe()
	}

	if s.cfg.Server.TLSConfig != nil {
		return s.http.ListenAndServeTLS(
			s.cfg.Server.TLSConfig.CertificateFilePath,
			s.cfg.Server.TLSConfig.KeyFilePath)
	}

	return errors.New("no tls config found")
}
