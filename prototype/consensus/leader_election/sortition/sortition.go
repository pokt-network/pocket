package sortition

import (
	"fmt"
	"log"
	"math/big"
	"pocket/consensus/leader_election/vrf"
	consensus_types "pocket/consensus/types"
	"strings"

	"gonum.org/v1/gonum/stat/distuv"
)

type SortitionResult uint64

// The denominator then mapping vrfOut to [0, 1)
var maxFloatString string = fmt.Sprintf("0x%s", strings.Repeat("f", vrf.VRFOutputSize*2))

// TODO(discuss):
// 	1. Instead of using `lastNLeaders`, we are using `prevBlockHash` (i.e. the app hash of the previous block).
// 	2. Why do we want to use the validator's `PubKey` in the seed. Uniquness in the block hash is sufficient.
//
// func formatSeed(h consensus.BlockHeight, r consensus.Round, prevBlockHash string, pubKey *types.PublicKey) []byte {
// 	return []byte(fmt.Sprintf("%d:%d:%s:%s", h, r, prevBlockHash, pubKey))
// }
func FormatSeed(
	h consensus_types.BlockHeight,
	r consensus_types.Round,
	prevBlockHash string,
) []byte {
	return []byte(fmt.Sprintf("%d:%d:%s", h, r, prevBlockHash))
}

// Reference Sesion 5 in the Algorand whitepaper: https://algorandcom.cdn.prismic.io/algorandcom%2Fa26acb80-b80c-46ff-a1ab-a8121f74f3a3_p51-gilad.pdf
func Sortition(validatorVotingPower, totalVotingPower, numExpectedCandidates float64, vrfOut vrf.VRFOutput) SortitionResult {
	p := numExpectedCandidates / totalVotingPower // Each POKT staked by validators has 1 / totalStakedValue to be selected.
	vrfRatio := vrfOutProb(vrfOut)                // Converts vrfOut to a random value in [0, 1).

	binomial := distuv.Binomial{
		N:   validatorVotingPower, // The number of Bernoulli trials is equal to the validator's stake.
		P:   p,                    // Each individual POKT has an equal probability of being selected.
		Src: nil,                  // TODO: Should we create a Source using crypto/rand?
	}

	return SortitionResult(sortitionBinomialCDFWalk(&binomial, vrfRatio, uint64(validatorVotingPower)))
}

/*
	Returns the number of tokens that were selected for the specific validator based on their VRF value.

	https://github.com/algorand/go-algorand/blob/master/data/committee/sortition/sortition.cpp

	TODO(olshansky): Make sure to fully understand what's going on here.
*/
func sortitionBinomialCDFWalk(binomial *distuv.Binomial, vrfRatio float64, validatorVotingPower uint64) uint64 {
	for j := uint64(0); j < validatorVotingPower; j++ {
		cdfBoundary := binomial.CDF(float64(j))
		if vrfRatio <= cdfBoundary {
			return j
		}
	}
	return validatorVotingPower
}

/*
	Converts vrfOut to a random value in [0, 1).

	Reference from Algorand: https://github.com/algorand/go-algorand/blob/master/data/committee/sortition/sortition.go
*/
func vrfOutProb(vrfOut vrf.VRFOutput) float64 {
	precision := uint(8 * (len(vrfOut) + 1)) // hashlen in bits

	max, b, err := big.ParseFloat(maxFloatString, 0, precision, big.ToNearestEven)
	if b != 16 || err != nil {
		log.Fatal("failed to parse big float constant for sortition")
	}

	t := &big.Int{}
	t.SetBytes(vrfOut[:])

	h := big.Float{}
	h.SetPrec(precision)
	h.SetInt(t)

	ratio := big.Float{}
	cratio, _ := ratio.Quo(&h, max).Float64()

	// hval, _ := h.Float64()
	// maxVal, _ := max.Float64()
	// log.Println("%f / %f", hval, maxVal)

	return cratio
}
