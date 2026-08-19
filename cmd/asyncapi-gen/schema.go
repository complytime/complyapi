// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/url"
	"strings"
)

// AsyncAPIDoc is the top-level AsyncAPI 3.0 document model.
type AsyncAPIDoc struct {
	AsyncAPI           string               `yaml:"asyncapi"`
	Info               Info                 `yaml:"info"`
	DefaultContentType string               `yaml:"defaultContentType"`
	Servers            map[string]Server    `yaml:"servers"`
	Channels           map[string]Channel   `yaml:"channels"`
	Operations         map[string]Operation `yaml:"operations"`
	Components         Components           `yaml:"components"`
}

// Info holds document metadata.
type Info struct {
	Title       string   `yaml:"title"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description,omitempty"`
	License     *License `yaml:"license,omitempty"`
	Contact     *Contact `yaml:"contact,omitempty"`
}

// License holds the license information for the AsyncAPI document.
type License struct {
	Name string `yaml:"name"`
}

// Contact holds the contact information for the AsyncAPI document.
type Contact struct {
	Name string `yaml:"name,omitempty"`
	URL  string `yaml:"url,omitempty"`
}

// Server describes a NATS server.
type Server struct {
	Host        string `yaml:"host"`
	Protocol    string `yaml:"protocol"`
	Description string `yaml:"description,omitempty"`
}

// Channel describes a NATS subject channel.
type Channel struct {
	Address     string               `yaml:"address"`
	Description string               `yaml:"description,omitempty"`
	Parameters  map[string]Parameter `yaml:"parameters,omitempty"`
	Messages    map[string]Ref       `yaml:"messages"`
}

// Parameter describes a channel address parameter.
type Parameter struct {
	Description string `yaml:"description"`
}

// Ref is an AsyncAPI $ref object.
type Ref struct {
	Ref string `yaml:"$ref"`
}

// Operation describes a send or receive operation.
type Operation struct {
	Action   string           `yaml:"action"`
	Summary  string           `yaml:"summary"`
	Channel  Ref              `yaml:"channel"`
	Bindings OperationBinding `yaml:"bindings,omitempty"`
}

// OperationBinding holds protocol-specific operation bindings.
type OperationBinding struct {
	NATS NATSOperationBinding `yaml:"nats,omitempty"`
}

// NATSOperationBinding holds NATS JetStream stream metadata.
// Stream is serialised as the AsyncAPI extension field x-stream because the
// official NATS binding 0.1.0 schema does not define a stream property;
// x-prefixed extensions are accepted by the AsyncAPI validator.
type NATSOperationBinding struct {
	Stream         string `yaml:"x-stream,omitempty"`
	BindingVersion string `yaml:"bindingVersion,omitempty"`
}

// Components holds reusable AsyncAPI components.
type Components struct {
	Messages map[string]Message `yaml:"messages"`
	Schemas  map[string]Schema  `yaml:"schemas"`
}

// Message describes an AsyncAPI message.
type Message struct {
	Name        string `yaml:"name"`
	Title       string `yaml:"title"`
	ContentType string `yaml:"contentType"`
	Payload     Ref    `yaml:"payload"`
}

// Schema is a simplified JSON Schema object for AsyncAPI.
type Schema struct {
	Type        string            `yaml:"type,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Required    []string          `yaml:"required,omitempty"`
	Properties  map[string]Schema `yaml:"properties,omitempty"`
	Const       string            `yaml:"const,omitempty"`
	Format      string            `yaml:"format,omitempty"`
	Ref         string            `yaml:"$ref,omitempty"`
}

// BuildDoc constructs an AsyncAPIDoc from the given specs and document metadata.
func BuildDoc(specs []EventSpec, title, version, description, licenseName, contactName, contactURL, serverURL string) AsyncAPIDoc {
	info := Info{Title: title, Version: version}
	if description != "" {
		info.Description = description
	}
	if licenseName != "" {
		info.License = &License{Name: licenseName}
	}
	if contactName != "" || contactURL != "" {
		info.Contact = &Contact{Name: contactName, URL: contactURL}
	}

	doc := AsyncAPIDoc{
		AsyncAPI:           "3.0.0",
		Info:               info,
		DefaultContentType: "application/cloudevents+json",
		Servers:            buildServers(serverURL),
		Channels:           make(map[string]Channel),
		Operations:         make(map[string]Operation),
		Components: Components{
			Messages: make(map[string]Message),
			Schemas:  make(map[string]Schema),
		},
	}

	for _, spec := range specs {
		chKey := channelName(spec.StructName)
		msgKey := messageKey(spec.StructName)
		envSchemaKey := strings.TrimSuffix(spec.StructName, "Data") + "CloudEvent"
		dataSchemaKey := spec.StructName

		// Channel
		params := make(map[string]Parameter)
		for k, v := range spec.Params {
			params[k] = Parameter{Description: v}
		}
		doc.Channels[chKey] = Channel{
			Address:     spec.Channel,
			Description: spec.ChannelDescription,
			Parameters:  params,
			Messages: map[string]Ref{
				msgKey: {Ref: fmt.Sprintf("#/components/messages/%s", msgKey)},
			},
		}

		// Send operation
		sendKey := "publish" + upperFirst(chKey)
		doc.Operations[sendKey] = Operation{
			Action:  "send",
			Summary: spec.SendSummary,
			Channel: Ref{Ref: fmt.Sprintf("#/channels/%s", chKey)},
			Bindings: OperationBinding{
				NATS: NATSOperationBinding{Stream: spec.Stream, BindingVersion: "0.1.0"},
			},
		}

		// Receive operation
		recvKey := "consume" + upperFirst(chKey)
		doc.Operations[recvKey] = Operation{
			Action:  "receive",
			Summary: spec.RecvSummary,
			Channel: Ref{Ref: fmt.Sprintf("#/channels/%s", chKey)},
		}

		// Message
		doc.Components.Messages[msgKey] = Message{
			Name:        msgKey,
			Title:       humanTitle(spec.StructName),
			ContentType: "application/cloudevents+json",
			Payload:     Ref{Ref: fmt.Sprintf("#/components/schemas/%s", envSchemaKey)},
		}

		// CloudEvents envelope schema
		doc.Components.Schemas[envSchemaKey] = buildEnvelopeSchema(spec)

		// Data schema
		doc.Components.Schemas[dataSchemaKey] = buildDataSchema(spec)
	}

	return doc
}

// buildServers parses the server URL and returns the servers map.
func buildServers(rawURL string) map[string]Server {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return map[string]Server{"nats": {Host: rawURL, Protocol: "nats"}}
	}
	return map[string]Server{
		"nats": {
			Host:     u.Host,
			Protocol: u.Scheme,
		},
	}
}

// buildEnvelopeSchema returns the CloudEvents envelope schema for a spec.
func buildEnvelopeSchema(spec EventSpec) Schema {
	return Schema{
		Type:        "object",
		Description: fmt.Sprintf("CloudEvents v1.0 envelope for %s", spec.CEType),
		Required:    []string{"specversion", "id", "type", "source", "subject", "time", "datacontenttype", "data"},
		Properties: map[string]Schema{
			"specversion":     {Type: "string", Const: "1.0"},
			"id":              {Type: "string", Format: "uuid"},
			"type":            {Type: "string", Const: spec.CEType},
			"source":          {Type: "string", Description: "URI identifying the producing service"},
			"subject":         {Type: "string", Description: "The compliance subject identifier"},
			"time":            {Type: "string", Format: "date-time"},
			"datacontenttype": {Type: "string", Const: "application/json"},
			"data":            {Ref: fmt.Sprintf("#/components/schemas/%s", spec.StructName)},
		},
	}
}

// buildDataSchema builds the data payload schema from struct fields.
func buildDataSchema(spec EventSpec) Schema {
	props := make(map[string]Schema)
	var required []string

	for _, f := range spec.Fields {
		s := Schema{Type: goTypeToJSONSchema(f.GoType)}
		if f.Description != "" {
			s.Description = f.Description
		}
		props[f.JSONName] = s
		if f.Required {
			required = append(required, f.JSONName)
		}
	}

	s := Schema{
		Type:       "object",
		Required:   required,
		Properties: props,
	}
	if spec.DocComment != "" {
		s.Description = spec.DocComment
	}
	return s
}

// goTypeToJSONSchema maps Go type strings to JSON Schema type strings.
func goTypeToJSONSchema(goType string) string {
	base := strings.TrimPrefix(goType, "*")
	switch base {
	case "string":
		return "string"
	case "int", "int32", "int64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "object"
	}
}

// channelName converts a struct name like "EvidenceIngestedData" to "evidenceIngested".
func channelName(structName string) string {
	name := strings.TrimSuffix(structName, "Data")
	if len(name) == 0 {
		return structName
	}
	return strings.ToLower(name[:1]) + name[1:]
}

// messageKey converts a struct name like "EvidenceIngestedData" to "EvidenceIngested".
func messageKey(structName string) string {
	return strings.TrimSuffix(structName, "Data")
}

// humanTitle converts a struct name like "EvidenceIngestedData" to "Evidence Ingested".
func humanTitle(structName string) string {
	name := strings.TrimSuffix(structName, "Data")
	var parts []string
	start := 0
	for i := 1; i < len(name); i++ {
		if name[i] >= 'A' && name[i] <= 'Z' {
			parts = append(parts, name[start:i])
			start = i
		}
	}
	parts = append(parts, name[start:])
	return strings.Join(parts, " ")
}

// upperFirst uppercases the first letter of s.
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
