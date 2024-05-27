package discovery

import (
	"fmt"

	"github.com/dogsays/mo/etcd"
)

type EtcdRegiser struct {
	etcd *etcd.Etcd
}

func NewEtcdRegiser(e *etcd.Etcd) *EtcdRegiser {
	return &EtcdRegiser{etcd: e}
}

func (e *EtcdRegiser) Regist(name, host string, port int) {

	k := fmt.Sprintf("%s/%s/%v", name, host, port)
	v := fmt.Sprintf("%s:%v", host, port)

	e.etcd.RegistAndKeepAlive(k, v)
}

func (e *EtcdRegiser) Revoke() {
	e.etcd.Close()
}
