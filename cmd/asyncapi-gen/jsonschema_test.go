// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildDataJSONSchema(t *testing.T) {
	schema := BuildDataJSONSchema(singleSpec())

	if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("$schema = %q, want Draft 2020-12 URI", schema["$schema"])
	}
	if schema["$id"] != "WidgetCreatedData.schema.json" {
		t.Errorf("$id = %q, want %q", schema["$id"], "WidgetCreatedData.schema.json")
	}
	if schema["type"] != "object" {
		t.Errorf("type = %q, want %q", schema["type"], "object")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required is not []string")
	}
	requiredSet := map[string]bool{}
	for _, r := range required {
		requiredSet[r] = true
	}
	if !requiredSet["widgetId"] {
		t.Error("widgetId should be required")
	}
	if !requiredSet["name"] {
		t.Error("name should be required")
	}
	if requiredSet["tag"] {
		t.Error("tag should not be required")
	}

	props, ok := schema["properties"].(JSONSchema)
	if !ok {
		t.Fatal("properties is not JSONSchema")
	}
	widgetID, ok := props["widgetId"].(JSONSchema)
	if !ok {
		t.Fatal("widgetId property not found")
	}
	if widgetID["type"] != "string" {
		t.Errorf("widgetId type = %q, want %q", widgetID["type"], "string")
	}
	if widgetID["description"] != "Unique widget identifier" {
		t.Errorf("widgetId description = %q, want %q", widgetID["description"], "Unique widget identifier")
	}

	// Data schema should have description from doc comment
	if schema["description"] != "WidgetCreatedData is the payload for widget.created events." {
		t.Errorf("data schema description = %q, want %q", schema["description"], "WidgetCreatedData is the payload for widget.created events.")
	}

	// Field without description should have no description key
	nameProp, ok := props["name"].(JSONSchema)
	if !ok {
		t.Fatal("name property not found")
	}
	if _, hasDesc := nameProp["description"]; hasDesc {
		t.Error("name should not have description key")
	}
}

func TestBuildEnvelopeJSONSchema(t *testing.T) {
	schema := BuildEnvelopeJSONSchema(singleSpec())

	if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("$schema = %q, want Draft 2020-12 URI", schema["$schema"])
	}
	if schema["$id"] != "WidgetCreatedCloudEvent.schema.json" {
		t.Errorf("$id = %q, want %q", schema["$id"], "WidgetCreatedCloudEvent.schema.json")
	}

	props, ok := schema["properties"].(JSONSchema)
	if !ok {
		t.Fatal("properties is not JSONSchema")
	}

	specversion, ok := props["specversion"].(JSONSchema)
	if !ok {
		t.Fatal("specversion property not found")
	}
	if specversion["const"] != "1.0" {
		t.Errorf("specversion const = %q, want %q", specversion["const"], "1.0")
	}

	ceType, ok := props["type"].(JSONSchema)
	if !ok {
		t.Fatal("type property not found")
	}
	if ceType["const"] != "dev.example.widget.created" {
		t.Errorf("type const = %q, want %q", ceType["const"], "dev.example.widget.created")
	}

	data, ok := props["data"].(JSONSchema)
	if !ok {
		t.Fatal("data property not found")
	}
	if data["$ref"] != "WidgetCreatedData.schema.json" {
		t.Errorf("data $ref = %q, want %q", data["$ref"], "WidgetCreatedData.schema.json")
	}
}

func TestWriteJSONSchemas(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "schemas")

	err := WriteJSONSchemas([]EventSpec{singleSpec()}, dir)
	if err != nil {
		t.Fatalf("WriteJSONSchemas: %v", err)
	}

	// Verify data schema file
	dataPath := filepath.Join(dir, "WidgetCreatedData.schema.json")
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("reading data schema: %v", err)
	}
	var dataSchema map[string]any
	if err := json.Unmarshal(dataBytes, &dataSchema); err != nil {
		t.Fatalf("parsing data schema JSON: %v", err)
	}
	if dataSchema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Error("data schema missing $schema")
	}

	// Verify envelope schema file
	envPath := filepath.Join(dir, "WidgetCreatedCloudEvent.schema.json")
	envBytes, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("reading envelope schema: %v", err)
	}
	var envSchema map[string]any
	if err := json.Unmarshal(envBytes, &envSchema); err != nil {
		t.Fatalf("parsing envelope schema JSON: %v", err)
	}
	if envSchema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Error("envelope schema missing $schema")
	}
}
