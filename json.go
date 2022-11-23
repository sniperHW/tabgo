package main

import "fmt"

func (a *Array) ToJsonString(s string) string {
	s += "["
	for i, vv := range a.value {
		s = vv.ToJsonString(s)
		if i != len(a.value)-1 {
			s += ","
		}
	}
	s += "]"
	return s
}

func (f *Field) ToJsonString(s string) string {
	s += fmt.Sprintf("\"%s\":", f.name)
	return f.value.ToJsonString(s)
}

func (ss *Struct) ToJsonString(s string) string {
	s += "{"
	for i, vv := range ss.fields {
		s = vv.ToJsonString(s)
		if i != len(ss.fields)-1 {
			s += ","
		}
	}
	s += "}"
	return s
}

func (v *Value) ToJsonString(s string) string {
	switch v.valueType {
	case typeArray:
		return v.value.(*Array).ToJsonString(s)
	case typeStruct:
		return v.value.(*Struct).ToJsonString(s)
	case typeString:
		return s + fmt.Sprintf("\"%v\"", v.value)
	default:
		return s + fmt.Sprintf("%v", v.value)
	}
}
