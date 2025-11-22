package main

import (
	"fmt"

	"github.com/sniperHW/tabgo/test"
)

// go读取json输出文件
func main() {
	err := test.LoadModelFromFile("./test/Model.json")
	if err != nil {
		panic(err)
	}

	test.ForEachModel(func(m *test.Model) bool {
		fmt.Println(*m)
		return true
	})
}
