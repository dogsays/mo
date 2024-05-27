package mongodb

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	Name   string
	Client *mongo.Client
}

func NewDBWithClient(dbName string, client *mongo.Client) *DB {
	return &DB{Name: dbName, Client: client}
}

func NewDB(dbName string) *DB {
	return &DB{Name: dbName}
}

func (db *DB) Connect(dburl string) error {
	cli, err := Connect(dburl)
	if err != nil {
		return err
	}
	db.Client = cli
	return nil
}

func (db *DB) Collection(collName string) *Collection {
	return &Collection{Name: collName, db: db}
}

type Collection struct {
	Name string
	db   *DB
}

func (c *Collection) Coll() *mongo.Collection {
	return c.db.Client.Database(c.db.Name).Collection(c.Name)
}

func (c *Collection) InsertOne(doc any) error {
	_, err := c.Coll().InsertOne(context.TODO(), doc)
	return err
}

func (c *Collection) FindId(id, doc any) error {
	return c.Coll().FindOne(context.TODO(), bson.M{"_id": id}).Decode(doc)
}

func (c *Collection) FindOne(filter, doc any) error {
	return c.Coll().FindOne(context.TODO(), filter).Decode(doc)
}

func (c *Collection) DeleteId(id any) error {
	_, err := c.Coll().DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

func (c *Collection) DeleteOne(filter any) error {
	_, err := c.Coll().DeleteOne(context.TODO(), filter)
	return err
}

func (c *Collection) FindAll(filter bson.M, doc any) error {

	cur, err := c.Coll().Find(context.TODO(), filter)
	if err != nil {
		return err
	}

	defer cur.Close(context.TODO())

	if err = cur.All(context.TODO(), doc); err != nil {
		return err
	}
	return nil
}

func (c *Collection) FindAllOpt(filter bson.M, doc any, opt *options.FindOptions) error {

	cur, err := c.Coll().Find(context.TODO(), filter, opt)
	if err != nil {
		return err
	}

	defer cur.Close(context.TODO())

	if err = cur.All(context.TODO(), doc); err != nil {
		return err
	}
	return nil
}

func (c *Collection) UpdateOne(id any, update any) error {
	_, err := c.Coll().UpdateOne(context.TODO(), id, update)
	return err
}

func (c *Collection) UpdateId(id any, update any) error {
	_, err := c.Coll().UpdateByID(context.TODO(), id, update)
	return err
}

func (c *Collection) Update(filter any, update any) error {
	_, err := c.Coll().UpdateMany(context.TODO(), filter, update)
	return err
}

func (c *Collection) UpsertOne(filter any, update any) error {
	_, err := c.Coll().UpdateOne(context.TODO(), filter, update,
		options.Update().SetUpsert(true))
	return err
}

func (c *Collection) UpsertId(id any, update any) error {
	_, err := c.Coll().UpdateByID(context.TODO(), id, update,
		options.Update().SetUpsert(true))
	return err
}

func (c *Collection) Upsert(filter any, update any) error {
	_, err := c.Coll().UpdateMany(context.TODO(), filter, update,
		options.Update().SetUpsert(true))
	return err
}

// 统计满足条件的文档数，会遍历整个表
// 判断条件是否存在请使用 Exists
func (c *Collection) CountDocuments(filter any) (int64, error) {
	return c.Coll().CountDocuments(context.TODO(), filter)
}

// 判断某条件是否存在于表中
func (c *Collection) Exists(filter bson.M) (bool, error) {

	opt := options.FindOne()
	opt.SetProjection(bson.M{"_id": 1}) // 只取出 _id 就好了。

	out := bson.M{}
	err := c.Coll().FindOne(context.TODO(), filter, opt).Decode(&out)

	if err == mongo.ErrNoDocuments {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

type FindPageOpt struct {
	Page     int64 // 从1开始的页码 default = 1
	PageSize int64 // 每页个数 default = 30
	Sort     any   // 排序
	Query    any   // 查找条件
}

func (c *Collection) FindPage(args FindPageOpt, doc any) (count int64, err error) {

	if args.Page < 1 {
		args.Page = 1
	}
	if args.PageSize <= 0 {
		args.PageSize = 30
	}

	op := options.Find()
	op.SetSkip((args.Page - 1) * args.PageSize)
	op.SetLimit(args.PageSize)
	if args.Sort != nil {
		op.SetSort(args.Sort)
	}

	coll := c.Coll()

	count, err = coll.CountDocuments(context.TODO(), args.Query)
	if err != nil {
		return
	}

	cursor, err := coll.Find(context.TODO(), args.Query, op)
	if err != nil {
		return
	}
	defer cursor.Close(context.TODO())

	err = cursor.All(context.TODO(), doc)

	return
}

// 返回假的count，避免用时过长
func (c *Collection) FindPageFakeCount(args FindPageOpt, doc any) (count int64, err error) {

	if args.Page < 1 {
		args.Page = 1
	}
	if args.PageSize <= 0 {
		args.PageSize = 30
	}

	op := options.Find()
	op.SetSkip((args.Page - 1) * args.PageSize)
	op.SetLimit(args.PageSize)
	if args.Sort != nil {
		op.SetSort(args.Sort)
	}

	coll := c.Coll()

	cursor, err := coll.Find(context.TODO(), args.Query, op)
	if err != nil {
		return
	}
	defer cursor.Close(context.TODO())

	err = cursor.All(context.TODO(), doc)

	ln := reflect.ValueOf(doc).Elem().Len()
	if ln >= int(args.PageSize) {
		count = 100 * args.PageSize
	} else {
		count = int64(ln)
	}

	return
}
