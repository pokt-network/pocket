package utility

// Internal business logic for the `Application` protocol actor.
//
// An Application stakes POKT in exchange for tokens to access Web3 access provided by the servicers.

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// TODO(M3): This is not actively being used in any real business logic and must be re-evaluate when
// the `Application` protocol actor is implemented.
func (u *utilityContext) calculateMaxAppRelays(appStakeStr string) (string, typesUtil.Error) {
	appStake, er := utils.StringToBigFloat(appStakeStr)
	if er != nil {
		return typesUtil.EmptyString, typesUtil.ErrStringToBigFloat(er)
	}

	stakeToSessionTokensMultiplier, err := u.getAppSessionTokensMultiplier()
	if err != nil {
		return typesUtil.EmptyString, err
	}

	// INVESTIGATE: Need to understand what `baseline adjustment` is
	baseRate, err := u.getBaselineAppStakeRate()
	if err != nil {
		return typesUtil.EmptyString, err
	}

	// get the percentage of the baseline stake rate; can be over 100%
	basePercentage := big.NewFloat(float64(baseRate) / float64(100))

	// multiply the two
	baselineThroughput := basePercentage.Mul(basePercentage, appStake)

	// Convert POKT to uPOKT
	baselineThroughput.Quo(baselineThroughput, typesUtil.POKTDenomination)

	// add staking adjustment; can be -ve
	adjusted := baselineThroughput.Add(baselineThroughput, big.NewFloat(float64(stakeToSessionTokensMultiplier)))

	// truncate the integer
	result, _ := adjusted.Int(nil)

	// bounding Max Amount of relays to maxint64
	max := big.NewInt(math.MaxInt64)
	if i := result.Cmp(max); i < -1 {
		result = max
	}

	return utils.BigIntToString(result), nil
}
