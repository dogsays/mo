package discovery

import (
	"net"
)

type SystemPortProvider struct {
}

func NewSystemPortProvider() *SystemPortProvider {
	return &SystemPortProvider{}
}

func (e *SystemPortProvider) Close() {
}

func (e *SystemPortProvider) GetPort(service string) (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
