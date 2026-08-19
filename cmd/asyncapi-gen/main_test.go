// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRun_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{"missing input", Options{Output: "out.yaml", Title: "T", Version: "1.0", Server: "nats://x"}},
		{"missing output", Options{Input: "in.go", Title: "T", Version: "1.0", Server: "nats://x"}},
		{"missing title", Options{Input: "in.go", Output: "out.yaml", Version: "1.0", Server: "nats://x"}},
		{"missing version", Options{Input: "in.go", Output: "out.yaml", Title: "T", Server: "nats://x"}},
		{"missing server", Options{Input: "in.go", Output: "out.yaml", Title: "T", Version: "1.0"}},
		{"all empty", Options{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tt.opts, &stdout, &stderr)
			if err == nil {
				t.Fatal("expected error for missing required flags")
			}
			if !strings.Contains(err.Error(), "required flags missing") {
				t.Errorf("error = %q, want mention of required flags", err)
			}
		})
	}
}

func TestRun_ParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	opts := Options{
		Input:   "nonexistent_file.go",
		Output:  "out.yaml",
		Title:   "Test",
		Version: "1.0",
		Server:  "nats://localhost:4222",
	}
	err := run(opts, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("error = %q, want mention of parse error", err)
	}
}

func TestRun_NoAnnotatedStructs(t *testing.T) {
	// Write a valid Go file with no asyncapi annotations
	dir := t.TempDir()
	input := dir + "/empty.go"
	content := "package testdata\n\ntype Plain struct {\n\tName string `json:\"name\"`\n}\n"
	if err := writeTestFile(input, content); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	opts := Options{
		Input:   input,
		Output:  dir + "/out.yaml",
		Title:   "Test",
		Version: "1.0",
		Server:  "nats://localhost:4222",
	}
	err := run(opts, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for no annotated structs")
	}
	if !strings.Contains(err.Error(), "no annotated structs") {
		t.Errorf("error = %q, want mention of no annotated structs", err)
	}
}

func TestRun_HappyPath(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	opts := Options{
		Input:   "testdata/fixture.go",
		Output:  dir + "/asyncapi.yaml",
		Title:   "Test API",
		Version: "1.0.0",
		Server:  "nats://localhost:4222",
	}
	err := run(opts, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "wrote") {
		t.Errorf("stdout = %q, want mention of wrote", stdout.String())
	}
}

func TestRun_HappyPathWithSchemas(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	opts := Options{
		Input:      "testdata/fixture.go",
		Output:     dir + "/asyncapi.yaml",
		Title:      "Test API",
		Version:    "1.0.0",
		Server:     "nats://localhost:4222",
		SchemasDir: dir + "/schemas",
	}
	err := run(opts, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "wrote JSON schemas") {
		t.Errorf("stdout = %q, want mention of wrote JSON schemas", output)
	}
	if !strings.Contains(output, "wrote") {
		t.Errorf("stdout = %q, want mention of wrote", output)
	}
}

func TestRun_WriteError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	opts := Options{
		Input:   "testdata/fixture.go",
		Output:  "/nonexistent/deeply/nested/dir/out.yaml",
		Title:   "Test API",
		Version: "1.0.0",
		Server:  "nats://localhost:4222",
	}
	err := run(opts, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid output path")
	}
	if !strings.Contains(err.Error(), "write error") {
		t.Errorf("error = %q, want mention of write error", err)
	}
}

func writeTestFile(path, content string) error {
	return writeFileForTest(path, []byte(content))
}

func writeFileForTest(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644) //nolint:gosec // 0o644 is correct for test fixture files (SC-005)
}
