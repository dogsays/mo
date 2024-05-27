package mongodb

import (
	"reflect"
	"strings"
	"sync"

	"github.com/dogsays/mo/logger"
	"github.com/dogsays/mo/ut2"

	"go.mongodb.org/mongo-driver/bson"
)

// 管理一张Mongo表。 并支持内存中缓存。自动以 _id 为 Key
type DocOf[K ut2.Comparable, V any] struct {
	lock    sync.Mutex
	idField string
	Coll    *Collection
	Ctor    func(*V) // 在自动创建的时候调用，初始化数据
	Map     ut2.IMap[K, *V]
}

func NewDocOfWith[K ut2.Comparable, V any](coll *Collection, mp ut2.IMap[K, *V]) *DocOf[K, V] {
	return &DocOf[K, V]{
		Coll:    coll,
		Map:     mp,
		idField: findId[V](),
	}
}

// 无 内存缓存
func NewDocOfWithoutCache[K ut2.Comparable, V any](coll *Collection) *DocOf[K, V] {
	return NewDocOfWith[K, V](coll, ut2.NewFakeMap[K, *V]())
}

// 有 内存缓存
func NewDocOf[K ut2.Comparable, V any](coll *Collection) *DocOf[K, V] {
	return NewDocOfWith[K, V](coll, ut2.NewSyncMap[K, *V]())
}

// 释放内存缓存
func (d *DocOf[K, V]) Release() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.Map.Clear()
}

// 更新数据
func (d *DocOf[K, V]) Call(key K, fn func(data *V)) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	data := d.getData(key)
	dataCopy := *data

	fn(&dataCopy)

	v1 := reflect.ValueOf(data).Elem()
	v2 := reflect.ValueOf(dataCopy)

	typ := reflect.TypeOf(dataCopy)

	setMap := bson.M{}

	for i := 0; i < typ.NumField(); i++ {
		f1 := v1.Field(i)
		f2 := v2.Field(i)

		tf := typ.Field(i)

		if !tf.IsExported() {
			continue
		}

		switch f1.Type().Kind() {

		case reflect.Map, reflect.Array, reflect.Slice, reflect.Struct, reflect.Ptr: // 直接设置

		case reflect.Func, reflect.Chan: // 跳过
			continue

		default: // 基础类型。变化才设置
			if f1.Interface() == f2.Interface() {
				continue
			}
		}

		tagName := tf.Tag.Get("bson")
		if tagName == "-" {
			continue
		}

		if tagName == "" {
			tagName = tf.Name
		} else {
			if strings.Contains(tagName, ",") {
				continue
			}
		}

		// logger.Info("setField", tf.Name, tagName)
		f1.Set(f2)

		setMap[tagName] = f2.Interface()
	}

	logger.Info("DocOf Commit", d.Coll.Name, key, ut2.ToJson(setMap))

	err := d.Coll.UpdateId(key, bson.M{"$set": setMap})
	if err != nil {
		logger.Info("DocOf Commit err: ", d.Coll.Name, err)
	}

	return err
}

// 获取数据拷贝
func (d *DocOf[K, V]) GetData(pid K) V {
	d.lock.Lock()
	defer d.lock.Unlock()

	return *d.getData(pid)
}

func (d *DocOf[K, V]) getData(key K) *V {

	doc, ok := d.Map.Load(key)
	if ok {
		return doc
	}

	doc = new(V)
	err := d.Coll.FindId(key, doc)
	if err != nil {

		val := reflect.ValueOf(doc).Elem()
		val.FieldByName(d.idField).Set(reflect.ValueOf(key))

		if d.Ctor != nil {
			d.Ctor(doc)
		}

		err = d.Coll.InsertOne(doc)
		if err != nil {
			logger.Err("DocOf Insert", d.Coll.Name, key, err)
		}
	}

	d.Map.Store(key, doc)

	return doc
}

func findId[V any]() string {
	typ := reflect.TypeOf(new(V)).Elem()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("bson")
		if tag == "_id" {
			return field.Name
		}
	}

	panic("no _id found: " + typ.Name())
}
