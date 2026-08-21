<!-- SPDX-License-Identifier: Apache-2.0 -->
# Event Versioning Strategy

This document defines how ComplyTime event contracts evolve without breaking
existing subscribers. It tells producers which field to change for a given kind
of change, and tells consumers what they must handle. Read it before making any
change to a type in [`events/events.go`](../events/events.go).

The strategy follows the accepted ADRs in `complytime/complytime`:
[ADR-0019](https://github.com/complytime/complytime/blob/main/docs/ADRs/0019-event-driven-ingestion.md)
(event-driven ingestion),
[ADR-0020](https://github.com/complytime/complytime/blob/main/docs/ADRs/0020-nats-jetstream-event-bus.md)
(NATS JetStream),
[ADR-0021](https://github.com/complytime/complytime/blob/main/docs/ADRs/0021-cloudevents-envelope.md)
(CloudEvents envelope), and
[ADR-0022](https://github.com/complytime/complytime/blob/main/docs/ADRs/0022-asyncapi-interface-docs.md)
(AsyncAPI contract).

## The four layers

Each layer versions a different thing. Changing the wrong one either breaks
subscribers or signals nothing.

| Layer | Field | Example | Versions | Reacts |
|-------|-------|---------|----------|--------|
| NATS subject | channel address | `core.evidence.ingested.{subjectId}` | Nothing — routing address only | Broker (routing) |
| CloudEvents type | `type` | `dev.complytime.evidence.ingested` | The event's payload contract (breaking changes) | Consumer per-message dispatch |
| AsyncAPI contract | `info.version` | `0.2.0` | The whole published contract | Humans, codegen tooling |
| CloudEvents envelope | `specversion` | `1.0` | The CNCF envelope format — not ours | Nobody (hands off) |

### NATS subject — frozen routing address

The subject is where a subscriber listens (ADR-0020). In this codebase that
hierarchy is `domain.entity.action.{param}` — e.g.
`core.evidence.ingested.{subjectId}`, `core.evidence.sealed.{subjectId}`,
`core.evidence.quarantined.{subjectId}`. The subject's `domain` segment
(`core`) is a routing namespace only — it does not need to match the
CloudEvents `type`'s reverse-DNS namespace (`dev.complytime`). The two
namespaces serve different audiences: the subject namespace scopes NATS
routing/permissions, the `type` namespace scopes payload-contract identity.
Do not infer one from the other; dispatch on `type`, subscribe on subject.

The subject MUST NOT carry a version segment. Version information belongs in
the CloudEvents `type` (see below), per ADR-0021, which routes and filters on
`type` and `source`.

Keeping version out of the subject means subscribers bind once, with a wildcard,
and never re-subscribe:

```
core.evidence.ingested.*        # every subjectId, ingested outcome only
core.evidence.sealed.*          # every subjectId, sealed outcome only
core.evidence.quarantined.*     # every subjectId, quarantined outcome only
core.evidence.>                 # every outcome, every subjectId, any deeper segments
```

That single binding survives both a new `subjectId` and a new payload version,
because neither changes the subject.

### CloudEvents `type` — the breaking-change signal

The `type` attribute identifies the event and its payload contract. A breaking
payload change bumps `type`; consumers switch on it:

```go
switch e.Type() {
case "dev.complytime.evidence.ingested":     // v1
    handleV1(e)
case "dev.complytime.evidence.ingested.v2":  // v2
    handleV2(e)
}
```

Both versions flow on the same subject. A v1-only consumer keeps matching the v1
`type` and ignores v2 messages; a v2 consumer handles both. No coordinated
cutover (ADR-0019: the API must not need to change when a consumer is added or
removed).

### AsyncAPI `info.version` — the human and codegen contract

The [AsyncAPI spec](../api/events/asyncapi.yaml) is the public, machine-readable
contract (ADR-0022). Consumers integrate from it and pin codegen to
`info.version`. It is generated from the Go types, so bump it in the
`//go:generate` directive in [`events/events.go`](../events/events.go), then
run `go generate ./events/...`.

### CloudEvents `specversion` — do not touch

`specversion` is the CNCF CloudEvents envelope version (currently `1.0`). It
changes only if CloudEvents itself releases a new envelope format. It is never
your schema-change signal. Do not conflate it with `type`.

## Pre-1.0 stability

While `info.version` is below `1.0.0`, the contract is considered unstable and
may change in place without a version bump. No downstream consumers should pin
to a pre-1.0 contract version for codegen stability. Once the contract reaches
`1.0.0`, the decision rules below become mandatory.

## Decision rule

Classify the change, then act:

### Additive / non-breaking

Adding an optional field, a new enum value, or a new event type.

1. Bump `info.version` minor (`0.1.0` → `0.2.0`) in the `//go:generate`
   directive in [`events/events.go`](../events/events.go).
2. Leave `type`, `specversion`, and the subject unchanged.
3. Run `go generate ./events/...` and commit the regenerated artifacts.

Old consumers ignore the new field (tolerant JSON reader). Nothing breaks.

### Breaking

Removing or renaming a field, changing a field's type, or changing the required
set.

1. Bump the CloudEvents `type` with a version suffix
   (`dev.complytime.evidence.ingested` → `dev.complytime.evidence.ingested.v2`).
   Update **both** the `Type…` constant and the `type:` key inside the `asyncapi`
   sentinel tag on the corresponding `*Data` struct in
   [`events/events.go`](../events/events.go). The generated `const` in
   `api/events/asyncapi.yaml` and `api/events/schemas/*.schema.json` comes from
   the tag, not the constant.
2. Bump `info.version` major (`1.0.0` → `2.0.0`) in the `//go:generate`
   directive.
3. Keep the subject unchanged.
4. Producer emits the new `type` on the same subject. Retire the old `type` only
   after consumers have migrated.
5. Run `go generate ./events/...` and commit the regenerated artifacts.

## Consumer contract

To stay compatible across versions, a subscriber MUST:

- Bind the subject with a wildcard, not a fixed `subjectId`.
- Dispatch on the CloudEvents `type`, not on subject text.
- Treat unknown fields as ignorable (tolerant reader).
- Treat an unknown `type` as skip-or-log, never as a fatal error.

## Reference

- Source of truth: [`events/events.go`](../events/events.go)
- Generated contract: [`api/events/asyncapi.yaml`](../api/events/asyncapi.yaml),
  [`api/events/schemas/`](../api/events/schemas/)
- CI validation: [`.github/workflows/ci_asyncapi.yml`](../.github/workflows/ci_asyncapi.yml)
- [CloudEvents specification](https://cloudevents.io/)
- [AsyncAPI specification](https://www.asyncapi.com/)
