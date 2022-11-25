package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	p, _ := MakeParser("int[][]")

	{
		v, err := p.Parse("[1,2],[3,4]")
		fmt.Println(v, err)
	}

	{
		b := strings.Builder{}
		v, _ := p.Parse("[[1,2],[3,4]]")
		v.ToJsonString(&b)
		fmt.Println(b.String())
	}

	p, _ = MakeParser("{x:int,y:int[][]}")

	{
		b := strings.Builder{}
		v, _ := p.Parse("{x:1,y:[[1,2],[3,4]]}")
		v.ToLuaString(&b)
		fmt.Println(b.String())
	}

}
