package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var filterToken []string = []string{
	" ",
	"\r",
	"\n",
	"\t",
}

var filterChar []byte = []byte{
	' ',
	'\r',
	'\n',
	'\t',
}

func isFilterChar(r byte) bool {
	for _, v := range filterChar {
		if r == v {
			return true
		}
	}
	return false
}

func trimLeft(s string) string {
	for _, v := range filterToken {
		s = strings.TrimLeft(s, v)
	}
	return s
}

func trimRight(s string) string {
	for _, v := range filterToken {
		s = strings.TrimRight(s, v)
	}
	return s
}

func trim(s string) string {
	return trimLeft(trimRight(s))
}

// 从字符串生成Value
type Parser interface {
	Parse(string) (*Value, error)
	ValueType() int
	GetGoType() string                    //获取go类型
	GenGoStruct(*strings.Builder, string) //生成go结构体
}

type ValueParser struct {
	valueType int
}

func (p *ValueParser) ValueType() int {
	return p.valueType
}

func (p *ValueParser) Parse(s string) (*Value, error) {
	var err error
	v := &Value{valueType: p.valueType}
	switch p.valueType {
	case typeInt:
		if s == "" {
			v.value = 0
		} else {
			s = trim(s)
			v.value, err = strconv.ParseInt(s, 10, 64)
		}
	case typeBool:
		if s == "" {
			v.value = false
		} else {
			s = trim(s)
			v.value, err = strconv.ParseBool(s)
		}
	case typeFloat:
		if s == "" {
			v.value = 0.0
		} else {
			s = trim(s)
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
	elements Parser
}

func (p *ArrayParser) ValueType() int {
	return typeArray
}

func (p *ArrayParser) splitCompose(s string, bracket string) (ret []string, err error) {
	if s == "" {
		return ret, nil
	} else {
		left := -1
		leftCount := 0
		for i := 0; i < len(s); i++ {
			switch s[i] {
			case bracket[0]:
				leftCount++
				if leftCount == 1 {
					left = i
				}
			case bracket[1]:
				leftCount--
				if leftCount < 0 {
					return nil, fmt.Errorf("ArrayParser.splitCompose left bracket mismatch")
				} else if leftCount == 0 {
					sub := s[left : i+1]
					if sub != bracket {
						ret = append(ret, sub)
					}
				}
			}
		}

		if leftCount != 0 {
			return nil, fmt.Errorf("ArrayParser.splitCompose right bracket mismatch")
		} else {
			return ret, nil
		}
	}
}

func (p *ArrayParser) split(s string) (ret []string, err error) {
	if s[0] != '[' || s[len(s)-1] != ']' {
		return ret, errors.New("ArrayParser.split bracket mismatch")
	}
	s = s[1 : len(s)-1] //去掉头尾括号
	switch p.elements.(type) {
	case *ArrayParser:
		ret, err = p.splitCompose(s, "[]")
	case *StructParser:
		ret, err = p.splitCompose(s, "{}")
	default:
		if p.elements.ValueType() == typeString {
			left := false
			//内嵌的string值必须用""包裹，如果内容包含"需要使用\转义
			var ss strings.Builder
			for i, v := range s {
				if v == '"' {
					if !left {
						left = true
						ss = strings.Builder{}
					} else if s[i-1] != '\\' {
						if i != len(s)-1 {
							i++
							for ; i < len(s); i++ {
								v := s[i]
								if v == ',' {
									break
								} else if !isFilterChar(v) {
									return ret, errors.New("ArrayParser.split error1")
								}
							}
						}
						ret = append(ret, ss.String())
						left = false
					} else {
						ss.WriteRune(v)
					}
				} else if left && !(v == '\\' && i < len(s)-1 && s[i+1] == '"') {
					ss.WriteRune(v)
				}
			}
			if left {
				//没有匹配的右"
				return ret, errors.New("ArrayParser.split error2")
			} else if len(s) > 0 && len(ret) == 0 {
				return ret, errors.New("ArrayParser.split error3")
			}

		} else {
			ret = strings.Split(s, ",")
		}
	}
	return ret, err
}

func (p *ArrayParser) Parse(s string) (*Value, error) {
	array := &Array{}
	s = trim(s)
	if !(s == "" || s == "[]") {
		if e, err := p.split(s); err != nil {
			return nil, err
		} else {
			for _, vv := range e {
				if value, err := p.elements.Parse(vv); err != nil {
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
	goType string
}

func (p *StructParser) ValueType() int {
	return typeStruct
}

func (p *StructParser) readFieldName(s string) (string, string, error) {
	o := -1
	for i := 0; i < len(s); i++ {
		if o == -1 && s[i] >= 'a' && s[i] <= 'z' {
			o = i
		} else if s[i] == ':' {
			if o == -1 {
				return "", "", errors.New("ErrReadFieldName1")
			} else {
				return trim(s[o:i]), trim(s[i+1:]), nil
			}
		}
	}
	return "", "", errors.New("ErrReadFieldName2")
}

func (p *StructParser) readComposeFiledValue(s string, parser Parser, bracket string) (*Value, string, error) {
	s = trim(s)
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
			return v, trim(s[i+1:]), nil
		}
	} else {
		return nil, "", errors.New("bracket mismatch")
	}
}

func (p *StructParser) readFieldValue(s string, parser Parser) (*Value, string, error) {
	switch parser.(type) {
	case *ValueParser:
		i := 0
		var value string
		if parser.ValueType() == typeString {
			left := false
			//内嵌的string值必须用""包裹，如果内容包含"需要使用\转义
			var ss strings.Builder
			for k, v := range s {
				if v == '"' {
					if !left {
						left = true
						ss = strings.Builder{}
					} else if s[k-1] != '\\' {
						if k != len(s)-1 {
							k++
							for ; k < len(s); k++ {
								v := s[k]
								if v == ',' {
									break
								} else if !isFilterChar(v) {
									return nil, "", errors.New("StructParser.readFieldValue error1")
								}
							}
						}
						value = ss.String()
						i = k
						break
					} else {
						ss.WriteRune(v)
					}
				} else if left && !(v == '\\' && i < len(s)-1 && s[k+1] == '"') {
					ss.WriteRune(v)
				}
			}

			if !left {
				return nil, "", errors.New("StructParser.readFieldValue error2")
			} else if i == 0 {
				return nil, "", errors.New("StructParser.readFieldValue error3")
			}

		} else {
			for ; i < len(s); i++ {
				if s[i] == ',' {
					break
				}
			}
			value = s[:i]
		}

		if v, err := parser.Parse(value); err != nil {
			return nil, "", err
		} else if i == len(s) {
			return v, "", nil
		} else {
			return v, s[i+1:], nil
		}

	case *ArrayParser:
		return p.readComposeFiledValue(s, parser, "[]")
	case *StructParser:
		return p.readComposeFiledValue(s, parser, "{}")
	default:
		return nil, "", errors.New("ErrReadFieldValue7")
	}
}

func (p *StructParser) Parse(s string) (*Value, error) {
	s = trim(s)
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
	s = trim(s)
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
	s = trim(s)
	switch s {
	case "int":
		return &ValueParser{valueType: typeInt}, nil
	case "string":
		return &ValueParser{valueType: typeString}, nil
	case "bool":
		return &ValueParser{valueType: typeBool}, nil
	case "float":
		return &ValueParser{valueType: typeFloat}, nil
	default:
		if strings.HasSuffix(s, "[]") {
			s = strings.TrimSuffix(s, "[]")
			p := &ArrayParser{}
			if p.elements, err = MakeParser(s); err != nil {
				return nil, err
			} else {
				return p, nil
			}
		} else if len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}' {
			s = s[1 : len(s)-1] //去掉头尾括号
			p := &StructParser{fields: map[string]Parser{}}
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
