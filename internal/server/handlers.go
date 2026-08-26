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

var interfaceByName = net.InterfaceByName
var arpProbe = arp.Probe
var wake = wol.Wake
var shutdown = ssh.Shutdown

func (s *Server) getMachineOrWriteError(w http.ResponseWriter, query url.Values) (*config.Machine, error) {
	machineName := query.Get("machine")

	if machineName == "" {
		if err := writeError(w, http.StatusBadRequest, api.ErrMissingMachineName, "missing machine name query param"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return nil, errors.New("machine name query is missing in query")
	}

	machine, ok := s.cfg.Machines[machineName]
	if !ok {
		if err := writeError(w, http.StatusBadRequest, api.ErrInvalidMachineName, "invalid machine name"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return nil, errors.New("no much for requested machine in config")
	}

	return machine, nil
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	machine, err := s.getMachineOrWriteError(w, query)
	if err != nil {
		return
	}

	ifi, err := interfaceByName(s.cfg.Server.ARPInterfaceName)
	if err != nil {
		if err = writeError(w, http.StatusInternalServerError, api.ErrInterfaceUnavailable, "configured interface is unavailable"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	online, err := arpProbe(net.HardwareAddr(machine.MAC), net.IP(machine.Address), ifi)
	if err != nil {
		slog.Error("ARP probe failed", "err", err)
		if err = writeError(w, http.StatusInternalServerError, api.ErrARPProbingFailed, "arp probing has failed"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	response := api.StatusResponse{Online: online}

	w.Header().Set("Content-Type", api.ApplicationJSON)
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		slog.Error("failed to write error response", "err", err)
		return
	}

	return
}

func (s *Server) wakeHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	machine, err := s.getMachineOrWriteError(w, query)
	if err != nil {
		return
	}

	err = wake(net.HardwareAddr(machine.MAC), net.IP(s.cfg.Server.Broadcast))
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
	machine, err := s.getMachineOrWriteError(w, query)
	if err != nil {
		return
	}

	err = shutdown(machine.User, machine.IdentityFilePath, net.IP(machine.Address), machine.SSHPort)
	if err != nil {
		slog.Debug("failed to shutdown machine", "err", err)
		if err = writeError(w, http.StatusBadRequest, api.ErrSSHFailed, "ssh to machine failed"); err != nil {
			slog.Error("failed to write error response", "err", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
