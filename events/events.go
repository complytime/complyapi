// SPDX-License-Identifier: Apache-2.0

// Package events defines CloudEvents data types for the ComplyTime
// evidence lifecycle.
package events

//go:generate go run ../cmd/asyncapi-gen -input ./events.go -output ../api/events/asyncapi.yaml -title "ComplyTime API Events" -version 0.1.0 -description "Event contract for the ComplyTime evidence lifecycle.\n\nAll public events use CloudEvents v1.0 envelope (JSON format).\nThe AsyncAPI spec is the source of truth for event contracts;\nGo types in the events package must match these schemas." -license Apache-2.0 -contact-name ComplyTime -contact-url https://github.com/complytime/complyapi -server nats://localhost:4222

import (
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
)

const TypeEvidenceIngested = "dev.complytime.evidence.ingested"

// EvidenceIngestedData is the CloudEvents data payload for
// evidence.ingested events.
type EvidenceIngestedData struct {
	//nolint:unused
	_ struct{} `asyncapi:"channel:core.evidence.ingested.{subjectId},param:subjectId=The compliance subject identifier,stream:EVIDENCE,type:dev.complytime.evidence.ingested,send:Published when evidence is accepted for processing,receive:Consume evidence-ingested events"`

	ContentDigest string  `json:"contentDigest"`
	ArtifactType  string  `json:"artifactType"`
	StorageRef    string  `json:"storageRef,omitempty"`
	SubjectID     string  `json:"subjectId"`
	ShardID       *string `json:"shardId,omitempty"`
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
