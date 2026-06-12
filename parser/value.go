package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Value 表示解析后的值
type Value struct {
	Type  string      // 类型：int, bool, float, string, array, struct
	Value interface{} // 实际值
}

// ToJsonStr 将 Value 转换为 JSON 字符串
func (v *Value) ToJsonStr() string {
	data, err := json.Marshal(v.toJSONValue())
	if err != nil {
		return "null"
	}
	return string(data)
}

// ToLuaStr 将 Value 转换为 Lua 格式字符串
func (v *Value) ToLuaStr() string {
	return v.toLuaValue(0)
}

// toJSONValue 将 Value 转换为 JSON 兼容的值
func (v *Value) toJSONValue() interface{} {
	switch v.Type {
	case "array":
		arr := v.Value.([]*Value)
		result := make([]interface{}, len(arr))
		for i, item := range arr {
			result[i] = item.toJSONValue()
		}
		return result
	case "struct":
		fields := v.Value.(map[string]*Value)
		result := make(map[string]interface{})
		for k, val := range fields {
			result[k] = val.toJSONValue()
		}
		return result
	default:
		return v.Value
	}
}

// toLuaValue 将 Value 转换为 Lua 格式字符串（单行格式，无换行）
func (v *Value) toLuaValue(indent int) string {
	switch v.Type {
	case "int", "float", "bool":
		return fmt.Sprintf("%v", v.Value)
	case "string":
		return fmt.Sprintf(`"%s"`, escapeString(v.Value.(string)))
	case "array":
		arr := v.Value.([]*Value)
		if len(arr) == 0 {
			return "{}"
		}
		var result strings.Builder
		result.Grow(len(arr) * 10) // 预分配空间，减少扩容
		result.WriteString("{")
		for i, item := range arr {
			if i > 0 {
				result.WriteString(",")
			}
			result.WriteString(item.toLuaValue(0))
		}
		result.WriteString("}")
		return result.String()
	case "struct":
		fields := v.Value.(map[string]*Value)
		if len(fields) == 0 {
			return "{}"
		}
		var result strings.Builder
		result.Grow(len(fields) * 20) // 预分配空间
		result.WriteString("{")
		first := true
		for k, val := range fields {
			if !first {
				result.WriteString(",")
			}
			first = false
			result.WriteString(k)
			result.WriteString(" = ")
			result.WriteString(val.toLuaValue(0))
		}
		result.WriteString("}")
		return result.String()
	default:
		return "nil"
	}
}

// escapeString 转义字符串中的特殊字符
func escapeString(s string) string {
	var result strings.Builder
	result.Grow(len(s) * 2) // 预分配空间，最坏情况下每个字符都需要转义
	for _, r := range s {
		switch r {
		case '"':
			result.WriteString("\\\"")
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}
