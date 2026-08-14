package arp

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"time"

	"github.com/mdlayher/arp"
)

const readTimeout = 300

func Probe(macAddr net.HardwareAddr, hostIP net.IP, ifi *net.Interface) (bool, error) {
	hostIP = hostIP.To4()
	if hostIP == nil {
		return false, errors.New("host IP is not IPv4")
	}

	arpClient, err := arp.Dial(ifi)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = arpClient.Close()
	}()

	addr, ok := netip.AddrFromSlice(hostIP) // mdlayher/arp uses netip.Addr instead of net.IP
	if !ok {
		return false, fmt.Errorf("invalid IPv4 address: %v", hostIP)
	}

	err = arpClient.Request(addr)
	if err != nil {
		return false, err
	}

	err = arpClient.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	if err != nil {
		return false, err
	}

	for {
		packet, _, err := arpClient.Read()
		if err != nil {
			var netError net.Error
			if errors.As(err, &netError) && netError.Timeout() {
				return false, nil
			}

			return false, err
		}

		if packet.Operation != arp.OperationReply {
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
