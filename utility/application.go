package utility

// Internal business logic for the `Application` protocol actor.
//
// An Application stakes POKT in exchange for tokens to access Web3 access provided by the servicers.

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// TODO(M3): This is not actively being used in any real business logic yet.
//
// calculateAppSessionTokens determines the number of "session tokens" an application gets at the beginning
// of every session. For example, 1 session token could equate to a quota of 1 relay.
func (u *utilityContext) calculateAppSessionTokens(appStakeStr string) (string, typesUtil.Error) {
	appStake, er := utils.StringToBigInt(appStakeStr)
	if er != nil {
		return typesUtil.EmptyString, typesUtil.ErrStringToBigInt(er)
	}

	stakeToSessionTokensMultiplier, err := u.getAppSessionTokensMultiplier()
	if err != nil {
		return typesUtil.EmptyString, err
	}

	stakeToSessionTokens := big.NewInt(int64(stakeToSessionTokensMultiplier))
	sessionTokens := appStake.Mul(appStake, stakeToSessionTokens)
	return sessionTokens.String(), nil
}
