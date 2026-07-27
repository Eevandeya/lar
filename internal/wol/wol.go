package wol

import (
	"net"
)

const wolPort = 9
const macAddressLength = 6
const magicPacketPrefixLength = 6
const macAddressRepeat = 16

func Wake(macAddr net.HardwareAddr, ip net.IP) error {

	magicPacket := make([]byte, magicPacketPrefixLength+macAddressLength*macAddressRepeat)

	for i := 0; i < magicPacketPrefixLength; i++ {
		magicPacket[i] = 0xFF
	}
	for i := 0; i < macAddressLength*macAddressRepeat; i++ {
		magicPacket[magicPacketPrefixLength+i] = macAddr[i%macAddressLength]
	}

	addr := net.UDPAddr{
		IP:   ip,
		Port: wolPort,
	}

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close() // escaping goland checks
	}()

	_, err = conn.Write(magicPacket)
	if err != nil {
		return err
	}

	return nil
}
