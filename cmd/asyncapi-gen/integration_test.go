// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegration_GeneratedMatchesCommitted regenerates asyncapi.yaml from
// events/events.go and verifies the output matches the committed file.
// This is the drift detector: it fails if the two are out of sync.
func TestIntegration_GeneratedMatchesCommitted(t *testing.T) {
	// Path to the real events source, relative to this test file location.
	inputPath := filepath.Join("..", "..", "events", "events.go")
	committedPath := filepath.Join("..", "..", "api", "events", "asyncapi.yaml")

	specs, err := ParseFile(inputPath)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	doc := BuildDoc(specs, "ComplyTime API Events", "0.1.0", "nats://localhost:4222")

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
