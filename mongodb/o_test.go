package mongodb

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Doc struct {
	ID   *ObjectID `bson:"_id"`
	AAA  string
	Time *TimeStamp
}

func TestOID(t *testing.T) {
	o := NewObjectID()
	fmt.Println("@@@", o.String(), o.Hex())

	str, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(str))

	var o1 *ObjectID
	err = json.Unmarshal(str, &o1)
	if err != nil {
		panic(err)
	}

	fmt.Println("@@@", o1.String(), o1.Hex())
}

func TestXxx(t *testing.T) {

	var db = NewDB("test")
	var collPlayers = db.Collection("test")

	if err := db.Connect("mongodb://127.0.0.1:27017"); err != nil {
		panic(err)
	}

	now := time.Now()
	bt, _ := json.Marshal(now)
	fmt.Println(string(bt))

	doc := &Doc{
		ID:   NewObjectID(),
		AAA:  "123123123",
		Time: NewTimeStamp(now),
	}

	err := collPlayers.InsertOne(doc)
	if err != nil {
		panic(err)
	}

	dest := Doc{}
	err = collPlayers.FindOne(bson.M{"_id": doc.ID}, &dest)
	if err != nil {
		panic(err)
	}

	jjbytes, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	fmt.Println("==>", string(jjbytes))
}
