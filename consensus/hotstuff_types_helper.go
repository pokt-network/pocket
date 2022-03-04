package consensus

import (
	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

const (
	NewRound  types_consensus.HotstuffStep = types_consensus.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	Prepare   types_consensus.HotstuffStep = types_consensus.HotstuffStep_HOTSTUFF_STEP_PREPARE
	PreCommit types_consensus.HotstuffStep = types_consensus.HotstuffStep_HOTSTUFF_STEP_PRECOMMIT
	Commit    types_consensus.HotstuffStep = types_consensus.HotstuffStep_HOTSTUFF_STEP_COMMIT
	Decide    types_consensus.HotstuffStep = types_consensus.HotstuffStep_HOTSTUFF_STEP_DECIDE
)

var HotstuffSteps = [...]types_consensus.HotstuffStep{NewRound, Prepare, PreCommit, Commit, Decide}

var StepToString map[types_consensus.HotstuffStep]string

func init() {
	StepToString = make(map[types_consensus.HotstuffStep]string, len(types_consensus.HotstuffStep_name))
	for i, step := range types_consensus.HotstuffStep_name {
		StepToString[types_consensus.HotstuffStep(i)] = step
	}
}
