package etcd

import (
	"context"
	"sync"
	"time"

	"github.com/dogsays/mo/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Etcd struct {
	*clientv3.Client

	lease   clientv3.Lease
	LeaseID clientv3.LeaseID
	gw      sync.WaitGroup
	ch      chan struct{}
}

func NewEtcd(addr string) (*Etcd, error) {

	if addr == "" {
		addr = "127.0.0.1:2379"
	}

	timeout := 3 * time.Second

	// 把etcd的log 写入到我们的logger中
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(logger.Default()), zapcore.WarnLevel)
	zapLogger := zap.New(core)

	config := clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: timeout,
		Logger:      zapLogger,
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.TODO(), timeout)

	_, err = client.Status(ctx, config.Endpoints[0])
	if err != nil {
		return nil, err
	}

	e := &Etcd{}
	e.Client = client
	e.ch = make(chan struct{}, 1)

	return e, nil
}

func (e *Etcd) Close() {
	e.ch <- struct{}{}
	e.gw.Wait()

	e.Client.Close()
}

func (e *Etcd) keepAlive() {
	e.gw.Add(1)

	var err error

	ticker := time.NewTicker(5 * time.Second) // 5秒 续约

	defer func() {

		ticker.Stop()

		_, err = e.lease.Revoke(context.TODO(), e.LeaseID)
		if err != nil {
			logger.Info("撤销租约错误", err)
		}

		err = e.lease.Close()
		if err != nil {
			logger.Info("关闭租约错误", err)
		}
		e.gw.Done()
		logger.Info("退出etcd keepalive")
	}()

	for {
		select {
		case <-ticker.C:
			_, err = e.lease.KeepAliveOnce(context.TODO(), e.LeaseID)
			if err != nil {
				logger.Info("Etcd 续约错误", err)
			}

		case <-e.ch:
			return
		}
	}
}

func (e *Etcd) RegistAndKeepAlive(k, value string) error {
	e.lease = clientv3.NewLease(e.Client)

	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	resp, err := e.lease.Grant(ctx, 20) // 20 秒 过期
	if err != nil {
		return err
	}

	e.LeaseID = resp.ID

	_, err = e.Put(context.TODO(), k, value, clientv3.WithLease(e.LeaseID))
	if err != nil {
		return err
	}

	go e.keepAlive()

	return nil
}

func (e *Etcd) NewWatcher() clientv3.Watcher {
	return clientv3.NewWatcher(e.Client)
}

func (e *Etcd) WatchPrefixAndBlock(prefix string, closeCh chan struct{}, fn func()) {
	watcher := clientv3.NewWatcher(e.Client)

	defer func() {
		err := watcher.Close()
		if err != nil {
			logger.Info("Etcd Watcher 关闭错误", err)
		}
	}()

	ch := watcher.Watch(context.TODO(), prefix, clientv3.WithPrefix())

	for {
		select {
		case <-ch:
			fn()

		case <-closeCh:
			return
		}
	}
}

func (e *Etcd) GetPrefix(prefix string) (ret []string, err error) {
	resp, err := e.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, v := range resp.Kvs {
		ret = append(ret, string(v.Value))
	}

	return ret, nil
}
