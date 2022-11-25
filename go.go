package main

import (
	"fmt"
	"os"
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
	case typeFloat:
		return "float64"
	default:
		panic("error")
	}
}

func (p ArrayParser) GenGoStruct(s string, s1 string) string {
	return p.child.GenGoStruct(s, s1)
}

func (p ArrayParser) GetGoType(s string) string {
	return "[]" + p.child.GetGoType(s)
}

func (p StructParser) GetGoType(s string) string {
	return s
}

func (p StructParser) GenGoStruct(s string, s1 string) string {
	goStructType := p.GetGoType(title(s1))
	//先遍历field生成所有嵌套类型
	for k, v := range p.fields {
		s = v.GenGoStruct(s, goStructType+title(k))
	}

	s += fmt.Sprintf("type %s struct {\n", goStructType)
	for k, v := range p.fields {
		if _, ok := v.(StructParser); ok {
			s += fmt.Sprintf("\t%s *%s `json:\"%s\"`\n", title(k), v.GetGoType(goStructType+title(k)), k)
		} else {
			s += fmt.Sprintf("\t%s %s `json:\"%s\"`\n", title(k), v.GetGoType(""), k)
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

	fmt.Printf("write %s ok\n", filename)
}

func (j *goStruct) outputGoJson(tmpl *template.Template, writePath string, colNames []string, rows [][]string, table *Table, idIndex int) {
	//先生成所有结构体类型
	for _, field := range table.fields {
		if field.parser != nil {
			j.str = field.parser.GenGoStruct(j.str, title(table.name)+title(field.name))
		}
	}

	j.str += fmt.Sprintf("type %s struct {\n", title(table.name))
	for _, field := range table.fields {
		if field.parser != nil {
			if _, ok := field.parser.(StructParser); ok {
				prefix := title(table.name) + title(field.name)
				j.str += fmt.Sprintf("\t%s *%s `json:\"%s\"` \n", title(field.name), field.parser.GetGoType(prefix), field.name)
			} else {
				j.str += fmt.Sprintf("\t%s %s `json:\"%s\"` \n", title(field.name), field.parser.GetGoType(""), field.name)
			}
		}
	}
	j.str += "}\n\n"
}
