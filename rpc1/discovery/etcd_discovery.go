package discovery

import (
	"net/url"
	"sync"

	"github.com/dogsays/mo/etcd"
	"github.com/dogsays/mo/logger"

	"google.golang.org/grpc/resolver"
)

type EtcdDiscovery struct {
	etcd *etcd.Etcd
}

func NewEtcdDiscovery(etcd *etcd.Etcd) *EtcdDiscovery {
	return &EtcdDiscovery{etcd: etcd}
}

func (mb *EtcdDiscovery) Scheme() string {
	return "etcd"
}

func (e *EtcdDiscovery) GetAddr(service string) ([]string, error) {
	return e.etcd.GetPrefix(service)
}

func (mb *EtcdDiscovery) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	r := newEtcdResolver(mb.etcd, target.URL, cc)
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

type etcdResolver struct {
	// etcd://serviceName
	// scheme://host/path
	URL     url.URL
	cc      resolver.ClientConn
	closeCh chan struct{}
	etcd    *etcd.Etcd
	gw      sync.WaitGroup
}

func newEtcdResolver(etcd *etcd.Etcd, u url.URL, cc resolver.ClientConn) *etcdResolver {
	r := &etcdResolver{
		URL:     u,
		cc:      cc,
		etcd:    etcd,
		closeCh: make(chan struct{}, 1),
	}

	r.gw.Add(1)
	go func() {
		r.etcd.WatchPrefixAndBlock(r.serviceName(), r.closeCh, func() {
			logger.Info("rpc地址发生改变")
			r.ResolveNow(resolver.ResolveNowOptions{})
		})

		logger.Info("watcher 退出")
		r.gw.Done()
	}()

	return r
}

func (r *etcdResolver) Close() {
	r.closeCh <- struct{}{}
	r.gw.Wait()
}

func (r *etcdResolver) serviceName() string {
	return r.URL.Host
}

func (r *etcdResolver) ResolveNow(opt resolver.ResolveNowOptions) {
	r.cc.UpdateState(resolver.State{Addresses: r.build()})
}

func (r *etcdResolver) build() (ret []resolver.Address) {

	resp, err := r.etcd.GetPrefix(r.serviceName())
	if err != nil {
		logger.Info("etcd Get", err)
		return
	}

	for _, v := range resp {
		ret = append(ret, resolver.Address{Addr: v})
	}

	return
}
