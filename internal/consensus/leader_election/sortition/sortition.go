package sortition

/*
This is an implementation of section 5 in the Algorand whitepaper [1], which heavily references
Algorand's own implementation of the algorithm [2].

[1] https://algorandcom.cdn.prismic.io/algorandcom%2Fa26acb80-b80c-46ff-a1ab-a8121f74f3a3_p51-gilad.pdf
[2] https://github.com/algorand/go-algorand/blob/master/data/committee/sortition/sortition.go
*/

import (
	crand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"

	"golang.org/x/exp/rand"

	"github.com/pokt-network/pocket/internal/consensus/leader_election/vrf"

	"gonum.org/v1/gonum/stat/distuv"
)

type SortitionResult uint64

const vrfOutFloatPrecision = uint(8 * (vrf.VRFOutputSize + 1)) // Hashlen in bits of vrfOut

var (
	maxVrfOutFloat *big.Float // Computed in `init()`
	maxRandomInt   *big.Int
)

// One time initialization of sortition related constants
func init() {
	// The maximum lexical value that vrfOut can be, based on the VRFOutputSize.
	// In other words, the denominator used to normalize vrfOut to a value in [0, 1).
	var maxVrfOutFloatString string = fmt.Sprintf("0x%s", strings.Repeat("f", vrf.VRFOutputSize*2))

	var base int
	var err error
	maxVrfOutFloat, base, err = big.ParseFloat(maxVrfOutFloatString, 0, vrfOutFloatPrecision, big.ToNearestEven)
	if base != 16 || err != nil {
		log.Fatal("failed to parse big float constant for sortition")
	}

	maxRandomInt = big.NewInt(^int64(0))
	maxRandomInt.Abs(maxRandomInt)
}

// Based on a validator's stake, the amount of staked uPOKT in the network, and the VRF output that
// the validator generated at some point in the past, this returns a value that can be used to rank
// potential view change leaders that is uniformally distributed and proportional to the validator's
// stake. See [3] for simpler explanation of the algorithm [1].
// [3] https://community.algorand.org/blog/the-intuition-behind-algorand-cryptographic-sortition/
func Sortition(validatorStake, networkStake, numExpectedCandidates uint64, vrfOut vrf.VRFOutput) SortitionResult {
	// Explanation: In Pocket Network's leader consensus algorithm, there is only going to be one leader
	// per round / view change. However, during sortition, several candidates are selected, to avoid
	// the chances of there being no leader at all. The chosen leaders (<= numCandidatesLeadersPerRound)
	// are sorted based on their sortition result and the top one is selected.
	//
	// Example: Assuming that all validators in the network staked `networkStake`
	// uPOKT, and each individual uPOKT has `1 / networkStake` probability of being selected, with
	// a total of `numExpectedCandidates` (i.e. # of uPOKT) to be selected as potential leaders.
	p := float64(numExpectedCandidates) / float64(networkStake)

	// Normalizes vrfOut to a uniformally distributed value in [0, 1)
	vrfProb := vrfOutProb(vrfOut)

	// Generate a random source using the crypto library
	f, err := crand.Int(crand.Reader, maxRandomInt)
	if err != nil {
		log.Fatal("failed to generate random integer for sortition")
	}
	src := rand.NewSource(f.Uint64())

	binomial := distuv.Binomial{
		N:   float64(validatorStake), // # of Bernoulli trials == validator's stake: 1 trial per uPOKT staked
		P:   p,                       // Each uPOKT has an equal probability of being selected
		Src: src,
	}

	return SortitionResult(sortitionBinomialCDFWalk(&binomial, vrfProb, uint64(validatorStake)))
}

/*
TODO(discuss): How should the seed be formatted?
Discrepancies from the original spec [4]:
	- Not using `lastNLeaders` as part of the seed
	- Not using the validator's `PubKey` as part of the seed
Reasoning:
	As long as the VRF keys are computed BEFORE some unpredictable seed, the security
	guarantees are maintained - this is provided by `prevBlockHash`.
[4] github.com/pokt-network/pocket-network-protocol/tree/main/consensus
*/
// Seed to be used for soritition when generating the vrfOut and vrfProof
func FormatSeed(height uint64, round uint64, prevBlockHash string) []byte {
	return []byte(fmt.Sprintf("%d:%d:%s", height, round, prevBlockHash))
}

// For a specific validator, at most `validatorStake` (i.e. # uof POKT staked) may be used to elect
// them as a leader for any specific round depending on the randomly distributed VRF Output
// probability. Therefore, return at most the amount of tokens they have staked.
func sortitionBinomialCDFWalk(binomial *distuv.Binomial, vrfProb float64, validatorStake uint64) uint64 {
	// TODO(olshansky): Fully understand this: github.com/algorand/go-algorand/pull/3558#issuecomment-1032186022.
	for j := uint64(0); j < validatorStake; j++ {
		jCDF := binomial.CDF(float64(j))
		if vrfProb <= jCDF {
			return j
		}
	}
	return validatorStake
}

// Normalizes vrfOut to a random value in [0, 1).
func vrfOutProb(vrfOut vrf.VRFOutput) float64 {
	t := &big.Int{}
	t.SetBytes(vrfOut[:])

	h := big.Float{}
	h.SetPrec(vrfOutFloatPrecision)
	h.SetInt(t)

	ratio := big.Float{}
	prob, _ := ratio.Quo(&h, maxVrfOutFloat).Float64()

	return prob
}
