# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**tabgo** is a game configuration table tool (游戏打表工具) that reads Excel (.xlsx) files and converts them into JSON, Lua, or Go code. Each Excel sheet defines a data table with column names (row 0), types (row 1), and data rows (starting from row 3). The first column must be named `id` and serves as the primary key.

## Build & Run

```bash
# Build
go build -o tabgo .

# Run examples (from project root)
./tabgo -input ./excel -output ./lua -mode lua
./tabgo -input ./excel -output ./json -mode json
./tabgo -input ./excel -output ./test -mode go -package test

# Run benchmarks
go test ./parser/ -bench=. -benchmem

# Run the example Go reader
cd example && go run gojson.go
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-input` | `./excel` | Path to directory containing .xlsx files |
| `-output` | `./lua` | Output directory for generated files |
| `-mode` | `json` | Output format: `lua`, `json`, or `go` |
| `-package` | `json` | Go package name (only for `-mode go`) |
| `-server` | `false` | Set to `true` to exclude columns tagged as `client` |

## Architecture

**Single-package CLI tool** (`package main`) with one sub-package (`parser`).

### Core pipeline (`tab.go`)

`Walker.walk()` reads all `.xlsx` files from the input directory concurrently (goroutines + `sync.WaitGroup`). For each file:
1. Reads column names (row 0), type definitions (row 1), data rows (row 3+)
2. Builds `Column` objects with name + `Parser` for each non-ignored column
3. Calls mode-specific callbacks: `funcTableProcessed` (Go mode only) then `funcOutput` (all modes)

Column names support tags: `name:tag` syntax. Columns with tags in the ignore set (e.g., `annotation`, `client` when `-server=true`) are skipped.

### Type system (`parser/`)

- `TypeDef` — recursive type definition: `int`, `bool`, `float`, `string`, `array` (with `ElemType`), `struct` (with `Fields` map + `FieldOrder`)
- `Parser` — created via `MakeParser(typeStr)`, parses cell string values into `Value` trees
- `Value` — parsed result with `Type` and `Value` fields; has `ToJsonStr()` and `ToLuaStr()` methods for output

Type definition syntax: `{field:type,field:type}` for structs, `[]type` for arrays, multidimensional via `[][]type`. The parser handles nested arrays, nested structs, and string escaping (quoted strings in arrays/structs, unquoted at top level).

### Output modes

- **`json.go`** (`outputJson`): Writes `{id: {field: value, ...}}` JSON files per table
- **`lua.go`** (`outputLua`): Writes `local TableName = {[id]={...}}` Lua files per table
- **`go.go`** (`goStruct`): Generates a `.go` file per table with struct definitions, `atomic.Value`-based map storage, `Load*`/`Get*`/`ForEach*` functions. Generated files are auto-formatted with `gofmt`. Column names use `name:tag` format where the part before `:` is the field name.
