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

type fakeInnerMdlayherClient struct {
	packet         *arp.Packet
	readErr        error
	requestIP      netip.Addr
	requestErr     error
	readDeadline   time.Time
	setDeadlineErr error
	closed         bool
	closeErr       error
}

func (c *fakeInnerMdlayherClient) Request(ip netip.Addr) error {
	if c.requestErr != nil {
		return c.requestErr
	}
	c.requestIP = ip
	return nil
}

func (c *fakeInnerMdlayherClient) SetReadDeadline(t time.Time) error {
	if c.setDeadlineErr != nil {
		return c.setDeadlineErr
	}
	c.readDeadline = t
	return nil
}

func (c *fakeInnerMdlayherClient) Read() (*arp.Packet, *ethernet.Frame, error) {
	if c.readErr != nil {
		return nil, nil, c.readErr
	}
	return c.packet, nil, nil
}

func (c *fakeInnerMdlayherClient) Close() error {
	if c.closeErr != nil {
		return c.closeErr
	}
	c.closed = true
	return nil
}

func TestMdlayherARPClientRequest(t *testing.T) {
	t.Run("Request error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeInnerMdlayherClient{requestErr: expectedErr}
		client := mdlayherARPClient{&fakeInnerClient}

		testIP := net.IP{32, 202, 40, 227}
		testNetIP, _ := netip.AddrFromSlice(testIP)

		err := client.Request(testNetIP)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("Reqeust success", func(t *testing.T) {
		fakeInnerClient := fakeInnerMdlayherClient{}
		client := mdlayherARPClient{&fakeInnerClient}

		testIP := net.IP{32, 202, 40, 227}
		testNetIP, _ := netip.AddrFromSlice(testIP)

		err := client.Request(testNetIP)
		require.NoError(t, err)
		require.Equal(t, testNetIP, fakeInnerClient.requestIP)
	})
}

func TestMdlayherARPClientSetReadDeadline(t *testing.T) {
	t.Run("Set deadline error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeInnerMdlayherClient{setDeadlineErr: expectedErr}
		client := mdlayherARPClient{&fakeInnerClient}

		deadline := time.Now().Add(time.Second)

		err := client.SetReadDeadline(deadline)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("Set deadline success", func(t *testing.T) {
		fakeInnerClient := fakeInnerMdlayherClient{}
		client := mdlayherARPClient{&fakeInnerClient}

		deadline := time.Now().Add(time.Second)

		err := client.SetReadDeadline(deadline)
		require.NoError(t, err)
		require.Equal(t, deadline, fakeInnerClient.readDeadline)
	})
}

func TestMdlayherARPClientClose(t *testing.T) {
	t.Run("Close error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeInnerMdlayherClient{closeErr: expectedErr}
		client := mdlayherARPClient{&fakeInnerClient}

		err := client.Close()
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("Close success", func(t *testing.T) {
		fakeInnerClient := fakeInnerMdlayherClient{}
		client := mdlayherARPClient{&fakeInnerClient}

		err := client.Close()
		require.NoError(t, err)
		require.True(t, fakeInnerClient.closed)
	})
}

func TestMdlayherARPClientRead(t *testing.T) {
	t.Run("Read error", func(t *testing.T) {
		expectedErr := errors.New("foo")
		fakeInnerClient := fakeInnerMdlayherClient{readErr: expectedErr}
		client := mdlayherARPClient{&fakeInnerClient}

		_, err := client.Read()

		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("Read success", func(t *testing.T) {
		fakeInnerClient := &fakeInnerMdlayherClient{
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
