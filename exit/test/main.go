package main

import (
	"fmt"
	"net/http"

	"github.com/dogsays/mo/exit"
)

func main() {
	exit.Callback("", func() { fmt.Println("要退出了1") })
	exit.Callback("", func() { fmt.Println("要退出了2") })
	exit.Callback("", func() { fmt.Println("要退出了3") })

	err := http.ListenAndServe(":9092", nil)
	if err != nil {
		fmt.Println(err)
	}
}
