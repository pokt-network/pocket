package p2p

import (
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *p2pModule) HandleEvent(event *anypb.Any) error {
	// no-op (for now... PRs are already cooked)
	return nil
}
