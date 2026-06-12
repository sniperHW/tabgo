package test

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

type Hero struct {
	Id                int64   `json:"id"`
	Name              string  `json:"name"`
	Attack_type       string  `json:"attack_type"`
	Camp              string  `json:"camp"`
	Quality           string  `json:"quality"`
	Max_level         int64   `json:"max_level"`
	Power             int64   `json:"power"`
	Building_hp       int64   `json:"building_hp"`
	Building_hp_extra int64   `json:"building_hp_extra"`
	Attack            int64   `json:"attack"`
	Attack_extra      int64   `json:"attack_extra"`
	Speed             float64 `json:"speed"`
	Time              float64 `json:"time"`
	Hp                int64   `json:"hp"`
	Hp_extra          int64   `json:"hp_extra"`
	Attack_range      float64 `json:"attack_range"`
}

type heroMap struct {
	heroMap      atomic.Value
	lastDigest   [sha256.Size]byte
	hasDigest    bool
	onLoadFinish []func()
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

func (m *heroMap) LoadFromFile(path string) (bool, error) {
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

func (m *heroMap) ForEach(fn func(m *Hero) bool) {
	for _, m := range m.getHeroMap() {
		if !fn(m) {
			break
		}
	}
}

func (m *heroMap) OnLoadFinish(fn func()) {
	m.onLoadFinish = append(m.onLoadFinish, fn)
}

func (m *heroMap) fireOnLoadFinish() {
	for _, fn := range m.onLoadFinish {
		fn()
	}
}
