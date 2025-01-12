package osutility

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestLiveHost_GetHostname(t *testing.T) {
	hostname, err := os.Hostname()
	hostname = strings.ToLower(hostname)
	if err != nil {
		t.Fail()
	}
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{name: "hostname", want: hostname, wantErr: false},
		{name: "InvalidHostname", want: hostname, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LiveHost{}
			got, err := l.GetHostname()
			if err != nil {
				t.Fail()
			}
			assert.Equalf(t, tt.want, got, "GetHostname()")
		})
	}
}
