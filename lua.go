package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

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

func outputLua(tmpl *template.Template, writePath string, colNames []string, types []string, rows [][]string, table *Table, idIndex int) {
	var builder strings.Builder
	rr := 0
	for rowNum, row := range rows {
		if row[idIndex] != "" {
			if rr > 0 {
				builder.WriteString(",\n")
			}
			fmt.Fprintf(&builder, "\t[%v]={", row[0])
			cc := 0
			for i, field := range table.fields {
				if field.parser != nil {
					if v, err := field.parser.Parse(row[i]); err != nil {
						panic(fmt.Errorf("parse err:(%v) table:(%s) columm:(%s) types:(%s) row:(%d) str:(%s)", err, table.name, colNames[i], types[i], rowNum+DatasRow+1, row[i]))
					} else {
						if cc > 0 {
							builder.WriteString(",")
						}
						fmt.Fprintf(&builder, "%s=", field.name)
						builder.WriteString(v.ToLuaStr())
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
