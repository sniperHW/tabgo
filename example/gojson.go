package main

import (
	"fmt"

	"github.com/sniperHW/tabgo/test"
)

// go读取json输出文件
func main() {

	test.TableLoader.OnLoadFinish("Model", func() {
		fmt.Println("OnLoadFinish")
	})

	test.TableLoader.Load("../test")

	test.ModelMap.ForEach(func(m *test.Model) bool {
		fmt.Println(*m)
		return true
	})
	test.TableLoader.Load("../test")
}
