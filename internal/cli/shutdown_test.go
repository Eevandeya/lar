package cli

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type ShutdownFunc func(string) error

func (f ShutdownFunc) Shutdown(machine string) error {
	return f(machine)
}

func TestNewShutdownCommand(t *testing.T) {
	tests := []struct {
		name                  string
		getClient             func() ShutdownClient
		args                  []string
		wantErrString         *string
		wantArgumentNumberErr *ArgumentNumberError
		wantOutput            *string
	}{
		{
			name: "Client error",
			getClient: func() ShutdownClient {
				return ShutdownFunc(func(machine string) error {
					require.Equal(t, "foo", machine)
					return errors.New("foo")
				})
			},
			args:          []string{"foo"},
			wantErrString: new("foo"),
		},
		{
			name: "Invalid arg number",
			args: []string{"foo", "bar"},
			wantArgumentNumberErr: new(ArgumentNumberError{
				Expected: 1,
				Got:      2,
			}),
		},
		{
			name: "Success",
			getClient: func() ShutdownClient {
				return ShutdownFunc(func(machine string) error {
					require.Equal(t, "foo", machine)
					return nil
				})
			},
			args:       []string{"foo"},
			wantOutput: new(fmt.Sprintf("foo: %sshutdown successful%s\n", Cyan, Reset)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			cmd := newShutdownCommand(&buf, tt.getClient)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			switch {
			case tt.wantErrString != nil:
				require.EqualError(t, err, *tt.wantErrString)
			case tt.wantArgumentNumberErr != nil:
				var argumentNumberErr *ArgumentNumberError
				require.ErrorAs(t, err, &argumentNumberErr)
				require.Equal(t, tt.wantArgumentNumberErr.Got, argumentNumberErr.Got)
				require.Equal(t, tt.wantArgumentNumberErr.Expected, argumentNumberErr.Expected)
			case tt.wantOutput != nil:
				require.Equal(t, *tt.wantOutput, buf.String())
			}
		})
	}
}
