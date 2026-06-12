package test

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

type ModelStructY struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type ModelStruct struct {
	X     int64        `json:"x"`
	Y     ModelStructY `json:"y"`
	Array []int64      `json:"array"`
}

type ModelArray_struct struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type Model struct {
	Id           int64               `json:"id"`
	Name         string              `json:"name"`
	Icon         string              `json:"icon"`
	Model        string              `json:"model"`
	Length       int64               `json:"length"`
	Width        int64               `json:"width"`
	Struct       ModelStruct         `json:"struct"`
	Array        []int64             `json:"array"`
	Array2d      [][]int64           `json:"array2d"`
	Array_struct []ModelArray_struct `json:"array_struct"`
}

type modelMap struct {
	modelMap     atomic.Value
	lastDigest   [sha256.Size]byte
	hasDigest    bool
	onLoadFinish []func()
}

var ModelMap modelMap

func (m *modelMap) getModelMap() map[int]*Model {
	v := m.modelMap.Load()
	if v == nil {
		return nil
	} else {
		return v.(map[int]*Model)
	}
}

func (m *modelMap) setModelMap(v map[int]*Model) {
	m.modelMap.Store(v)
}

func (m *modelMap) Get(id int) (mo *Model, ok bool) {
	if v := m.getModelMap(); v == nil {
		return
	} else {
		mo, ok = v[id]
		return
	}
}

func (m *modelMap) loadFromBytes(s []byte) error {
	v := make(map[int]*Model)
	err := json.Unmarshal(s, &v)
	if err != nil {
		return err
	}
	m.setModelMap(v)
	return nil
}

func (m *modelMap) LoadFromString(s string) error {
	return m.loadFromBytes([]byte(s))
}

func (m *modelMap) LoadFromFile(path string) (bool, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return false, err
	}
	digest := sha256.Sum256(jsonData)
	if m.hasDigest && m.lastDigest == digest {
		return false, nil
	}
	m.lastDigest = digest
	m.hasDigest = true
	return true, m.loadFromBytes(jsonData)
}

func (m *modelMap) ForEach(fn func(m *Model) bool) {
	for _, m := range m.getModelMap() {
		if !fn(m) {
			break
		}
	}
}

func (m *modelMap) OnLoadFinish(fn func()) {
	m.onLoadFinish = append(m.onLoadFinish, fn)
}

func (m *modelMap) fireOnLoadFinish() {
	for _, fn := range m.onLoadFinish {
		fn()
	}
}
