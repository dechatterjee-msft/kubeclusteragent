package cri

import (
	"context"
	"kubeclusteragent/pkg/util/osutility"
	"reflect"
	"testing"
)

func TestAirgapRegistry_Add(t *testing.T) {
	type fields struct {
		RegistryFQDN         string
		RegistryEndpoint     string
		SkipVerify           bool
		RegistryCapabilities []string
		Ou                   osutility.OSUtil
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "Add-Private-Registry", fields: struct {
			RegistryFQDN         string
			RegistryEndpoint     string
			SkipVerify           bool
			RegistryCapabilities []string
			Ou                   osutility.OSUtil
		}{RegistryFQDN: "172-1-0-2.cdc.airgap",
			RegistryEndpoint:     "",
			SkipVerify:           true,
			RegistryCapabilities: []string{"push,pull"},
			Ou:                   osutility.NewDryRun()},
			args:    struct{ ctx context.Context }{ctx: context.Background()},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := PrivateRegistry{
				RegistryFQDN:         tt.fields.RegistryFQDN,
				RegistryEndpoint:     tt.fields.RegistryEndpoint,
				SkipVerify:           tt.fields.SkipVerify,
				RegistryCapabilities: tt.fields.RegistryCapabilities,
				Ou:                   tt.fields.Ou,
			}
			if err := a.Add(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAirgapRegistry_Del(t *testing.T) {
	type fields struct {
		RegistryFQDN         string
		RegistryEndpoint     string
		SkipVerify           bool
		RegistryCapabilities []string
		Ou                   osutility.OSUtil
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "Del-Private-Registry", fields: struct {
			RegistryFQDN         string
			RegistryEndpoint     string
			SkipVerify           bool
			RegistryCapabilities []string
			Ou                   osutility.OSUtil
		}{RegistryFQDN: "172-1-0-2.cdc.airgap",
			RegistryEndpoint:     "",
			SkipVerify:           true,
			RegistryCapabilities: []string{"push,pull"},
			Ou:                   osutility.NewDryRun()},
			args:    struct{ ctx context.Context }{ctx: context.Background()},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := PrivateRegistry{
				RegistryFQDN:         tt.fields.RegistryFQDN,
				RegistryEndpoint:     tt.fields.RegistryEndpoint,
				SkipVerify:           tt.fields.SkipVerify,
				RegistryCapabilities: tt.fields.RegistryCapabilities,
				Ou:                   tt.fields.Ou,
			}
			if err := a.Del(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Del() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAirgapRegistry_Get(t *testing.T) {
	type fields struct {
		RegistryFQDN         string
		RegistryEndpoint     string
		SkipVerify           bool
		RegistryCapabilities []string
		Ou                   osutility.OSUtil
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "Get-Private-Registry", fields: struct {
			RegistryFQDN         string
			RegistryEndpoint     string
			SkipVerify           bool
			RegistryCapabilities []string
			Ou                   osutility.OSUtil
		}{RegistryFQDN: "172-1-0-2.cdc.airgap",
			RegistryEndpoint:     "",
			SkipVerify:           true,
			RegistryCapabilities: []string{"push,pull"},
			Ou:                   osutility.NewDryRun()},
			args:    struct{ ctx context.Context }{ctx: context.Background()},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := PrivateRegistry{
				RegistryFQDN:         tt.fields.RegistryFQDN,
				RegistryEndpoint:     tt.fields.RegistryEndpoint,
				SkipVerify:           tt.fields.SkipVerify,
				RegistryCapabilities: tt.fields.RegistryCapabilities,
				Ou:                   tt.fields.Ou,
			}
			got, err := a.Get(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAirgapRegistry_Update(t *testing.T) {
	type fields struct {
		RegistryFQDN         string
		RegistryEndpoint     string
		SkipVerify           bool
		RegistryCapabilities []string
		Ou                   osutility.OSUtil
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "Update-Private-Registry", fields: struct {
			RegistryFQDN         string
			RegistryEndpoint     string
			SkipVerify           bool
			RegistryCapabilities []string
			Ou                   osutility.OSUtil
		}{RegistryFQDN: "172-1-0-2.cdc.airgap",
			RegistryEndpoint:     "",
			SkipVerify:           true,
			RegistryCapabilities: []string{"push,pull"},
			Ou:                   osutility.NewDryRun()},
			args:    struct{ ctx context.Context }{ctx: context.Background()},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := PrivateRegistry{
				RegistryFQDN:         tt.fields.RegistryFQDN,
				RegistryEndpoint:     tt.fields.RegistryEndpoint,
				SkipVerify:           tt.fields.SkipVerify,
				RegistryCapabilities: tt.fields.RegistryCapabilities,
				Ou:                   tt.fields.Ou,
			}
			if err := a.Update(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
