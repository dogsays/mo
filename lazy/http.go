package lazy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/dogsays/mo/logger"
	"github.com/dogsays/mo/rpc1"
)

type bytesMarshaler struct {
	Bytes []byte
}

func (b *bytesMarshaler) MarshalJSON() ([]byte, error) {
	return b.Bytes, nil
}

func MakeRouteMsg(value any, route any) []byte {
	return MakeMsg(value, map[string]any{"id": route})
}

func WrapMsg(value any) map[string]any {
	resp := map[string]any{}

	switch msg := value.(type) {
	case error:
		resp["error"] = rpc1.GetErrorMessage(msg)

	case []byte: // 这里我们认为他是json
		resp["data"] = &bytesMarshaler{Bytes: msg}

	default:
		resp["data"] = value
	}

	return resp
}

func MakeMsg(value any, override map[string]any) []byte {
	resp := WrapMsg(value)

	// unnecessary nil check around range
	for k, v := range override {
		resp[k] = v
	}

	buf, err := json.Marshal(resp)
	if err != nil {
		logger.Info("json错误")
		logger.Info(debug.Stack())
	}

	return buf
}

func HttpReturn(w http.ResponseWriter, value interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(MakeMsg(value, nil))
}

type handler struct {
	val reflect.Value
	typ reflect.Type
}

type HttpContext struct {
	context.Context
	Request *http.Request
}

func (h *handler) doHandleHttp(path string, server any, httpReq *http.Request) any {

	defer func() {
		// 捕捉崩溃。并返回错误到调用者
		if x := recover(); x != nil {
			logger.Info("PANIC", httpReq.URL.Path, x)
			logger.Info(debug.Stack())
		}
	}()

	mtype, ok := h.typ.MethodByName(path)

	if !ok {
		return errors.New("no method")
	}

	bytes, err := io.ReadAll(httpReq.Body)
	if err != nil {
		return err
	}

	reqType := mtype.Type.In(2)
	req := reflect.New(reqType.Elem())
	if err := json.Unmarshal(bytes, req.Interface()); err != nil {
		return err
	}

	ctx := &HttpContext{
		Context: context.Background(),
		Request: httpReq,
	}

	vs := make([]reflect.Value, 2)
	vs[0] = reflect.ValueOf(ctx)
	vs[1] = req

	method := h.val.MethodByName(path)
	ret := method.Call(vs)

	if !ret[1].IsNil() {
		return ret[1].Interface()
	}

	return ret[0].Interface()
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := h.doHandleHttp(strings.TrimLeft(r.URL.Path, "/"), nil, r)
	HttpReturn(w, resp)
}

func GrpcHttpHandleFunc(grpcServer any) http.HandlerFunc {
	h := &handler{
		val: reflect.ValueOf(grpcServer),
		typ: reflect.TypeOf(grpcServer),
	}
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func StartGrpcTestHttpServer(addr string, grpcServer any) {
	http.Handle("/", GrpcHttpHandleFunc(grpcServer))
	http.ListenAndServe(addr, nil)
}
