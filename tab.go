package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/sniperHW/tabgo/parser"
)

type Array struct {
	value []*parser.Value
}

type Field struct {
	name  string
	value *parser.Value
}

type Struct struct {
	fields []*Field
}

type Value struct {
	valueType int
	value     interface{}
}

type Column struct {
	name   string
	parser *parser.Parser
}

type Table struct {
	name   string
	fields []*Column
}

type Walker struct {
	loadPath           string
	writePath          string
	tmpl               *template.Template
	funcOutput         func(*template.Template, string, []string, []string, [][]string, *Table, int)
	funcTableProcessed func([]string, []string, *Table)
	funcOk             func(string, *template.Template)
	serverOnly         bool
}

const NamesRow = 0  //名字定义所在的行
const TypesRow = 1  //类型定义所在行
const TagsRow = 2   //标记所在行
const DatasRow = 3  //数据起始行
const IdName = "id" //索引列的名字

func (w *Walker) checkColumn(name string, tag string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	tag = strings.TrimSpace(tag)
	if tag == "ignore" {
		return "", false
	}
	if name == IdName {
		return name, true
	}
	if w.serverOnly {
		return name, tag != "c"
	}
	return name, tag != "s"
}

func (w *Walker) walk() {
	var wait sync.WaitGroup
	if err := filepath.Walk(w.loadPath, func(filePath string, f os.FileInfo, _ error) error {
		if f != nil && !f.IsDir() {
			wait.Add(1)
			go func() {
				filename := f.Name()
				defer func() {
					wait.Done()
				}()
				if strings.Contains(filename, ".xlsx") {
					table := &Table{
						name: strings.TrimSuffix(filename, ".xlsx"),
					}

					xlsx, err := excelize.OpenFile(path.Join(w.loadPath, filename))
					if err != nil {
						panic(err)
					}

					rows := xlsx.GetRows(xlsx.GetSheetName(xlsx.GetActiveSheetIndex()))

					names := rows[NamesRow]
					types := rows[TypesRow]
					var tags []string
					if len(rows) > TagsRow {
						tags = rows[TagsRow]
					}
					if len(rows) <= DatasRow {
						return
					}
					rows = rows[DatasRow:]

					idIndex := -1

					for i := 0; i < len(names); i++ {
						var tag string
						if i < len(tags) {
							tag = tags[i]
						}
						if colName, ok := w.checkColumn(names[i], tag); ok {
							if colName == IdName {
								idIndex = i
							}
							if parser, err := parser.MakeParser(types[i]); err != nil {
								panic(fmt.Sprintf("MakeParserError:%v file:%v column:%v", err, filename, names[i]))
							} else {
								col := &Column{
									name:   colName,
									parser: parser,
								}
								table.fields = append(table.fields, col)
							}
						} else {
							table.fields = append(table.fields, &Column{})
						}
					}

					if idIndex < 0 {
						panic("not id field")
					}

					// 如果有表结构处理回调（Go 模式），先调用它生成类型定义
					if w.funcTableProcessed != nil {
						w.funcTableProcessed(names, types, table)
					}

					// 如果有输出回调（JSON/Lua 模式），处理数据行
					if w.funcOutput != nil {
						w.funcOutput(w.tmpl, w.writePath, names, types, rows, table, idIndex)
					}
				}
			}()
		}
		return nil
	}); err != nil {
		panic(err)
	}
	wait.Wait()
	if w.funcOk != nil {
		w.funcOk(w.writePath, w.tmpl)
	}
}

func main() {
	input := flag.String("input", "./excel", "path of xlsx")
	output := flag.String("output", "./lua", "path of output files")
	gopackage := flag.String("package", "json", "package of go")
	mode := flag.String("mode", "json", "lua|json|go")
	serverOnly := flag.String("server", "false", "true|false")
	flag.Parse()

	var fn func(tmpl *template.Template, writePath string, colNames []string, types []string, rows [][]string, tab *Table, idIdx int)
	var tableProcessed func(colNames []string, types []string, tab *Table)
	var walkOk func(writePath string, tmpl *template.Template)
	var tmpl *template.Template
	var err error

	switch *mode {
	case "lua":
		fn = outputLua
		tmpl, err = template.New("test").Parse(luaTemplate)
		if err != nil {
			panic(err)
		}
	case "json":
		fn = outputJson
		tmpl, err = template.New("test").Parse(jsonTemplate)
		if err != nil {
			panic(err)
		}
	case "go":
		j := &goStruct{
			Package: *gopackage,
			str:     strings.Builder{},
		}
		tableProcessed = j.processTable
		tmpl, err = template.New("test").Parse(goTemplate)
		if err != nil {
			panic(err)
		}
		// 创建一个闭包来传递 tmpl 参数
		finalTmpl := tmpl
		walkOk = func(writePath string, t *template.Template) {
			j.walkOk(writePath, finalTmpl)
		}
	default:
		panic("unsupport mode")
	}

	w := &Walker{
		loadPath:           *input,
		writePath:          *output,
		tmpl:               tmpl,
		funcOutput:         fn,
		funcTableProcessed: tableProcessed,
		funcOk:             walkOk,
		serverOnly:         *serverOnly == "true",
	}

	w.walk()
}
