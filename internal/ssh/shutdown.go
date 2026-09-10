package ssh

import (
	"net"
	"os"
	"strconv"
)

const poweroffCmd = "sudo poweroff"

func Shutdown(user, keyPath string, machineIP net.IP, sshPort uint16) error {
	return shutdown(user, keyPath, machineIP, sshPort, ClientProvider(NewSSHClient))
}

func shutdown(user, keyPath string, machineIP net.IP, sshPort uint16, provider ClientProvider) error {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(machineIP.String(), strconv.FormatUint(uint64(sshPort), 10))
	client, err := provider(addr, user, key)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	err = session.Run(poweroffCmd)
	if err != nil {
		return err
	}

	return nil
}
