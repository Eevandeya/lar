package server

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/gateway/arp"
)

func (s *Server) writeError(w http.ResponseWriter, statusCode int, code api.ErrorCode, message string) error {
	respErr := api.ErrorResponse{
		Error: &api.Error{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(respErr)
}

func (s *Server) checkAuth(r *http.Request) bool {
	authHeader, ok := r.Header["Authorization"]
	if !ok || len(authHeader) == 0 {
		return false
	}

	// NOTE: simplification for now
	authPayload := authHeader[0]

	if !strings.HasPrefix(authPayload, "Bearer ") {
		return false
	}

	return strings.TrimPrefix(authPayload, "Bearer ") == s.cfg.Secret
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		if err := s.writeError(w, http.StatusUnauthorized, api.ErrUnauthorized, "invalid credentials"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}

		return
	}

	queryParams := r.URL.Query()
	hostName := queryParams.Get("host")

	if hostName == "" {
		if err := s.writeError(w, http.StatusBadRequest, api.ErrMissingHostName, "missing host name query param"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	host, ok := s.cfg.Hosts[hostName]
	if !ok {
		if err := s.writeError(w, http.StatusBadRequest, api.ErrInvalidHostName, "invalid host name"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	iface, err := net.InterfaceByName(s.cfg.InterfaceName)
	if err != nil {
		if err = s.writeError(w, http.StatusInternalServerError, api.ErrInterfaceUnavailable, "configured interface is unavailable"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	online, err := arp.Probe(net.HardwareAddr(host.MAC), net.IP(host.Address), iface)
	if err != nil {
		slog.Error("ARP probe failed", "err", err)
		if err = s.writeError(w, http.StatusInternalServerError, api.ErrARPProbingFailed, "arp probing has failed"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	response := api.StatusResponse{Online: online}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		slog.Error("failed to write error response", "err", err)
		return
	}

	return
}
