// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// WriteYAML marshals doc to YAML and writes it to path, prepending the
// SPDX license header. The file is created with 0o644 permissions.
func WriteYAML(doc AsyncAPIDoc, path string) error {
	b, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshaling AsyncAPI document: %w", err)
	}

	header := "# SPDX-License-Identifier: Apache-2.0\n"
	out := append([]byte(header), b...)

	if err := os.WriteFile(path, out, 0o644); err != nil { //nolint:gosec // 0o644 is correct for generated YAML output files (SC-005)
		return fmt.Errorf("writing output file: %w", err)
	}
	return nil
}
