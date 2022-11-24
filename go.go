package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

func (p ValueParser) GenGoStruct(s string, _ string) string {
	return s
}

func (p ValueParser) GetGoType(_ string) string {
	switch p.valueType {
	case typeInt:
		return "int"
	case typeString:
		return "string"
	case typeBool:
		return "bool"
	default:
		panic("error")
	}
}

func (p ArrayParser) GenGoStruct(s string, s1 string) string {
	return p.subParser.GenGoStruct(s, s1)
}

func (p ArrayParser) GetGoType(s string) string {
	return "[]" + p.subParser.GetGoType(s)
}

func (p StructParser) GetGoType(s string) string {
	return s + strings.Title(p.structName)
}

func (p StructParser) GenGoStruct(s string, s1 string) string {
	goStructType := p.GetGoType(strings.Title(s1))
	//先遍历field生成所有嵌套类型
	for k, v := range p.fields {
		s = v.GenGoStruct(s, goStructType+strings.Title(k))
	}

	s += fmt.Sprintf("type %s struct {\n", goStructType)
	for k, v := range p.fields {
		if _, ok := v.(StructParser); ok {
			s += fmt.Sprintf(" %s *%s `json:\"%s\"`\n", strings.Title(k), v.GetGoType(goStructType+strings.Title(k)), k)
		} else {
			s += fmt.Sprintf(" %s %s `json:\"%s\"`\n", strings.Title(k), v.GetGoType(""), k)
		}
	}
	s += "}\n\n"
	return s
}

type goStruct struct {
	gopackage string
	str       string
}

func (j *goStruct) walkOk(writePath string) {
	path := fmt.Sprintf("%s/%s", writePath, j.gopackage)
	filename := fmt.Sprintf("%s/%s.go", path, j.gopackage)
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
	defer f.Close()

	err = os.Truncate(filename, 0)
	if err != nil {
		panic(err)
	}

	f.WriteString(j.str)

	fmt.Println(fmt.Sprintf("write %s ok", filename))
}

func (j *goStruct) outputGoJson(tmpl *template.Template, writePath string, rows [][]string, table *Table, idIndex int) {
	//先生成所有结构体类型
	for _, field := range table.fields {
		if field.parser != nil {
			j.str = field.parser.GenGoStruct(j.str, strings.Title(table.name)+strings.Title(field.name))
		}
	}

	j.str += fmt.Sprintf("type %s struct {\n", strings.Title(table.name))
	for _, field := range table.fields {
		if field.parser != nil {
			if _, ok := field.parser.(StructParser); ok {
				prefix := strings.Title(table.name) + strings.Title(field.name)
				j.str += fmt.Sprintf(" %s *%s `json:\"%s\"` \n", strings.Title(field.name), field.parser.GetGoType(prefix), field.name)
			} else {
				j.str += fmt.Sprintf(" %s %s `json:\"%s\"` \n", strings.Title(field.name), field.parser.GetGoType(""), field.name)
			}
		}
	}
	j.str += "}\n\n"
}
