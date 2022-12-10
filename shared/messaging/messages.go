package messaging

import "strings"

const (
	DebugMessageEventType = "pocket.DebugMessage"
)

func (envelope *PocketEnvelope) GetContentType() string {
	return strings.Split(envelope.Content.GetTypeUrl(), "/")[1]
}
