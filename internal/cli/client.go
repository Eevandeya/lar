package cli

type MachineClient interface {
	Wake(machine string) error
	Status(machine string) (bool, error)
	Shutdown(machine string) error
}

type WakeClient interface {
	Wake(machine string) error
}

type StatusClient interface {
	Status(machine string) (bool, error)
}

type ShutdownClient interface {
	Shutdown(machine string) error
}
