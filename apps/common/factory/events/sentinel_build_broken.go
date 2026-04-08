package events

import "github.com/yolo-hq/yolo/core/event"

// SentinelBuildBrokenEvent is emitted when sentinel detects a broken build.
type SentinelBuildBrokenEvent struct {
	event.CustomEvent[SentinelPayload]
}

// SentinelPayload carries sentinel finding data.
type SentinelPayload struct {
	ProjectID string `entity:"Project"`
	Error     string `json:"error"`
	Severity  string `json:"severity"`
}
