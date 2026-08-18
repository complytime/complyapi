// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
)

func singleSpec() EventSpec {
	return EventSpec{
		StructName:  "WidgetCreatedData",
		Channel:     "core.widget.created.{ownerId}",
		Params:      map[string]string{"ownerId": "The widget owner identifier"},
		Stream:      "WIDGETS",
		CEType:      "dev.example.widget.created",
		SendSummary: "Published when a widget is created",
		RecvSummary: "Consume widget-created events",
		Fields: []FieldSpec{
			{JSONName: "widgetId", GoType: "string", Required: true, Description: "Unique widget identifier"},
			{JSONName: "name", GoType: "string", Required: true},
			{JSONName: "tag", GoType: "string", Required: false},
			{JSONName: "parentId", GoType: "*string", Required: false, Description: "Parent widget ID"},
		},
	}
}

func TestBuildDoc_InfoFields(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	if doc.AsyncAPI != "3.0.0" {
		t.Errorf("AsyncAPI = %q, want %q", doc.AsyncAPI, "3.0.0")
	}
	if doc.Info.Title != "Test API" {
		t.Errorf("Title = %q, want %q", doc.Info.Title, "Test API")
	}
	if doc.Info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", doc.Info.Version, "1.0.0")
	}
	if doc.DefaultContentType != "application/cloudevents+json" {
		t.Errorf("DefaultContentType = %q, want %q", doc.DefaultContentType, "application/cloudevents+json")
	}
}

func TestBuildDoc_ServerURL(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	if len(doc.Servers) != 1 {
		t.Fatalf("len(Servers) = %d, want 1", len(doc.Servers))
	}
	srv := doc.Servers["nats"]
	if srv.Host != "localhost:4222" {
		t.Errorf("Host = %q, want %q", srv.Host, "localhost:4222")
	}
	if srv.Protocol != "nats" {
		t.Errorf("Protocol = %q, want %q", srv.Protocol, "nats")
	}
}

func TestBuildDoc_Channel(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	ch, ok := doc.Channels["widgetCreated"]
	if !ok {
		t.Fatal("channel widgetCreated not found")
	}
	if ch.Address != "core.widget.created.{ownerId}" {
		t.Errorf("Address = %q, want %q", ch.Address, "core.widget.created.{ownerId}")
	}
	param, ok := ch.Parameters["ownerId"]
	if !ok {
		t.Fatal("parameter ownerId not found")
	}
	if param.Description != "The widget owner identifier" {
		t.Errorf("param description = %q, want %q", param.Description, "The widget owner identifier")
	}
}

func TestBuildDoc_Operations(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	sendOp, ok := doc.Operations["publishWidgetCreated"]
	if !ok {
		t.Fatal("operation publishWidgetCreated not found")
	}
	if sendOp.Action != "send" {
		t.Errorf("send action = %q, want %q", sendOp.Action, "send")
	}
	if sendOp.Summary != "Published when a widget is created" {
		t.Errorf("send summary = %q, want %q", sendOp.Summary, "Published when a widget is created")
	}

	recvOp, ok := doc.Operations["consumeWidgetCreated"]
	if !ok {
		t.Fatal("operation consumeWidgetCreated not found")
	}
	if recvOp.Action != "receive" {
		t.Errorf("receive action = %q, want %q", recvOp.Action, "receive")
	}
	if recvOp.Bindings.NATS.Stream != "" {
		t.Errorf("receive op NATS stream = %q, want empty (no binding)", recvOp.Bindings.NATS.Stream)
	}
}

func TestBuildDoc_DataSchemaFields(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	dataSchema, ok := doc.Components.Schemas["WidgetCreatedData"]
	if !ok {
		t.Fatal("schema WidgetCreatedData not found")
	}

	widgetIDProp, ok := dataSchema.Properties["widgetId"]
	if !ok {
		t.Fatal("property widgetId not found")
	}
	if widgetIDProp.Type != "string" {
		t.Errorf("widgetId type = %q, want %q", widgetIDProp.Type, "string")
	}
	if widgetIDProp.Description != "Unique widget identifier" {
		t.Errorf("widgetId description = %q, want %q", widgetIDProp.Description, "Unique widget identifier")
	}

	// name has no description tag — should be empty
	nameProp, ok := dataSchema.Properties["name"]
	if !ok {
		t.Fatal("property name not found")
	}
	if nameProp.Description != "" {
		t.Errorf("name description = %q, want empty", nameProp.Description)
	}

	// widgetId and name must be in required list
	required := map[string]bool{}
	for _, r := range dataSchema.Required {
		required[r] = true
	}
	if !required["widgetId"] {
		t.Error("widgetId should be required")
	}
	if !required["name"] {
		t.Error("name should be required")
	}
	if required["tag"] {
		t.Error("tag should not be required")
	}
	if required["parentId"] {
		t.Error("parentId should not be required")
	}
}

func TestBuildDoc_CloudEventsEnvelope(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	env, ok := doc.Components.Schemas["WidgetCreatedCloudEvent"]
	if !ok {
		t.Fatal("schema WidgetCreatedCloudEvent not found")
	}

	specversion, ok := env.Properties["specversion"]
	if !ok {
		t.Fatal("specversion property not found")
	}
	if specversion.Const != "1.0" {
		t.Errorf("specversion const = %q, want %q", specversion.Const, "1.0")
	}

	ceType, ok := env.Properties["type"]
	if !ok {
		t.Fatal("type property not found")
	}
	if ceType.Const != "dev.example.widget.created" {
		t.Errorf("type const = %q, want %q", ceType.Const, "dev.example.widget.created")
	}
}

func TestBuildDoc_NATSBinding(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	op, ok := doc.Operations["publishWidgetCreated"]
	if !ok {
		t.Fatal("operation publishWidgetCreated not found")
	}
	if op.Bindings.NATS.Stream != "WIDGETS" {
		t.Errorf("NATS stream = %q, want %q", op.Bindings.NATS.Stream, "WIDGETS")
	}
}
