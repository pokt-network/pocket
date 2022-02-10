package consensus

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"

	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/types"
	"pocket/shared/context"
)

func (m *consensusModule) HandleTelemetryMessage(ctx *context.PocketContext, networkConnection net.Conn) {
	defer networkConnection.Close()

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)

	if err := enc.Encode(m.consensusModuleState()); err != nil {
		log.Println("Error during encoding:", err)
	}

	networkConnection.Write(buff.Bytes())
}

func (m *consensusModule) GetNodeState() consensus_types.ConsensusNodeState {
	return m.consensusModuleState()
}

func (m *consensusModule) consensusModuleState() consensus_types.ConsensusNodeState {
	leaderId := types.NodeId(0)
	if m.LeaderId != nil {
		leaderId = *m.LeaderId
	}
	return consensus_types.ConsensusNodeState{
		NodeId:   types.NodeId(m.NodeId),
		Height:   uint64(m.Height),
		Round:    uint8(m.Round),
		Step:     uint8(m.Step),
		IsLeader: m.isLeader(),
		LeaderId: types.NodeId(leaderId),
	}
}
