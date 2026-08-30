package cli

type MachineClientProvider struct {
	getClient func() MachineClient
}

func (p MachineClientProvider) Client() MachineClient {
	return p.getClient()
}

func (p MachineClientProvider) WakeClient() WakeClient {
	return p.Client()
}

func (p MachineClientProvider) StatusClient() StatusClient {
	return p.Client()

}

func (p MachineClientProvider) ShutdownClient() ShutdownClient {
	return p.Client()
}
