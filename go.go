package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

func title(s string) string {
	if len(s) > 0 && s[0] >= 'a' && s[0] <= 'z' {
		b := []byte(s)
		b[0] -= ('a' - 'A')
		return string(b)
	} else {
		return s
	}
}

func (p *ValueParser) GenGoStruct(s *strings.Builder, _ string) {
}

func (p *ValueParser) GetGoType() string {
	switch p.valueType {
	case typeInt:
		return "int"
	case typeString:
		return "string"
	case typeBool:
		return "bool"
	case typeFloat:
		return "float64"
	default:
		panic("error")
	}
}

func (p *ArrayParser) GenGoStruct(s *strings.Builder, s1 string) {
	p.elements.GenGoStruct(s, s1)
}

func (p *ArrayParser) GetGoType() string {
	return "[]" + p.elements.GetGoType()
}

func (p *StructParser) GetGoType() string {
	return p.goType
}

func (p *StructParser) GenGoStruct(s *strings.Builder, s1 string) {
	goStructType := title(s1)
	p.goType = goStructType
	//先遍历field生成所有嵌套类型
	for _, v := range p.fieldsArray {
		f := p.fields[v]
		f.GenGoStruct(s, goStructType+title(v))
	}

	fmt.Fprintf(s, "type %s struct {\n", goStructType)
	for _, v := range p.fieldsArray {
		f := p.fields[v]
		fmt.Fprintf(s, "\t%s %s `json:\"%s\"`\n", title(v), f.GetGoType(), v)
	}
	s.WriteString("}\n\n")
}

type goStruct struct {
	TableName string
	Data      string
	Package   string
	tmpl      *template.Template
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

func (j *goStruct) walkOk(writePath string) {
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
	err = j.tmpl.Execute(f, j)
	if err != nil {
		panic(err)
	} else {
		log.Printf("%s Write ok\n", filename)
	}
}

func (j *goStruct) outputGoJson(tmpl *template.Template, writePath string, colNames []string, types []string, rows [][]string, table *Table, idIndex int) {
	j.TableName = table.name
	j.tmpl = tmpl
	fields := []string{}
	for i := 0; i < len(colNames); i++ {
		fields = append(fields, fmt.Sprintf("%s:%s", strings.Split(colNames[i], ":")[0], types[i]))
	}
	str := "{" + strings.Join(fields, ",") + "}"
	p, err := MakeParser(str)
	if err != nil {
		panic(err)
	}
	p.GenGoStruct(&j.str, title(table.name))
}
