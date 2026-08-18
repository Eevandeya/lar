package config

import (
	"net"
)

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

	*m = MACAddr(addr)
	return nil
}
