package sortition

import (
	"crypto/rand"
	"pocket/consensus/leader_election/vrf"
	"testing"

	"github.com/stretchr/testify/assert"
)

// make test_sortition
func TestSortitionBasic(t *testing.T) {
	minValidatorStake := float64(15000)
	numValidators := float64(1000)
	totalVotingPower := minValidatorStake * numValidators
	maxStakePerOneValidator := float64(0.1) * totalVotingPower // 10% of total voting power
	errThreshold := 0.05                                       // 5% error threshold. TODO: is this too high?
	numCandidatesLeadersPerRound := float64(3)

	for validatorVotingPower := minValidatorStake; validatorVotingPower < maxStakePerOneValidator; validatorVotingPower += minValidatorStake {
		numViewChanges := float64(1000)
		// errorTolerance :=

		hitCount := SortitionResult(0)
		for i := float64(0); i < numViewChanges; i++ {
			var vrfOutput [vrf.VRFOutputSize]byte
			rand.Read(vrfOutput[:])

			sortitionResult := Sortition(validatorVotingPower, totalVotingPower, numCandidatesLeadersPerRound, vrfOutput[:])
			hitCount += sortitionResult
		}

		// NOTE: Originally `errTolerance` was set to `float64(expectedSelections) * errThreshold`, but this failed
		// for very small values. E.g.: "Max difference between 4 and 3 allowed is 0.2, but difference was 1"
		errTolerance := numViewChanges * errThreshold
		expectedSelections := uint64(numViewChanges * numCandidatesLeadersPerRound * validatorVotingPower / totalVotingPower)
		assert.InDelta(t, expectedSelections, uint64(hitCount), errTolerance)
	}
}

// make benchmark_sortition
func BenchmarkSortition(b *testing.B) {
	b.StopTimer()

	vrfOutputs := make([]vrf.VRFOutput, b.N)
	for i := 0; i < b.N; i++ {
		rand.Read(vrfOutputs[i][:])
	}

	b.StartTimer()
	validatorVotingPower := float64(1000000)
	totalVotingPower := float64(1000000000000)
	numCandidatesLeadersPerRound := float64(3)
	for i := 0; i < b.N; i++ {
		Sortition(validatorVotingPower, totalVotingPower, numCandidatesLeadersPerRound, vrfOutputs[i])
	}
}
