package utility

// Internal business logic for the `Application` protocol actor.
//
// An Application stakes POKT in exchange for tokens to access Web3 access provided by the servicers.

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// TODO(M3): Not bThis is not actively being used in any real business logic and must be re-evaluate when
// the `Application` protocol actor is implemented.
func (u *utilityContext) calculateAppSessionTokens(appStakeStr string) (string, typesUtil.Error) {
	appStake, er := utils.StringToBigFloat(appStakeStr)
	if er != nil {
		return typesUtil.EmptyString, typesUtil.ErrStringToBigFloat(er)
	}

	stakeToSessionTokensMultiplier, err := u.getAppSessionTokensMultiplier()
	if err != nil {
		return typesUtil.EmptyString, err
	}
	multiplier := big.NewFloat(float64(stakeToSessionTokensMultiplier))

	return appStake.Mul(appStake, multiplier).String(), nil
}
