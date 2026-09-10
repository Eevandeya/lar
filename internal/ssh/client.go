package ssh

type ClientProvider func(addr string, user string, key []byte) (Client, error)

type Session interface {
	Run(string) error
	Close() error
}

type Client interface {
	Close() error
	NewSession() (Session, error)
}
