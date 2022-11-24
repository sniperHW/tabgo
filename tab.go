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
	typeArray  = 4
	typeStruct = 5
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

func test() {
	/*{

		p := ArrayParser{
			subParser: ValueParser{
				valueType: typeInt,
			},
		}

		v := p.Parse("")

		fmt.Println(v.ToString(""))

	}

	{

		p := ArrayParser{
			subParser: ValueParser{
				valueType: typeInt,
			},
		}

		v := p.Parse("[]")

		fmt.Println(v.ToString(""))

	}

	{

		p := ArrayParser{
			subParser: ValueParser{
				valueType: typeInt,
			},
		}

		v := p.Parse("[1,2,3]")

		fmt.Println(v.ToString(""))

	}

	{

		p := ArrayParser{
			subParser: ArrayParser{
				subParser: ValueParser{
					valueType: typeInt,
				},
			},
		}

		v := p.Parse("[[1,2,3],[4,5,6]]")

		fmt.Println(v.ToString(""))

	}

	{
		p := StructParser{
			fields: map[string]Parser{},
		}

		p.fields["x"] = ValueParser{
			valueType: typeInt,
		}

		p.fields["y"] = ValueParser{
			valueType: typeString,
		}

		v := p.Parse("{x:1,y:hello}")
		fmt.Println(v.ToString(""))
	}

	{
		p := StructParser{
			fields: map[string]Parser{},
		}

		p.fields["x"] = ValueParser{
			valueType: typeInt,
		}

		p.fields["y"] = ArrayParser{
			subParser: ArrayParser{
				subParser: ValueParser{valueType: typeInt},
			},
		}

		v := p.Parse("{x:1,y:[[1,2],[3,4]]}")
		fmt.Println(v.ToString(""))
	}

	{

		p := StructParser{
			fields: map[string]Parser{},
		}

		p.fields["x"] = ValueParser{
			valueType: typeInt,
		}

		nest := StructParser{
			fields: map[string]Parser{},
		}

		nest.fields["nestX"] = ValueParser{
			valueType: typeInt,
		}

		nest.fields["nestY"] = ArrayParser{
			subParser: ArrayParser{
				subParser: ValueParser{valueType: typeInt},
			},
		}

		p.fields["y"] = nest

		v := p.Parse("{x:1,y:{nestX:100,nestY:[[1,2],[3,4]]}}")
		fmt.Println(v.ToString(""))

	}

	{
		if p, err := MakeParser("int[][]"); err == nil {
			v := p.Parse("[[1,2],[3,4]]")
			fmt.Println(v.ToString(""))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("test{x:int,y:int}"); err == nil {
			v := p.Parse("{x:1,y:2}")
			fmt.Println(v.ToString(""))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("test{x:int,y:int[]}"); err == nil {
			v := p.Parse("{x:1,y:[1,2,3]}")
			fmt.Println(v.ToString(""))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("{x:string,y:test{xx:int,yy:int[]}}"); err == nil {
			v := p.Parse("{x:1,y:{xx:2,yy:[1,2]}}")
			fmt.Println(v.ToLuaString(""))
			fmt.Println(v.ToJsonString(""))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("test{x:int,y:int}[]"); err == nil {
			v := p.Parse("[{x:1,y:11},{x:2,y:22}]")
			fmt.Println(v.ToLuaString(""))
			fmt.Println(v.ToJsonString(""))

			fmt.Println(p.(ArrayParser).subParser.(StructParser).structName)

		} else {
			fmt.Println(err)
		}
	}*/

	/*{
		if p, err := MakeParser("test{x:int,y:int}"); err == nil {
			fmt.Println(p.(StructParser).GenGoStruct("", "build"))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("test{x:int,y:int}[][]"); err == nil {
			fmt.Println(p.GetGoType(""))
		} else {
			fmt.Println(err)
		}
	}*/

	/*{
		if p, err := MakeParser("test{x:int,y:{x:int,y:int}}"); err == nil {
			fmt.Println(p.(StructParser).GenGoStruct("", "build"))
		} else {
			fmt.Println(err)
		}
	}*/
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
	funcOutput func(tmpl *template.Template, writePath string, rows [][]string, tab *Table, idIdx int)
	funcOk     func(string)
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
					name := strings.TrimSuffix(filename, ".xlsx")

					table := &Table{
						name: name,
					}

					xlsx, err := excelize.OpenFile(path.Join(w.loadPath, filename))
					if err != nil {
						panic(err)
					}

					rows := xlsx.GetRows(xlsx.GetSheetName(xlsx.GetActiveSheetIndex()))

					names := rows[0]
					types := rows[1]
					if len(rows) < 4 {
						return
					}
					rows = rows[3:]

					idIndex := -1

					for i := 0; i < len(names); i++ {
						if names[i] == "" {
							field := &Column{
								name: names[i],
							}
							table.fields = append(table.fields, field)
						} else if names[i] == "annotation" {
							field := &Column{
								name: names[i],
							}
							table.fields = append(table.fields, field)
						} else {
							if names[i] == "id" {
								idIndex = i
							}

							if parser, err := MakeParser(types[i]); err != nil {
								panic(fmt.Sprintf("MakeParserError:%v file:%s column:%s", err, filename, name[i]))
							} else {
								col := &Column{
									name:   names[i],
									parser: parser,
								}
								table.fields = append(table.fields, col)
							}
						}
					}

					if idIndex < 0 {
						panic("not id field")
					}

					w.funcOutput(w.tmpl, w.writePath, rows, table, idIndex)
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
	input := flag.String("input", "./table", "path of xlsx")
	output := flag.String("output", "./lua", "path of output files")
	gopackage := flag.String("package", "json", "package of go")
	mode := flag.String("mode", "lua", "lua|json|go")
	flag.Parse()

	var fn func(tmpl *template.Template, writePath string, rows [][]string, tab *Table, idIdx int)
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
	}

	w.walk()
}
