package cfgmgr

import (
	"context"
	"errors"

	"github.com/dogsays/mo/etcd"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdWatcher struct {
	Etcd    *etcd.Etcd
	watcher clientv3.Watcher
}

func NewEtcdWatcher(et *etcd.Etcd) *EtcdWatcher {
	return &EtcdWatcher{
		Etcd: et,
	}
}

const etcd_prefix = "#config/"

func (m *EtcdWatcher) basename(file string) string {
	return file[len(etcd_prefix):]
}

func (m *EtcdWatcher) WriteFile(name string, content []byte) error {
	_, err := m.Etcd.Put(context.TODO(), etcd_prefix+name, string(content))
	return err
}

func (m *EtcdWatcher) IsExists(name string) bool {
	resp, err := m.Etcd.Get(context.TODO(), etcd_prefix+name)
	if err != nil {
		return false
	}

	return resp.Count != 0
}

func (m *EtcdWatcher) ReadFile(name string) ([]byte, error) {
	resp, err := m.Etcd.Get(context.TODO(), etcd_prefix+name)
	if err != nil {
		return nil, err
	}

	if resp.Count > 0 {
		return resp.Kvs[0].Value, nil
	}

	return nil, errors.New("file empty")
}

func (m *EtcdWatcher) GetFileNames() ([]string, error) {
	resp, err := m.Etcd.Get(context.TODO(), etcd_prefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for _, v := range resp.Kvs {
		ret = append(ret, m.basename(string(v.Key)))
	}
	return ret, nil
}

func (m *EtcdWatcher) Init() error {
	m.watcher = m.Etcd.NewWatcher()
	return nil
}

func (m *EtcdWatcher) Stop() error {
	return m.watcher.Close()
}

func (m *EtcdWatcher) Start(closeCh chan struct{}, dispatch func(string, []byte)) {

	ch := m.watcher.Watch(context.TODO(), etcd_prefix, clientv3.WithPrefix())

	for {
		select {
		case resp := <-ch:
			for _, v := range resp.Events {
				if v.Type == mvccpb.PUT {
					dispatch(m.basename(string(v.Kv.Key)), v.Kv.Value)
				}
			}

		case <-closeCh:
			return
		}
	}
}
