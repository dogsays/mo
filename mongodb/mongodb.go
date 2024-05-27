package mongodb

import (
	"context"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonoptions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var timeCodec = bsoncodec.NewTimeCodec(
	bsonoptions.TimeCodec().SetUseLocalTimeZone(true),
)

func build() *bsoncodec.Registry {
	bsoncodec.DefaultStructTagParser = uppercaseStructTagParser

	rb := bson.NewRegistryBuilder()

	// 解决数据库里面是UTC时间。我们这里设置为使用本地时间

	rb.RegisterTypeEncoder(reflect.TypeOf(time.Time{}), timeCodec)
	rb.RegisterTypeDecoder(reflect.TypeOf(time.Time{}), timeCodec)

	oidCodec := &ObjectIDCodec{}
	rb.RegisterTypeDecoder(tOID, oidCodec)
	rb.RegisterTypeEncoder(tOID, oidCodec)

	tsCodec := &TimeStampCodec{}
	rb.RegisterTypeDecoder(tTimeStamp, tsCodec)
	rb.RegisterTypeEncoder(tTimeStamp, tsCodec)

	return rb.Build()
}

var DefaultRegistry = build()

func Connect(mongoUrl string) (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(mongoUrl),
		options.Client().SetRegistry(DefaultRegistry), // 必须覆盖默认的Registry
	)
	if err != nil {
		return nil, err
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

// copy from: mongo-driver@v1.10.3/bson/bsoncodec/struct_tag_parser.go
// 解决 struct field 默认小写问题
func uppercaseStructTagParser(sf reflect.StructField) (bsoncodec.StructTags, error) {
	key := sf.Name // modify
	tag, ok := sf.Tag.Lookup("bson")
	if !ok && !strings.Contains(string(sf.Tag), ":") && len(sf.Tag) > 0 {
		tag = string(sf.Tag)
	}
	return parseTags(key, tag)
}

func parseTags(key string, tag string) (bsoncodec.StructTags, error) {
	var st bsoncodec.StructTags
	if tag == "-" {
		st.Skip = true
		return st, nil
	}

	for idx, str := range strings.Split(tag, ",") {
		if idx == 0 && str != "" {
			key = str
		}
		switch str {
		case "omitempty":
			st.OmitEmpty = true
		case "minsize":
			st.MinSize = true
		case "truncate":
			st.Truncate = true
		case "inline":
			st.Inline = true
		}
	}

	st.Name = key

	return st, nil
}
