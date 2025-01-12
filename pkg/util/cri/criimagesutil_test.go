package cri

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/magiconair/properties"
	"kubeclusteragent/pkg/constants"
	"os"
	"reflect"
	"testing"
)

var propFileName = "versions.properties"

type FakeContainerdClient struct {
	Address    string
	Namespace  string
	Connection *containerd.Client
}

func (f FakeContainerdClient) ListImages(ctx context.Context, imageTag string) ([]string, error) {
	return []string{"/kube-apiserver:v1.26.5",
		"/kube-controller-manager:v1.26.5",
		"/kube-proxy:v1.26.5",
		"/kube-scheduler:v1.26.5"}, nil
}

func (f FakeContainerdClient) DeleteImage(ctx context.Context, image string) error {
	return nil
}

func (f FakeContainerdClient) Close(ctx context.Context) error {
	return nil
}

func generateFile(i int) error {
	prop := properties.NewProperties()
	f, err := os.OpenFile(propFileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if i == 0 {
		_, _, err = prop.Set(constants.CoreDNSPropFile, "1.9.3")
		if err != nil {
			return err
		}
		_, _, err = prop.Set(constants.EtcdPropFile, "3.5.6")
		if err != nil {
			return err
		}
		_, _, err = prop.Set(constants.PausePropFile, "3.7")
		if err != nil {
			return err
		}
	} else if i == 1 {
		_, _, err = prop.Set(constants.CoreDNSPropFile, "1.9.3")
		if err != nil {
			return err
		}
		_, _, err = prop.Set(constants.EtcdPropFile, "3.5.6")
		if err != nil {
			return err
		}
	} else if i == 2 {
		_, _, err = prop.Set(constants.CoreDNSPropFile, "1.9.3")
		if err != nil {
			return err
		}
		_, _, err = prop.Set(constants.PausePropFile, "3.7")
		if err != nil {
			return err
		}
	} else if i == 3 {
		_, _, err = prop.Set(constants.EtcdPropFile, "3.5.6")
		if err != nil {
			return err
		}
		_, _, err = prop.Set(constants.PausePropFile, "3.7")
		if err != nil {
			return err
		}
	}
	_, err = f.WriteString(prop.String())
	if err != nil {
		return err
	}
	return nil
}

func TestGetK8sControlPlaneImagesFromPropertiesFile(t *testing.T) {
	propFileLocation = propFileName
	tests := []struct {
		name     string
		want     map[string]string
		wantErr  bool
		skipfile bool
	}{
		{name: "happy", want: map[string]string{
			constants.EtcdImage:    "v3.5.6",
			constants.CoreDNSImage: "v1.9.3",
			constants.PauseImage:   "3.7",
		}, wantErr: false},
		{name: "PauseMissing", want: nil, wantErr: true},
		{name: "EtcdMissing", want: nil, wantErr: true},
		{name: "CoreDNSMissing", want: nil, wantErr: true},
		{name: "PropFileMissing", want: nil, wantErr: true, skipfile: true},
	}
	for i := 0; i < len(tests); i++ {
		t.Run(tests[i].name, func(t *testing.T) {
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {

				}
			}(propFileName)
			if !tests[i].skipfile {
				err := generateFile(i)
				if err != nil {
					t.Fail()
				}
			}
			got, err := GetK8sControlPlaneImagesFromPropertiesFile()
			if (err != nil) != tests[i].wantErr {
				t.Errorf("GetK8sControlPlaneImagesFromPropertiesFile() error = %v, wantErr %v", err, tests[i].wantErr)
				return
			}
			if !reflect.DeepEqual(got, tests[i].want) {
				t.Errorf("GetK8sControlPlaneImagesFromPropertiesFile() got = %v, want %v", got, tests[i].want)
			}
		})

	}
}
