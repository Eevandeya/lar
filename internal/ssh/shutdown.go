package ssh

import (
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
)

const poweroffCmd = "sudo poweroff"

func loadSigner(keyPath string) (ssh.Signer, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func Shutdown(user, keyPath string, hostIP net.IP, sshPort uint16) error {
	signer, err := loadSigner(keyPath)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: figure it out
	}

	addr := net.JoinHostPort(hostIP.String(), strconv.FormatUint(uint64(sshPort), 10))
	client, err := ssh.Dial("tcp", addr, config)
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
