package rpc1

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dogsays/mo/logger"
	"github.com/dogsays/mo/rpc1/discovery"
	"github.com/dogsays/mo/ut2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// grpc健康检查
// https://zhuanlan.zhihu.com/p/451936968

// GRPC 优雅的停止
// https://zhuanlan.zhihu.com/p/438808354

// gRPC源码——keepalive
// https://zhuanlan.zhihu.com/p/530266840

type ServerOption struct {
	Name             string // [必填]本服务名称
	Host             string // 用于给别人调用本服务的地址，会注册到注册中心，默认="127.0.0.1"
	UnaryInterceptor grpc.UnaryServerInterceptor
}

type Server struct {
	*grpc.Server

	Config         *ServerOption
	Register       discovery.Register
	PortProvider   discovery.PortProvider
	ServiceDescMap ut2.IMap[string, *DescAndImpl]
}

type DescAndImpl struct {
	Desc *grpc.ServiceDesc
	Impl any
}

func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.ServiceDescMap.Store(desc.ServiceName, &DescAndImpl{Desc: desc, Impl: impl})
	s.Server.RegisterService(desc, impl)
}

func (s *Server) Close() {
	s.Register.Revoke()
	s.Server.GracefulStop()
}

func NewServer(cfg *ServerOption, p discovery.PortProvider, r discovery.Register) *Server {
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}

	gx := &Server{
		Config:         cfg,
		Register:       r,
		PortProvider:   p,
		ServiceDescMap: ut2.NewSyncMap[string, *DescAndImpl](),
	}

	var kaep = keepalive.EnforcementPolicy{
		// 客户端ping的间隔应该不小于这个时长，默认是5分钟。
		MinTime: 5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		// 服务端是否允许在没有RPC调用时发送PING，默认不允许。
		// 在不允许的情况下，客户端发送了PING，服务端将发送GOAWAY帧，关闭连接。
		PermitWithoutStream: true, // Allow pings even when there are no active streams
	}

	// 默认值在 grpc@v1.50.1/internal/transport/defaults.go
	var kasp = keepalive.ServerParameters{

		// 当连接处于idle的时长超过 MaxConnectionIdle时，服务端就发送GOAWAY，关闭连接。默认值为无限大
		// 如果一个 client 空闲超过该值, 发送一个 GOAWAY, 为了防止同一时间发送大量 GOAWAY,
		// 会在此时间间隔上下浮动 10%, 例如设置为15s，即 15+1.5 或者 15-1.5
		// MaxConnectionIdle: 15 * time.Second,

		// 一个连接只能使用 MaxConnectionAge 这么长的时间，服务端就会关闭这个连接。默认无限大
		// 如果任意连接存活时间超过该值, 发送一个 GOAWAY
		// MaxConnectionAge: 10 * time.Minute,

		// 服务端优雅关闭连接时长（关闭grpc的时候等待未完成的rpc的时长）
		// 在强制关闭连接之间, 允许有该值的时间完成 pending 的 rpc 请求
		// Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		MaxConnectionAgeGrace: 15 * time.Second,

		// 如果一个 client 空闲超过该值, 则发送一个 ping 请求
		// Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Time: 5 * time.Second,

		// 默认值为20秒
		// 如果 ping 请求该时间段内未收到回复, 则认为该连接已断开
		// Wait 1 second for the ping ack before assuming the connection is dead
		Timeout: 3 * time.Second,
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	}

	if cfg.UnaryInterceptor != nil {
		opts = append(opts, grpc.UnaryInterceptor(cfg.UnaryInterceptor))
	} else {
		opts = append(opts, grpc.UnaryInterceptor(DefaultUnaryInterceptor))
	}

	gx.Server = grpc.NewServer(opts...)

	// 服务端默认开启json格式
	encoding.RegisterCodec(JsonCodec{})

	return gx
}

// 我们自己报的错 使用996 这个错误码
const not_grpc_err_code = codes.Code(996)

type gprcStatusError interface {
	GRPCStatus() *status.Status
}

func Err(s any) error {
	switch m := s.(type) {
	case string:
		return status.Error(not_grpc_err_code, m)
	case error:
		// 已经是通过status.Error创建的错误，则不需要包装了
		if _, ok := m.(gprcStatusError); !ok {
			return status.Error(not_grpc_err_code, m.Error())
		}
		return m
	default:
		return status.Error(not_grpc_err_code, fmt.Sprintf("%v", m))
	}
}

func GetErrorCode(err error) int {

	if se, ok := err.(gprcStatusError); ok {
		return int(se.GRPCStatus().Code())
	}

	return int(codes.Unknown)
}

// 获取错误消息。逻辑错误不显示错误码。
func GetErrorMessage(err error) string {

	if se, ok := err.(gprcStatusError); ok {
		stat := se.GRPCStatus()
		// 逻辑错误
		if stat.Code() == not_grpc_err_code {
			return fmt.Sprintf("%v", stat.Message())
		}

		return fmt.Sprintf("%v:%v", stat.Code(), stat.Message())
	}

	return err.Error()
}

func tojson(a any) string {
	bytes, _ := json.Marshal(a)
	return string(bytes)
}

func UnaryInterceptorWithLogPrefix(logpfx string, skipLogResp bool, ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	logpfx += "@" + info.FullMethod

	logger.Info("GRPC Begin", logpfx, tojson(req))

	before := time.Now()

	defer func() {
		use := fmt.Sprintf("GrpcUse=%v", time.Since(before))
		use = strings.ReplaceAll(use, "µ", "u") // loki 查不到 这个字符。特殊处理

		// 捕捉grpc崩溃。并返回错误到调用者
		if x := recover(); x != nil {
			logger.Err("GRPC Panic", logpfx, use, x)
			logger.Err(string(debug.Stack()))
			err = Err(fmt.Sprintf("PANIC:%v", x))
		} else if err != nil {
			logger.Err("GRPC Error", logpfx, use, GetErrorMessage(err))
		} else {
			respStr := "<skip>"
			if !skipLogResp {
				respStr = tojson(resp)
			}

			logger.Info("GRPC End  ", logpfx, use, respStr)
		}
	}()

	resp, err = handler(ctx, req)

	if err != nil {
		return nil, Err(err)
	}

	return
}

var DefaultUnaryInterceptor grpc.UnaryServerInterceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return UnaryInterceptorWithLogPrefix(ut2.RandomString(10), false, ctx, req, info, handler)
}

func (gx *Server) Serve() error {

	reflection.Register(gx.Server)

	name := gx.Config.Name
	host := gx.Config.Host
	port, err := gx.PortProvider.GetPort(gx.Config.Name)
	if err != nil {
		return err
	}

	gx.Register.Regist(name, host, port)

	addr := fmt.Sprintf(":%d", port)

	logger.Info("start grpc", name, addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return gx.Server.Serve(listener)
}
