package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sniperHW/tabgo/test"
)

// go读取json输出文件
func main() {

	var buildMap map[int]*test.Model

	f, err := os.Open("../test/Model.Json")
	if err != nil {
		panic(err)
	}

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	f.Close()

	err = json.Unmarshal(bts, &buildMap)

	if err != nil {
		panic(err)
	}

	for _, v := range buildMap {
		fmt.Println(*v)
	}

}
