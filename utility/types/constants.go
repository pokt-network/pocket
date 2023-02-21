package types

// CLEANUP: Consider moving these into a shared location or eliminating altogether
const (
	ZeroInt = 0
	// IMPROVE: -1 is returned when retrieving the paused height of an unpaused actor. Consider a more user friendly and semantic way of managing this.
	HeightNotUsed = int64(-1)
	EmptyString   = ""
)
