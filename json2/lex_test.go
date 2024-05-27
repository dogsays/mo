package json2

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func getJson() string {
	return `
	{
		"false": true,
		number: 123,
		"n:umber1": 123,
		nullvalue: null, //fuckyou
		ba: 1.33,
		/**/ /***sadf你说的发生豆腐块***/
		_arr /***sadf你说的发生豆腐块***/: [1, 2, 3, 4

			// sdfkas;ldkflkj
			, 5,
		],
		Substruct: [{
			A: 110,
			Map: {
				"123": 66666,
				"中文Key": "as???",
			}
		}],
		a: -1.233,
		Map: {
			"A": {
				F1 : 1,
				F2 : 2,
			},

			"B": {
				F1 : 2,
				F2 : 3,
			},
		}
	}
		`
}

func TestA(t *testing.T) {
	str := getJson()

	mp := map[string]any{}
	unmarshalTo(str, &mp)

	type A struct {
		A   int
		Map map[string]any
	}

	type F12 struct {
		F1 int
		F2 int8
	}

	var tar struct {
		False     bool    `json:"false"`
		Number    int     `json:"number"`
		Ba        float32 `json:"ba"`
		Arr       []int   `json:"_arr"`
		Substruct []A
		Map       map[string]F12
	}
	unmarshalTo(str, &tar)

	var it any
	unmarshalTo(str, &it)
}

func TestArray(t *testing.T) {
	str := `[
		{
			"hello":123,
			"World":123,
		},
		{
			hello:666,
		}
	]`

	type HelloWorld struct {
		Hello int `json:"hello"`
		World int
	}
	arr := []HelloWorld{}
	unmarshalTo(str, &arr)

	arrMp := []map[string]any{}
	unmarshalTo(str, &arrMp)
}

func unmarshalTo(str string, to any) {
	err := Unmarshal([]byte(str), to)
	if err != nil {
		panic(err)
	}

	buf, _ := json.MarshalIndent(to, "", " ")
	fmt.Println("@@@@@@@@@")
	fmt.Println(string(buf))
}

func TestP(t *testing.T) {

	buf, err := Standardize([]byte(getJson()))
	if err != nil {
		panic(err)
	}

	times := 10000

	fn := func(str string, UnmarshalFN func(data []byte, v any) error) {
		before := time.Now()
		for i := 0; i < times; i++ {
			var one any
			UnmarshalFN(buf, &one)
		}
		after := time.Now()
		fmt.Println("speed", str, after.Sub(before))
	}

	fn("Unmarshal", Unmarshal)
	fn("json.Unmarshal", json.Unmarshal)
}
