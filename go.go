package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/sniperHW/tabgo/parser"
)

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
	defer func() {
		f.Close()
		cmd := exec.Command("gofmt", "-w", filename)
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
		}
	}()

	err = os.Truncate(filename, 0)
	if err != nil {
		panic(err)
	}

	f.WriteString(j.str.String())

	fmt.Printf("write %s ok\n", filename)
}

func (j *goStruct) outputGoJson(tmpl *template.Template, writePath string, colNames []string, types []string, rows [][]string, table *Table, idIndex int) {
	fields := []string{}
	for i := 0; i < len(colNames); i++ {
		fields = append(fields, fmt.Sprintf("%s:%s", strings.Split(colNames[i], ":")[0], types[i]))
	}
	str := "{" + strings.Join(fields, ",") + "}"
	p, err := parser.MakeParser(str)
	if err != nil {
		panic(err)
	}
	j.str.WriteString(p.GenGoDefine(strings.Title(table.name)))
}
