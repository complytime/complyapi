# complyapi

Go library providing event contract types for the
[ComplyTime](https://github.com/complytime) ecosystem.

Events use [CloudEvents](https://cloudevents.io/) v1.0 envelopes with
JSON-encoded payloads. Go types in `events/events.go` are the source of
truth for event contracts; the [AsyncAPI 3.0 spec](api/events/asyncapi.yaml)
and [JSON Schema files](api/events/schemas/) are generated from those types
via `go generate`. Do not edit the generated files manually.

## Installation

```bash
go get github.com/complytime/complyapi
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/complytime/complyapi/events"
)

func main() {
	data := events.EvidenceIngestedData{
		ContentDigest: "sha256:abc123...",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
	}

	e, err := events.NewEvidenceIngestedEvent("my-service", "my-app-v1", data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(e.Type()) // dev.complytime.evidence.ingested
}
```

## Event Types

| Type | Constant | Description |
|------|----------|-------------|
| `dev.complytime.evidence.ingested` | `events.TypeEvidenceIngested` | Evidence accepted for processing |

## Development

After modifying Go structs in `events/events.go`, regenerate derived artifacts:

```bash
task generate
```

This runs `go generate ./events/...` which rebuilds `api/events/asyncapi.yaml`
and the JSON Schema files in `api/events/schemas/`.

To validate the generated AsyncAPI spec:

```bash
task asyncapi-lint
```

To run all checks (lint, vet, test, asyncapi validation):

```bash
task check
```

### Adding a new event type

1. Define a new `*Data` struct in `events/events.go` with a sentinel blank
   field carrying the `asyncapi` tag (channel, params, stream, type, send,
   receive metadata).
2. Add `asyncapi-field:"description:..."` tags on each struct field for
   schema descriptions.
3. Run `task generate` to regenerate all derived artifacts.
4. Add a constructor function (e.g., `NewYourEventEvent()`) following the
   existing pattern.

### Evolving an event contract

Before changing an existing event's payload, read the
[Event Versioning Strategy](docs/versioning.md). It defines which field to bump
for additive versus breaking changes (CloudEvents `type` and AsyncAPI
`info.version`), and why the NATS subject stays stable so subscribers never
re-subscribe.

## License

[Apache-2.0](LICENSE)
