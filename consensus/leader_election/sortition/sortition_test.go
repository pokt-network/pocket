package sortition

import (
	"crypto/rand"
	"testing"

	"github.com/pokt-network/pocket/consensus/leader_election/vrf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	uPOKTMinValidatorStake = uint64(15000) * 1e6 // TODO(discuss): Will we have a codebase wide constant for this?
	// Since leader election is non-deterministic, we allow for a small error threshold
	// even in the unit tests for the chances of a certain validator to be elected as a leader.
	// 5% error threshold.
	errThreshold = 0.05 // NOTE: The 5% value is somewhat arbitrary.
)

// This test iterates of different network configurations but calls `SingleSortitionTest` under
// the hood.
func TestSortition(t *testing.T) {
	// The number of validators in the network
	numValidators := uint64(1000) // NOTE: This value is somewhat arbitrary.
	uPOKTNetworkStake := uPOKTMinValidatorStake * numValidators

	testParameters := []struct {
		uPOKTValidatorStake, uPOKTTotalStaked, numViewChanges, numCandidates uint64
	}{
		{uPOKTMinValidatorStake, uPOKTNetworkStake, 100, 1},
		{uPOKTMinValidatorStake, uPOKTNetworkStake, 1000, 1},
		{uPOKTMinValidatorStake, uPOKTNetworkStake, 10000, 1},
		{uPOKTMinValidatorStake * 5, uPOKTNetworkStake, 100, 3},
		{uPOKTMinValidatorStake * 10, uPOKTNetworkStake, 1000, 3},
		{uPOKTMinValidatorStake * 100, uPOKTNetworkStake, 10000, 3},
		{uPOKTMinValidatorStake * 5, uPOKTNetworkStake, 100, 10},
		{uPOKTMinValidatorStake * 10, uPOKTNetworkStake, 1000, 10},
		{uPOKTMinValidatorStake * 100, uPOKTNetworkStake, 10000, 10},
	}

	for _, test := range testParameters {
		SingleSortitionTest(t, test.uPOKTValidatorStake, test.uPOKTTotalStaked, test.numViewChanges, test.numCandidates)
	}
}

// The changes of a validator getting selected is non-deterministic, but is also uniformally distributed
// and proportional to their stake. As such, we iterate over a large number of view changes for a
// provided network configuration, and check if the validator was selected some number of times,
// within a predefined error threshold.
// EXAMPLE: If a single validator has 10% of the total stake in the network, and there are 1000 view
// changes, we expect that validator to be selected as a leader 100±5 times.
func SingleSortitionTest(t *testing.T, uPOKTValidatorStake, uPOKTNetworkStake, numViewChanges, numCandidates uint64) {
	selectCount := SortitionResult(0)
	for i := uint64(0); i < numViewChanges; i++ {
		var vrfOutput [vrf.VRFOutputSize]byte
		_, err := rand.Read(vrfOutput[:])
		require.NoError(t, err)

		sortitionResult := Sortition(uPOKTValidatorStake, uPOKTNetworkStake, numCandidates, vrfOutput[:])
		selectCount += sortitionResult
	}

	errTolerance := float64(numViewChanges) * errThreshold
	expectedSelections := uint64(numViewChanges * numCandidates * uPOKTValidatorStake / uPOKTNetworkStake)
	assert.InDelta(t, expectedSelections, uint64(selectCount), errTolerance)
}

func BenchmarkSortition(b *testing.B) {
	b.StopTimer()

	vrfOutputs := make([]vrf.VRFOutput, b.N)
	for i := 0; i < b.N; i++ {
		_, err := rand.Read(vrfOutputs[i][:])
		require.NoError(b, err)
	}

	b.StartTimer()
	uPOKTValidatorStake := uint64(1000000)
	uPOKTNetworkStake := uint64(1000000000000)
	numCandidatesLeadersPerRound := uint64(3)
	for i := 0; i < b.N; i++ {
		Sortition(uPOKTValidatorStake, uPOKTNetworkStake, numCandidatesLeadersPerRound, vrfOutputs[i])
	}
}
