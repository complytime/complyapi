// SPDX-License-Identifier: Apache-2.0

// Command asyncapi-gen generates an AsyncAPI 3.0 document from annotated
// Go event structs. Run via go generate in the events package.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	input := flag.String("input", "", "Path to Go source file containing annotated event structs (required)")
	output := flag.String("output", "", "Path to write the generated asyncapi.yaml (required)")
	title := flag.String("title", "", "AsyncAPI document title (required)")
	version := flag.String("version", "", "AsyncAPI document version (required)")
	server := flag.String("server", "", "NATS server URL, e.g. nats://localhost:4222 (required)")
	flag.Parse()

	if *input == "" || *output == "" || *title == "" || *version == "" || *server == "" {
		fmt.Fprintln(os.Stderr, "asyncapi-gen: all flags are required: -input -output -title -version -server")
		flag.Usage()
		os.Exit(1)
	}

	specs, err := ParseFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "asyncapi-gen: parse error: %v\n", err)
		os.Exit(1)
	}
	if len(specs) == 0 {
		fmt.Fprintln(os.Stderr, "asyncapi-gen: no annotated structs found in input file")
		os.Exit(1)
	}

	doc := BuildDoc(specs, *title, *version, *server)

	if err := WriteYAML(doc, *output); err != nil {
		fmt.Fprintf(os.Stderr, "asyncapi-gen: write error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("asyncapi-gen: wrote %s (%d event(s))\n", *output, len(specs))
}
