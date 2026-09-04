package config

import (
	"errors"
	"net"
)

var ErrInvalidIPAddress = errors.New("invalid ip address")
var ErrUnsupportedIPVersion = errors.New("only IPv4 is supported")
var ErrInvalidMacAddressFormat = errors.New("invalid MAC address format, only Ethernet MAC (MAC-48) can be used")

type IP net.IP
type MACAddr net.HardwareAddr

func (i *IP) UnmarshalText(text []byte) error {
	ip := net.ParseIP(string(text))
	if ip == nil {
		return ErrInvalidIPAddress
	}

	if ip.To4() == nil {
		return ErrUnsupportedIPVersion
	}

	*i = IP(ip)
	return nil
}

func (m *MACAddr) UnmarshalText(text []byte) error {
	addr, err := net.ParseMAC(string(text))
	if err != nil {
		return err
	}

	if len(addr) != 6 {
		return ErrInvalidMacAddressFormat
	}

	*m = MACAddr(addr)
	return nil
}
