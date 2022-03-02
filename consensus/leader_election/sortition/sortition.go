package sortition

// This is an implementation of section 5 in the Algorand whitepaper: https://algorandcom.cdn.prismic.io/algorandcom%2Fa26acb80-b80c-46ff-a1ab-a8121f74f3a3_p51-gilad.pdf
// It heavily references the Algorand implementation here: https://github.com/algorand/go-algorand/blob/master/data/committee/sortition/sortition.go

import (
	crand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"

	"golang.org/x/exp/rand"

	"github.com/pokt-network/pocket/consensus/leader_election/vrf"
	types_consensus "github.com/pokt-network/pocket/consensus/types"

	"gonum.org/v1/gonum/stat/distuv"
)

type SortitionResult uint64

// Hashlen in bits of vrfOut
const precision = uint(8 * (vrf.VRFOutputSize + 1))

// Initialized via init() below and
var maxFloat *big.Float
var maxInt *big.Int

func Sortition(validatorStakeAmount, totalStakedAmount, numExpectedCandidates float64, vrfOut vrf.VRFOutput) SortitionResult {
	// Assuming that `totalStakedAmount` POKT is staked on the entired network,
	// each individual POKT has has 1 / totalStakedValue to be selected, with a total
	// of `numExpectedCandidates` (i.e. number of tokens) to be selected as potential
	// leaders.
	p := numExpectedCandidates / totalStakedAmount

	// Normalizes vrfOut to a uniformally distributed value in [0, 1)
	vrfProb := vrfOutProb(vrfOut)

	// Generate a random source use the crypto library.
	f, err := crand.Int(crand.Reader, maxInt)
	if err != nil {
		log.Fatal("failed to generate random number for sortition")
	}
	src := rand.NewSource(f.Uint64())

	binomial := distuv.Binomial{
		N:   validatorStakeAmount, // The number of Bernoulli trials is equal to the validator's specific stake; 1 entry per POKT staked.
		P:   p,                    // Each individual POKT has an equal probability of being selected.
		Src: src,
	}

	return SortitionResult(sortitionBinomialCDFWalk(&binomial, vrfProb, uint64(validatorStakeAmount)))
}

/*
TODO(discuss): Discrepancies here from the original spec (github.com/pokt-network/pocket-network-protocol/tree/main/consensus):
	- Not using `lastNLeaders` as part of the seed
	- Not using the validator's `PubKey` as part of the seed
	* Reasoning: As long as the VRF keys are computed BEFORE some unpredictable seed, the security
		       guarantees are maintained - this is provided by prevBlockHash.
*/

// Seed to be used for soritition when generating the vrfOut and vrfProof. Exposed publically for optimization purposes.
func FormatSeed(
	h types_consensus.BlockHeight,
	r types_consensus.Round,
	prevBlockHash string,
) []byte {
	return []byte(fmt.Sprintf("%d:%d:%s", h, r, prevBlockHash))
}

// For a specific validator, at most `validatorStakeAmount` (i.e. # of POKT staked) may be used to elect
// them as a leader for any specific round depending on the randomly distributed VRF Output probability.
// Therefore, return at most the amount of tokens they have staked.
// TODO(olshansky): Full understand this; github.com/algorand/go-algorand/pull/3558#issuecomment-1032186022.
func sortitionBinomialCDFWalk(binomial *distuv.Binomial, vrfProb float64, validatorStakeAmount uint64) uint64 {
	for j := uint64(0); j < validatorStakeAmount; j++ {
		jCDF := binomial.CDF(float64(j))
		if vrfProb <= jCDF {
			return j
		}
	}
	return validatorStakeAmount
}

// Normalizes vrfOut to a random value in [0, 1).
func vrfOutProb(vrfOut vrf.VRFOutput) float64 {
	t := &big.Int{}
	t.SetBytes(vrfOut[:])

	h := big.Float{}
	h.SetPrec(precision)
	h.SetInt(t)

	ratio := big.Float{}
	cratio, _ := ratio.Quo(&h, maxFloat).Float64()

	return cratio
}

// One time initialization of sortition related variables
func init() {
	// The maximum lexical value that vrfOut can be based on the VRF output size
	var maxFloatString string = fmt.Sprintf("0x%s", strings.Repeat("f", vrf.VRFOutputSize*2))

	// The denominator used to normalize vrfOut to a value in [0, 1).
	var b int
	var err error
	maxFloat, b, err = big.ParseFloat(maxFloatString, 0, precision, big.ToNearestEven)
	if b != 16 || err != nil {
		log.Fatal("failed to parse big float constant for sortition")
	}

	maxInt = big.NewInt(^int64(0))
	maxInt.Abs(maxInt)
}
