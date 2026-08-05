// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
)

// EventSpec holds the generator metadata extracted from one event data struct.
type EventSpec struct {
	StructName  string
	Channel     string
	Params      map[string]string // param name → description
	Stream      string
	CEType      string
	SendSummary string
	RecvSummary string
	Fields      []FieldSpec
}

// FieldSpec describes one data field extracted from a struct.
type FieldSpec struct {
	JSONName string
	GoType   string // e.g. "string", "*string", "int"
	Required bool   // false when omitempty or pointer type
}

// ParseFile parses the Go source file at path and returns one EventSpec
// per annotated struct. Returns an error if the file cannot be parsed or
// a required tag key is missing.
func ParseFile(path string) ([]EventSpec, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	var specs []EventSpec
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			es, ok, err := extractEventSpec(typeSpec.Name.Name, structType)
			if err != nil {
				return nil, err
			}
			if ok {
				specs = append(specs, es)
			}
		}
	}
	return specs, nil
}

// extractEventSpec extracts an EventSpec from a struct type if it has a
// sentinel blank field with an asyncapi tag. Returns ok=false if the struct
// is not annotated.
func extractEventSpec(name string, st *ast.StructType) (EventSpec, bool, error) {
	var asyncapiTag string
	var dataFields []FieldSpec

	for _, field := range st.Fields.List {
		// Sentinel blank field: unnamed or named "_", type struct{}
		isSentinel := false
		if len(field.Names) == 0 {
			isSentinel = true
		} else if len(field.Names) == 1 && field.Names[0].Name == "_" {
			isSentinel = true
		}

		if isSentinel {
			if field.Tag == nil {
				continue
			}
			raw := strings.Trim(field.Tag.Value, "`")
			tag := reflect.StructTag(raw)
			val := tag.Get("asyncapi")
			if val != "" {
				asyncapiTag = val
			}
			continue
		}

		// Data field
		if field.Tag == nil {
			continue
		}
		raw := strings.Trim(field.Tag.Value, "`")
		tag := reflect.StructTag(raw)
		jsonVal := tag.Get("json")
		if jsonVal == "" || jsonVal == "-" {
			continue
		}
		parts := strings.SplitN(jsonVal, ",", 2)
		jsonName := parts[0]
		omitempty := len(parts) > 1 && strings.Contains(parts[1], "omitempty")

		goType := fieldGoType(field.Type)
		required := !omitempty && !strings.HasPrefix(goType, "*")

		dataFields = append(dataFields, FieldSpec{
			JSONName: jsonName,
			GoType:   goType,
			Required: required,
		})
	}

	if asyncapiTag == "" {
		return EventSpec{}, false, nil
	}

	es, err := parseAsyncAPITag(name, asyncapiTag)
	if err != nil {
		return EventSpec{}, false, err
	}
	es.Fields = dataFields
	return es, true, nil
}

// fieldGoType returns a string representation of the Go type expression.
func fieldGoType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + fieldGoType(t.X)
	case *ast.ArrayType:
		return "[]" + fieldGoType(t.Elt)
	default:
		return "interface{}"
	}
}

// parseAsyncAPITag parses a comma-separated key:value asyncapi tag string.
// Values may contain spaces but not commas. Multiple param entries are
// supported by repeating the param key.
func parseAsyncAPITag(structName, tag string) (EventSpec, error) {
	es := EventSpec{
		StructName: structName,
		Params:     make(map[string]string),
	}

	pairs := strings.Split(tag, ",")
	for _, pair := range pairs {
		idx := strings.IndexByte(pair, ':')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(pair[:idx])
		val := strings.TrimSpace(pair[idx+1:])
		switch key {
		case "channel":
			es.Channel = val
		case "param":
			eqIdx := strings.IndexByte(val, '=')
			if eqIdx < 0 {
				return EventSpec{}, fmt.Errorf("struct %s: param tag %q missing '='", structName, val)
			}
			es.Params[val[:eqIdx]] = val[eqIdx+1:]
		case "stream":
			es.Stream = val
		case "type":
			es.CEType = val
		case "send":
			es.SendSummary = val
		case "receive":
			es.RecvSummary = val
		}
	}

	required := []string{"channel", "stream", "type", "send", "receive"}
	var missing []string
	for _, r := range required {
		switch r {
		case "channel":
			if es.Channel == "" {
				missing = append(missing, r)
			}
		case "stream":
			if es.Stream == "" {
				missing = append(missing, r)
			}
		case "type":
			if es.CEType == "" {
				missing = append(missing, r)
			}
		case "send":
			if es.SendSummary == "" {
				missing = append(missing, r)
			}
		case "receive":
			if es.RecvSummary == "" {
				missing = append(missing, r)
			}
		}
	}
	if len(missing) > 0 {
		return EventSpec{}, fmt.Errorf("struct %s: asyncapi tag missing required keys: %s", structName, strings.Join(missing, ", "))
	}

	return es, nil
}
