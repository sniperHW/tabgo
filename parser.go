package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 从字符串生成Value
type Parser interface {
	Parse(string) *Value
	ValueType() int
	GetGoType(string) string           //获取go类型
	GenGoStruct(string, string) string //生成go结构体
}

type ValueParser struct {
	valueType int
}

func (p ValueParser) ValueType() int {
	return p.valueType
}

func (p ValueParser) Parse(s string) *Value {
	v := &Value{valueType: p.valueType}
	switch p.valueType {
	case typeInt:
		if s == "" {
			v.value = 0
		} else {
			vv, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("err:%v value:%s", err, s))
			}
			v.value = vv
		}
	case typeBool:
		if s == "" {
			v.value = false
		} else {
			vv, err := strconv.ParseBool(s)
			if err != nil {
				panic(fmt.Sprintf("err:%v value:%s", err, s))
			}
			v.value = vv
		}
	case typeString:
		v.value = s
	default:
		panic("invaild type")
	}

	return v
}

type ArrayParser struct {
	subParser Parser
}

func (p ArrayParser) ValueType() int {
	return typeArray
}

func (p ArrayParser) splitCompose(s string, bracket string) (ret []string) {
	left := -1
	for i := 0; i < len(s); i++ {
		if s[i] == bracket[0] {
			if left != -1 {
				panic("error1")
			} else {
				left = i
			}
		} else if s[i] == bracket[1] {
			if left == -1 {
				panic("error2")
			} else {
				sub := s[left : i+1]
				if sub != bracket {
					ret = append(ret, sub)
				}
				left = -1
			}
		}
	}
	return ret
}

func (p ArrayParser) split(s string) (ret []string) {
	if s[0] != '[' || s[len(s)-1] != ']' {
		panic("error0")
	}
	s = s[1 : len(s)-1] //去掉头尾括号
	switch p.subParser.(type) {
	case ArrayParser:
		ret = p.splitCompose(s, "[]")
	case StructParser:
		ret = p.splitCompose(s, "{}")
	default:
		ret = strings.Split(s, ",")
	}
	return ret
}

func (p ArrayParser) Parse(s string) *Value {
	v := &Value{valueType: typeArray}
	array := &Array{}
	if !(s == "" || s == "[]") {
		e := p.split(s)
		for _, vv := range e {
			array.value = append(array.value, p.subParser.Parse(vv))
		}
	}
	v.value = array
	return v
}

type StructParser struct {
	structName string
	fields     map[string]Parser
}

func (p StructParser) ValueType() int {
	return typeStruct
}

func (p StructParser) readFieldName(s string) (string, string, error) {
	o := -1
	for i := 0; i < len(s); i++ {
		if o == -1 && s[i] >= 'a' && s[i] <= 'z' {
			o = i
		} else if s[i] == ':' {
			if o == -1 {
				return "", "", errors.New("ErrReadFieldName")
			} else {
				return s[o:i], s[i+1:], nil
			}
		}
	}
	return "", "", errors.New("ErrReadFieldName")
}

func (p StructParser) readComposeFiledValue(s string, parser Parser, bracket string) (*Value, string, error) {
	if s[0] != bracket[0] {
		return nil, "", errors.New("ErrReadFieldValue1")
	}
	i := 1
	leftCount := 1
	for ; i < len(s); i++ {
		if s[i] == bracket[0] {
			leftCount++
		} else if s[i] == bracket[1] {
			leftCount--
		} else if s[i] == ',' && leftCount == 0 {
			return parser.Parse(s[:i]), s[:i+1], nil
		}
	}

	if leftCount == 0 {
		return parser.Parse(s[:i]), "", nil
	} else {
		return nil, "", errors.New("ErrReadFieldValue3")
	}
}

func (p StructParser) readFieldValue(s string, parser Parser) (*Value, string, error) {
	switch parser.(type) {
	case ValueParser:
		i := 0
		for ; i < len(s); i++ {
			if s[i] == ',' {
				return parser.Parse(s[:i]), s[i+1:], nil
			}
		}
		return parser.Parse(s[:i]), "", nil
	case ArrayParser:
		return p.readComposeFiledValue(s, parser, "[]")
	case StructParser:
		return p.readComposeFiledValue(s, parser, "{}")
	default:
		return nil, "", errors.New("ErrReadFieldValue7")
	}
}

func (p StructParser) Parse(s string) *Value {
	v := &Value{valueType: typeStruct}
	st := &Struct{}
	if !(s == "" || s == "{}") {
		if s[0] != '{' || s[len(s)-1] != '}' {
			panic("error0")
		}
		s = s[1 : len(s)-1] //去掉头尾括号
		for s != "" {
			var name string
			var err error
			var field *Value

			name, s, err = p.readFieldName(s)
			if err != nil {
				panic(err)
			}

			if fieldParser, ok := p.fields[name]; !ok {
				panic(fmt.Sprintf("invaild field:%s", name))
			} else {

				field, s, err = p.readFieldValue(s, fieldParser)
				if err != nil {
					panic(err)
				}
				st.fields = append(st.fields, &Field{
					name:  name,
					value: field,
				})
			}
		}
	}
	v.value = st
	return v
}

func splitNameType(s string) (string, string, error) {
	idx := strings.Index(s, ":")
	if idx > 0 && idx < len(s)-1 {
		return s[:idx], s[idx+1:], nil
	} else {
		return "", "", errors.New("invaild define")
	}
}

func readStructField(s string) (string, string, string, error) {
	leftCount := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			leftCount++
		} else if s[i] == '}' {
			leftCount--
		} else if s[i] == ',' && leftCount == 0 {
			name, value, err := splitNameType(s[:i])
			return name, value, s[i+1:], err
		}
	}

	if leftCount == 0 {
		name, value, err := splitNameType(s)
		return name, value, "", err
	} else {
		return "", "", "", errors.New("invaild define")
	}
}

// 分离结构名和结构体定义
func splitStructName(s string) (string, string, error) {
	if idx := strings.Index(s, "{"); idx > 0 && idx < len(s) {
		name := s[:idx]
		typeStr := s[idx:]
		if name == "" || !(typeStr[0] == '{' && typeStr[len(typeStr)-1] == '}') {
			return "", "", errors.New("invaild define")
		} else {
			return name, typeStr, nil
		}
	} else {
		return "", "", errors.New("invaild define")
	}
}

func MakeParser(s string) (Parser, error) {
	switch s {
	case "int":
		return ValueParser{valueType: typeInt}, nil
	case "string":
		return ValueParser{valueType: typeString}, nil
	case "bool":
		return ValueParser{valueType: typeBool}, nil
	default:
		if strings.HasSuffix(s, "[]") {
			s = strings.TrimSuffix(s, "[]")
			var err error
			p := ArrayParser{}
			if p.subParser, err = MakeParser(s); err != nil {
				return nil, err
			} else {
				return p, nil
			}
		} else {
			structName, typeStr, err := splitStructName(s)
			if err != nil {
				return nil, err
			}
			s = typeStr
			s = s[1 : len(s)-1] //去掉头尾括号
			p := StructParser{structName: structName, fields: map[string]Parser{}}
			for s != "" {
				var name string
				var typeStr string
				if name, typeStr, s, err = readStructField(s); err != nil {
					return nil, err
				} else {
					fieldParser, err := MakeParser(typeStr)
					if err != nil {
						return nil, err
					}
					p.fields[name] = fieldParser
				}
			}
			return p, nil
		}
	}
}
