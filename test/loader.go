package test

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type tableLoader struct{}

var TableLoader tableLoader

func (l *tableLoader) Load(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Println("read dir ", path, "error:", err)
		return
	}
	loaded := []func(){}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			if fn := l.load(filepath.Join(path, entry.Name())); fn != nil {
				loaded = append(loaded, fn)
			}
		}
	}
	for _, fn := range loaded {
		fn()
	}
}

func (l *tableLoader) load(file string) func() {
	name := filepath.Base(file)
	name = strings.TrimSuffix(name, ".json")
	switch name {

	case "Model":
		if loaded, err := ModelMap.LoadFromFile(file); err != nil {
			log.Println("load ", file, "error:", err)
		} else if loaded {
			return ModelMap.fireOnLoadFinish
		}

	case "hero":
		if loaded, err := HeroMap.LoadFromFile(file); err != nil {
			log.Println("load ", file, "error:", err)
		} else if loaded {
			return HeroMap.fireOnLoadFinish
		}

	}
	return nil
}
