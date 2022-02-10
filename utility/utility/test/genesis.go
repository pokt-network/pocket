package test

import (
	"fmt"
	"github.com/pokt-network/utility-pre-prototype/utility"
	"github.com/pokt-network/utility-pre-prototype/utility/types"
)

func InitGenesis(u *utility.UtilityContext, state *GenesisState) types.Error {
	if err := InsertParams(u, state.Params); err != nil {
		return err
	}
	for _, account := range state.Accounts {
		if err := u.SetAccountWithAmountString(account.Address, account.Amount); err != nil {
			return err
		}
	}
	for _, p := range state.Pools {
		if err := u.InsertPool(p.Name, p.Account.Address, p.Account.Amount); err != nil {
			return err
		}
	}
	for _, validator := range state.Validators {
		err := u.InsertValidator(validator.Address, validator.PublicKey, validator.Output, validator.ServiceURL, validator.StakedTokens)
		if err != nil {
			return err
		}
	}
	for _, fisherman := range state.Fishermen {
		err := u.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, fisherman.ServiceURL, fisherman.StakedTokens, fisherman.Chains)
		if err != nil {
			return err
		}
	}
	for _, serviceNode := range state.ServiceNodes {
		err := u.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, serviceNode.ServiceURL, serviceNode.StakedTokens, serviceNode.Chains)
		if err != nil {
			return err
		}
	}
	for _, application := range state.Apps {
		maxRelays, err := u.CalculateAppRelays(application.StakedTokens)
		if err != nil {
			return err
		}
		err = u.InsertApplication(application.Address, application.PublicKey, application.Output, maxRelays, application.StakedTokens, application.Chains)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExportState(u *utility.UtilityContext) (*GenesisState, types.Error) {
	c, ok := u.Context.PersistenceContext.(*MockPersistenceContext)
	if !ok {
		return nil, types.ErrExportState(fmt.Errorf("couldn't convert context to `mock persistence context`"))
	}
	return c.ExportState()
}
