package test

import (
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

type Hero struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Quality   string `json:"quality"`
	Max_level int64  `json:"max_level"`
}

type heroMap struct {
	heroMap atomic.Value
}

var HeroMap heroMap

func (m *heroMap) getHeroMap() map[int]*Hero {
	v := m.heroMap.Load()
	if v == nil {
		return nil
	} else {
		return v.(map[int]*Hero)
	}
}

func (m *heroMap) setHeroMap(v map[int]*Hero) {
	m.heroMap.Store(v)
}

func (m *heroMap) Get(id int) (mo *Hero, ok bool) {
	if v := m.getHeroMap(); v == nil {
		return
	} else {
		mo, ok = v[id]
		return
	}
}

func (m *heroMap) loadFromBytes(s []byte) error {
	v := make(map[int]*Hero)
	err := json.Unmarshal(s, &v)
	if err != nil {
		return err
	}
	m.setHeroMap(v)
	return nil
}

func (m *heroMap) LoadFromString(s string) error {
	return m.loadFromBytes([]byte(s))
}

func (m *heroMap) LoadFromFile(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	return m.loadFromBytes(jsonData)
}

func (m *heroMap) ForEach(fn func(m *Hero) bool) {
	for _, m := range m.getHeroMap() {
		if !fn(m) {
			break
		}
	}
}
