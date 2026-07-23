// SPDX-License-Identifier: Apache-2.0

// Package events defines CloudEvents data types for the ComplyTime
// evidence lifecycle.
package events

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
