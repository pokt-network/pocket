package consensus

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

func (m *consensusModule) HandleTelemetryMessage(networkConnection net.Conn) {
	defer networkConnection.Close()

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)

	if err := enc.Encode(m.consensusModuleState()); err != nil {
		log.Println("Error during encoding:", err)
	}

	networkConnection.Write(buff.Bytes())
}

func (m *consensusModule) GetNodeState() types_consensus.ConsensusNodeState {
	return m.consensusModuleState()
}

func (m *consensusModule) consensusModuleState() types_consensus.ConsensusNodeState {
	leaderId := types_consensus.NodeId(0)
	if m.LeaderId != nil {
		leaderId = *m.LeaderId
	}
	return types_consensus.ConsensusNodeState{
		NodeId:   m.NodeId,
		Height:   uint64(m.Height),
		Round:    uint8(m.Round),
		Step:     uint8(m.Step),
		IsLeader: m.isLeader(),
		LeaderId: leaderId,
	}
}
