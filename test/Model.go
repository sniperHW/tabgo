package test

import (
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

type ModelStructY struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ModelStruct struct {
	X     int          `json:"x"`
	Y     ModelStructY `json:"y"`
	Array []int        `json:"array"`
}

type ModelArray_struct struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Model struct {
	Id           int                 `json:"id"`
	Name         string              `json:"name"`
	Icon         string              `json:"icon"`
	Model        string              `json:"model"`
	Length       int                 `json:"length"`
	Width        int                 `json:"width"`
	Struct       ModelStruct         `json:"struct"`
	Array        []int               `json:"array"`
	Array2d      [][]int             `json:"array2d"`
	Array_struct []ModelArray_struct `json:"array_struct"`
}

type _ModelMap map[int]*Model

var __ModelMap atomic.Value

func init() {
	__ModelMap.Store(make(_ModelMap))
}

func getModelMap() _ModelMap {
	return __ModelMap.Load().(_ModelMap)
}

func setModelMap(m _ModelMap) {
	__ModelMap.Store(m)
}

func GetModel(id int) (*Model, bool) {
	m, ok := getModelMap()[id]
	return m, ok
}

func loadModelFromBytes(s []byte) error {
	m := make(_ModelMap)
	err := json.Unmarshal(s, &m)
	if err != nil {
		return err
	}
	setModelMap(m)
	return nil
}

func LoadModelFromString(s string) error {
	return loadModelFromBytes([]byte(s))
}

func LoadModelFromFile(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	return loadModelFromBytes(jsonData)
}

func ForEachModel(fn func(m *Model) bool) {
	for _, m := range getModelMap() {
		if !fn(m) {
			break
		}
	}
}
