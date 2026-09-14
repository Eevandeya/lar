package ssh

import "golang.org/x/crypto/ssh"

type sshClient struct {
	client *ssh.Client
}

func (c sshClient) Close() error {
	return c.client.Close()
}

// TODO: probably should test it...
func (c sshClient) NewSession() (Session, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	return Session(session), err
}

func NewSSHClient(addr string, user string, key []byte) (Client, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: figure it out
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	return sshClient{client: client}, nil
}
