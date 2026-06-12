package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	"github.com/sniperHW/tabgo/parser"
)

type tableEntry struct {
	OrigName  string
	Name      string
	NameLower string
	Data      string
}

type goStruct struct {
	Package string
	tables  []tableEntry
	mu      sync.Mutex
}

var goTemplate string = `
package {{.Package}}

import(
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

{{.Data}}

type {{.TableNameLower}}Map struct {
	{{.TableNameLower}}Map atomic.Value
	lastDigest             [sha256.Size]byte
	hasDigest              bool
	onLoadFinish           []func()
}

var {{.TableName}}Map {{.TableNameLower}}Map

func (m *{{.TableNameLower}}Map) get{{.TableName}}Map() map[int]*{{.TableName}} {
	v := m.{{.TableNameLower}}Map.Load()
	if v == nil {
		return nil
	} else {
		return v.(map[int]*{{.TableName}})
	}
}

func (m *{{.TableNameLower}}Map) set{{.TableName}}Map(v map[int]*{{.TableName}}) {
	m.{{.TableNameLower}}Map.Store(v)
}

func (m *{{.TableNameLower}}Map) Get(id int) (mo *{{.TableName}}, ok bool) {
	if v := m.get{{.TableName}}Map(); v == nil {
		return
	} else {
		mo, ok = v[id]
		return
	}
}

func (m *{{.TableNameLower}}Map) loadFromBytes(s []byte) error {
	v := make(map[int]*{{.TableName}})
	err := json.Unmarshal(s, &v)
	if err != nil {
		return err
	}
	m.set{{.TableName}}Map(v)
	return nil
}

func (m *{{.TableNameLower}}Map) LoadFromString(s string) error {
	return m.loadFromBytes([]byte(s))
}

func (m *{{.TableNameLower}}Map) LoadFromFile(path string) (bool, error) {
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

func (m *{{.TableNameLower}}Map) ForEach(fn func(m *{{.TableName}}) bool) {
	for _, m := range m.get{{.TableName}}Map() {
		if !fn(m) {
			break
		}
	}
}

func (m *{{.TableNameLower}}Map) OnLoadFinish(fn func()) {
	m.onLoadFinish = append(m.onLoadFinish, fn)
}

func (m *{{.TableNameLower}}Map) fireOnLoadFinish() {
	for _, fn := range m.onLoadFinish {
		fn()
	}
}
`

var loaderTemplate string = `package {{.Package}}

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
{{range .Tables}}
	case "{{.OrigName}}":
		if loaded, err := {{.Name}}Map.LoadFromFile(file); err != nil {
			log.Println("load ", file, "error:", err)
		} else if loaded {
			return {{.Name}}Map.fireOnLoadFinish
		}
{{end}}
	}
	return nil
}
`

type tableTemplateData struct {
	Package        string
	Data           string
	TableName      string
	TableNameLower string
}

type loaderTableInfo struct {
	OrigName string
	Name     string
}

type loaderTemplateData struct {
	Package string
	Tables  []loaderTableInfo
}

func writeGoFile(filename string, tmpl *template.Template, data interface{}) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(f, data)
	f.Close()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("gofmt", "-w", filename)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	log.Printf("%s Write ok\n", filename)
}

func (j *goStruct) walkOk(writePath string, tmpl *template.Template) {
	dir := fmt.Sprintf("%s/%s", writePath, j.Package)
	os.MkdirAll(dir, os.ModePerm)

	tableNames := make([]loaderTableInfo, 0, len(j.tables))
	for _, t := range j.tables {
		filename := fmt.Sprintf("%s/%s.go", dir, t.OrigName)
		writeGoFile(filename, tmpl, tableTemplateData{
			Package:        j.Package,
			Data:           t.Data,
			TableName:      t.Name,
			TableNameLower: t.NameLower,
		})
		tableNames = append(tableNames, loaderTableInfo{OrigName: t.OrigName, Name: t.Name})
	}

	loaderFile := fmt.Sprintf("%s/loader.go", dir)
	loaderTmpl, err := template.New("loader").Parse(loaderTemplate)
	if err != nil {
		panic(err)
	}
	writeGoFile(loaderFile, loaderTmpl, loaderTemplateData{
		Package: j.Package,
		Tables:  tableNames,
	})
}

// processTable 处理表结构，生成 Go 类型定义
func (j *goStruct) processTable(colNames []string, types []string, table *Table) {
	origName := table.name
	name := strings.Title(table.name)
	nameLower := strings.ToLower(table.name)
	fields := []string{}
	for i := 0; i < len(colNames); i++ {
		if table.fields[i].parser != nil {
			fields = append(fields, fmt.Sprintf("%s:%s", colNames[i], types[i]))
		}
	}
	str := "{" + strings.Join(fields, ",") + "}"
	p, err := parser.MakeParser(str)
	if err != nil {
		panic(err)
	}
	data := p.GenGoDefine(name)
	j.mu.Lock()
	j.tables = append(j.tables, tableEntry{
		OrigName:  origName,
		Name:      name,
		NameLower: nameLower,
		Data:      data,
	})
	j.mu.Unlock()
}
