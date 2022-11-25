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
)

const (
	typeInt    = 1
	typeString = 2
	typeBool   = 3
	typeFloat  = 4
	typeArray  = 5
	typeStruct = 6
)

type Array struct {
	value []*Value
}

type Field struct {
	name  string
	value *Value
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
	parser Parser
}

type Table struct {
	name   string
	fields []*Column
}

type Walker struct {
	loadPath   string
	writePath  string
	tmpl       *template.Template
	funcOutput func(*template.Template, string, []string, [][]string, *Table, int)
	funcOk     func(string)
	ignore     map[string]bool
}

const NamesRow = 0  //名字定义所在的行
const TypesRow = 1  //类型定义所在行
const DatasRow = 3  //数据起始行
const IdName = "id" //索引列的名字

func (w *Walker) checkColumn(s string) (string, bool) {
	v := strings.Split(s, ":")
	if v[0] == "" {
		//名字为空字符串
		return "", false
	} else if len(v) > 1 && w.ignore[v[1]] {
		//标记在忽略列表中
		return "", false
	} else {
		return v[0], true
	}
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
					if len(rows) <= DatasRow {
						return
					}
					rows = rows[DatasRow:]

					idIndex := -1

					for i := 0; i < len(names); i++ {

						if colName, ok := w.checkColumn(names[i]); ok {
							if colName == IdName {
								idIndex = i
							}
							if parser, err := MakeParser(types[i]); err != nil {
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

					w.funcOutput(w.tmpl, w.writePath, names, rows, table, idIndex)
				}
			}()
		}
		return nil
	}); err != nil {
		panic(err)
	}
	wait.Wait()
	if w.funcOk != nil {
		w.funcOk(w.writePath)
	}
}

func main() {
	input := flag.String("input", "./excel", "path of xlsx")
	output := flag.String("output", "./lua", "path of output files")
	gopackage := flag.String("package", "json", "package of go")
	mode := flag.String("mode", "json", "lua|json|go")
	serverOnly := flag.String("server", "false", "true|false")
	flag.Parse()

	var fn func(tmpl *template.Template, writePath string, colNames []string, rows [][]string, tab *Table, idIdx int)
	var walkOk func(writePath string)
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
			gopackage: *gopackage,
			str:       fmt.Sprintf("package %s\n\n", *gopackage),
		}
		fn = j.outputGoJson
		walkOk = j.walkOk
	default:
		panic("unsupport mode")
	}

	w := &Walker{
		loadPath:   *input,
		writePath:  *output,
		tmpl:       tmpl,
		funcOutput: fn,
		funcOk:     walkOk,
		ignore:     map[string]bool{"annotation": true},
	}

	if *serverOnly == "true" {
		//打服务端表，将所有标记为client的字段加入忽略列表
		w.ignore["client"] = true
	}

	w.walk()
}
