// SPDX-License-Identifier: Apache-2.0

package testdata

// WidgetCreatedData is the payload for widget.created events.
type WidgetCreatedData struct {
	_ struct{} `asyncapi:"channel:core.widget.created.{ownerId},param:ownerId=The widget owner identifier,stream:WIDGETS,type:dev.example.widget.created,send:Published when a widget is created,receive:Consume widget-created events"`

	WidgetID string  `json:"widgetId"`
	Name     string  `json:"name"`
	Tag      string  `json:"tag,omitempty"`
	ParentID *string `json:"parentId,omitempty"`
}
