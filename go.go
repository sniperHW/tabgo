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
	TableName string
	Data      string
	Package   string
	str       strings.Builder
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

type _{{.TableName}}Map map[int]*{{.TableName}}

var __{{.TableName}}Map atomic.Value

func init() {
	__{{.TableName}}Map.Store(make(_{{.TableName}}Map))
}

func get{{.TableName}}Map() _{{.TableName}}Map {
	return __{{.TableName}}Map.Load().(_{{.TableName}}Map)
}

func set{{.TableName}}Map(m _{{.TableName}}Map) {
	__{{.TableName}}Map.Store(m)
}

func Get{{.TableName}}(id int) (*{{.TableName}}, bool) {
	m, ok := get{{.TableName}}Map()[id]
	return m, ok
}

func load{{.TableName}}FromBytes(s []byte) error {
	m := make(_{{.TableName}}Map)
	err := json.Unmarshal(s, &m)
	if err != nil {
		return err
	}
	set{{.TableName}}Map(m)
	return nil
}

func Load{{.TableName}}FromString(s string) error {
	return load{{.TableName}}FromBytes([]byte(s))
}

func Load{{.TableName}}FromFile(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	return load{{.TableName}}FromBytes(jsonData)
}

func ForEach{{.TableName}}(fn func(m *{{.TableName}}) bool) {
	for _, m := range get{{.TableName}}Map() {
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
	j.TableName = table.name
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
	j.str.WriteString(p.GenGoDefine(strings.Title(table.name)))
}
