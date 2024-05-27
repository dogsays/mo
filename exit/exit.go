package exit

import (
	"os"
	"os/signal"
	"sort"
	"sync/atomic"
	"syscall"

	"github.com/dogsays/mo/logger"
)

var ch chan os.Signal

type namefn struct {
	name     string
	fn       func()
	priority int
}

// var arr *ut2.ArrayAnySync[func()]
var arr []namefn

var exiting int32 = 0

// 在程序 kill之前通知你，注意不要使用 kill -9
func Callback(name string, fn func()) {
	CallbackWithPriority(name, fn, 0)
}

func CallbackWithPriority(name string, fn func(), priority int) {
	arr = append(arr, namefn{name: name, fn: fn, priority: priority})
}

type CloseAble interface {
	Close()
}

func Close(name string, o CloseAble) {
	CallbackWithPriority(name, func() { o.Close() }, 0)
}

func CloseWithPriority(name string, o CloseAble, priority int) {
	CallbackWithPriority(name, func() { o.Close() }, priority)
}

type WaitAble interface {
	Wait()
}

func Wait(name string, gp WaitAble) {
	Callback(name, func() { gp.Wait() })
}

func Exit() {

	v := atomic.LoadInt32(&exiting)
	if v == 1 {
		return
	}

	atomic.StoreInt32(&exiting, 1)

	logger.Info("################ 程序退出开始 ##############")

	sort.Slice(arr, func(i, j int) bool { return arr[i].priority > arr[j].priority })

	// for i := len(arr) - 1; i > 0; i-- {
	for i := 0; i < len(arr); i++ {
		nf := arr[i]
		logger.Info(nf.name)
		nf.fn()
		logger.Info(nf.name, "ok")
	}

	logger.Info("################ 程序退出结束 ##############")

	os.Exit(0)
}

func init() {

	ch = make(chan os.Signal)

	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-ch
		Exit()
	}()
}
