// 在proto文件中使用ObjectID

package mongodb

import (
	"encoding/hex"
	"encoding/json"
	reflect "reflect"
	sync "sync"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	// "gopkg.in/mgo.v2/bson"
	// "gopkg.in/mgo.v2/bson"
)

type ObjectID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=Data,proto3" `
}

///

func NewObjectID() *ObjectID {
	oid := primitive.NewObjectID()
	return &ObjectID{Data: oid[:]}
}

func ObjectIDFromHex(s string) (*ObjectID, error) {
	if len(s) != 24 {
		return nil, primitive.ErrInvalidHex
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var oid [12]byte
	copy(oid[:], b)

	return &ObjectID{Data: oid[:]}, nil
}

func (id *ObjectID) Hex() string {
	var buf [24]byte
	hex.Encode(buf[:], id.Data[:])
	return string(buf[:])
}

func (id *ObjectID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.Hex())
}
func (id *ObjectID) UnmarshalJSON(b []byte) error {

	var oid primitive.ObjectID
	err := oid.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	id.Data = oid[:]
	return nil
}

type ObjectIDCodec struct{}

var tOID = reflect.TypeOf(&ObjectID{})

// DecodeValue is the ValueDecoderFunc for time.Time.
func (tc *ObjectIDCodec) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tOID {
		return bsoncodec.ValueEncoderError{Name: "ObjectIDEncodeValue", Types: []reflect.Type{tOID}, Received: val}
	}

	oid, err := vr.ReadObjectID()
	if err != nil {
		return err
	}

	dest := val.Interface().(*ObjectID)
	dest.Data = oid[:]
	return nil
}

// EncodeValue is the ValueEncoderFunc for time.TIme.
func (tc *ObjectIDCodec) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tOID {
		return bsoncodec.ValueEncoderError{Name: "ObjectIDEncodeValue", Types: []reflect.Type{tOID}, Received: val}
	}

	id := val.Interface().(*ObjectID)
	var oid primitive.ObjectID
	copy(oid[:], id.Data[:])

	return vw.WriteObjectID(primitive.ObjectID(oid))
}

// func (id *ObjectID) MarshalBSON() ([]byte, error) {
// 	var oid primitive.ObjectID
// 	copy(oid[:], id.Data[:])
// 	return bson.Marshal(oid)
// }
// func (id *ObjectID) UnmarshalBSON(b []byte) error {
// 	var oid primitive.ObjectID
// 	err := bson.Unmarshal(b, &oid)
// 	if err != nil {
// 		return err
// 	}
// 	id.Data = oid[:]
// 	return nil
// }

func (x *ObjectID) String() string {
	return x.Hex()
}

////

func (x *ObjectID) Reset() {
	*x = ObjectID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_ObjectID_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

// func (x *ObjectID) String() string {
// 	return protoimpl.X.MessageStringOf(x)
// }

func (*ObjectID) ProtoMessage() {}

func (x *ObjectID) ProtoReflect() protoreflect.Message {
	mi := &file_pb_ObjectID_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObjectID.ProtoReflect.Descriptor instead.
func (*ObjectID) Descriptor() ([]byte, []int) {
	return file_pb_ObjectID_proto_rawDescGZIP(), []int{0}
}

// func (x *ObjectID) GetData() []byte {
// 	if x != nil {
// 		return x.Data
// 	}
// 	return nil
// }

var File_pb_ObjectID_proto protoreflect.FileDescriptor

var file_pb_ObjectID_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x62, 0x2f, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x1e, 0x0a, 0x08, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x12,
	0x12, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x44,
	0x61, 0x74, 0x61, 0x42, 0x16, 0x5a, 0x14, 0x64, 0x75, 0x63, 0x6b, 0x2f, 0x6d, 0x6f, 0x6e, 0x67,
	0x6f, 0x64, 0x62, 0x3b, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_pb_ObjectID_proto_rawDescOnce sync.Once
	file_pb_ObjectID_proto_rawDescData = file_pb_ObjectID_proto_rawDesc
)

func file_pb_ObjectID_proto_rawDescGZIP() []byte {
	file_pb_ObjectID_proto_rawDescOnce.Do(func() {
		file_pb_ObjectID_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_ObjectID_proto_rawDescData)
	})
	return file_pb_ObjectID_proto_rawDescData
}

var file_pb_ObjectID_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_pb_ObjectID_proto_goTypes = []interface{}{
	(*ObjectID)(nil), // 0: ObjectID
}
var file_pb_ObjectID_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pb_ObjectID_proto_init() }
func file_pb_ObjectID_proto_init() {
	if File_pb_ObjectID_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pb_ObjectID_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObjectID); i {
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
			RawDescriptor: file_pb_ObjectID_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pb_ObjectID_proto_goTypes,
		DependencyIndexes: file_pb_ObjectID_proto_depIdxs,
		MessageInfos:      file_pb_ObjectID_proto_msgTypes,
	}.Build()
	File_pb_ObjectID_proto = out.File
	file_pb_ObjectID_proto_rawDesc = nil
	file_pb_ObjectID_proto_goTypes = nil
	file_pb_ObjectID_proto_depIdxs = nil
}
