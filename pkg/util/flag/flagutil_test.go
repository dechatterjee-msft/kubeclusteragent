package flag_test

import (
	"flag"
	"fmt"
	"kubeclusteragent/pkg/util/flag"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvStringVar(t *testing.T) {
	s := "from env"

	flagSet := flag.CommandLine
	defer func() {
		flag.CommandLine = flagSet
	}()

	type args struct {
		s        *string
		key      string
		flagName string
	}

	tests := []struct {
		name string
		args args
		env  []string
		want string
	}{
		{
			name: "value provided in environment",
			args: args{
				s:        &s,
				key:      "TEST_KEY",
				flagName: "key",
			},
			env: []string{
				"TEST_KEY=value",
			},
			want: "value",
		},
		{
			name: "value not provided in environment",
			args: args{
				s:        &s,
				key:      "TEST_KEY",
				flagName: "key",
			},
			env:  nil,
			want: "default",
		},
	}
	for _, tt := range tests {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		t.Run(
			tt.name, func(t *testing.T) {
				defer func() {
					require.NoError(t, os.Unsetenv(tt.args.key))
				}()

				require.NoError(t, setEnvFromSlice(tt.env))
				flag.EnvStringVar(tt.args.s, tt.args.key, tt.args.flagName, "default", "usage")

				require.Equal(t, tt.want, *tt.args.s)
			},
		)

	}
}

func setEnvFromSlice(sl []string) error {
	for _, s := range sl {
		parts := strings.SplitN(s, "=", 2)
		if err := os.Setenv(parts[0], parts[1]); err != nil {
			return fmt.Errorf("resetting environment %s: %w", parts, err)
		}
	}

	return nil
}
