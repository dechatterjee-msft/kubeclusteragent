package calico

import (
	"context"
	"kubeclusteragent/pkg/util/osutility"
	"testing"
)

func TestConfigurePodIPReservation(t *testing.T) {
	type args struct {
		ctx context.Context
		ou  osutility.OSUtil
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "ConfigurePodIP", args: struct {
			ctx context.Context
			ou  osutility.OSUtil
		}{ctx: context.Background(), ou: osutility.NewDryRun()}, want: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConfigurePodIPReservation(tt.args.ctx, tt.args.ou)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigurePodIPReservation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConfigurePodIPReservation() got = %v, want %v", got, tt.want)
			}
		})
	}
}
