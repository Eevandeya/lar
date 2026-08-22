package config

import "testing"

func TestIncompleteTLSConfigErrError(t *testing.T) {
	tests := []struct {
		name  string
		input TLSConfigErrorType
		want  string
	}{
		{
			name:  "Certificate path missing",
			input: CertificatePathMissing,
			want:  "invalid tls config: private key path is provided, but certificate path is missing",
		},
		{
			name:  "Private key missing",
			input: PrivateKeyPathMissing,
			want:  "invalid tls config: certificate path is provided, but private key path is missing",
		},
		{
			name:  "Unknown error type",
			input: 3,
			want:  "invalid tls config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IncompleteTLSConfigErr{
				Type: tt.input,
			}
			if err.Error() != tt.want {
				t.Fatalf("expected error message '%s', but got '%s'", tt.want, err.Error())
			}
		})
	}
}

func TestMissingConfigValuesError(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{
			name:  "One value",
			input: []string{"server.config"},
			want:  "missing mandatory values in config: server.config",
		},
		{
			name:  "Multiple values",
			input: []string{"server.config", "machines.pc1.ssh-port"},
			want:  "missing mandatory values in config: server.config, machines.pc1.ssh-port",
		},
		{
			name:  "No values",
			input: nil,
			want:  "missing mandatory values in config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MissingConfigValuesErr(tt.input)
			if err.Error() != tt.want {
				t.Fatalf("expected error message '%s', but got '%s'", tt.want, err.Error())
			}
		})
	}
}
