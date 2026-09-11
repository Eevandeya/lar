package wol

import (
	"net"
)

const wolPort = 9

const macAddressLength = 6
const magicPacketPrefixLength = 6
const macAddressRepeat = 16

func Wake(macAddr net.HardwareAddr, ip net.IP) error {
	return wake(macAddr, ip, wolPort)
}

func wake(macAddr net.HardwareAddr, ip net.IP, port uint16) error {

	magicPacket := make([]byte, magicPacketPrefixLength+macAddressLength*macAddressRepeat)

	for i := range magicPacketPrefixLength {
		magicPacket[i] = 0xFF
	}
	for i := range macAddressLength * macAddressRepeat {
		magicPacket[magicPacketPrefixLength+i] = macAddr[i%macAddressLength]
	}

	addr := net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close() // escaping goland checks
		// TODO: get rid of those escapes in project
	}()

	_, err = conn.Write(magicPacket)
	if err != nil {
		return err
	}

	return nil
}
