package config

import (
	"errors"
	"net"
	"testing"
)

func TestIPUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "IPv4",
			input: "192.168.1.1",
			want:  "192.168.1.1",
		},
		{
			name:    "Invalid IP",
			input:   "hello",
			wantErr: ErrInvalidIPAddress,
		},
		{
			name:    "Empty",
			input:   "",
			wantErr: ErrInvalidIPAddress,
		},
		{
			name:    "IPv6",
			input:   "db8b:a8e8:2019:943d:e8ec:92f3:d364:f06c",
			wantErr: ErrUnsupportedIPVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ip IP
			err := ip.UnmarshalText([]byte(tt.input))

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}

			if net.IP(ip).String() != tt.want && err == nil {
				t.Fatalf("unexpected ip: %v", ip)
			}
		})
	}
}

func TestMACAddrUnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantErr     error
		wantAddrErr bool
	}{
		{
			name:  "Valid MAC",
			input: "2b:a6:f9:2a:cf:2e",
			want:  "2b:a6:f9:2a:cf:2e",
		},
		{
			name:        "Empty",
			input:       "",
			wantAddrErr: true,
		},
		{
			name:        "invalid MAC",
			input:       "2b:a6:f9:2a:cf:2e:4b",
			wantAddrErr: true,
		},
		{
			name:    "invalid MAC format",
			input:   "00:14:22:ff:fe:01:23:45",
			wantErr: ErrInvalidMacAddressFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mac MACAddr
			err := mac.UnmarshalText([]byte(tt.input))

			if tt.wantAddrErr {
				if _, ok := errors.AsType[*net.AddrError](err); !ok {
					t.Fatalf("unexpected *net.AddrError, got %T", err)
				}
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}

			if net.HardwareAddr(mac).String() != tt.want && err == nil {
				t.Fatalf("unexpected mac address: %v", mac)
			}
		})
	}
}
