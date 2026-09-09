package arp

import (
	"net"
	"net/netip"
	"time"

	"github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
)

type innerMdlayherARPClient interface {
	Request(netip.Addr) error
	SetReadDeadline(time.Time) error
	Read() (*arp.Packet, *ethernet.Frame, error)
	Close() error
}

type mdlayherARPClient struct {
	client innerMdlayherARPClient
}

func (c *mdlayherARPClient) Request(ip netip.Addr) error {
	return c.client.Request(ip)
}

func (c *mdlayherARPClient) SetReadDeadline(t time.Time) error {
	return c.client.SetReadDeadline(t)
}

func (c *mdlayherARPClient) Read() (*Packet, error) {
	packet, _, err := c.client.Read()
	if err != nil {
		return nil, err
	}

	return &Packet{
		SenderHardwareAddr: packet.SenderHardwareAddr,
		SenderIP:           packet.SenderIP,
		Operation:          Operation(packet.Operation),
	}, err
}

func (c *mdlayherARPClient) Close() error {
	return c.client.Close()
}

func NewMdlayherClient(ifi *net.Interface) (Client, error) {
	client, err := arp.Dial(ifi)
	if err != nil {
		return nil, err
	}
	return &mdlayherARPClient{client}, nil
}
