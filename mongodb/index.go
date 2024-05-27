package mongodb

import (
	"context"

	"github.com/dogsays/mo/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexOrder int32

const (
	IndexAscending  IndexOrder = 0 // 升序 = 1 默认
	IndexDescending            = 1 // 降序 = -1
)

type Index struct {
	Key        string
	Background bool
	SetUnique  bool
	Sparse     bool
	Order      IndexOrder
}

// 创建索引，不支持联合索引
func (coll *Collection) CreateIndex(list []*Index) ([]string, error) {
	indexs := []mongo.IndexModel{}

	for _, v := range list {
		order := 1
		if v.Order == IndexDescending {
			order = -1
		}

		opt := options.Index()
		opt.SetBackground(v.Background)
		opt.SetUnique(v.SetUnique)
		opt.SetSparse(v.Sparse)

		indexs = append(indexs, mongo.IndexModel{
			Keys:    bson.D{{Key: v.Key, Value: order}},
			Options: opt,
		})
	}

	return coll.Coll().Indexes().CreateMany(context.TODO(), indexs)
}

// 偷懒写法
func (coll *Collection) CreateIndexByKeys(keys ...string) {
	list := []*Index{}
	for _, v := range keys {
		list = append(list, &Index{Key: v})
	}

	_, err := coll.CreateIndex(list)
	if err != nil {
		logger.Info("创建索引失败", err)
	}
}
