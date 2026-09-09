package arp

import (
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const tolerance = 30 * time.Millisecond

type fakeARPClient struct {
	requestFn         func(ip netip.Addr) error
	setReadDeadlineFn func(t time.Time) error
	readFn            func() (*Packet, error)
	closeFn           func() error
	closed            bool
}

func (c *fakeARPClient) Request(ip netip.Addr) error {
	return c.requestFn(ip)
}

func (c *fakeARPClient) SetReadDeadline(t time.Time) error {
	return c.setReadDeadlineFn(t)
}

func (c *fakeARPClient) Read() (*Packet, error) {
	return c.readFn()
}
func (c *fakeARPClient) Close() error {
	c.closed = true
	return c.closeFn()
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return false }

func TestProbe(t *testing.T) {
	testMac := net.HardwareAddr{0x1A, 0x2B, 0x3C, 0x4D, 0x5E, 0x6F}
	testIP := net.IP{32, 202, 40, 227}
	testNetIP, _ := netip.AddrFromSlice(testIP)
	testInterface := &net.Interface{}

	provider := func(client *fakeARPClient) ClientProvider {
		return ClientProvider(func(ifi *net.Interface) (Client, error) {
			require.Equal(t, testInterface, ifi)
			return client, nil
		})
	}

	closeFn := func() error {
		return nil
	}

	requestFn := func(ip netip.Addr) error {
		require.Equal(t, testNetIP, ip)
		return nil
	}

	setReadDeadlineFn := func(tm time.Time) error {
		maxTimoutTiming := time.Now().Add(readTimeout * time.Millisecond)
		require.WithinDuration(t, maxTimoutTiming, tm, tolerance)
		return nil
	}

	t.Run("IPv6 address", func(t *testing.T) {
		ipv6 := net.ParseIP("2001:db8:85a3::8a2e:370:7334")

		_, err := probe(testMac, ipv6, testInterface, ClientProvider(nil))
		require.EqualError(t, err, "machine IP is not IPv4")
	})

	t.Run("Provider error", func(t *testing.T) {
		_, err := probe(testMac, testIP, testInterface, ClientProvider(func(ifi *net.Interface) (Client, error) {
			require.Equal(t, testInterface, ifi)
			return nil, errors.New("foo")
		}))
		require.EqualError(t, err, "foo")
	})

	t.Run("Request error", func(t *testing.T) {
		client := fakeARPClient{
			closeFn: closeFn,
			requestFn: func(ip netip.Addr) error {
				require.Equal(t, testNetIP, ip)
				return errors.New("foo")
			},
		}

		_, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.EqualError(t, err, "foo")
	})

	t.Run("Set deadline error", func(t *testing.T) {
		client := fakeARPClient{
			closeFn:   closeFn,
			requestFn: requestFn,
			setReadDeadlineFn: func(tm time.Time) error {
				maxTimoutTiming := time.Now().Add(readTimeout * time.Millisecond)
				require.WithinDuration(t, maxTimoutTiming, tm, tolerance)
				return errors.New("foo")
			},
		}

		_, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.EqualError(t, err, "foo")
	})

	t.Run("Read timeout", func(t *testing.T) {
		client := fakeARPClient{
			closeFn:           closeFn,
			requestFn:         requestFn,
			setReadDeadlineFn: setReadDeadlineFn,
			readFn: func() (*Packet, error) {
				return nil, timeoutError{}
			},
		}

		online, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.NoError(t, err)
		require.False(t, online)
	})

	t.Run("Read error", func(t *testing.T) {
		client := fakeARPClient{
			closeFn:           closeFn,
			requestFn:         requestFn,
			setReadDeadlineFn: setReadDeadlineFn,
			readFn: func() (*Packet, error) {
				return nil, errors.New("foo")
			},
		}

		_, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.EqualError(t, err, "foo")
	})

	t.Run("Non reply packet", func(t *testing.T) {
		var counter = 0
		client := fakeARPClient{
			closeFn:           closeFn,
			requestFn:         requestFn,
			setReadDeadlineFn: setReadDeadlineFn,
			readFn: func() (*Packet, error) {
				if counter == 0 {
					counter++
					return &Packet{
						SenderHardwareAddr: testMac,
						SenderIP:           testNetIP,
						Operation:          OperationRequest,
					}, nil
				}
				return nil, errors.New("foo")
			},
		}

		_, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.EqualError(t, err, "foo")
	})

	t.Run("Other mac packet", func(t *testing.T) {
		var counter = 0
		otherMac := net.HardwareAddr{0xA1, 0xB2, 0xC3, 0xD4, 0xE5, 0xF6}
		client := fakeARPClient{
			closeFn:           closeFn,
			requestFn:         requestFn,
			setReadDeadlineFn: setReadDeadlineFn,
			readFn: func() (*Packet, error) {
				if counter == 0 {
					counter++
					return &Packet{
						SenderHardwareAddr: otherMac,
						SenderIP:           testNetIP,
						Operation:          OperationReply,
					}, nil
				}
				return nil, errors.New("foo")
			},
		}

		_, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.EqualError(t, err, "foo")
	})

	t.Run("Other ip packet", func(t *testing.T) {
		var counter = 0
		otherNetIP := netip.AddrFrom4([4]byte{1, 2, 3, 4})
		client := fakeARPClient{
			closeFn:           closeFn,
			requestFn:         requestFn,
			setReadDeadlineFn: setReadDeadlineFn,
			readFn: func() (*Packet, error) {
				if counter == 0 {
					counter++
					return &Packet{
						SenderHardwareAddr: testMac,
						SenderIP:           otherNetIP,
						Operation:          OperationReply,
					}, nil
				}
				return nil, errors.New("foo")
			},
		}

		_, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.EqualError(t, err, "foo")
	})

	t.Run("Machine online", func(t *testing.T) {
		client := fakeARPClient{
			closeFn:           closeFn,
			requestFn:         requestFn,
			setReadDeadlineFn: setReadDeadlineFn,
			readFn: func() (*Packet, error) {
				return &Packet{
					SenderHardwareAddr: testMac,
					SenderIP:           testNetIP,
					Operation:          OperationReply,
				}, nil
			},
		}

		online, err := probe(testMac, testIP, testInterface, provider(&client))
		require.True(t, client.closed)
		require.NoError(t, err)
		require.True(t, online)
	})
}
