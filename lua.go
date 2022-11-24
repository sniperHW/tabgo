package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

func (a *Array) ToLuaString(s string) string {
	s += "{"
	for i, vv := range a.value {
		s = vv.ToLuaString(s)
		if i != len(a.value)-1 {
			s += ","
		}
	}
	s += "}"
	return s
}

func (f *Field) ToLuaString(s string) string {
	s += (f.name + "=")
	return f.value.ToLuaString(s)
}

func (ss *Struct) ToLuaString(s string) string {
	s += "{"
	for i, vv := range ss.fields {
		s = vv.ToLuaString(s)
		if i != len(ss.fields)-1 {
			s += ","
		}
	}
	s += "}"
	return s
}

func (v *Value) ToLuaString(s string) string {
	switch v.valueType {
	case typeArray:
		return v.value.(*Array).ToLuaString(s)
	case typeStruct:
		return v.value.(*Struct).ToLuaString(s)
	case typeString:
		return s + fmt.Sprintf("\"%v\"", v.value)
	default:
		return s + fmt.Sprintf("%v", v.value)
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
	rowsStr := []string{}
	for rowNum, row := range rows {
		if row[idIndex] != "" {
			fieldsStr := []string{}
			for i, field := range table.fields {
				if field.parser != nil {
					if v, err := field.parser.Parse(row[i]); err != nil {
						panic(fmt.Errorf("parse err:%v table:%s columm:%d row:%d str:%s", err, table.name, i, rowNum+3, row[i]))
					} else {
						if v.valueType == typeStruct && len(v.value.(*Struct).fields) == 0 {
							continue
						}
						fieldsStr = append(fieldsStr, v.ToLuaString(fmt.Sprintf("%s=", field.name)))
					}
				}
			}
			rowsStr = append(rowsStr, fmt.Sprintf("	[%v]={%s}", row[0], strings.Join(fieldsStr, ",")))
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

	err = tmpl.Execute(f, lua{table.name, strings.Join(rowsStr, ",\n")})
	if err != nil {
		panic(err)
	} else {
		log.Printf("%s Write ok\n", filename)
	}
}
