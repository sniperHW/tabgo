package main

import (
	"fmt"
	"strings"
)

func (p ValueParser) GenGoStruct(s string, _ string) string {
	return s
}

func (p ValueParser) GetGoType(_ string) string {
	switch p.valueType {
	case typeInt:
		return "int"
	case typeString:
		return "string"
	case typeBool:
		return "bool"
	default:
		panic("error")
	}
}

func (p ArrayParser) GenGoStruct(s string, s1 string) string {
	return p.subParser.GenGoStruct(s, s1)
}

func (p ArrayParser) GetGoType(s string) string {
	return "[]" + p.subParser.GetGoType(s)
}

func (p StructParser) GetGoType(s string) string {
	return s + strings.Title(p.structName)
}

func (p StructParser) GenGoStruct(s string, s1 string) string {
	goStructType := p.GetGoType(strings.Title(s1))
	//先遍历field生成所有嵌套类型
	for k, v := range p.fields {
		s = v.GenGoStruct(s, goStructType+strings.Title(k))
	}

	s += fmt.Sprintf("type %s struct {\n", goStructType)
	for k, v := range p.fields {
		s += fmt.Sprintf(" %s %s `json:\"%s\"`\n", strings.Title(k), v.GetGoType(goStructType+strings.Title(k)), k)
	}
	s += "}\n\n"
	return s
}
