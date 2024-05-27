package rpc1

import (
	"github.com/dogsays/mo/ut2"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

type ClientManager struct {
	dis         resolver.Builder
	clientConns *ut2.SyncMap[string, *grpc.ClientConn]
}

func NewClientManager(d resolver.Builder) *ClientManager {

	cm := &ClientManager{dis: d}
	cm.clientConns = ut2.NewSyncMap[string, *grpc.ClientConn]()

	resolver.Register(d)
	// encoding.RegisterCodec(JsonBytesCodec{})

	return cm
}

func (gx *ClientManager) Close() {

	gx.clientConns.Each(func(k string, v *grpc.ClientConn) {
		v.Close()
	})
}

func (gx *ClientManager) GetClient(serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, ok := gx.clientConns.Load(serviceName)
	if ok {
		return conn, nil
	}

	var kacp = keepalive.ClientParameters{
		// 如果没有 activity， 则每隔此值发送一个 ping 包
		// send pings every 10 seconds if there is no activity
		Time: 10 * time.Second,

		// 如果 ping ack 该值之内未返回则认为连接已断开
		// wait 1 second for ping ack before considering the connection dead
		Timeout: time.Second,

		// 如果没有 active 的 stream， 是否允许发送 ping
		// send pings even without active streams
		PermitWithoutStream: true,
	}

	// 连接失败后，客户端会进行重试连接。
	// 重试次数越多，等待下一次连接时间也会变长，但不能超过MaxDelay值。
	var backoff = grpc.BackoffConfig{
		MaxDelay: 2 * time.Second,
	}

	opts = append(opts,
		grpc.WithKeepaliveParams(kacp),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBackoffConfig(backoff),
		grpc.WithMaxMsgSize(40*1024*1024), // 默认4Mb,这里设置为40Mb
	)

	addr := gx.dis.Scheme() + "://" + serviceName

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	gx.clientConns.Store(serviceName, conn)

	return conn, nil
}

// grpc 内置 负载均衡器，客户端随机调用多个服务端
func RoundRobin() grpc.DialOption {
	return grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`)
}
