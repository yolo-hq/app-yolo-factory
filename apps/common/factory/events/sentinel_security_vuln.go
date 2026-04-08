package events

import "github.com/yolo-hq/yolo/core/event"

// SentinelSecurityVulnEvent is emitted when sentinel detects a security vulnerability.
type SentinelSecurityVulnEvent struct {
	event.CustomEvent[SentinelPayload]
}
