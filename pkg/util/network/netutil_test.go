package network

import (
	"context"
	"kubeclusteragent/pkg/util/osutility"
	"reflect"
	"testing"
)

func TestNewSecondaryNetwork(t *testing.T) {
	type args struct {
		ctx                     context.Context
		PrimaryNetworkInterface string
		Label                   string
		IPv4Address             string
		IPv4Gateway             string
		ou                      osutility.OSUtil
	}
	tests := []struct {
		name string
		args args
		want *SecondaryNetworkIPv4
	}{
		{
			name: "new-secondary-network", args: struct {
				ctx                     context.Context
				PrimaryNetworkInterface string
				Label                   string
				IPv4Address             string
				IPv4Gateway             string
				ou                      osutility.OSUtil
			}{
				ctx:                     context.Background(),
				PrimaryNetworkInterface: "eth0",
				Label:                   "1",
				IPv4Address:             "100.102.1.1",
				IPv4Gateway:             "",
				ou:                      osutility.NewDryRun(),
			},
			want: &SecondaryNetworkIPv4{
				PrimaryNetworkInterface: "eth0",
				IPv4Address:             "100.102.1.1",
				IPv4Gateway:             "",
				AddressLabel:            "1",
				Osutil:                  osutility.NewDryRun(),
				networkInterfacePath:    "/etc/systemd/network/10-eth0.network.d",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSecondaryNetworkIPv4(tt.args.PrimaryNetworkInterface, tt.args.Label, tt.args.IPv4Address, tt.args.IPv4Gateway, tt.args.ou); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSecondaryNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecondaryNetwork_Add(t *testing.T) {
	type fields struct {
		PrimaryNetworkInterface string
		IPv4Address             string
		IPv4Gateway             string
		AddressLabel            string
		Osutil                  osutility.OSUtil
		networkInterfacePath    string
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
		{name: "add-net-interface", fields: struct {
			PrimaryNetworkInterface string
			IPv4Address             string
			IPv4Gateway             string
			AddressLabel            string
			Osutil                  osutility.OSUtil
			networkInterfacePath    string
		}{
			PrimaryNetworkInterface: "eth0", IPv4Address: "100.102.1.1",
			IPv4Gateway:          "",
			AddressLabel:         "1",
			Osutil:               osutility.NewDryRun(),
			networkInterfacePath: "/etc/systemd/network/10-eth0.network.d"},
			args: struct {
				ctx context.Context
			}{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secondaryNetwork := &SecondaryNetworkIPv4{
				PrimaryNetworkInterface: tt.fields.PrimaryNetworkInterface,
				IPv4Address:             tt.fields.IPv4Address,
				IPv4Gateway:             tt.fields.IPv4Gateway,
				AddressLabel:            tt.fields.AddressLabel,
				Osutil:                  tt.fields.Osutil,
				networkInterfacePath:    tt.fields.networkInterfacePath,
			}
			if err := secondaryNetwork.Add(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecondaryNetwork_Del(t *testing.T) {
	type fields struct {
		PrimaryNetworkInterface string
		IPv4Address             string
		IPv4Gateway             string
		AddressLabel            string
		Osutil                  osutility.OSUtil
		networkInterfacePath    string
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
		{name: "del-net-interface", fields: struct {
			PrimaryNetworkInterface string
			IPv4Address             string
			IPv4Gateway             string
			AddressLabel            string
			Osutil                  osutility.OSUtil
			networkInterfacePath    string
		}{
			PrimaryNetworkInterface: "eth0", IPv4Address: "100.102.1.1",
			IPv4Gateway:          "",
			AddressLabel:         "1",
			Osutil:               osutility.NewDryRun(),
			networkInterfacePath: "/etc/systemd/network/10-eth0.network.d"},
			args: struct {
				ctx context.Context
			}{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secondaryNetwork := &SecondaryNetworkIPv4{
				PrimaryNetworkInterface: tt.fields.PrimaryNetworkInterface,
				IPv4Address:             tt.fields.IPv4Address,
				IPv4Gateway:             tt.fields.IPv4Gateway,
				AddressLabel:            tt.fields.AddressLabel,
				Osutil:                  tt.fields.Osutil,
				networkInterfacePath:    tt.fields.networkInterfacePath,
			}
			if err := secondaryNetwork.Del(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Del() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecondaryNetwork_Update(t *testing.T) {
	type fields struct {
		PrimaryNetworkInterface string
		IPv4Address             string
		IPv4Gateway             string
		AddressLabel            string
		Osutil                  osutility.OSUtil
		networkInterfacePath    string
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
		{name: "update-network", fields: struct {
			PrimaryNetworkInterface string
			IPv4Address             string
			IPv4Gateway             string
			AddressLabel            string
			Osutil                  osutility.OSUtil
			networkInterfacePath    string
		}{
			PrimaryNetworkInterface: "eth0", IPv4Address: "100.102.1.1",
			IPv4Gateway:          "",
			AddressLabel:         "1",
			Osutil:               osutility.NewDryRun(),
			networkInterfacePath: "/etc/systemd/network/10-eth0.network.d"},
			args: struct {
				ctx context.Context
			}{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secondaryNetwork := &SecondaryNetworkIPv4{
				PrimaryNetworkInterface: tt.fields.PrimaryNetworkInterface,
				IPv4Address:             tt.fields.IPv4Address,
				IPv4Gateway:             tt.fields.IPv4Gateway,
				AddressLabel:            tt.fields.AddressLabel,
				Osutil:                  tt.fields.Osutil,
				networkInterfacePath:    tt.fields.networkInterfacePath,
			}
			if err := secondaryNetwork.Update(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
