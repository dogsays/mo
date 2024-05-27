package env

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"

	"github.com/dogsays/mo/logger"
	"github.com/dogsays/mo/rpc1/discovery"
)

// 使用环境变量的方式配置
type envStruct struct {
	JwtKey        string // jwt密码
	ConfigWatcher string // 配置读取模式
	Logger        string // logger行为 daily | console
	Pprof         string
}

var Default = envStruct{
	ConfigWatcher: "file://config", // etcd://127.0.0.1:8080
	Pprof:         "",
	Logger:        "console",
}

func initFromEnv() {
	val := reflect.ValueOf(&Default).Elem()
	typ := reflect.TypeOf(Default)

	for i := 0; i < val.NumField(); i++ {
		vf := val.Field(i)
		tf := typ.Field(i)

		ev := os.Getenv(tf.Name)

		if ev != "" {
			vf.SetString(ev)
		}
	}
}

func init() {
	initFromEnv()

	if Default.Pprof != "" {
		sp := discovery.NewSystemPortProvider()
		p, _ := sp.GetPort("_")
		addr := fmt.Sprintf(":%d", p)

		go http.ListenAndServe(addr, nil)

		logger.Info("Pprof 监听在", addr)
	}
}
