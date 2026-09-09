package arp

import (
	"net"
	"net/netip"
	"time"

	"github.com/mdlayher/arp"
)

type Operation uint16

const (
	OperationRequest Operation = Operation(arp.OperationRequest)
	OperationReply   Operation = Operation(arp.OperationReply)
)

type Packet struct {
	SenderHardwareAddr net.HardwareAddr
	SenderIP           netip.Addr
	Operation          Operation
}

type Client interface {
	Request(netip.Addr) error
	SetReadDeadline(time.Time) error
	Read() (*Packet, error)
	Close() error
}

type ClientProvider func(ifi *net.Interface) (Client, error)
