package consensus

import (
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) handleTransaction(anyMessage *anypb.Any) error {
	module := m.GetBus().GetUtilityModule()
	m.utilityContext, _ = module.NewContext(int64(m.Height))
	// TODO(olshansky): integrate with utility
	// if err := m.utilityContext.CheckTransaction(messageProto.Data); err != nil {
	// 	m.nodeLogError(err.Error())
	// }
	m.utilityContext.ReleaseContext()
	return nil
}
