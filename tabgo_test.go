package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	p, _ := MakeParser("int[][]")

	{
		_, err := p.Parse("[1,2],[3,4]") //正确形式[[1,2],[3,4]]
		assert.NotNil(t, err)
	}

	{
		b := strings.Builder{}
		v, _ := p.Parse("[[1,2]]")
		v.ToJsonString(&b)
		assert.Equal(t, b.String(), "[[1,2]]")
	}

	{
		_, err := p.Parse("[[[1,2],[3,4]]]") //多了最外面一层[]
		assert.NotNil(t, err)
	}

	//结构嵌套数组
	p, _ = MakeParser("{x:int,y:int[][]}")

	{
		b := strings.Builder{}
		v, _ := p.Parse("{x:1,y:[[1,2],[3,4]]}")
		v.ToJsonString(&b)
		assert.Equal(t, b.String(), "{\"x\":1,\"y\":[[1,2],[3,4]]}")
	}

	p, _ = MakeParser("{x:int,y:{xx:int,yy:int}}")

	//结构嵌套结构
	{
		b := strings.Builder{}
		v, _ := p.Parse("{x:1,y:{xx:10,yy:11}}")
		v.ToJsonString(&b)
		assert.Equal(t, b.String(), "{\"x\":1,\"y\":{\"xx\":10,\"yy\":11}}")
	}

	//结构数组
	p, _ = MakeParser("{x:int,y:int}[]")
	{
		b := strings.Builder{}
		v, _ := p.Parse("[{x:1,y:11},{x:2,y:22}]")
		v.ToJsonString(&b)
		assert.Equal(t, b.String(), "[{\"x\":1,\"y\":11},{\"x\":2,\"y\":22}]")
	}

	//数组套结构套数组
	p, _ = MakeParser("{x:int,y:int[]}[]")
	{
		b := strings.Builder{}
		v, _ := p.Parse("[{x:1,y:[2,3,4]},{x:2,y:[22,23,24]}]")
		v.ToJsonString(&b)
		assert.Equal(t, b.String(), "[{\"x\":1,\"y\":[2,3,4]},{\"x\":2,\"y\":[22,23,24]}]")
	}
}
