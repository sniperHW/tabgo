package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

func (a *Array) ToJsonString(s string) string {
	s += "["
	for i, vv := range a.value {
		s = vv.ToJsonString(s)
		if i != len(a.value)-1 {
			s += ","
		}
	}
	s += "]"
	return s
}

func (f *Field) ToJsonString(s string) string {
	s += fmt.Sprintf("\"%s\":", f.name)
	return f.value.ToJsonString(s)
}

func (ss *Struct) ToJsonString(s string) string {
	s += "{"
	for i, vv := range ss.fields {
		s = vv.ToJsonString(s)
		if i != len(ss.fields)-1 {
			s += ","
		}
	}
	s += "}"
	return s
}

func (v *Value) ToJsonString(s string) string {
	switch v.valueType {
	case typeArray:
		return v.value.(*Array).ToJsonString(s)
	case typeStruct:
		return v.value.(*Struct).ToJsonString(s)
	case typeString:
		return s + fmt.Sprintf("\"%v\"", v.value)
	default:
		return s + fmt.Sprintf("%v", v.value)
	}
}

type json struct {
	Data string
}

var jsonTemplate string = `
{
{{.Data}}	
}
`

func outputJson(tmpl *template.Template, writePath string, rows [][]string, table *Table, idIndex int) {
	rowsStr := []string{}
	for rowNum, row := range rows {
		if row[idIndex] != "" {
			fieldsStr := []string{}
			for i, field := range table.fields {
				if field.parser != nil {
					if v, err := field.parser.Parse(row[i]); err != nil {
						panic(fmt.Errorf("parse err:%v table:%s columm:%d row:%d", err, table.name, i, rowNum+3))
					} else {
						if v.valueType == typeStruct && len(v.value.(*Struct).fields) == 0 {
							continue
						}
						fieldsStr = append(fieldsStr, v.ToJsonString(fmt.Sprintf("\"%s\":", field.name)))
					}
				}
			}
			rowsStr = append(rowsStr, fmt.Sprintf("	\"%v\":{%s}", row[0], strings.Join(fieldsStr, ",")))
		}
	}
	filename := fmt.Sprintf("%s/%s.json", writePath, table.name)
	os.MkdirAll(writePath, os.ModePerm)
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

	err = tmpl.Execute(f, json{strings.Join(rowsStr, ",\n")})
	if err != nil {
		panic(err)
	} else {
		log.Printf("%s Write ok\n", filename)
	}
}
