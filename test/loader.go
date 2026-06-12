package test

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type tableLoader struct {
	onLoadFinish []func() //加载完毕后回调
}

var TableLoader tableLoader

func (l *tableLoader) OnLoadFinish(fn func()) {
	l.onLoadFinish = append(l.onLoadFinish, fn)
}

func (l *tableLoader) Load(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Println("read dir ", path, "error:", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			l.load(filepath.Join(path, entry.Name()))
		}
	}
	for _, v := range l.onLoadFinish {
		v()
	}
}

func (l *tableLoader) load(file string) {
	name := filepath.Base(file)
	name = strings.TrimSuffix(name, ".json")
	switch name {

	case "Model":
		if err := ModelMap.LoadFromFile(file); err != nil {
			log.Println("load ", file, "error:", err)
		}

	case "hero":
		if err := HeroMap.LoadFromFile(file); err != nil {
			log.Println("load ", file, "error:", err)
		}

	}
}
