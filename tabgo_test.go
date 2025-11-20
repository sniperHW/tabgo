package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeParser(t *testing.T) {
	{
		_, err := MakeParser(" int[][] ")
		assert.Nil(t, err)
	}
	{
		_, err := MakeParser("int[ ][]")
		assert.NotNil(t, err)
		fmt.Println(err)
	}
	{
		_, err := MakeParser("int[] []")
		assert.Nil(t, err)
	}

	{
		_, err := MakeParser(" { x : int, y : int[] [] } ")
		assert.Nil(t, err)
	}

	{
		_, err := MakeParser(`{ 
								x : int , 
								y : int[] [] 
								} `)
		assert.Nil(t, err)
	}
}

func TestParseGo(t *testing.T) {
	{
		p, _ := MakeParser("{x:int,y:{x:int,y:int},array:{x:int,y:{xx:int}[]}[]}")
		s := p.GenGoStruct("", "f")
		fmt.Println(s)
	}
}

func TestParse(t *testing.T) {

	{
		p, _ := MakeParser("{x:string}")
		v, err := p.Parse("{x:\"你好\"}") //正确形式[[1,2],[3,4]]
		sb := strings.Builder{}
		fmt.Println(err)
		v.ToJsonString(&sb)
		fmt.Println(sb.String())
	}

	{
		p, _ := MakeParser("string[]")
		v, err := p.Parse("[\"你好\",\"世界\"]") //正确形式[[1,2],[3,4]]
		sb := strings.Builder{}
		fmt.Println(err)
		v.ToJsonString(&sb)
		fmt.Println(sb.String())
	}

	{
		p, _ := MakeParser("{x:int,y:{x:int,y:int},array:int[]}")
		v, err := p.Parse("{x:1,y:{x:2,y:3},array:[1,2,3,4]}") //正确形式[[1,2],[3,4]]
		sb := strings.Builder{}
		fmt.Println(err)
		v.ToJsonString(&sb)
		fmt.Println(sb.String())
	}

	{
		p, err := MakeParser("{x:int,y:{x:int,y:int},array:{x:int,y:int}[]}")
		fmt.Println(p, err)
		v, err := p.Parse("{x:1,y:{x:2,y:3},array:[{x:1,y:2},{x:3,y:4}]}")
		sb := strings.Builder{}
		fmt.Println(err)
		v.ToJsonString(&sb)
		fmt.Println(sb.String())
	}

	{
		p, _ := MakeParser(" int[][] ")

		{
			_, err := p.Parse(" [1,2],[3,4] ") //正确形式[[1,2],[3,4]]
			fmt.Println(err)
			assert.NotNil(t, err)
		}

		{
			b := strings.Builder{}
			v, _ := p.Parse(" [ [ 1 , 2 ] ] ")
			v.ToJsonString(&b)
			assert.Equal(t, b.String(), "[[1,2]]")
		}

		{
			_, err := p.Parse("[[[1,2],[3,4]]]") //多了最外面一层[]
			assert.NotNil(t, err)
		}
	}
	{
		//结构嵌套数组
		p, _ := MakeParser("{x:int,y:int[][]}")

		{
			b := strings.Builder{}
			v, _ := p.Parse(`{
							x: 1 ,
							y: [[1,2],[3,4]] 
							}`)

			v.ToJsonString(&b)
			assert.Equal(t, b.String(), "{\"x\":1,\"y\":[[1,2],[3,4]]}")
		}
	}

	p, _ := MakeParser("{x:int,y:{xx:int,yy:int}}")

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

	p, _ = MakeParser("{x:int,y:string,z:int}")

	{
		b := strings.Builder{}
		v, err := p.Parse("{x:1,y:  \"hello\\\"\"  ,z:10}")
		if nil != err {
			fmt.Println(err)
		} else {
			v.ToJsonString(&b)
			fmt.Println(b.String())
		}
	}

	p, _ = MakeParser("string[]")
	{
		v, err := p.Parse("[a,b,c]")
		fmt.Println("p.Parse(\"[a,b,c]\")", v, err)
	}

	{
		b := strings.Builder{}
		v, _ := p.Parse("[\"a,b,c\"]")
		v.ToJsonString(&b)
		fmt.Println(b.String())
	}

	{
		b := strings.Builder{}
		v, _ := p.Parse("[\"a,b,c\\\"\",\"e,f,g\"]")
		v.ToJsonString(&b)
		fmt.Println(b.String())
	}

	{
		b := strings.Builder{}
		v, _ := p.Parse("[ \"a\" , \"b\" ]")
		v.ToJsonString(&b)
		fmt.Println(b.String())
	}

}
