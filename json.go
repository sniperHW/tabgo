package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

func (a *Array) ToJsonString(s *strings.Builder) {
	s.WriteString("[")
	for i, vv := range a.value {
		vv.ToJsonString(s)
		if i != len(a.value)-1 {
			s.WriteString(",")
		}
	}
	s.WriteString("]")
}

func (f *Field) ToJsonString(s *strings.Builder) {
	s.WriteString(fmt.Sprintf("\"%s\":", f.name))
	f.value.ToJsonString(s)
}

func (ss *Struct) ToJsonString(s *strings.Builder) {
	s.WriteString("{")
	for i, vv := range ss.fields {
		vv.ToJsonString(s)
		if i != len(ss.fields)-1 {
			s.WriteString(",")
		}
	}
	s.WriteString("}")
}

func (v *Value) ToJsonString(s *strings.Builder) {
	switch v.valueType {
	case typeArray:
		v.value.(*Array).ToJsonString(s)
	case typeStruct:
		v.value.(*Struct).ToJsonString(s)
	case typeString:
		s.WriteString(fmt.Sprintf("\"%v\"", v.value))
	default:
		s.WriteString(fmt.Sprintf("%v", v.value))
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
	var builder strings.Builder
	rr := 0
	for rowNum, row := range rows {
		if row[idIndex] != "" {

			if rr > 0 {
				builder.WriteString(",\n")
			}

			builder.WriteString(fmt.Sprintf("	\"%v\":{", row[0]))
			cc := 0
			for i, field := range table.fields {
				if field.parser != nil {
					if v, err := field.parser.Parse(row[i]); err != nil {
						panic(fmt.Errorf("parse err:%v table:%s columm:%d row:%d", err, table.name, i, rowNum+3))
					} else {
						if v.valueType == typeStruct && len(v.value.(*Struct).fields) == 0 {
							continue
						}
						if cc > 0 {
							builder.WriteString(",")
						}
						builder.WriteString(fmt.Sprintf("\"%s\":", field.name))
						v.ToJsonString(&builder)
						cc++
					}
				}
			}

			builder.WriteString("}")
			rr++
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

	err = tmpl.Execute(f, json{Data: builder.String()})
	if err != nil {
		panic(err)
	} else {
		log.Printf("%s Write ok\n", filename)
	}
}
