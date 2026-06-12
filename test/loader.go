package test

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type tableLoader struct {
	onLoadFinish map[string][]func()
}

var TableLoader tableLoader

func (l *tableLoader) OnLoadFinish(tabname string, fn func()) {
	if l.onLoadFinish == nil {
		l.onLoadFinish = make(map[string][]func())
	}
	l.onLoadFinish[tabname] = append(l.onLoadFinish[tabname], fn)
}

func (l *tableLoader) Load(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Println("read dir ", path, "error:", err)
		return
	}
	loaded := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			if name, ok := l.load(filepath.Join(path, entry.Name())); ok {
				loaded[name] = true
			}
		}
	}
	for name, fns := range l.onLoadFinish {
		if loaded[name] {
			for _, fn := range fns {
				fn()
			}
		}
	}
}

func (l *tableLoader) load(file string) (string, bool) {
	name := filepath.Base(file)
	name = strings.TrimSuffix(name, ".json")
	switch name {

	case "Model":
		loaded, err := ModelMap.LoadFromFile(file)
		if err != nil {
			log.Println("load ", file, "error:", err)
		}
		return name, loaded

	case "hero":
		loaded, err := HeroMap.LoadFromFile(file)
		if err != nil {
			log.Println("load ", file, "error:", err)
		}
		return name, loaded

	}
	return name, false
}
