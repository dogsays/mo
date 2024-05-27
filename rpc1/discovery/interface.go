package discovery

import (
	"google.golang.org/grpc/resolver"
)

type Discovery interface {
	resolver.Builder
	GetAddr(name string) ([]string, error)
}

type Register interface {
	Regist(name, host string, port int)
	Revoke()
}

type PortProvider interface {
	GetPort(name string) (int, error)
	Close()
}
