package arp

import (
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
	"github.com/stretchr/testify/require"
)

type fakeMdlayherClient struct {
	packet       *arp.Packet
	readErr      error
	requestIP    netip.Addr
	requestErr error
	readDeadline time.Time
	setDeadlineErr error
	closed       bool
	closeErr error
}

func (c *fakeMdlayherClient) Request(ip netip.Addr) error {
	c.requestIP = ip
	return nil
}

func (c *fakeMdlayherClient) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *fakeMdlayherClient) Read() (*arp.Packet, *ethernet.Frame, error) {
	if c.readErr != nil {
		return nil, nil, c.readErr
	}
	return c.packet, nil, nil
}

func (c *fakeMdlayherClient) Close() error {
	c.closed = true
	return nil
}

func TestMdlayherARPClientRequest(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeMdlayherClient{requestErr: expectedErr}
		client := mdlayherARPClient{&fakeInnerClient}
	
		testIP := net.IP{32, 202, 40, 227}
		testNetIP, _ := netip.AddrFromSlice(testIP)
		
		err := client.Request(testNetIP)
		require.ErrorIs(t, err, expectedErr)
	})
	
	t.Run("No error", func(t *testing.T) {
		fakeInnerClient := fakeMdlayherClient{}
		client := mdlayherARPClient{&fakeInnerClient}
	
		testIP := net.IP{32, 202, 40, 227}
		testNetIP, _ := netip.AddrFromSlice(testIP)
	
		err := client.Request(testNetIP)
		require.NoError(t, err)
		require.Equal(t, testNetIP, fakeInnerClient.requestIP)
	})
}

func TestMdlayherARPClientSetReadDeadline(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeMdlayherClient{setDeadlineErr: expectedErr}
		client := mdlayherARPClient{&fakeInnerClient}
	
		deadline := time.Now().Add(time.Second)
	
		err := client.SetReadDeadline(deadline)
		require.ErrorIs(t, err, expectedErr)
	})
	
	t.Run("No error", func(t *testing.T) {
		fakeInnerClient := fakeMdlayherClient{}
		client := mdlayherARPClient{&fakeInnerClient}
	
		deadline := time.Now().Add(time.Second)
	
		err := client.SetReadDeadline(deadline)
		require.NoError(t, err)
		require.Equal(t, deadline, fakeInnerClient.readDeadline)
	})
}

func TestMdlayherARPClientClose(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeMdlayherClient{}
		client := mdlayherARPClient{&fakeInnerClient}
	
		err := client.Close()
		require.ErrorIs(t, err, expectedErr)
	})
	
	t.Run("No error", func(t *testing.T) {
		fakeInnerClient := fakeMdlayherClient{}
		client := mdlayherARPClient{&fakeInnerClient}
	
		err := client.Close()
		require.NoError(t, err)
		require.True(t, fakeInnerClient.closed)
	})
}

func TestMdlayherARPClientRead(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeMdlayherClient{readErr: expectedErr}
		client := mdlayherARPClient{&fakeInnerClient}

		_, err := client.Read()

		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("No error", func(t *testing.T) {
		fakeInnerClient := &fakeMdlayherClient{
			packet: &arp.Packet{
				Operation:          arp.OperationReply,
				SenderIP:           netip.MustParseAddr("192.168.1.10"),
				SenderHardwareAddr: net.HardwareAddr{1, 2, 3, 4, 5, 6},
			},
		}
		client := mdlayherARPClient{fakeInnerClient}

		packet, err := client.Read()

		require.NoError(t, err)
		require.Equal(t, OperationReply, packet.Operation)
		require.Equal(t, netip.MustParseAddr("192.168.1.10"), packet.SenderIP)
		require.Equal(t, net.HardwareAddr{1, 2, 3, 4, 5, 6}, packet.SenderHardwareAddr)
	})
}
