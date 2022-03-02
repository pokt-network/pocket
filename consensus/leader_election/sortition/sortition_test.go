package sortition

import (
	"crypto/rand"
	"testing"

	"github.com/pokt-network/pocket/consensus/leader_election/vrf"

	"github.com/stretchr/testify/assert"
)

func TestSortitionBasic(t *testing.T) {
	minValidatorStake := float64(15000)
	numValidators := float64(1000)
	totalStakedAmount := minValidatorStake * numValidators
	maxStakePerOneValidator := float64(0.1) * totalStakedAmount // 10% of total voting power
	errThreshold := 0.05                                        // 5% error threshold due to randomness
	numCandidatesLeadersPerRound := float64(3)

	numViewChanges := float64(1000)
	for validatorStakeAmount := minValidatorStake; validatorStakeAmount < maxStakePerOneValidator; validatorStakeAmount += minValidatorStake {
		selectCount := SortitionResult(0)
		for i := float64(0); i < numViewChanges; i++ {
			var vrfOutput [vrf.VRFOutputSize]byte
			rand.Read(vrfOutput[:])

			sortitionResult := Sortition(validatorStakeAmount, totalStakedAmount, numCandidatesLeadersPerRound, vrfOutput[:])
			selectCount += sortitionResult
		}

		errTolerance := numViewChanges * errThreshold
		expectedSelections := uint64(numViewChanges * numCandidatesLeadersPerRound * validatorStakeAmount / totalStakedAmount)
		assert.InDelta(t, expectedSelections, uint64(selectCount), errTolerance)
		// log.Printf("Stake %%: %0.3f%%; ExpectedCount vs SelectedCount: %d vs %d\n", (validatorStakeAmount / totalStakedAmount * 100), selectCount, expectedSelections)
	}
}

func BenchmarkSortition(b *testing.B) {
	b.StopTimer()

	vrfOutputs := make([]vrf.VRFOutput, b.N)
	for i := 0; i < b.N; i++ {
		rand.Read(vrfOutputs[i][:])
	}

	b.StartTimer()
	validatorStakeAmount := float64(1000000)
	totalStakedAmount := float64(1000000000000)
	numCandidatesLeadersPerRound := float64(3)
	for i := 0; i < b.N; i++ {
		Sortition(validatorStakeAmount, totalStakedAmount, numCandidatesLeadersPerRound, vrfOutputs[i])
	}
}
