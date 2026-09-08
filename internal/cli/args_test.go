package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequireOneArg(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantErr *ArgumentNumberError
	}{
		{
			name:  "0 args",
			input: nil,
			wantErr: &ArgumentNumberError{
				Expected: 1,
				Got:      0,
			},
		},
		{
			name: "1 arg",
			input: []string{
				"foo",
			},
		},
		{
			name: "Multiple args",
			input: []string{
				"foo",
				"bar",
				"baz",
			},
			wantErr: &ArgumentNumberError{
				Expected: 1,
				Got:      3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireOneArg(nil, tt.input)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				var argNumErr *ArgumentNumberError
				require.ErrorAs(t, err, &argNumErr)
				require.Equal(t, tt.wantErr.Got, argNumErr.Got)
				require.Equal(t, tt.wantErr.Expected, argNumErr.Expected)
			}
		})
	}
}
