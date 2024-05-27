package rpc1

import (
	"encoding/json"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

/////

// JSON: 这种方式会把int64 转为 string。不采取
// opt := protojson.MarshalOptions{
// 	EmitUnpopulated: true, // 不处理 omitempty
// 	UseEnumNumbers:  true, // 枚举用数字
// }
// return opt.Marshal(pbMsg)

/////

type JsonCodec struct{}

func (c JsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
func (c JsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
func (c JsonCodec) Name() string {
	return "myxjson" //
}

// 直接转发无需解析
type JsonBytesCodec struct{}

func (c JsonBytesCodec) Marshal(v interface{}) ([]byte, error) {
	switch msg := v.(type) {
	case []byte:
		return msg, nil
	default:
		return json.Marshal(v)
	}
}
func (c JsonBytesCodec) Unmarshal(data []byte, v interface{}) error {
	switch v.(type) {
	case *[]byte:
		val := reflect.ValueOf(v)
		val = val.Elem()
		val.Set(reflect.ValueOf(data))
	default:
		return json.Unmarshal(data, v)
	}
	return nil
}

func (c JsonBytesCodec) Name() string {
	return "myxjson" //
}

// 客户端添加到DialOption中，则默认使用指定协议传输
func UseCodec(codec encoding.Codec) grpc.DialOption {
	return grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codec.Name()))
}
