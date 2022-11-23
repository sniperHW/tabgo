package main

import "fmt"

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

func (a *Array) ToString(s string) string {
	s += "["
	for i, vv := range a.value {
		s = vv.ToString(s)
		if i != len(a.value)-1 {
			s += ","
		}
	}
	s += "]"
	return s
}

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

type Field struct {
	name  string
	value *Value
}

func (f *Field) ToString(s string) string {
	s += (f.name + ":")
	return f.value.ToString(s)
}

func (f *Field) ToLuaString(s string) string {
	s += (f.name + "=")
	return f.value.ToLuaString(s)
}

func (f *Field) ToJsonString(s string) string {
	s += fmt.Sprintf("\"%s\":", f.name)
	return f.value.ToJsonString(s)
}

type Struct struct {
	fields []*Field
}

func (ss *Struct) ToString(s string) string {
	s += "{"
	for i, vv := range ss.fields {
		s = vv.ToString(s)
		if i != len(ss.fields)-1 {
			s += ","
		}
	}
	s += "}"
	return s
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

type Value struct {
	valueType int
	value     interface{}
}

func (v *Value) ToString(s string) string {
	switch v.valueType {
	case typeArray:
		return v.value.(*Array).ToString(s)
	case typeStruct:
		return v.value.(*Struct).ToString(s)
	default:
		return s + fmt.Sprintf("%v", v.value)
	}
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

func main() {
	{

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
		if p, err := MakeParser("{x:int,y:int}"); err == nil {
			v := p.Parse("{x:1,y:2}")
			fmt.Println(v.ToString(""))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("{x:int,y:int[]}"); err == nil {
			v := p.Parse("{x:1,y:[1,2,3]}")
			fmt.Println(v.ToString(""))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("{x:string,y:{xx:int,yy:int[]}}"); err == nil {
			v := p.Parse("{x:1,y:{xx:2,yy:[1,2]}}")
			fmt.Println(v.ToLuaString(""))
			fmt.Println(v.ToJsonString(""))
		} else {
			fmt.Println(err)
		}
	}

	{
		if p, err := MakeParser("{x:int,y:int}[]"); err == nil {
			v := p.Parse("[{x:1,y:11},{x:2,y:22}]")
			fmt.Println(v.ToLuaString(""))
			fmt.Println(v.ToJsonString(""))
		} else {
			fmt.Println(err)
		}
	}

}
