package consensus

import (
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) handleTransaction(anyMessage *anypb.Any) error {
	module := m.GetBus().GetUtilityModule()
	m.utilityContext, _ = module.NewContext(int64(m.Height))
	// TODO(olshansky): decode data, basic validation, send to utility module.
	// if err := m.utilityContext.CheckTransaction(messageProto.Data); err != nil {
	// 	m.nodeLogError(err.Error())
	// }
	// fmt.Println("TRANSACTION IS CHECKED")
	m.utilityContext.ReleaseContext()
	return nil
}
