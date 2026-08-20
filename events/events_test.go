// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestTypeEvidenceIngestedConstant(t *testing.T) {
	if TypeEvidenceIngested != "dev.complytime.evidence.ingested" {
		t.Errorf("TypeEvidenceIngested = %q, want %q", TypeEvidenceIngested, "dev.complytime.evidence.ingested")
	}
}

func TestEvidenceIngestedDataJSON(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		StorageRef:    "ref/123",
		SubjectID:     "my-app-v1",
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got EvidenceIngestedData
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.ContentDigest != data.ContentDigest {
		t.Errorf("ContentDigest = %q, want %q", got.ContentDigest, data.ContentDigest)
	}
	if got.ArtifactType != data.ArtifactType {
		t.Errorf("ArtifactType = %q, want %q", got.ArtifactType, data.ArtifactType)
	}
	if got.StorageRef != data.StorageRef {
		t.Errorf("StorageRef = %q, want %q", got.StorageRef, data.StorageRef)
	}
	if got.SubjectID != data.SubjectID {
		t.Errorf("SubjectID = %q, want %q", got.SubjectID, data.SubjectID)
	}
}

func TestEvidenceIngestedDataJSONOmitsOptionalFields(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	raw := make(map[string]interface{})
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}

	if _, ok := raw["storageRef"]; ok {
		t.Error("storageRef should be omitted when empty")
	}
	if _, ok := raw["shardId"]; ok {
		t.Error("shardId should be omitted when nil")
	}
}

func TestEvidenceIngestedDataJSONIncludesShardID(t *testing.T) {
	shard := "shard-1"
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
		ShardID:       &shard,
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	raw := make(map[string]interface{})
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}

	if raw["shardId"] != "shard-1" {
		t.Errorf("shardId = %v, want %q", raw["shardId"], "shard-1")
	}
}

func TestNewEvidenceIngestedEvent(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
	}

	e, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data)
	if err != nil {
		t.Fatalf("NewEvidenceIngestedEvent: %v", err)
	}

	if e.Type() != TypeEvidenceIngested {
		t.Errorf("Type() = %q, want %q", e.Type(), TypeEvidenceIngested)
	}
	if e.Source() != "complytime-gateway" {
		t.Errorf("Source() = %q, want %q", e.Source(), "complytime-gateway")
	}
	if e.Subject() != "my-app-v1" {
		t.Errorf("Subject() = %q, want %q", e.Subject(), "my-app-v1")
	}
	if e.SpecVersion() != cloudevents.VersionV1 {
		t.Errorf("SpecVersion() = %q, want %q", e.SpecVersion(), cloudevents.VersionV1)
	}
	if e.DataContentType() != "application/json" {
		t.Errorf("DataContentType() = %q, want %q", e.DataContentType(), "application/json")
	}
	if e.Time().IsZero() {
		t.Error("Time() should not be zero")
	}
	if e.ID() == "" {
		t.Error("ID() should not be empty")
	}

	var got EvidenceIngestedData
	if err := e.DataAs(&got); err != nil {
		t.Fatalf("DataAs: %v", err)
	}
	if got.ContentDigest != data.ContentDigest {
		t.Errorf("data.ContentDigest = %q, want %q", got.ContentDigest, data.ContentDigest)
	}
	if got.ArtifactType != data.ArtifactType {
		t.Errorf("data.ArtifactType = %q, want %q", got.ArtifactType, data.ArtifactType)
	}
	if got.SubjectID != data.SubjectID {
		t.Errorf("data.SubjectID = %q, want %q", got.SubjectID, data.SubjectID)
	}
}

func TestNewEvidenceIngestedEventEmptySource(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
	}

	_, err := NewEvidenceIngestedEvent("", "my-app-v1", data)
	if err == nil {
		t.Error("expected error for empty source")
	}
}

func TestNewEvidenceIngestedEventEmptySubject(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
	}

	_, err := NewEvidenceIngestedEvent("complytime-gateway", "", data)
	if err == nil {
		t.Error("expected error for empty subject")
	}
}

func TestNewEvidenceIngestedEventWireFormatRoundTrip(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		StorageRef:    "ref/456",
		SubjectID:     "my-app-v1",
	}

	e, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data)
	if err != nil {
		t.Fatalf("NewEvidenceIngestedEvent: %v", err)
	}

	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal event: %v", err)
	}

	var restored cloudevents.Event
	if err := json.Unmarshal(b, &restored); err != nil {
		t.Fatalf("Unmarshal event: %v", err)
	}

	if restored.Type() != TypeEvidenceIngested {
		t.Errorf("restored Type() = %q, want %q", restored.Type(), TypeEvidenceIngested)
	}
	if restored.Source() != "complytime-gateway" {
		t.Errorf("restored Source() = %q, want %q", restored.Source(), "complytime-gateway")
	}
	if restored.Subject() != "my-app-v1" {
		t.Errorf("restored Subject() = %q, want %q", restored.Subject(), "my-app-v1")
	}
	if restored.SpecVersion() != cloudevents.VersionV1 {
		t.Errorf("restored SpecVersion() = %q, want %q", restored.SpecVersion(), cloudevents.VersionV1)
	}

	var got EvidenceIngestedData
	if err := restored.DataAs(&got); err != nil {
		t.Fatalf("restored DataAs: %v", err)
	}
	if got.ContentDigest != data.ContentDigest {
		t.Errorf("ContentDigest = %q, want %q", got.ContentDigest, data.ContentDigest)
	}
	if got.ArtifactType != data.ArtifactType {
		t.Errorf("ArtifactType = %q, want %q", got.ArtifactType, data.ArtifactType)
	}
	if got.StorageRef != data.StorageRef {
		t.Errorf("StorageRef = %q, want %q", got.StorageRef, data.StorageRef)
	}
	if got.SubjectID != data.SubjectID {
		t.Errorf("SubjectID = %q, want %q", got.SubjectID, data.SubjectID)
	}
}

func TestNewEvidenceIngestedEventInvalidSubjectID(t *testing.T) {
	cases := []struct {
		name      string
		subjectID string
	}{
		{"contains dot", "my-app.v1"},
		{"contains star wildcard", "my-app-*"},
		{"contains gt wildcard", "my-app->"},
		{"contains space", "my app"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := EvidenceIngestedData{
				ContentDigest: "sha256:abc123",
				ArtifactType:  "application/vnd.gemara.evaluation-log+json",
				SubjectID:     tc.subjectID,
			}
			_, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data)
			if err == nil {
				t.Errorf("expected error for subjectId %q", tc.subjectID)
			}
		})
	}
}

func TestNewEvidenceIngestedEventValidSubjectIDCharset(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app_v1-2",
	}
	if _, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data); err != nil {
		t.Errorf("unexpected error for valid subjectId: %v", err)
	}
}

func TestExamplePayloads_ConformToSchema(t *testing.T) {
	examplesDir := filepath.Join("..", "api", "events", "examples")
	files, err := filepath.Glob(filepath.Join(examplesDir, "*.json"))
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no example files found in api/events/examples/")
	}

	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			// Validate the full CloudEvents envelope deserializes.
			var envelope struct {
				SpecVersion     string               `json:"specversion"`
				ID              string               `json:"id"`
				Type            string               `json:"type"`
				Source          string               `json:"source"`
				Subject         string               `json:"subject"`
				Time            string               `json:"time"`
				DataContentType string               `json:"datacontenttype"`
				Data            EvidenceIngestedData `json:"data"`
			}
			if err := json.Unmarshal(b, &envelope); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			// CloudEvents envelope const values.
			if envelope.SpecVersion != "1.0" {
				t.Errorf("specversion = %q, want %q", envelope.SpecVersion, "1.0")
			}
			if envelope.Type != TypeEvidenceIngested {
				t.Errorf("type = %q, want %q", envelope.Type, TypeEvidenceIngested)
			}
			if envelope.DataContentType != "application/json" {
				t.Errorf("datacontenttype = %q, want %q", envelope.DataContentType, "application/json")
			}

			// Required envelope fields must be non-empty.
			if envelope.ID == "" {
				t.Error("id must not be empty")
			}
			if envelope.Source == "" {
				t.Error("source must not be empty")
			}
			if envelope.Subject == "" {
				t.Error("subject must not be empty")
			}
			if envelope.Time == "" {
				t.Error("time must not be empty")
			}

			// Required data fields must be non-empty.
			if envelope.Data.ContentDigest == "" {
				t.Error("data.contentDigest must not be empty")
			}
			if envelope.Data.ArtifactType == "" {
				t.Error("data.artifactType must not be empty")
			}
			if envelope.Data.SubjectID == "" {
				t.Error("data.subjectId must not be empty")
			}
		})
	}
}
