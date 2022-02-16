package consensus

// The Pocket Network block height.
type BlockHeight uint64 // TODO: Move this into `consensus_types`.

// The number of times the node was interrupted at the current height; always 0 in the "happy path".
type Round uint8 // TODO: Move this into `consensus_types`.

// Smallest logical unit in a single round; the`type` in the Hotstuff whitepaper.
type Step uint8

const (
	NewRound Step = iota
	Prepare
	PreCommit
	Commit
	Decide
)

var HotstuffSteps = [...]Step{NewRound, Prepare, PreCommit, Commit, Decide}

var StepToString = map[Step]string{
	NewRound:  "NEW_ROUND",
	Prepare:   "PREPARE",
	PreCommit: "PRECOMMIT",
	Commit:    "COMMIT",
	Decide:    "DECIDE",
}
