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
| `dev.complytime.evidence.ingested` | `events.TypeEvidenceIngested` | Evidence accepted for processing, before validation |
| `dev.complytime.evidence.sealed` | `events.TypeEvidenceSealed` | Evidence validated and sealed into a unit of work |
| `dev.complytime.evidence.quarantined` | `events.TypeEvidenceQuarantined` | Evidence failed validation and was quarantined |

### Payload Examples

Example CloudEvents JSON payloads are in
[`api/events/examples/`](api/events/examples/):

| File | Description |
|------|-------------|
| [`evidence-ingested.json`](api/events/examples/evidence-ingested.json) | Common payload with required fields and `storageRef` |
| [`evidence-ingested-minimal.json`](api/events/examples/evidence-ingested-minimal.json) | Required data fields only (no optional fields) |
| [`evidence-ingested-with-shard.json`](api/events/examples/evidence-ingested-with-shard.json) | All fields including optional `shardId` |
| [`evidence-sealed.json`](api/events/examples/evidence-sealed.json) | Sealed outcome after successful validation |
| [`evidence-quarantined.json`](api/events/examples/evidence-quarantined.json) | Quarantined outcome after failed validation, with `reason` |

These examples conform to the [JSON Schema](api/events/schemas/) and
[AsyncAPI spec](api/events/asyncapi.yaml). They are hand-maintained
reference payloads; the generated schemas remain the source of truth
for validation.

## Development

### Prerequisites

- [Go](https://go.dev/) (version per `go.mod`)
- [Task](https://taskfile.dev/) (task runner)
- [golangci-lint](https://golangci-lint.run/)
- [Node.js](https://nodejs.org/) with `npx` (required by `task asyncapi-lint`; first run downloads the AsyncAPI CLI from npm)

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
   field carrying the `asyncapi` tag. The tag is a comma-separated list of
   `key:value` pairs. **Values must not contain commas** (the parser splits
   on commas, and a comma inside a value silently truncates it). Use
   semicolons for natural pauses. Recognised keys:

   | Key | Required | Format | Description |
   |-----|----------|--------|-------------|
   | `channel` | yes | NATS subject with `{param}` placeholders | Channel address |
   | `param` | no | `name=description` (repeatable) | Channel parameter |
   | `stream` | yes | Upper-case stream name | NATS JetStream stream |
   | `type` | yes | Reverse-DNS CloudEvents type | CloudEvents `type` attribute |
   | `send` | yes | Free text (no commas) | Send operation summary |
   | `receive` | yes | Free text (no commas) | Receive operation summary |
   | `description` | no | Free text (no commas) | Channel description |

   Example sentinel field:
   ```go
   _ struct{} `asyncapi:"channel:complyapi.widget.created.{ownerId},param:ownerId=The widget owner,stream:WIDGETS,type:dev.complytime.widget.created,send:Published when a widget is created,receive:Consume widget-created events,description:Widget creation pipeline"`
   ```

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
