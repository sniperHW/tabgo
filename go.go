package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/sniperHW/tabgo/parser"
)

type goStruct struct {
	TableName      string
	TableNameLower string
	Data           string
	Package        string
	str            strings.Builder
}

var goTemplate string = `
package {{.Package}}

import(
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

{{.Data}}

type {{.TableNameLower}}Map struct {
	{{.TableNameLower}}Map atomic.Value
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

func (m *{{.TableNameLower}}Map) LoadFromFile(path string) error {
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

func (m *{{.TableNameLower}}Map) ForEach(fn func(m *{{.TableName}}) bool) {
	for _, m := range m.get{{.TableName}}Map() {
		if !fn(m) {
			break
		}
	}
}
`

func (j *goStruct) walkOk(writePath string, tmpl *template.Template) {
	path := fmt.Sprintf("%s/%s", writePath, j.Package)
	filename := fmt.Sprintf("%s/%s.go", path, j.TableName)
	os.MkdirAll(path, os.ModePerm)
	f, err := os.OpenFile(filename, os.O_RDWR, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(filename)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	defer func() {
		f.Close()
		cmd := exec.Command("gofmt", "-w", filename)
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
		}
	}()

	err = os.Truncate(filename, 0)
	if err != nil {
		panic(err)
	}

	j.Data = j.str.String()
	err = tmpl.Execute(f, j)
	if err != nil {
		panic(err)
	} else {
		log.Printf("%s Write ok\n", filename)
	}
}

// processTable 处理表结构，生成 Go 类型定义
func (j *goStruct) processTable(colNames []string, types []string, table *Table) {
	j.TableName = strings.Title(table.name)
	j.TableNameLower = strings.ToLower(table.name)
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
	j.str.WriteString(p.GenGoDefine(j.TableName))
}
