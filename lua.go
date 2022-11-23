package main

import "fmt"

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
