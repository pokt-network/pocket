package leader_election

import (
	"math/rand"
	"strconv"
	"testing"

	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test_leader_election
func TestLeaderElection(t *testing.T) {
	// Prepare configurations.
	testValidators := make([]*TestValidatorConfigs, 10)
	for i := uint(0); i < uint(len(testValidators)); i++ {
		testValidators[i] = &TestValidatorConfigs{
			NodeId: i + 1, // NodeId = 0 is invalid.
			UPokt:  uint(rand.Intn(10) * 1e6),
		}
	}
	valMap, totalVotingPower := prepareTestValidators(t, testValidators)
	numViewChanges := 1000
	numCandidatesLeadersPerRound := float64(3)

	// Run leader election for many different block heights.
	candidateCounter := make(map[types.NodeId]uint64, len(testValidators))
	numViewsNoLeader := 0

	mapNumCandidates := make(map[uint64]uint64)
	for h := 0; h < numViewChanges; h++ {
		height := consensus_types.BlockHeight(h)
		round := consensus_types.Round(rand.Intn(10))
		prevBlockHash := strconv.Itoa(rand.Intn(1e10)) // TODO: Create a general purpose utility for random strings?

		candidates := make([]*LeaderCandidate, 0)
		for _, v := range valMap {
			leaderCandidate, err := IsLeaderCandidate(
				v.validator,
				height,
				round,
				prevBlockHash,
				float64(v.validator.UPokt),
				float64(totalVotingPower),
				numCandidatesLeadersPerRound,
				v.secretKey,
			)
			require.NoError(t, err)

			if leaderCandidate != nil {
				candidateCounter[v.validator.NodeId]++
				candidates = append(candidates, leaderCandidate)
			}
		}
		// Guarantee that a leader was selected.
		leaderId, err := ElectLeader(candidates, height, round, prevBlockHash)

		mapNumCandidates[uint64(len(candidates))]++

		// If the error is nil, a leader with a non-zero ID must be elected.
		if err != nil {
			numViewsNoLeader++
			continue
		}

		assert.Greater(t, int(leaderId), 0)
		_, ok := valMap[leaderId]
		assert.True(t, ok)
	}

	errThreshold := 0.07 // 7% error threshold. TODO: is this too high? Emperically determined to pass almost all the time to avoid flaky tests.

	// Validate that each validator was elected as a candidate the expected number of times based on their stake.
	errTolerance := numCandidatesLeadersPerRound * float64(numViewChanges) * errThreshold
	for nodeId, numTimesCandidate := range candidateCounter {
		validatorStakeFraction := float64(valMap[nodeId].validator.UPokt) / float64(totalVotingPower)
		expected := numCandidatesLeadersPerRound * float64(numViewChanges) * validatorStakeFraction
		assert.InDelta(t, expected, numTimesCandidate, errTolerance)
	}

	// Validate that the number of times no leader was elected via the VRF/CDF (default)
	expected := float64(numViewChanges) * errThreshold
	assert.InDelta(t, expected, numViewsNoLeader, errTolerance, "TODO: Investigate why this is so high.")

	/* Useful for debugging and understanding the data */
	// fmt.Println("num view changes: ", numViewChanges)
	// fmt.Println("num_candidates:count: ", mapNumCandidates)
	// for nodeId, numTimesCandidate := range candidateCounter {
	// 	fmt.Println("node_id: ", nodeId, " num_times_candidate: ", numTimesCandidate, "relative stake: ", (float64(valMap[nodeId].validator.UPokt) / float64(totalVotingPower)))
	// }
}
