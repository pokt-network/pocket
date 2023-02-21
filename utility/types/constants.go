package types

import "math/big"

// CLEANUP: Consider moving these into a shared location or eliminating altogether
const (
	ZeroInt = 0
	// IMPROVE: -1 is returned when retrieving the paused height of an unpaused actor. Consider a more user friendly and semantic way of managing this.
	HeightNotUsed = int64(-1)
	EmptyString   = ""
)

var (
	// TECHDEBT: Re-evalute the denomination of tokens used throughout the codebase. `POKTDenomination` is
	// currently used to convert POKT to uPOKT but this is not clear throughout the codebase.
	POKTDenomination = big.NewFloat(1e6)
)
