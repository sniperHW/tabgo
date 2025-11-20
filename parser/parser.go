package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// TypeDef 表示类型定义
type TypeDef struct {
	Type       string              // 基本类型：int, bool, float, string, array, struct
	ElemType   *TypeDef            // 数组元素类型（仅当 Type == "array" 时使用）
	Fields     map[string]*TypeDef // 结构体字段（仅当 Type == "struct" 时使用）
	FieldOrder []string            // 结构体字段的顺序（仅当 Type == "struct" 时使用）
}

// Parser 解析器
type Parser struct {
	typeDef *TypeDef
}

// MakeParser 根据类型定义字符串创建解析器
func MakeParser(typeStr string) (*Parser, error) {
	typeDef, err := parseTypeDef(typeStr)
	if err != nil {
		return nil, err
	}
	return &Parser{typeDef: typeDef}, nil
}

// Parse 解析输入字符串并返回 Value
func (p *Parser) Parse(input string) (*Value, error) {
	// 过滤输入字符串中的空白字符（\t、\r、\n），但保留字符串值内部的转义字符
	input = filterWhitespace(input)
	input = strings.TrimSpace(input)
	return p.parseValue(input, p.typeDef)
}

// filterWhitespace 过滤输入字符串中的空白字符（\t、\r、\n）
// 但保留字符串值内部的转义字符（引号内的内容不受影响）
func filterWhitespace(input string) string {
	var result strings.Builder
	inString := false
	escape := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		if escape {
			// 转义字符后的字符，直接保留
			result.WriteByte(char)
			escape = false
			continue
		}

		if char == '\\' && inString {
			// 在字符串内的转义字符，保留
			result.WriteByte(char)
			escape = true
			continue
		}

		if char == '"' {
			// 切换字符串状态
			inString = !inString
			result.WriteByte(char)
			continue
		}

		if inString {
			// 字符串内的所有字符都保留
			result.WriteByte(char)
		} else {
			// 字符串外的空白字符过滤掉
			if char != '\t' && char != '\r' && char != '\n' {
				result.WriteByte(char)
			}
		}
	}

	return result.String()
}

// parseTypeDef 解析类型定义字符串
func parseTypeDef(typeStr string) (*TypeDef, error) {
	typeStr = strings.TrimSpace(typeStr)

	// 检查是否是数组类型
	if strings.HasSuffix(typeStr, "[]") {
		elemTypeStr := typeStr[:len(typeStr)-2]
		elemType, err := parseTypeDef(elemTypeStr)
		if err != nil {
			return nil, err
		}
		return &TypeDef{
			Type:     "array",
			ElemType: elemType,
		}, nil
	}

	// 检查是否是结构体类型
	if strings.HasPrefix(typeStr, "{") && strings.HasSuffix(typeStr, "}") {
		content := typeStr[1 : len(typeStr)-1]
		fields := make(map[string]*TypeDef)
		fieldOrder := make([]string, 0)

		// 解析字段
		fieldStrs := splitStructFields(content)
		for _, fieldStr := range fieldStrs {
			fieldStr = strings.TrimSpace(fieldStr)
			parts := strings.SplitN(fieldStr, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid struct field: %s", fieldStr)
			}

			fieldName := strings.TrimSpace(parts[0])
			fieldTypeStr := strings.TrimSpace(parts[1])

			fieldType, err := parseTypeDef(fieldTypeStr)
			if err != nil {
				return nil, err
			}

			fields[fieldName] = fieldType
			fieldOrder = append(fieldOrder, fieldName)
		}

		return &TypeDef{
			Type:       "struct",
			Fields:     fields,
			FieldOrder: fieldOrder,
		}, nil
	}

	// 基本类型
	switch typeStr {
	case "int", "bool", "float", "string":
		return &TypeDef{Type: typeStr}, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", typeStr)
	}
}

// splitStructFields 分割结构体字段字符串
func splitStructFields(content string) []string {
	var fields []string
	var current strings.Builder
	depth := 0
	inString := false
	escape := false

	for i, r := range content {
		if escape {
			current.WriteRune(r)
			escape = false
			continue
		}

		if r == '\\' && inString {
			escape = true
			current.WriteRune(r)
			continue
		}

		if r == '"' {
			inString = !inString
			current.WriteRune(r)
			continue
		}

		if inString {
			current.WriteRune(r)
			continue
		}

		switch r {
		case '{', '[':
			depth++
			current.WriteRune(r)
		case '}', ']':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				fields = append(fields, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}

		// 处理最后一个字段
		if i == len(content)-1 && current.Len() > 0 {
			fields = append(fields, current.String())
		}
	}

	return fields
}

// parseValue 根据类型定义解析值
func (p *Parser) parseValue(input string, typeDef *TypeDef) (*Value, error) {
	switch typeDef.Type {
	case "int":
		return parseInt(input)
	case "bool":
		return parseBool(input)
	case "float":
		return parseFloat(input)
	case "string":
		return parseString(input)
	case "array":
		return p.parseArray(input, typeDef.ElemType)
	case "struct":
		return p.parseStruct(input, typeDef.Fields)
	default:
		return nil, fmt.Errorf("unknown type: %s", typeDef.Type)
	}
}

// parseInt 解析整数
func parseInt(input string) (*Value, error) {
	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid int: %s", input)
	}
	return &Value{Type: "int", Value: int(val)}, nil
}

// parseBool 解析布尔值
func parseBool(input string) (*Value, error) {
	val, err := strconv.ParseBool(input)
	if err != nil {
		return nil, fmt.Errorf("invalid bool: %s", input)
	}
	return &Value{Type: "bool", Value: val}, nil
}

// parseFloat 解析浮点数
func parseFloat(input string) (*Value, error) {
	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid float: %s", input)
	}
	return &Value{Type: "float", Value: val}, nil
}

// parseString 解析字符串（支持转义字符）
// 对于非嵌套的 string，可以不用引号包裹；对于嵌套的 string（数组、结构体中），必须用引号包裹
func parseString(input string) (*Value, error) {
	input = strings.TrimSpace(input)

	// 如果以引号开头和结尾，按带引号的方式处理（支持转义字符）
	if strings.HasPrefix(input, `"`) && strings.HasSuffix(input, `"`) {
		// 移除首尾引号
		content := input[1 : len(input)-1]

		// 处理转义字符
		result := strings.Builder{}
		i := 0
		for i < len(content) {
			if content[i] == '\\' && i+1 < len(content) {
				switch content[i+1] {
				case 'n':
					result.WriteByte('\n')
					i += 2
				case 'r':
					result.WriteByte('\r')
					i += 2
				case 't':
					result.WriteByte('\t')
					i += 2
				case '\\':
					result.WriteByte('\\')
					i += 2
				case '"':
					result.WriteByte('"')
					i += 2
				default:
					result.WriteByte(content[i])
					i++
				}
			} else {
				result.WriteByte(content[i])
				i++
			}
		}

		return &Value{Type: "string", Value: result.String()}, nil
	}

	// 如果不带引号，直接作为字符串内容（非嵌套情况）
	return &Value{Type: "string", Value: input}, nil
}

// parseArray 解析数组
func (p *Parser) parseArray(input string, elemType *TypeDef) (*Value, error) {
	input = strings.TrimSpace(input)
	// 如果输入为空字符串，按空数组处理
	if input == "" {
		return &Value{Type: "array", Value: []*Value{}}, nil
	}

	if !strings.HasPrefix(input, "[") || !strings.HasSuffix(input, "]") {
		return nil, fmt.Errorf("array must be wrapped in brackets: %s", input)
	}

	content := strings.TrimSpace(input[1 : len(input)-1])
	if content == "" {
		return &Value{Type: "array", Value: []*Value{}}, nil
	}

	// 分割数组元素
	elements := splitArrayElements(content, elemType)

	values := make([]*Value, 0, len(elements))
	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem == "" {
			continue
		}

		val, err := p.parseValue(elem, elemType)
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	return &Value{Type: "array", Value: values}, nil
}

// splitArrayElements 分割数组元素
func splitArrayElements(content string, elemType *TypeDef) []string {
	var elements []string
	var current strings.Builder
	depth := 0
	inString := false
	escape := false

	for _, r := range content {
		if escape {
			current.WriteRune(r)
			escape = false
			continue
		}

		if r == '\\' && inString {
			escape = true
			current.WriteRune(r)
			continue
		}

		if r == '"' {
			inString = !inString
			current.WriteRune(r)
			continue
		}

		if inString {
			current.WriteRune(r)
			continue
		}

		switch r {
		case '{':
			depth++
			current.WriteRune(r)
		case '}':
			depth--
			current.WriteRune(r)
		case '[':
			depth++
			current.WriteRune(r)
		case ']':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				elements = append(elements, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	// 处理最后一个元素
	if current.Len() > 0 {
		elements = append(elements, current.String())
	}

	return elements
}

// parseStruct 解析结构体
func (p *Parser) parseStruct(input string, fields map[string]*TypeDef) (*Value, error) {
	input = strings.TrimSpace(input)
	// 如果输入为空字符串，按空结构体处理
	if input == "" {
		return &Value{Type: "struct", Value: make(map[string]*Value)}, nil
	}

	if !strings.HasPrefix(input, "{") || !strings.HasSuffix(input, "}") {
		return nil, fmt.Errorf("struct must be wrapped in braces: %s", input)
	}

	content := strings.TrimSpace(input[1 : len(input)-1])
	if content == "" {
		return &Value{Type: "struct", Value: make(map[string]*Value)}, nil
	}

	// 分割结构体字段
	fieldStrs := splitStructFields(content)

	values := make(map[string]*Value)
	for _, fieldStr := range fieldStrs {
		fieldStr = strings.TrimSpace(fieldStr)
		if fieldStr == "" {
			continue
		}

		// 查找冒号分隔符（需要考虑嵌套结构）
		colonIdx := findFieldColon(fieldStr)
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid struct field format: %s", fieldStr)
		}

		fieldName := strings.TrimSpace(fieldStr[:colonIdx])
		fieldValueStr := strings.TrimSpace(fieldStr[colonIdx+1:])

		fieldType, ok := fields[fieldName]
		if !ok {
			return nil, fmt.Errorf("unknown field: %s", fieldName)
		}

		val, err := p.parseValue(fieldValueStr, fieldType)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s: %v", fieldName, err)
		}

		values[fieldName] = val
	}

	return &Value{Type: "struct", Value: values}, nil
}

// findFieldColon 查找字段名和值之间的冒号位置
func findFieldColon(fieldStr string) int {
	depth := 0
	inString := false
	escape := false

	for i, r := range fieldStr {
		if escape {
			escape = false
			continue
		}

		if r == '\\' && inString {
			escape = true
			continue
		}

		if r == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch r {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		case ':':
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// GenGoDefine 生成 Go 语言类型定义
func (p *Parser) GenGoDefine(name string) string {
	var result strings.Builder
	structNameMap := make(map[*TypeDef]string) // 记录每个 TypeDef 对应的 Go 结构体名称

	// 先建立结构体名称映射
	p.buildStructNameMap(p.typeDef, name, structNameMap)

	// 生成嵌套结构体定义（按依赖顺序）
	visited := make(map[*TypeDef]bool)
	p.genNestedStructs(p.typeDef, name, &result, structNameMap, visited, true)

	// 生成主类型定义
	if p.typeDef.Type == "struct" {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		structName := structNameMap[p.typeDef]
		result.WriteString(p.genStructDefinition(p.typeDef, structName, structNameMap, true))
	} else {
		// 基本类型或数组类型
		goFieldName := toGoFieldName(name)
		goType := p.typeDefToGoType(p.typeDef, structNameMap, false)
		result.WriteString(goFieldName)
		result.WriteString(" ")
		result.WriteString(goType)
	}

	return result.String()
}

// buildStructNameMap 建立 TypeDef 到 Go 结构体名称的映射
func (p *Parser) buildStructNameMap(typeDef *TypeDef, prefix string, nameMap map[*TypeDef]string) {
	if typeDef.Type == "struct" {
		structName := toGoStructName(prefix)
		nameMap[typeDef] = structName

		// 递归处理字段中的嵌套结构体
		for fieldName, fieldType := range typeDef.Fields {
			if fieldType.Type == "struct" {
				nestedPrefix := prefix + toGoFieldName(fieldName)
				p.buildStructNameMap(fieldType, nestedPrefix, nameMap)
			} else if fieldType.Type == "array" && fieldType.ElemType != nil && fieldType.ElemType.Type == "struct" {
				nestedPrefix := prefix + toGoFieldName(fieldName)
				p.buildStructNameMap(fieldType.ElemType, nestedPrefix, nameMap)
			}
		}
	} else if typeDef.Type == "array" && typeDef.ElemType != nil {
		p.buildStructNameMap(typeDef.ElemType, prefix, nameMap)
	}
}

// genNestedStructs 递归生成所有嵌套结构体定义
func (p *Parser) genNestedStructs(typeDef *TypeDef, prefix string, result *strings.Builder, nameMap map[*TypeDef]string, visited map[*TypeDef]bool, isMain bool) {
	if typeDef.Type == "struct" {
		// 检查是否已经生成过
		if visited[typeDef] {
			return
		}

		// 先递归处理字段中的嵌套结构体
		for fieldName, fieldType := range typeDef.Fields {
			if fieldType.Type == "struct" {
				nestedPrefix := prefix + toGoFieldName(fieldName)
				p.genNestedStructs(fieldType, nestedPrefix, result, nameMap, visited, false)
			} else if fieldType.Type == "array" && fieldType.ElemType != nil && fieldType.ElemType.Type == "struct" {
				nestedPrefix := prefix + toGoFieldName(fieldName)
				p.genNestedStructs(fieldType.ElemType, nestedPrefix, result, nameMap, visited, false)
			}
		}

		// 生成当前结构体定义（跳过主结构体，它会在最后生成）
		if !isMain {
			structName := nameMap[typeDef]
			result.WriteString(p.genStructDefinition(typeDef, structName, nameMap, true))
			result.WriteString("\n\n")
		}
		visited[typeDef] = true
	} else if typeDef.Type == "array" && typeDef.ElemType != nil {
		p.genNestedStructs(typeDef.ElemType, prefix, result, nameMap, visited, false)
	}
}

// genStructDefinition 生成结构体定义
func (p *Parser) genStructDefinition(typeDef *TypeDef, structName string, nameMap map[*TypeDef]string, isTopLevel bool) string {
	if typeDef.Type != "struct" {
		return ""
	}

	var result strings.Builder
	result.WriteString("type ")
	result.WriteString(structName)
	result.WriteString(" struct{\n")

	// 使用保存的字段顺序
	fieldNames := typeDef.FieldOrder
	if len(fieldNames) == 0 {
		// 如果没有保存顺序，则从 map 中获取（向后兼容）
		fieldNames = make([]string, 0, len(typeDef.Fields))
		for fieldName := range typeDef.Fields {
			fieldNames = append(fieldNames, fieldName)
		}
	}

	for _, fieldName := range fieldNames {
		fieldType := typeDef.Fields[fieldName]
		goFieldName := toGoFieldName(fieldName)
		goType := p.typeDefToGoType(fieldType, nameMap, isTopLevel)

		result.WriteString("    ")
		result.WriteString(goFieldName)
		result.WriteString(" ")
		result.WriteString(goType)
		result.WriteString(" `json:\"")
		result.WriteString(fieldName)
		result.WriteString("\"`\n")
	}

	result.WriteString("}")
	return result.String()
}

// typeDefToGoType 将类型定义转换为 Go 类型字符串
func (p *Parser) typeDefToGoType(typeDef *TypeDef, nameMap map[*TypeDef]string, isTopLevel bool) string {
	switch typeDef.Type {
	case "int":
		return "int64"
	case "bool":
		return "bool"
	case "float":
		return "float64"
	case "string":
		return "string"
	case "array":
		if typeDef.ElemType != nil {
			elemType := p.typeDefToGoType(typeDef.ElemType, nameMap, isTopLevel)
			return "[]" + elemType
		}
		return "[]interface{}"
	case "struct":
		// 使用结构体名称（无论是主结构体还是嵌套结构体定义）
		if structName, ok := nameMap[typeDef]; ok {
			return structName
		}
		return "interface{}"
	default:
		return "interface{}"
	}
}

// toGoFieldName 将字段名转换为 Go 命名规范（首字母大写）
func toGoFieldName(name string) string {
	if name == "" {
		return ""
	}
	// 将首字母转为大写
	first := strings.ToUpper(string(name[0]))
	if len(name) == 1 {
		return first
	}
	return first + name[1:]
}

// toGoStructName 将名称转换为 Go 结构体名称（首字母大写）
func toGoStructName(name string) string {
	return toGoFieldName(name)
}
