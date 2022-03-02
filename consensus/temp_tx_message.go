package consensus

import (
	"github.com/pokt-network/pocket/consensus/types"
)

type TxWrapperMessage struct {
	types.GenericConsensusMessage

	Data []byte
}

func (m TxWrapperMessage) GetType() types.ConsensusMessageType {
	return types.TxWrapperMessageType
}

func (m *TxWrapperMessage) Encode() ([]byte, error) {
	bytes, err := types.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *TxWrapperMessage) Decode(data []byte) error {
	err := types.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
