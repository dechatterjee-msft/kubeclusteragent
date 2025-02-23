package linux

import (
	"fmt"
	"os"
	"strings"
)

type Host interface {
	GetHostname() (string, error)
}

type LiveHost struct{}
type FakeHost struct{}
type FakeHostWithErr struct{}

func (l *LiveHost) GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return strings.ToLower(hostname), nil
}

func (l *FakeHost) GetHostname() (string, error) {
	return "testutil", nil
}

func (l *FakeHostWithErr) GetHostname() (string, error) {
	return "", fmt.Errorf("error for testing")
}
