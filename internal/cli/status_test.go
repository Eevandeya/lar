package cli

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type StatusFunc func(string) (bool, error)

func (f StatusFunc) Status(machine string) (bool, error) {
	return f(machine)
}

func TestNewStatusCommand(t *testing.T) {
	tests := []struct {
		name                  string
		getClient             func() StatusClient
		args                  []string
		wantErrString         *string
		wantExitCodeErr       *ExitCodeError
		wantArgumentNumberErr *ArgumentNumberError
		wantOutput            *string
	}{
		{
			name: "Client error",
			getClient: func() StatusClient {
				return StatusFunc(func(machine string) (bool, error) {
					require.Equal(t, "foo", machine)
					return false, errors.New("foo")
				})
			},
			args:          []string{"foo"},
			wantErrString: new("foo"),
		},
		{
			name: "Quiet success online",
			getClient: func() StatusClient {
				return StatusFunc(func(machine string) (bool, error) {
					require.Equal(t, "foo", machine)
					return true, nil
				})
			},
			args:            []string{"foo", "--quiet"},
			wantExitCodeErr: new(ExitCodeError(0)),
		},
		{
			name: "Quiet success offline",
			getClient: func() StatusClient {
				return StatusFunc(func(machine string) (bool, error) {
					require.Equal(t, "foo", machine)
					return false, nil
				})
			},
			args:            []string{"foo", "--quiet"},
			wantExitCodeErr: new(ExitCodeError(1)),
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
			name: "Success online",
			getClient: func() StatusClient {
				return StatusFunc(func(machine string) (bool, error) {
					require.Equal(t, "foo", machine)
					return true, nil
				})
			},
			args:       []string{"foo"},
			wantOutput: new(fmt.Sprintf("foo: %sonline%s\n", Green, Reset)),
		},
		{
			name: "Success offline",
			getClient: func() StatusClient {
				return StatusFunc(func(machine string) (bool, error) {
					require.Equal(t, "foo", machine)
					return false, nil
				})
			},
			args:       []string{"foo"},
			wantOutput: new(fmt.Sprintf("foo: %soffline%s\n", Red, Reset)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			cmd := newStatusCommand(&buf, tt.getClient)
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
			case tt.wantExitCodeErr != nil:
				var exitCodeErr ExitCodeError
				require.ErrorAs(t, err, &exitCodeErr)
				require.Equal(t, int(*tt.wantExitCodeErr), int(exitCodeErr))
			case tt.wantOutput != nil:
				require.NoError(t, err)
				require.Equal(t, *tt.wantOutput, buf.String())
			}
		})
	}
}
