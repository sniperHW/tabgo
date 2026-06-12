package parser

import (
	"strings"
	"testing"
)

// BenchmarkEscapeStringOld 使用 += 操作的性能
func BenchmarkEscapeStringOld(b *testing.B) {
	s := "Test string with \"quotes\" and \\backslashes\\ and\nnewlines\t"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := ""
		for _, r := range s {
			switch r {
			case '"':
				result += "\\\""
			case '\\':
				result += "\\\\"
			case '\n':
				result += "\\n"
			case '\r':
				result += "\\r"
			case '\t':
				result += "\\t"
			default:
				result += string(r)
			}
		}
		_ = result
	}
}

// BenchmarkEscapeStringNew 使用 strings.Builder 的性能
func BenchmarkEscapeStringNew(b *testing.B) {
	s := "Test string with \"quotes\" and \\backslashes\\ and\nnewlines\t"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result strings.Builder
		result.Grow(len(s) * 2)
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
		_ = result.String()
	}
}

// BenchmarkToLuaValue 复杂结构的 Lua 转换性能测试
func BenchmarkToLuaValue(b *testing.B) {
	// 创建一个复杂的嵌套结构
	v := &Value{
		Type: "struct",
		Value: map[string]*Value{
			"id":    {Type: "int", Value: 123},
			"name":  {Type: "string", Value: "test\nstring\"with\tescapes"},
			"array": {Type: "array", Value: []*Value{
				{Type: "int", Value: 1},
				{Type: "int", Value: 2},
				{Type: "int", Value: 3},
			}},
			"nested": {Type: "struct", Value: map[string]*Value{
				"x": {Type: "int", Value: 10},
				"y": {Type: "int", Value: 20},
			}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ToLuaStr()
	}
}
