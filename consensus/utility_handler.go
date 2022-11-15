package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

func (m *consensusModule) handleUtilityMessage(msg *typesCons.UtilityMessage) error {
	switch msg.GetType() {
	case typesCons.UtilityMessageType_UTILITY_MESSAGE_TRANSACTION:
		if m.utilityContext == nil {
			m.refreshUtilityContext()

		}
		if err := m.utilityContext.CheckTransaction(msg.GetData()); err != nil {
			return err
		} else {
			m.nodeLog("Successfully checked transaction")
		}
	default:
		panic("unknown utility message type")
	}
	return nil
}
