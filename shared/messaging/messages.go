package messaging

import "strings"

const (
	DebugMessageEventType = "pocket.DebugMessage"
)

func (x *PocketEnvelope) GetContentType() string {
	return strings.Split(x.Content.GetTypeUrl(), "/")[1]
}
