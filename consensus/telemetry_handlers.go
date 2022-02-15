package consensus

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"
	consensus_types "pocket/consensus/types"
)

func (m *ConsensusModule) HandleTelemetryMessage(networkConnection net.Conn) {
	defer networkConnection.Close()

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)

	if err := enc.Encode(m.consensusModuleState()); err != nil {
		log.Println("Error during encoding:", err)
	}

	networkConnection.Write(buff.Bytes())
}

func (m *ConsensusModule) GetNodeState() consensus_types.ConsensusNodeState {
	return m.consensusModuleState()
}

func (m *ConsensusModule) consensusModuleState() consensus_types.ConsensusNodeState {
	leaderId := consensus_types.NodeId(0)
	if m.LeaderId != nil {
		leaderId = *m.LeaderId
	}
	return consensus_types.ConsensusNodeState{
		NodeId:   m.NodeId,
		Height:   uint64(m.Height),
		Round:    uint8(m.Round),
		Step:     uint8(m.Step),
		IsLeader: m.isLeader(),
		LeaderId: leaderId,
	}
}
