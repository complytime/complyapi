// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// parseGenerateDirective reads the go:generate directive from events.go
// and returns the flag arguments (everything after "go run ../cmd/asyncapi-gen").
// This ensures the integration test always uses the same metadata as the
// real go:generate invocation — no hardcoded duplication.
func parseGenerateDirective(t *testing.T, eventsPath string) []string {
	t.Helper()
	f, err := os.Open(eventsPath)
	if err != nil {
		t.Fatalf("opening %s: %v", eventsPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Increase buffer for long go:generate lines.
	scanner.Buffer(make([]byte, 0, 64*1024), 64*1024)
	const prefix = "//go:generate go run ../cmd/asyncapi-gen "
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			argStr := line[len(prefix):]
			return splitArgs(argStr)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning %s: %v", eventsPath, err)
	}
	t.Fatalf("no //go:generate directive found in %s", eventsPath)
	return nil
}

// splitArgs splits a flag string respecting quoted values (for -description etc).
// Inside quoted segments, \n is interpreted as a real newline to match go generate
// behaviour (go generate processes escape sequences in quoted arguments).
func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		if ch == ' ' && !inQuote {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}
		// Handle escape sequences inside quoted strings.
		if inQuote && ch == '\\' && i+1 < len(s) {
			next := s[i+1]
			switch next {
			case 'n':
				current.WriteByte('\n')
				i++
				continue
			case 't':
				current.WriteByte('\t')
				i++
				continue
			case '\\':
				current.WriteByte('\\')
				i++
				continue
			case '"':
				current.WriteByte('"')
				i++
				continue
			}
		}
		current.WriteByte(ch)
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"simple", "-a foo -b bar", []string{"-a", "foo", "-b", "bar"}},
		{"quoted with spaces", `-title "My API Title"`, []string{"-title", "My API Title"}},
		{"newline escape", `-desc "line1\nline2"`, []string{"-desc", "line1\nline2"}},
		{"tab escape", `-desc "col1\tcol2"`, []string{"-desc", "col1\tcol2"}},
		{"backslash escape", `-path "c:\\dir"`, []string{"-path", "c:\\dir"}},
		{"escaped quote", `-msg "say \"hi\""`, []string{"-msg", `say "hi"`}},
		{"empty input", "", nil},
		{"trailing spaces", "-a foo  ", []string{"-a", "foo"}},
		{"adjacent values", `-a "x" -b "y"`, []string{"-a", "x", "-b", "y"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitArgs(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitArgs(%q) = %v (len %d), want %v (len %d)", tt.in, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitArgs(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestIntegration_GeneratedMatchesCommitted regenerates asyncapi.yaml from
// events/events.go and verifies the output matches the committed file.
// This is the drift detector: it fails if the two are out of sync.
func TestIntegration_GeneratedMatchesCommitted(t *testing.T) {
	inputPath := filepath.Join("..", "..", "events", "events.go")
	committedPath := filepath.Join("..", "..", "api", "events", "asyncapi.yaml")

	// Parse flags from the go:generate directive — single source of truth.
	args := parseGenerateDirective(t, inputPath)
	opts := parseFlags(args)

	specs, err := ParseFile(inputPath)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	doc := BuildDoc(specs, DocMeta{
		Title:       opts.Title,
		Version:     opts.Version,
		Description: opts.Description,
		LicenseName: opts.LicenseName,
		ContactName: opts.ContactName,
		ContactURL:  opts.ContactURL,
		ServerURL:   opts.Server,
	})

	outPath := filepath.Join(t.TempDir(), "asyncapi.yaml")
	if err := WriteYAML(doc, outPath); err != nil {
		t.Fatalf("WriteYAML: %v", err)
	}

	generated, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	committed, err := os.ReadFile(committedPath)
	if err != nil {
		t.Fatalf("reading committed file: %v", err)
	}

	if string(generated) != string(committed) {
		t.Errorf("generated asyncapi.yaml does not match committed file.\n"+
			"Run `go generate ./events/...` to update it.\n\n"+
			"--- committed\n+++ generated\n%s",
			diffStrings(string(committed), string(generated)),
		)
	}
}

// TestIntegration_JSONSchemasMatchCommitted regenerates JSON Schema files
// and verifies they match the committed versions.
func TestIntegration_JSONSchemasMatchCommitted(t *testing.T) {
	inputPath := filepath.Join("..", "..", "events", "events.go")
	committedDir := filepath.Join("..", "..", "api", "events", "schemas")

	specs, err := ParseFile(inputPath)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "schemas")
	if err := WriteJSONSchemas(specs, outDir); err != nil {
		t.Fatalf("WriteJSONSchemas: %v", err)
	}

	for _, spec := range specs {
		files := []string{
			spec.StructName + ".schema.json",
			envelopeSchemaName(spec.StructName) + ".schema.json",
		}
		for _, name := range files {
			generated, err := os.ReadFile(filepath.Join(outDir, name))
			if err != nil {
				t.Fatalf("reading generated %s: %v", name, err)
			}
			committed, err := os.ReadFile(filepath.Join(committedDir, name))
			if err != nil {
				t.Fatalf("reading committed %s: %v", name, err)
			}
			if string(generated) != string(committed) {
				t.Errorf("generated %s does not match committed file.\n"+
					"Run `go generate ./events/...` to update it.\n\n"+
					"--- committed\n+++ generated\n%s",
					name, diffStrings(string(committed), string(generated)),
				)
			}
		}
	}
}

// diffStrings returns a simple line-diff between a and b.
func diffStrings(a, b string) string {
	aLines := splitLines(a)
	bLines := splitLines(b)
	var out []string
	max := len(aLines)
	if len(bLines) > max {
		max = len(bLines)
	}
	for i := 0; i < max; i++ {
		var al, bl string
		if i < len(aLines) {
			al = aLines[i]
		}
		if i < len(bLines) {
			bl = bLines[i]
		}
		if al != bl {
			out = append(out, fmt.Sprintf("line %d:\n  committed:  %q\n  generated:  %q", i+1, al, bl))
		}
	}
	if len(out) == 0 {
		return "(no line differences found — may be whitespace)"
	}
	return strings.Join(out, "\n")
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
