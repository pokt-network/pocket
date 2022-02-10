package types

import (
	"pocket/consensus/pkg/types"

	"github.com/mindstand/gogm/v2"
)

type ConsensusNodeState struct {
	NodeId   types.NodeId `gogm:"name=NodeId"`
	Height   uint64       `gogm:"name=Height"` // TODO: Change to proper type
	Round    uint8        `gogm:"name=Round"`  // TODO: Change to proper type
	Step     uint8        `gogm:"name=Step"`   // TODO: Change to proper type
	IsLeader bool         `gogm:"name=IsLeader"`
	LeaderId types.NodeId `gogm:"name=Leader"`

	gogm.BaseNode // Provides required node fields for neo4j DB
}
