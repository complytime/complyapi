// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"testing"
)

func TestParseFile_ReturnsEventSpec(t *testing.T) {
	specs, err := ParseFile("testdata/fixture.go")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("len(specs) = %d, want 1", len(specs))
	}

	s := specs[0]

	if s.StructName != "WidgetCreatedData" {
		t.Errorf("StructName = %q, want %q", s.StructName, "WidgetCreatedData")
	}
	if s.Channel != "core.widget.created.{ownerId}" {
		t.Errorf("Channel = %q, want %q", s.Channel, "core.widget.created.{ownerId}")
	}
	if s.Params["ownerId"] != "The widget owner identifier" {
		t.Errorf("Params[ownerId] = %q, want %q", s.Params["ownerId"], "The widget owner identifier")
	}
	if s.Stream != "WIDGETS" {
		t.Errorf("Stream = %q, want %q", s.Stream, "WIDGETS")
	}
	if s.CEType != "dev.example.widget.created" {
		t.Errorf("CEType = %q, want %q", s.CEType, "dev.example.widget.created")
	}
	if s.SendSummary != "Published when a widget is created" {
		t.Errorf("SendSummary = %q, want %q", s.SendSummary, "Published when a widget is created")
	}
	if s.RecvSummary != "Consume widget-created events" {
		t.Errorf("RecvSummary = %q, want %q", s.RecvSummary, "Consume widget-created events")
	}
}

func TestParseFile_ReturnsFields(t *testing.T) {
	specs, err := ParseFile("testdata/fixture.go")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	fields := specs[0].Fields

	// sentinel blank field must be excluded
	for _, f := range fields {
		if f.JSONName == "" || f.JSONName == "_" {
			t.Errorf("sentinel field leaked into Fields: %+v", f)
		}
	}

	// widgetId — required (no omitempty, not pointer)
	widgetID := findField(fields, "widgetId")
	if widgetID == nil {
		t.Fatal("field widgetId not found")
	}
	if !widgetID.Required {
		t.Error("widgetId should be required")
	}
	if widgetID.GoType != "string" {
		t.Errorf("widgetId GoType = %q, want %q", widgetID.GoType, "string")
	}

	// tag — optional (omitempty)
	tag := findField(fields, "tag")
	if tag == nil {
		t.Fatal("field tag not found")
	}
	if tag.Required {
		t.Error("tag should not be required (omitempty)")
	}

	// parentId — optional (pointer)
	parentID := findField(fields, "parentId")
	if parentID == nil {
		t.Fatal("field parentId not found")
	}
	if parentID.Required {
		t.Error("parentId should not be required (pointer)")
	}
	if parentID.GoType != "*string" {
		t.Errorf("parentId GoType = %q, want %q", parentID.GoType, "*string")
	}
}

func TestParseFile_MissingRequiredTag_ReturnsError(t *testing.T) {
	// Write a temp file with a missing required tag key
	content := `package testdata
type BadData struct {
	_ struct{} ` + "`" + `asyncapi:"channel:core.bad.{id}"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
}`
	path := t.TempDir() + "/bad.go"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := ParseFile(path)
	if err == nil {
		t.Error("expected error for missing required tag keys")
	}
}

func findField(fields []FieldSpec, jsonName string) *FieldSpec {
	for i := range fields {
		if fields[i].JSONName == jsonName {
			return &fields[i]
		}
	}
	return nil
}
