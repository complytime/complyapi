// SPDX-License-Identifier: Apache-2.0

// Command asyncapi-gen generates an AsyncAPI 3.0 document from annotated
// Go event structs. Run via go generate in the events package.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// Options holds the parsed command-line flags for asyncapi-gen.
type Options struct {
	Input       string
	Output      string
	Title       string
	Version     string
	Server      string
	Description string
	LicenseName string
	ContactName string
	ContactURL  string
	SchemasDir  string
}

func main() {
	opts := parseFlags(os.Args[1:])
	if err := run(opts, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "asyncapi-gen: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags parses command-line arguments into Options.
func parseFlags(args []string) Options {
	fs := flag.NewFlagSet("asyncapi-gen", flag.ExitOnError)
	var opts Options
	fs.StringVar(&opts.Input, "input", "", "Path to Go source file containing annotated event structs (required)")
	fs.StringVar(&opts.Output, "output", "", "Path to write the generated asyncapi.yaml (required)")
	fs.StringVar(&opts.Title, "title", "", "AsyncAPI document title (required)")
	fs.StringVar(&opts.Version, "version", "", "AsyncAPI document version (required)")
	fs.StringVar(&opts.Server, "server", "", "NATS server URL, e.g. nats://localhost:4222 (required)")
	fs.StringVar(&opts.Description, "description", "", "AsyncAPI document description (optional)")
	fs.StringVar(&opts.LicenseName, "license", "", "License name, e.g. Apache-2.0 (optional)")
	fs.StringVar(&opts.ContactName, "contact-name", "", "Contact name (optional)")
	fs.StringVar(&opts.ContactURL, "contact-url", "", "Contact URL (optional)")
	fs.StringVar(&opts.SchemasDir, "schemas-dir", "", "Directory to write JSON Schema files (optional)")
	_ = fs.Parse(args)
	return opts
}

// run executes the asyncapi-gen pipeline with the given options.
func run(opts Options, stdout, stderr io.Writer) error {
	if opts.Input == "" || opts.Output == "" || opts.Title == "" || opts.Version == "" || opts.Server == "" {
		return fmt.Errorf("required flags missing: -input -output -title -version -server")
	}

	specs, err := ParseFile(opts.Input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if len(specs) == 0 {
		return fmt.Errorf("no annotated structs found in %s", opts.Input)
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

	if err := WriteYAML(doc, opts.Output); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	if opts.SchemasDir != "" {
		if err := WriteJSONSchemas(specs, opts.SchemasDir); err != nil {
			return fmt.Errorf("schema write error: %w", err)
		}
		fmt.Fprintf(stdout, "asyncapi-gen: wrote JSON schemas to %s\n", opts.SchemasDir) //nolint:gosec // G705: stdout is a CLI status stream, not an HTML/browser sink; XSS does not apply
	}

	fmt.Fprintf(stdout, "asyncapi-gen: wrote %s (%d event(s))\n", opts.Output, len(specs)) //nolint:gosec // G705: stdout is a CLI status stream, not an HTML/browser sink; XSS does not apply
	return nil
}
