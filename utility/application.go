package utility

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

var (
	// TECHDEBT: Re-evalute the denomination of tokens used throughout the codebase. `MillionInt` is
	// currently used to convert POKT to uPOKT but this is not clear throughout the codebase.
	MillionInt = big.NewFloat(1000000)
)

// TODO(M3): Re-evaluate the implementation in this function when implementing the Application Protocol
// and rate limiting
func (u *utilityContext) calculateMaxAppRelays(appStakeStr string) (string, typesUtil.Error) {
	appStakeBigInt, er := converters.StringToBigInt(appStakeStr)
	if er != nil {
		return typesUtil.EmptyString, typesUtil.ErrStringToBigInt(er)
	}

	stabilityAdjustment, err := u.getStabilityAdjustment()
	if err != nil {
		return typesUtil.EmptyString, err
	}

	// INVESTIGATE: Need to understand what `baseline adjustment` is
	baseRate, err := u.getBaselineAppStakeRate()
	if err != nil {
		return typesUtil.EmptyString, err
	}

	// convert amount to float64
	appStake := big.NewFloat(float64(appStakeBigInt.Int64()))

	// get the percentage of the baseline stake rate; can be over 100%
	basePercentage := big.NewFloat(float64(baseRate) / float64(100))

	// multiply the two
	baselineThroughput := basePercentage.Mul(basePercentage, appStake)

	// Convert POKT to uPOKT
	baselineThroughput.Quo(baselineThroughput, MillionInt)

	// add staking adjustment; can be -ve
	adjusted := baselineThroughput.Add(baselineThroughput, big.NewFloat(float64(stabilityAdjustment)))

	// truncate the integer
	result, _ := adjusted.Int(nil)

	// bounding Max Amount of relays to maxint64
	max := big.NewInt(math.MaxInt64)
	if i := result.Cmp(max); i < -1 {
		result = max
	}

	return converters.BigIntToString(result), nil
}
