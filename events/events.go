// SPDX-License-Identifier: Apache-2.0

// Package events defines CloudEvents data types for the ComplyTime
// evidence lifecycle.
package events

//go:generate go run ../cmd/asyncapi-gen -input ./events.go -output ../api/events/asyncapi.yaml -schemas-dir ../api/events/schemas -title "ComplyTime API Events" -version 0.1.0 -description "Event contract for the ComplyTime evidence lifecycle.\n\nAll public events use CloudEvents v1.0 envelope (JSON format).\nThis spec is generated from Go types in the events package via cmd/asyncapi-gen.\nDo not edit manually — run 'go generate ./events/...' to regenerate." -license Apache-2.0 -contact-name ComplyTime -contact-url https://github.com/complytime/complyapi -server nats://localhost:4222

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
)

// subjectIDPattern restricts subjectId to characters that are safe as a
// literal NATS subject token: alphanumerics, dash, and underscore. It
// excludes ".", "*", and ">" — the NATS token separator and wildcards — so
// a subjectId can never reshape or collide with a subscriber's wildcard
// binding (e.g. "core.evidence.ingested.*").
var subjectIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// storageRefPattern requires storageRef to carry a URI-style scheme prefix
// (e.g. "s3://", "gcp://", "locker://") per RFC 3986 scheme syntax. The set
// of valid backends is not fixed, so this validates general shape rather
// than an enumerated allowlist.
var storageRefPattern = regexp.MustCompile(`^[a-z][a-z0-9+.-]*://`)

// validateSubjectID returns an error if subjectID is empty or contains any
// character unsafe for use as a literal NATS subject token.
func validateSubjectID(subjectID string) error {
	if subjectID == "" {
		return errors.New("subjectId must not be empty")
	}
	if !subjectIDPattern.MatchString(subjectID) {
		return fmt.Errorf("subjectId %q must match %s", subjectID, subjectIDPattern.String())
	}
	return nil
}

const TypeEvidenceIngested = "dev.complytime.evidence.ingested"

// EvidenceIngestedData is the CloudEvents data payload for
// evidence.ingested events.
type EvidenceIngestedData struct {
	//nolint:unused
	_ struct{} `asyncapi:"channel:core.evidence.ingested.{subjectId},param:subjectId=The compliance subject identifier,stream:EVIDENCE,type:dev.complytime.evidence.ingested,send:Published when evidence is accepted for processing; before sealing,receive:Consume evidence-ingested events,description:Evidence ingestion pipeline for compliance artifacts"`

	ContentDigest string  `json:"contentDigest" asyncapi-field:"description:SHA-256 digest of the evidence artifact"`
	ArtifactType  string  `json:"artifactType" asyncapi-field:"description:Gemara artifact type"`
	StorageRef    string  `json:"storageRef" asyncapi-field:"description:URI-style storage reference consumers use to fetch the evidence artifact (must include a scheme prefix, e.g. s3://, gcp://, locker://)"`
	SubjectID     string  `json:"subjectId" asyncapi-field:"description:Compliance subject identifier"`
	ShardID       *string `json:"shardId,omitempty" asyncapi-field:"description:Subject shard identifier (null when sharding is not configured)"`
}

// NewEvidenceIngestedEvent constructs a CloudEvents v1.0 event with
// the given source, subject, and data payload.
func NewEvidenceIngestedEvent(source, subject string, data EvidenceIngestedData) (cloudevents.Event, error) {
	if source == "" {
		return cloudevents.Event{}, errors.New("source must not be empty")
	}
	if subject == "" {
		return cloudevents.Event{}, errors.New("subject must not be empty")
	}
	if err := validateSubjectID(data.SubjectID); err != nil {
		return cloudevents.Event{}, err
	}
	if data.StorageRef == "" {
		return cloudevents.Event{}, errors.New("storageRef must not be empty")
	}
	if !storageRefPattern.MatchString(data.StorageRef) {
		return cloudevents.Event{}, fmt.Errorf("storageRef %q must have a URI scheme prefix (e.g. s3://, gcp://, locker://)", data.StorageRef)
	}

	e := event.New(cloudevents.VersionV1)
	e.SetID(uuid.New().String())
	e.SetType(TypeEvidenceIngested)
	e.SetSource(source)
	e.SetSubject(subject)
	e.SetTime(time.Now())
	e.SetDataContentType("application/json")
	if err := e.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return cloudevents.Event{}, err
	}
	return e, nil
}
