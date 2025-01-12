package osutility

import (
	"context"
	"kubeclusteragent/pkg/util/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLiveExec_Command(t *testing.T) {
	tests := []struct {
		name        string
		args        string
		wantStatus  int
		wantContent string
		wantError   bool
	}{
		{
			name:        "exit 0",
			args:        "echo 'good';",
			wantStatus:  0,
			wantContent: "good\n",
			wantError:   false,
		},
		{
			name:        "exit 1",
			args:        "echo 'bad'; exit 1",
			wantStatus:  1,
			wantContent: "bad\n",
			wantError:   false,
		},
		{
			name:        "capture stdout and stderr",
			args:        "echo 'stdout'; echo 'stderr' 1>&2",
			wantStatus:  0,
			wantContent: "stdout\nstderr\n",
			wantError:   false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewLiveExec()
			status, data, err := e.Command(context.Background(), "sh", nil, append([]string{"-c"}, test.args)...)

			test.CheckError(t, test.wantError, err, func() {
				require.Equal(t, test.wantStatus, status)
				require.Equal(t, test.wantContent, string(data))
			})

		})
	}
}
