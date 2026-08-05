// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteYAML_CreatesFile(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "nats://localhost:4222")
	out := filepath.Join(t.TempDir(), "asyncapi.yaml")

	if err := WriteYAML(doc, out); err != nil {
		t.Fatalf("WriteYAML: %v", err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(b)

	if !strings.Contains(content, "asyncapi: 3.0.0") {
		t.Error("output missing 'asyncapi: 3.0.0'")
	}
	if !strings.Contains(content, "application/cloudevents+json") {
		t.Error("output missing defaultContentType")
	}
	if !strings.Contains(content, "core.widget.created.{ownerId}") {
		t.Error("output missing channel address")
	}
	if !strings.Contains(content, "WIDGETS") {
		t.Error("output missing NATS stream name")
	}
	if !strings.Contains(content, "specversion") {
		t.Error("output missing CloudEvents envelope field")
	}
}

func TestWriteYAML_SpdxHeader(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "nats://localhost:4222")
	out := filepath.Join(t.TempDir(), "asyncapi.yaml")

	if err := WriteYAML(doc, out); err != nil {
		t.Fatalf("WriteYAML: %v", err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.HasPrefix(string(b), "# SPDX-License-Identifier: Apache-2.0") {
		t.Error("output missing SPDX header")
	}
}
