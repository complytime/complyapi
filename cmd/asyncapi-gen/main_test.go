// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
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
	dir := t.TempDir()
	input := filepath.Join(dir, "empty.go")
	content := "package testdata\n\ntype Plain struct {\n\tName string `json:\"name\"`\n}\n"
	if err := os.WriteFile(input, []byte(content), 0o644); err != nil { //nolint:gosec // 0o644 is correct for test fixture files (SC-005)
		t.Fatalf("writing test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	opts := Options{
		Input:   input,
		Output:  filepath.Join(dir, "out.yaml"),
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
		Output:  filepath.Join(dir, "asyncapi.yaml"),
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
		Output:     filepath.Join(dir, "asyncapi.yaml"),
		Title:      "Test API",
		Version:    "1.0.0",
		Server:     "nats://localhost:4222",
		SchemasDir: filepath.Join(dir, "schemas"),
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

func TestRun_SchemaWriteError(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	opts := Options{
		Input:      "testdata/fixture.go",
		Output:     filepath.Join(dir, "asyncapi.yaml"),
		Title:      "Test API",
		Version:    "1.0.0",
		Server:     "nats://localhost:4222",
		SchemasDir: "/nonexistent/deeply/nested/dir/schemas",
	}
	err := run(opts, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid schemas directory")
	}
	if !strings.Contains(err.Error(), "schema write error") {
		t.Errorf("error = %q, want mention of schema write error", err)
	}
}

func TestParseFlags_AllFlags(t *testing.T) {
	args := []string{
		"-input", "events.go",
		"-output", "out.yaml",
		"-title", "My API",
		"-version", "2.0.0",
		"-server", "nats://host:4222",
		"-description", "A description",
		"-license", "MIT",
		"-contact-name", "Alice",
		"-contact-url", "https://example.com",
		"-schemas-dir", "/tmp/schemas",
	}
	opts := parseFlags(args)

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"Input", opts.Input, "events.go"},
		{"Output", opts.Output, "out.yaml"},
		{"Title", opts.Title, "My API"},
		{"Version", opts.Version, "2.0.0"},
		{"Server", opts.Server, "nats://host:4222"},
		{"Description", opts.Description, "A description"},
		{"LicenseName", opts.LicenseName, "MIT"},
		{"ContactName", opts.ContactName, "Alice"},
		{"ContactURL", opts.ContactURL, "https://example.com"},
		{"SchemasDir", opts.SchemasDir, "/tmp/schemas"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	opts := parseFlags([]string{})
	if opts.Input != "" {
		t.Errorf("Input default = %q, want empty", opts.Input)
	}
	if opts.SchemasDir != "" {
		t.Errorf("SchemasDir default = %q, want empty", opts.SchemasDir)
	}
}
