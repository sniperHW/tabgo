package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 从字符串生成Value
type Parser interface {
	Parse(string) (*Value, error)
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

func (p ValueParser) Parse(s string) (*Value, error) {
	var err error
	v := &Value{valueType: p.valueType}
	switch p.valueType {
	case typeInt:
		if s == "" {
			v.value = 0
		} else {
			v.value, err = strconv.ParseInt(s, 10, 64)
		}
	case typeBool:
		if s == "" {
			v.value = false
		} else {
			v.value, err = strconv.ParseBool(s)
		}
	case typeFloat:
		if s == "" {
			v.value = 0.0
		} else {
			v.value, err = strconv.ParseFloat(s, 64)
		}
	case typeString:
		v.value = s
	default:
		err = fmt.Errorf("invaild type str:%s", s)
	}

	return v, err
}

type ArrayParser struct {
	child Parser
}

func (p ArrayParser) ValueType() int {
	return typeArray
}

func (p ArrayParser) splitCompose(s string, bracket string) (ret []string, err error) {
	left := -1
	for i := 0; i < len(s); i++ {
		if s[i] == bracket[0] {
			if left != -1 {
				return nil, fmt.Errorf("ArrayParser.splitCompose left bracket mismatch")
			} else {
				left = i
			}
		} else if s[i] == bracket[1] {
			if left == -1 {
				return nil, fmt.Errorf("ArrayParser.splitCompose right bracket mismatch")
			} else {
				sub := s[left : i+1]
				if sub != bracket {
					ret = append(ret, sub)
				}
				left = -1
			}
		}
	}
	return ret, nil
}

func (p ArrayParser) split(s string) (ret []string, err error) {
	if s[0] != '[' || s[len(s)-1] != ']' {
		return ret, errors.New("ArrayParser.split bracket mismatch")
	}
	s = s[1 : len(s)-1] //去掉头尾括号
	switch p.child.(type) {
	case ArrayParser:
		ret, err = p.splitCompose(s, "[]")
	case StructParser:
		ret, err = p.splitCompose(s, "{}")
	default:
		ret = strings.Split(s, ",")
	}
	return ret, err
}

func (p ArrayParser) Parse(s string) (*Value, error) {
	array := &Array{}
	if !(s == "" || s == "[]") {
		if e, err := p.split(s); err != nil {
			return nil, err
		} else {
			for _, vv := range e {
				if value, err := p.child.Parse(vv); err != nil {
					return nil, err
				} else {
					array.value = append(array.value, value)
				}
			}
		}
	}
	return &Value{valueType: typeArray, value: array}, nil
}

type StructParser struct {
	fields map[string]Parser
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
				return "", "", errors.New("ErrReadFieldName1")
			} else {
				return s[o:i], s[i+1:], nil
			}
		}
	}
	return "", "", errors.New("ErrReadFieldName2")
}

func (p StructParser) readComposeFiledValue(s string, parser Parser, bracket string) (*Value, string, error) {
	if s[0] != bracket[0] {
		return nil, "", errors.New("StructParser.readComposeFiledValue left bracket mismatch")
	}
	i := 1
	leftCount := 1
	for ; i < len(s); i++ {
		if s[i] == bracket[0] {
			leftCount++
		} else if s[i] == bracket[1] {
			leftCount--
		} else if s[i] == ',' && leftCount == 0 {
			break
		}
	}

	if leftCount == 0 {
		if v, err := parser.Parse(s[:i]); err != nil {
			return nil, "", err
		} else if i == len(s) {
			return v, "", nil
		} else {
			return v, s[i+1:], nil
		}
	} else {
		return nil, "", errors.New("bracket mismatch")
	}
}

func (p StructParser) readFieldValue(s string, parser Parser) (*Value, string, error) {
	switch parser.(type) {
	case ValueParser:
		i := 0
		for ; i < len(s); i++ {
			if s[i] == ',' {
				break
			}
		}

		if v, err := parser.Parse(s[:i]); err != nil {
			return nil, "", err
		} else if i == len(s) {
			return v, "", nil
		} else {
			return v, s[i+1:], nil
		}
	case ArrayParser:
		return p.readComposeFiledValue(s, parser, "[]")
	case StructParser:
		return p.readComposeFiledValue(s, parser, "{}")
	default:
		return nil, "", errors.New("ErrReadFieldValue7")
	}
}

func (p StructParser) Parse(s string) (*Value, error) {
	v := &Value{valueType: typeStruct}
	st := &Struct{}
	if !(s == "" || s == "{}") {
		if s[0] != '{' || s[len(s)-1] != '}' {
			return nil, fmt.Errorf("invaild struct")
		}
		s = s[1 : len(s)-1] //去掉头尾括号
		for s != "" {
			var name string
			var err error
			var field *Value

			name, s, err = p.readFieldName(s)
			if err != nil {
				return nil, err
			}

			if fieldParser, ok := p.fields[name]; !ok {
				return nil, fmt.Errorf("invaild field:%s", name)
			} else {

				field, s, err = p.readFieldValue(s, fieldParser)
				if err != nil {
					return nil, err
				}
				st.fields = append(st.fields, &Field{
					name:  name,
					value: field,
				})
			}
		}
	}
	v.value = st
	return v, nil
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

func MakeParser(s string) (Parser, error) {
	var name string
	var typeStr string
	var err error
	switch s {
	case "int":
		return ValueParser{valueType: typeInt}, nil
	case "string":
		return ValueParser{valueType: typeString}, nil
	case "bool":
		return ValueParser{valueType: typeBool}, nil
	case "float":
		return ValueParser{valueType: typeFloat}, nil
	default:
		if strings.HasSuffix(s, "[]") {
			s = strings.TrimSuffix(s, "[]")
			p := ArrayParser{}
			if p.child, err = MakeParser(s); err != nil {
				return nil, err
			} else {
				return p, nil
			}
		} else if len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}' {
			s = s[1 : len(s)-1] //去掉头尾括号
			p := StructParser{fields: map[string]Parser{}}
			for s != "" {
				if name, typeStr, s, err = readStructField(s); err != nil {
					return nil, err
				} else if fieldParser, err := MakeParser(typeStr); err == nil {
					p.fields[name] = fieldParser
				} else {
					return nil, err
				}
			}
			return p, nil
		} else {
			return nil, fmt.Errorf("invaild type define:%s", s)
		}
	}
}
