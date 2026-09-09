package arp

import (
	"errors"
	"net"
	"net/netip"
	"slices"
	"time"
)

const readTimeout = 300

func Probe(macAddr net.HardwareAddr, machineIP net.IP, ifi *net.Interface) (bool, error) {
	return probe(macAddr, machineIP, ifi, ClientProvider(NewMdlayherClient))
}

func probe(macAddr net.HardwareAddr, machineIP net.IP, ifi *net.Interface, provider ClientProvider) (bool, error) {
	machineIP = machineIP.To4()
	if machineIP == nil {
		return false, errors.New("machine IP is not IPv4")
	}

	arpClient, err := provider(ifi)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = arpClient.Close()
	}()

	// To4() above guarantees that machineIP is a valid IPv4 address
	addr, _ := netip.AddrFromSlice(machineIP)

	err = arpClient.Request(addr)
	if err != nil {
		return false, err
	}

	err = arpClient.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	if err != nil {
		return false, err
	}

	for {
		packet, err := arpClient.Read()
		if err != nil {
			var netError net.Error
			if errors.As(err, &netError) && netError.Timeout() {
				return false, nil
			}

			return false, err
		}

		if packet.Operation != OperationReply {
			continue
		}
		if packet.SenderIP.Compare(addr) != 0 {
			continue
		}
		if !slices.Equal(packet.SenderHardwareAddr, macAddr) {
			continue
		}

		return true, nil
	}
}
