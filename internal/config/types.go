package config

import (
	"errors"
	"net"
)

type IP net.IP
type MACAddr net.HardwareAddr

var ErrInvalidIPAddress = errors.New("invalid ip address")
var ErrUnsupportedIPVersion = errors.New("only IPv4 is supported")

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

	*m = MACAddr(addr)
	return nil
}
