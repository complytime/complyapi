// SPDX-License-Identifier: Apache-2.0

package events

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"
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

func TestNewEvidenceIngestedEvent(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		StorageRef:    "locker://ref/123",
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

func TestNewEvidenceIngestedEventEmptyStorageRef(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
	}
	if _, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data); err == nil {
		t.Error("expected error for empty storageRef")
	}
}

func TestNewEvidenceIngestedEventStorageRefMissingScheme(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		StorageRef:    "store/evidence/2026/08/20/a1b2c3d4",
		SubjectID:     "my-app-v1",
	}
	if _, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data); err == nil {
		t.Error("expected error for storageRef missing URI scheme prefix")
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

func TestNewEvidenceIngestedEventEmptyContentDigest(t *testing.T) {
	data := EvidenceIngestedData{
		ArtifactType: "application/vnd.gemara.evaluation-log+json",
		StorageRef:   "locker://ref/123",
		SubjectID:    "my-app-v1",
	}
	if _, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data); err == nil {
		t.Error("expected error for empty contentDigest")
	}
}

func TestNewEvidenceIngestedEventEmptyArtifactType(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		StorageRef:    "locker://ref/123",
		SubjectID:     "my-app-v1",
	}
	if _, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data); err == nil {
		t.Error("expected error for empty artifactType")
	}
}

func TestNewEvidenceIngestedEventWireFormatRoundTrip(t *testing.T) {
	data := EvidenceIngestedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		StorageRef:    "locker://ref/456",
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
		StorageRef:    "locker://ref/123",
		SubjectID:     "my-app_v1-2",
	}
	if _, err := NewEvidenceIngestedEvent("complytime-gateway", "my-app-v1", data); err != nil {
		t.Errorf("unexpected error for valid subjectId: %v", err)
	}
}

func TestNewEvidenceSealedEvent(t *testing.T) {
	data := EvidenceSealedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		StorageRef:    "s3://evidence/abc123",
		SubjectID:     "my-app-v1",
	}

	e, err := NewEvidenceSealedEvent("complytime-worker", "my-app-v1", data)
	if err != nil {
		t.Fatalf("NewEvidenceSealedEvent: %v", err)
	}

	if e.Type() != TypeEvidenceSealed {
		t.Errorf("Type() = %q, want %q", e.Type(), TypeEvidenceSealed)
	}
	if e.Subject() != "my-app-v1" {
		t.Errorf("Subject() = %q, want %q", e.Subject(), "my-app-v1")
	}

	var got EvidenceSealedData
	if err := e.DataAs(&got); err != nil {
		t.Fatalf("DataAs: %v", err)
	}
	if got.SubjectID != data.SubjectID {
		t.Errorf("data.SubjectID = %q, want %q", got.SubjectID, data.SubjectID)
	}
	if got.StorageRef != data.StorageRef {
		t.Errorf("data.StorageRef = %q, want %q", got.StorageRef, data.StorageRef)
	}
}

func TestNewEvidenceSealedEventEmptyStorageRef(t *testing.T) {
	data := EvidenceSealedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
	}
	if _, err := NewEvidenceSealedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for empty storageRef")
	}
}

func TestNewEvidenceSealedEventInvalidStorageRef(t *testing.T) {
	data := EvidenceSealedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		StorageRef:    "evidence-bucket/abc123",
		SubjectID:     "my-app-v1",
	}
	if _, err := NewEvidenceSealedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for storageRef without a URI scheme prefix")
	}
}

func TestNewEvidenceSealedEventInvalidSubjectID(t *testing.T) {
	data := EvidenceSealedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app.v1",
	}
	if _, err := NewEvidenceSealedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for invalid subjectId")
	}
}

func TestNewEvidenceSealedEventEmptyContentDigest(t *testing.T) {
	data := EvidenceSealedData{
		ArtifactType: "application/vnd.gemara.evaluation-log+json",
		SubjectID:    "my-app-v1",
	}
	if _, err := NewEvidenceSealedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for empty contentDigest")
	}
}

func TestNewEvidenceSealedEventEmptyArtifactType(t *testing.T) {
	data := EvidenceSealedData{
		ContentDigest: "sha256:abc123",
		SubjectID:     "my-app-v1",
	}
	if _, err := NewEvidenceSealedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for empty artifactType")
	}
}

func TestNewEvidenceQuarantinedEvent(t *testing.T) {
	data := EvidenceQuarantinedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
		Reason:        "content digest mismatch",
	}

	e, err := NewEvidenceQuarantinedEvent("complytime-worker", "my-app-v1", data)
	if err != nil {
		t.Fatalf("NewEvidenceQuarantinedEvent: %v", err)
	}

	if e.Type() != TypeEvidenceQuarantined {
		t.Errorf("Type() = %q, want %q", e.Type(), TypeEvidenceQuarantined)
	}

	var got EvidenceQuarantinedData
	if err := e.DataAs(&got); err != nil {
		t.Fatalf("DataAs: %v", err)
	}
	if got.Reason != data.Reason {
		t.Errorf("data.Reason = %q, want %q", got.Reason, data.Reason)
	}
}

func TestNewEvidenceQuarantinedEventEmptyReason(t *testing.T) {
	data := EvidenceQuarantinedData{
		ContentDigest: "sha256:abc123",
		ArtifactType:  "application/vnd.gemara.evaluation-log+json",
		SubjectID:     "my-app-v1",
		Reason:        "",
	}
	if _, err := NewEvidenceQuarantinedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for empty reason")
	}
}

func TestNewEvidenceQuarantinedEventEmptyContentDigest(t *testing.T) {
	data := EvidenceQuarantinedData{
		ArtifactType: "application/vnd.gemara.evaluation-log+json",
		SubjectID:    "my-app-v1",
		Reason:       "content digest mismatch",
	}
	if _, err := NewEvidenceQuarantinedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for empty contentDigest")
	}
}

func TestNewEvidenceQuarantinedEventEmptyArtifactType(t *testing.T) {
	data := EvidenceQuarantinedData{
		ContentDigest: "sha256:abc123",
		SubjectID:     "my-app-v1",
		Reason:        "content digest mismatch",
	}
	if _, err := NewEvidenceQuarantinedEvent("complytime-worker", "my-app-v1", data); err == nil {
		t.Error("expected error for empty artifactType")
	}
}

func TestExamplePayloads_ConformToSchema(t *testing.T) {
	schemasDir := filepath.Join("..", "api", "events", "schemas")
	examplesDir := filepath.Join("..", "api", "events", "examples")

	// Register every generated schema with the compiler under a stable
	// in-memory URL. The relative $id in each schema resolves against its
	// registration URL, so the CloudEvent envelope's cross-file $ref to its
	// *Data schema (e.g. "EvidenceIngestedData.schema.json") resolves too.
	const base = "mem:///"
	schemaFiles, err := filepath.Glob(filepath.Join(schemasDir, "*.schema.json"))
	if err != nil {
		t.Fatalf("glob schemas: %v", err)
	}
	if len(schemaFiles) == 0 {
		t.Fatal("no schema files found in api/events/schemas/")
	}
	compiler := jsonschema.NewCompiler()
	for _, path := range schemaFiles {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read schema %s: %v", path, err)
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("parse schema %s: %v", path, err)
		}
		if err := compiler.AddResource(base+filepath.Base(path), doc); err != nil {
			t.Fatalf("add schema %s: %v", path, err)
		}
	}

	// envelopeSchema maps a CloudEvents type to the envelope schema that
	// validates a full example payload (envelope plus data) for that type.
	envelopeSchema := map[string]string{
		TypeEvidenceIngested:    "EvidenceIngestedCloudEvent.schema.json",
		TypeEvidenceSealed:      "EvidenceSealedCloudEvent.schema.json",
		TypeEvidenceQuarantined: "EvidenceQuarantinedCloudEvent.schema.json",
	}

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

			// The CloudEvents type selects which envelope schema applies.
			var envelope struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(b, &envelope); err != nil {
				t.Fatalf("unmarshal type: %v", err)
			}
			schemaFile, ok := envelopeSchema[envelope.Type]
			if !ok {
				t.Fatalf("unhandled example type %q — add it to envelopeSchema", envelope.Type)
			}

			sch, err := compiler.Compile(base + schemaFile)
			if err != nil {
				t.Fatalf("compile %s: %v", schemaFile, err)
			}

			inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
			if err != nil {
				t.Fatalf("parse example %s: %v", path, err)
			}
			if err := sch.Validate(inst); err != nil {
				t.Errorf("%s does not conform to %s:\n%v", filepath.Base(path), schemaFile, err)
			}
		})
	}
}
