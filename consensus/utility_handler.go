package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
)

func (m *consensusModule) handleUtilityMessage(msg *typesCons.UtilityMessage) error {
	switch msg.GetType() {
	case typesCons.UtilityMessageType_UTILITY_MESSAGE_TRANSACTION:
		if err := m.GetBus().GetUtilityModule().CheckTransaction(msg.GetData()); err != nil {
			return err
		} else {
			m.nodeLog("Successfully checked transaction")
		}
	default:
		return fmt.Errorf("unknown utility message type: %v", msg.GetType())
	}
	return nil
}
