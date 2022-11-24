package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

func (a *Array) ToLuaString(s *strings.Builder) {
	s.WriteString("{")
	for i, vv := range a.value {
		vv.ToLuaString(s)
		if i != len(a.value)-1 {
			s.WriteString(",")
		}
	}
	s.WriteString("}")
}

func (f *Field) ToLuaString(s *strings.Builder) {
	s.WriteString(f.name + "=")
	f.value.ToLuaString(s)
}

func (ss *Struct) ToLuaString(s *strings.Builder) {
	s.WriteString("{")
	for i, vv := range ss.fields {
		vv.ToLuaString(s)
		if i != len(ss.fields)-1 {
			s.WriteString(",")
		}
	}
	s.WriteString("}")
}

func (v *Value) ToLuaString(s *strings.Builder) {
	switch v.valueType {
	case typeArray:
		v.value.(*Array).ToLuaString(s)
	case typeStruct:
		v.value.(*Struct).ToLuaString(s)
	case typeString:
		s.WriteString(fmt.Sprintf("\"%v\"", v.value))
	default:
		s.WriteString(fmt.Sprintf("%v", v.value))
	}
}

type lua struct {
	TableName string
	Data      string
}

var luaTemplate string = `
local {{.TableName}} = {
{{.Data}}	
}

return {{.TableName}}
`

func outputLua(tmpl *template.Template, writePath string, rows [][]string, table *Table, idIndex int) {
	var builder strings.Builder
	rr := 0
	for rowNum, row := range rows {
		if row[idIndex] != "" {
			if rr > 0 {
				builder.WriteString(",\n")
			}
			builder.WriteString(fmt.Sprintf("\t[%v]={", row[0]))
			cc := 0
			for i, field := range table.fields {
				if field.parser != nil {
					if v, err := field.parser.Parse(row[i]); err != nil {
						panic(fmt.Errorf("parse err:%v table:%s columm:%d row:%d str:%s", err, table.name, i, rowNum+3, row[i]))
					} else {
						if v.valueType == typeStruct && len(v.value.(*Struct).fields) == 0 {
							continue
						}
						if cc > 0 {
							builder.WriteString(",")
						}
						builder.WriteString(fmt.Sprintf("%s=", field.name))
						v.ToJsonString(&builder)
						cc++

					}
				}
			}

			builder.WriteString("}")
			rr++
		}
	}

	filename := fmt.Sprintf("%s/%s.lua", writePath, table.name)
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

	err = tmpl.Execute(f, lua{table.name, builder.String()})
	if err != nil {
		panic(err)
	} else {
		log.Printf("%s Write ok\n", filename)
	}
}
