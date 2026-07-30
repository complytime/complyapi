# complyapi

Go library providing event contract types for the
[ComplyTime](https://github.com/complytime) ecosystem.

Events use [CloudEvents](https://cloudevents.io/) v1.0 envelopes with
JSON-encoded payloads. The canonical event contract is defined in the
[AsyncAPI 3.0 spec](api/events/asyncapi.yaml); the Go types in this
library must match that spec.

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

## License

[Apache-2.0](LICENSE)
