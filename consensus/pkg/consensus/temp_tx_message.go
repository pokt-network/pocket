package consensus

import (
	"pocket/consensus/pkg/consensus/types"
	"pocket/shared"
)

type TxWrapperMessage struct {
	types.GenericConsensusMessage

	Data []byte
}

func (m TxWrapperMessage) GetType() types.ConsensusMessageType {
	return types.TxWrapperMessageType
}

func (m *TxWrapperMessage) Encode() ([]byte, error) {
	bytes, err := shared.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *TxWrapperMessage) Decode(data []byte) error {
	err := shared.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
