package main

import (
	"fmt"
	"os"
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
	p.goType = "*" + goStructType
	//先遍历field生成所有嵌套类型
	for k, v := range p.fields {
		v.GenGoStruct(s, goStructType+title(k))
	}

	fmt.Fprintf(s, "type %s struct {\n", goStructType)
	for k, v := range p.fields {
		switch v.(type) {
		case *StructParser:
			fmt.Fprintf(s, "\t%s %s `json:\"%s\"`\n", title(k), v.GetGoType(), k)
		case *ArrayParser:
			switch v.(*ArrayParser).elements.(type) {
			case *StructParser:
				fmt.Fprintf(s, "\t%s %s `json:\"%s\"`\n", title(k), v.GetGoType(), k)
			default:
				fmt.Fprintf(s, "\t%s %s `json:\"%s\"`\n", title(k), v.GetGoType(), k)
			}
		default:
			fmt.Fprintf(s, "\t%s %s `json:\"%s\"`\n", title(k), v.GetGoType(), k)
		}
	}
	s.WriteString("}\n\n")
}

type goStruct struct {
	gopackage string
	str       strings.Builder
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

	f.WriteString(j.str.String())

	fmt.Printf("write %s ok\n", filename)
}

func (j *goStruct) outputGoJson(tmpl *template.Template, writePath string, colNames []string, types []string, rows [][]string, table *Table, idIndex int) {
	//先生成所有结构体类型
	for _, field := range table.fields {
		if field.parser != nil {
			field.parser.GenGoStruct(&j.str, title(table.name)+title(field.name))
		}
	}

	fmt.Fprintf(&j.str, "type %s struct {\n", title(table.name))
	for _, field := range table.fields {
		if field.parser != nil {
			switch field.parser.(type) {
			case *StructParser:
				fmt.Fprintf(&j.str, "\t%s %s `json:\"%s\"` \n", title(field.name), field.parser.GetGoType(), field.name)
			case *ArrayParser:
				switch field.parser.(*ArrayParser).elements.(type) {
				case *StructParser:
					fmt.Fprintf(&j.str, "\t%s %s `json:\"%s\"` \n", title(field.name), field.parser.GetGoType(), field.name)
				default:
					fmt.Fprintf(&j.str, "\t%s %s `json:\"%s\"` \n", title(field.name), field.parser.GetGoType(), field.name)
				}
			default:
				fmt.Fprintf(&j.str, "\t%s %s `json:\"%s\"` \n", title(field.name), field.parser.GetGoType(), field.name)
			}
		}
	}
	j.str.WriteString("}\n\n")
}
