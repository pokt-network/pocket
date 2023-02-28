package types

// CLEANUP: Consider moving these into a shared location or seeing if they can be eliminated altogether
const (
	EmptyString = ""
	// IMPROVE: -1 is used for defining unused heights (e.g. unpaused actor has pausedHeight=-1).
	// Consider a more user friendly and semantic way of managing this.
	HeightNotUsed = int64(-1)
)
