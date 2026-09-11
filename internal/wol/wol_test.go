package wol

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWake(t *testing.T) {
	mac := net.HardwareAddr{0x1A, 0x2B, 0x3C, 0x4D, 0x5E, 0x6F}
	ip := net.IP{127, 0, 0, 1}

	listener, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   ip,
		Port: 0,
	})
	require.NoError(t, err)
	defer listener.Close()

	bufCh := make(chan []byte)
	errCh := make(chan error)

	go func() {
		buf := make([]byte, 1024)

		listener.SetReadDeadline(time.Now().Add(time.Second))

		n, _, err := listener.ReadFromUDP(buf)

		if err != nil {
			errCh <- err
			return
		}

		bufCh <- buf[:n]
	}()

	port := listener.LocalAddr().(*net.UDPAddr).Port
	err = wake(mac, ip, uint16(port))
	require.NoError(t, err)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case packet := <-bufCh:
		require.Len(t, packet, 6+6*16)
		require.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, packet[:6])

		for i := range 6 * 16 {
			require.Equal(t, mac[i%6], packet[6+i])
		}
	}
}
