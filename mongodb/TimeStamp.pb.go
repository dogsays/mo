// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: pb/Time.proto

package mongodb

import (
	reflect "reflect"
	sync "sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type TimeStamp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Seconds int64 `protobuf:"varint,1,opt,name=Seconds,proto3" `
	Nanos   int32 `protobuf:"varint,2,opt,name=Nanos,proto3" `
}

///

// New constructs a new Timestamp from the provided time.Time.
func NewTimeStamp(t time.Time) *TimeStamp {
	r := &TimeStamp{}
	r.SetTime(t)
	return r
}

// AsTime converts x to a time.Time.
func (x *TimeStamp) AsTime() time.Time {
	return time.Unix(int64(x.GetSeconds()), int64(x.GetNanos())).Local()
}

func (x *TimeStamp) SetTime(t time.Time) {
	x.Seconds = int64(t.Unix())
	x.Nanos = int32(t.Nanosecond())
}

func (x *TimeStamp) MarshalJSON() ([]byte, error) {
	return x.AsTime().MarshalJSON()
}

func (x *TimeStamp) UnmarshalJSON(b []byte) error {
	var t time.Time

	err := t.UnmarshalJSON(b)
	if err != nil {
		return err
	}

	x.Seconds = int64(t.Unix())
	x.Nanos = int32(t.Nanosecond())

	return nil
}

type TimeStampCodec struct{}

var tTimeStamp = reflect.TypeOf(&TimeStamp{})

// DecodeValue is the ValueDecoderFunc for time.Time.
func (tc *TimeStampCodec) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tTimeStamp {
		return bsoncodec.ValueEncoderError{Name: "TimeStampEncodeValue", Types: []reflect.Type{tTimeStamp}, Received: val}
	}

	var t time.Time
	// valueof(t) 是t的拷贝，我们这里反射必须先取地址
	err := timeCodec.DecodeValue(dc, vr, reflect.ValueOf(&t).Elem())
	if err != nil {
		return err
	}

	ts := val.Interface().(*TimeStamp)
	ts.SetTime(t)

	return nil
}

// EncodeValue is the ValueEncoderFunc for time.TIme.
func (tc *TimeStampCodec) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tTimeStamp {
		return bsoncodec.ValueEncoderError{Name: "TimeStampEncodeValue", Types: []reflect.Type{tTimeStamp}, Received: val}
	}
	t := val.Interface().(*TimeStamp).AsTime()
	return timeCodec.EncodeValue(ec, vw, reflect.ValueOf(t))
}

///

func (x *TimeStamp) Reset() {
	*x = TimeStamp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_TimeStamp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimeStamp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimeStamp) ProtoMessage() {}

func (x *TimeStamp) ProtoReflect() protoreflect.Message {
	mi := &file_pb_TimeStamp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimeStamp.ProtoReflect.Descriptor instead.
func (*TimeStamp) Descriptor() ([]byte, []int) {
	return file_pb_TimeStamp_proto_rawDescGZIP(), []int{0}
}

func (x *TimeStamp) GetSeconds() int64 {
	if x != nil {
		return x.Seconds
	}
	return 0
}

func (x *TimeStamp) GetNanos() int32 {
	if x != nil {
		return x.Nanos
	}
	return 0
}

var File_pb_TimeStamp_proto protoreflect.FileDescriptor

var file_pb_TimeStamp_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x62, 0x2f, 0x54, 0x69, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3b, 0x0a, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x18, 0x0a, 0x07, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x07, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x4e,
	0x61, 0x6e, 0x6f, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x4e, 0x61, 0x6e, 0x6f,
	0x73, 0x42, 0x16, 0x5a, 0x14, 0x64, 0x75, 0x63, 0x6b, 0x2f, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64,
	0x62, 0x3b, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_pb_TimeStamp_proto_rawDescOnce sync.Once
	file_pb_TimeStamp_proto_rawDescData = file_pb_TimeStamp_proto_rawDesc
)

func file_pb_TimeStamp_proto_rawDescGZIP() []byte {
	file_pb_TimeStamp_proto_rawDescOnce.Do(func() {
		file_pb_TimeStamp_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_TimeStamp_proto_rawDescData)
	})
	return file_pb_TimeStamp_proto_rawDescData
}

var file_pb_TimeStamp_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_pb_TimeStamp_proto_goTypes = []interface{}{
	(*TimeStamp)(nil), // 0: TimeStamp
}
var file_pb_TimeStamp_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pb_TimeStamp_proto_init() }
func file_pb_TimeStamp_proto_init() {
	if File_pb_TimeStamp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pb_TimeStamp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimeStamp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pb_TimeStamp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pb_TimeStamp_proto_goTypes,
		DependencyIndexes: file_pb_TimeStamp_proto_depIdxs,
		MessageInfos:      file_pb_TimeStamp_proto_msgTypes,
	}.Build()
	File_pb_TimeStamp_proto = out.File
	file_pb_TimeStamp_proto_rawDesc = nil
	file_pb_TimeStamp_proto_goTypes = nil
	file_pb_TimeStamp_proto_depIdxs = nil
}
