package network

import (
	"bytes"
	"context"
	"fmt"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

type NetworkInterface struct {
	InterfaceName string
	InterfaceType string
	IPv4Address   string
	IPv6Address   string
	Osutil        linux.OSUtil
}

const fileExits = "File exists"

func NewNetworkInterface(interfaceName, interfaceType, ipv4Address, ipv6Address string, ou linux.OSUtil) *NetworkInterface {
	return &NetworkInterface{
		InterfaceName: interfaceName,
		InterfaceType: interfaceType,
		IPv4Address:   ipv4Address,
		IPv6Address:   ipv6Address,
		Osutil:        ou,
	}
}

func (n NetworkInterface) Add(ctx context.Context) error {
	logger := log.From(ctx)
	_, _, err := n.Osutil.Exec().Command(ctx, constants.IPCommand, nil, "link", "add", "name", n.InterfaceName, "type", n.InterfaceType)
	if err != nil {
		if strings.Contains(err.Error(), fileExits) {
			logger.Info("network interface already exists", "name", n.InterfaceName)
		} else {
			return err
		}
	}

	_, _, err = n.Osutil.Exec().Command(ctx, constants.IPCommand, nil, "addr", "add", n.IPv4Address, "dev", n.InterfaceName)
	if err != nil {
		if strings.Contains(err.Error(), fileExits) {
			logger.Info("network interface already has given address",
				"name", n.InterfaceName,
				"address", n.IPv4Address)
		} else {
			return err
		}
	}
	_, _, err = n.Osutil.Exec().Command(ctx, constants.IPCommand, nil, "addr", "add", n.IPv6Address, "dev", n.InterfaceName)
	if err != nil {
		if strings.Contains(err.Error(), fileExits) {
			logger.Info("network interface already has given address",
				"name", n.InterfaceName,
				"address", n.IPv6Address)
		} else {
			return err
		}
	}
	return nil
}

func (n NetworkInterface) Del(ctx context.Context) error {
	_, _, err := n.Osutil.Exec().Command(ctx, constants.IPCommand, nil, "link", "delete", n.InterfaceName)
	if err != nil {
		return err
	}
	return nil
}

func (n NetworkInterface) Update(ctx context.Context) error {
	err := n.Del(ctx)
	if err != nil {
		return err
	}
	err = n.Add(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (n NetworkInterface) Get(ctx context.Context) (addresses map[string]bool, err error) {
	logger := log.From(ctx)
	logger.Info("Network interface name", "name", n.InterfaceName)
	netInterface, err := net.InterfaceByName(n.InterfaceName)
	if err != nil {
		logger.Error(err, "getting network interface by name")
		return nil, err
	}

	netInterfaceAddrs, err := netInterface.Addrs()
	if err != nil {
		logger.Error(err, "getting network interface addresses")
		return nil, err
	}
	results := make(map[string]bool)
	for _, netInterfaceAddr := range netInterfaceAddrs {
		results[netInterfaceAddr.String()] = true
	}
	return results, nil
}

type SecondaryNetworkIPv4 struct {
	PrimaryNetworkInterface string
	IPv4Address             string
	IPv4Gateway             string
	AddressLabel            string
	Osutil                  linux.OSUtil
	networkInterfacePath    string
}

var NetworkFolderPath = constants.DefaultNetworkFolder

var secondaryNetworkTemplate = `[Match]
Name={{ .PrimaryNetworkInterface }}
[Address]
Label={{ .PrimaryNetworkInterface }}:{{ .AddressLabel }}
Address={{ .IPv4Address }}`

type NetworkUtility interface {
	Add(ctx context.Context) error
	Del(ctx context.Context) error
	Update(ctx context.Context) error
	Get(ctx context.Context) (addresses map[string]bool, err error)
}

var mutex = &sync.Mutex{}

func NewSecondaryNetworkIPv4(primaryNetworkInterface string, label string,
	ipv4Address string,
	ipv4Gateway string,
	ou linux.OSUtil) *SecondaryNetworkIPv4 {
	interfacePath := fmt.Sprintf("%s/10-%s.network.d", NetworkFolderPath, primaryNetworkInterface)
	return &SecondaryNetworkIPv4{
		PrimaryNetworkInterface: primaryNetworkInterface,
		AddressLabel:            label,
		IPv4Address:             ipv4Address,
		IPv4Gateway:             ipv4Gateway,
		Osutil:                  ou,
		networkInterfacePath:    interfacePath,
	}
}

func (secondaryNetwork *SecondaryNetworkIPv4) Add(ctx context.Context) error {
	mutex.Lock()
	defer mutex.Unlock()
	exists, err := secondaryNetwork.Osutil.Filesystem().Exists(ctx, secondaryNetwork.networkInterfacePath)
	if err != nil {
		return err
	}
	if !exists {
		err = secondaryNetwork.Osutil.Filesystem().MkdirAll(ctx, secondaryNetwork.networkInterfacePath, constants.DirPerm)
		if err != nil {
			return err
		}
	}
	parse, err := template.New("private-ip").Parse(secondaryNetworkTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = parse.Execute(&buf, secondaryNetwork)
	if err != nil {
		return err
	}
	secondaryIPFilePath := filepath.Join(secondaryNetwork.networkInterfacePath, constants.SecondaryIPfileName)
	f, err := secondaryNetwork.Osutil.Filesystem().OpenFileWithPermission(ctx, secondaryIPFilePath, os.O_RDWR|os.O_CREATE, constants.FilePerm)
	defer func(f *os.File) {
		closeErr := f.Close()
		if closeErr != nil {
			return
		}
	}(f)
	if err != nil {
		return err
	}
	if err := secondaryNetwork.Osutil.Filesystem().WriteFile(ctx, secondaryIPFilePath, buf.Bytes(), constants.FilePerm); err != nil {
		return fmt.Errorf("write secondary ip config file: %w", err)
	}
	_, _, err = secondaryNetwork.Osutil.Exec().Command(ctx, "ifconfig", nil,
		fmt.Sprintf("%s:%s", secondaryNetwork.PrimaryNetworkInterface, secondaryNetwork.AddressLabel),
		constants.PrivateIPv4AddressMaskDigest)
	if err != nil {
		return err
	}
	_, _, err = secondaryNetwork.Osutil.Exec().Command(ctx, "ifconfig", nil, secondaryNetwork.PrimaryNetworkInterface, "up")
	if err != nil {
		return err
	}
	err = secondaryNetwork.Osutil.Systemd().Restart(ctx, constants.SystemdNetworkProcess)
	// sleeping for 10s for the new secondary interface to come-up
	time.Sleep(10 * time.Second)
	if err != nil {
		return err
	}
	return nil
}

func (secondaryNetwork *SecondaryNetworkIPv4) Del(ctx context.Context) error {
	return secondaryNetwork.Osutil.Filesystem().RemoveAll(ctx, secondaryNetwork.networkInterfacePath)
}

func (secondaryNetwork *SecondaryNetworkIPv4) Update(ctx context.Context) error {
	err := secondaryNetwork.Del(ctx)
	if err != nil {
		return err
	}
	err = secondaryNetwork.Add(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (secondaryNetwork *SecondaryNetworkIPv4) Get(ctx context.Context) (map[string]bool, error) {
	logger := log.From(ctx)
	logger.Info("Network interface name", "name", secondaryNetwork.PrimaryNetworkInterface)
	netInterface, err := net.InterfaceByName(secondaryNetwork.PrimaryNetworkInterface)
	if err != nil {
		logger.Error(err, "getting network interface by name")
		return nil, err
	}

	netInterfaceAddrs, err := netInterface.Addrs()
	if err != nil {
		logger.Error(err, "getting network interface addresses")
		return nil, err
	}
	results := make(map[string]bool)
	for _, netInterfaceAddr := range netInterfaceAddrs {
		results[netInterfaceAddr.String()] = true
	}
	return results, nil
}
