package messaging

import "strings"

const (
	DebugMessageEventType = "pocket.DebugMessage"

	AddressBookAtHeightEventType = "pocket.AddressBookAtHeight"
)

func (envelope *PocketEnvelope) GetContentType() string {
	return strings.Split(envelope.Content.GetTypeUrl(), "/")[1]
}
