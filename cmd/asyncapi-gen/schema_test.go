// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
)

func singleSpec() EventSpec {
	return EventSpec{
		StructName:         "WidgetCreatedData",
		Channel:            "core.widget.created.{ownerId}",
		Params:             map[string]string{"ownerId": "The widget owner identifier"},
		Stream:             "WIDGETS",
		CEType:             "dev.example.widget.created",
		SendSummary:        "Published when a widget is created",
		RecvSummary:        "Consume widget-created events",
		ChannelDescription: "Widget creation pipeline",
		DocComment:         "WidgetCreatedData is the payload for widget.created events.",
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

func TestBuildDoc_ChannelDescription(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	ch, ok := doc.Channels["widgetCreated"]
	if !ok {
		t.Fatal("channel widgetCreated not found")
	}
	if ch.Description != "Widget creation pipeline" {
		t.Errorf("channel description = %q, want %q", ch.Description, "Widget creation pipeline")
	}
}

func TestBuildDoc_ChannelDescription_Empty(t *testing.T) {
	spec := singleSpec()
	spec.ChannelDescription = ""
	doc := BuildDoc([]EventSpec{spec}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	ch := doc.Channels["widgetCreated"]
	if ch.Description != "" {
		t.Errorf("channel description = %q, want empty", ch.Description)
	}
}

func TestBuildDoc_DataSchemaDescription(t *testing.T) {
	doc := BuildDoc([]EventSpec{singleSpec()}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	dataSchema, ok := doc.Components.Schemas["WidgetCreatedData"]
	if !ok {
		t.Fatal("schema WidgetCreatedData not found")
	}
	if dataSchema.Description != "WidgetCreatedData is the payload for widget.created events." {
		t.Errorf("data schema description = %q, want %q", dataSchema.Description, "WidgetCreatedData is the payload for widget.created events.")
	}
}

func TestBuildDoc_DataSchemaDescription_Empty(t *testing.T) {
	spec := singleSpec()
	spec.DocComment = ""
	doc := BuildDoc([]EventSpec{spec}, "Test API", "1.0.0", "", "", "", "", "nats://localhost:4222")

	dataSchema := doc.Components.Schemas["WidgetCreatedData"]
	if dataSchema.Description != "" {
		t.Errorf("data schema description = %q, want empty", dataSchema.Description)
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

func TestGoTypeToJSONSchema(t *testing.T) {
	tests := []struct {
		goType string
		want   string
	}{
		{"string", "string"},
		{"*string", "string"},
		{"int", "integer"},
		{"int32", "integer"},
		{"int64", "integer"},
		{"float32", "number"},
		{"float64", "number"},
		{"bool", "boolean"},
		{"SomeStruct", "object"},
		{"*int", "integer"},
		{"*bool", "boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			got := goTypeToJSONSchema(tt.goType)
			if got != tt.want {
				t.Errorf("goTypeToJSONSchema(%q) = %q, want %q", tt.goType, got, tt.want)
			}
		})
	}
}

func TestBuildServers_ValidURL(t *testing.T) {
	servers := buildServers("nats://myhost:4222")
	srv, ok := servers["nats"]
	if !ok {
		t.Fatal("nats server not found")
	}
	if srv.Host != "myhost:4222" {
		t.Errorf("Host = %q, want %q", srv.Host, "myhost:4222")
	}
	if srv.Protocol != "nats" {
		t.Errorf("Protocol = %q, want %q", srv.Protocol, "nats")
	}
}

func TestBuildServers_MalformedURL(t *testing.T) {
	// URL without scheme — Host will be empty, falls back to raw string
	servers := buildServers("host:4222")
	srv, ok := servers["nats"]
	if !ok {
		t.Fatal("nats server not found")
	}
	// Fallback: raw URL used as host
	if srv.Host != "host:4222" {
		t.Errorf("Host = %q, want %q", srv.Host, "host:4222")
	}
	if srv.Protocol != "nats" {
		t.Errorf("Protocol = %q, want %q", srv.Protocol, "nats")
	}
}

func TestBuildServers_EmptyURL(t *testing.T) {
	servers := buildServers("")
	srv, ok := servers["nats"]
	if !ok {
		t.Fatal("nats server not found")
	}
	if srv.Host != "" {
		t.Errorf("Host = %q, want empty", srv.Host)
	}
	if srv.Protocol != "nats" {
		t.Errorf("Protocol = %q, want %q", srv.Protocol, "nats")
	}
}

func TestChannelName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"EvidenceIngestedData", "evidenceIngested"},
		{"WidgetCreatedData", "widgetCreated"},
		{"Data", "Data"},
		// edge: name becomes empty after trim, returns original
		{"SimpleData", "simple"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := channelName(tt.in)
			if got != tt.want {
				t.Errorf("channelName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestHumanTitle(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"EvidenceIngestedData", "Evidence Ingested"},
		{"WidgetCreatedData", "Widget Created"},
		{"SimpleData", "Simple"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := humanTitle(tt.in)
			if got != tt.want {
				t.Errorf("humanTitle(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestUpperFirst(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"", ""},
		{"a", "A"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := upperFirst(tt.in)
			if got != tt.want {
				t.Errorf("upperFirst(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
