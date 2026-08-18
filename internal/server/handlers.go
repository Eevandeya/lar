package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/arp"
	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/ssh"
	"github.com/eevandeya/lar/internal/wol"
)

func (s *Server) getHostOrWriteError(w http.ResponseWriter, query url.Values) (*config.Machine, error) {
	hostName := query.Get("host")

	if hostName == "" {
		if err := writeError(w, http.StatusBadRequest, api.ErrMissingHostName, "missing host name query param"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return nil, errors.New("host name query is missing in query")
	}

	host, ok := s.cfg.Machines[hostName]
	if !ok {
		if err := writeError(w, http.StatusBadRequest, api.ErrInvalidHostName, "invalid host name"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return nil, errors.New("no much for requested host in config")
	}

	return host, nil
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	host, err := s.getHostOrWriteError(w, query)
	if err != nil {
		return
	}

	iface, err := net.InterfaceByName(s.cfg.Server.ARPInterfaceName)
	if err != nil {
		if err = writeError(w, http.StatusInternalServerError, api.ErrInterfaceUnavailable, "configured interface is unavailable"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	online, err := arp.Probe(net.HardwareAddr(host.MAC), net.IP(host.Address), iface)
	if err != nil {
		slog.Error("ARP probe failed", "err", err)
		if err = writeError(w, http.StatusInternalServerError, api.ErrARPProbingFailed, "arp probing has failed"); err != nil {
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

func (s *Server) wakeHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	host, err := s.getHostOrWriteError(w, query)
	if err != nil {
		return
	}

	err = wol.Wake(net.HardwareAddr(host.MAC), net.IP(s.cfg.Server.Broadcast))
	if err != nil {
		slog.Error("Wake-on-Lan failed", "err", err)
		if err = writeError(w, http.StatusInternalServerError, api.ErrWOLFailed, "wake-on-lan has failed"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

func (s *Server) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	host, err := s.getHostOrWriteError(w, query)
	if err != nil {
		return
	}

	err = ssh.Shutdown(host.User, host.IdentityFilePath, net.IP(host.Address), host.SSHPort)
	if err != nil {
		slog.Debug("failed to shutdown host", "err", err)
		if err = writeError(w, http.StatusBadRequest, api.ErrSSHFailed, "ssh to host failed"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
