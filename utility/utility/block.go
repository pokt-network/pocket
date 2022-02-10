package utility

import (
	"pocket/utility/utility/types"
)

func (u *UtilityContext) ApplyBlock(latestHeight int64, proposer []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error) {
	u.LatestHeight = latestHeight
	// begin block lifecycle phase
	if err := u.BeginBlock(lastBlockByzantineValidators); err != nil {
		return nil, err
	}
	// deliver txs lifecycle phase
	for _, transaction := range transactions {
		tx, err := types.TransactionFromBytes(transaction)
		if err != nil {
			return nil, err
		}
		if err := tx.ValidateBasic(); err != nil {
			return nil, err
		}
		if err := u.ApplyTransaction(tx); err != nil {
			return nil, err
		}
		// if found, remove transaction from mempool
		//if err := u.Mempool.DeleteTransaction(tx); err != nil {
		//	return nil, err
		//}
	}
	// end block lifecycle phase
	if err := u.EndBlock(proposer); err != nil {
		return nil, err
	}
	// return the app hash (consensus module will get the validator set directly
	return u.GetAppHash()
}

func (u *UtilityContext) BeginBlock(previousBlockByzantineValidators [][]byte) types.Error {
	if err := u.HandleByzantineValidators(previousBlockByzantineValidators); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) EndBlock(proposer []byte) types.Error {
	if err := u.HandleProposal(proposer); err != nil {
		return err
	}
	if err := u.UnstakeActorsThatAreReady(); err != nil {
		return err
	}
	if err := u.BeginUnstakingMaxPausedActors(); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPausedActors() types.Error {
	if err := u.BeginUnstakingMaxPausedApps(); err != nil {
		return err
	}
	if err := u.BeginUnstakingMaxPausedFishermans(); err != nil {
		return err
	}
	if err := u.BeginUnstakingMaxPausedValidators(); err != nil {
		return err
	}
	if err := u.BeginUnstakingMaxPausedServiceNodes(); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeActorsThatAreReady() types.Error {
	if err := u.UnstakeAppsThatAreReady(); err != nil {
		return err
	}
	if err := u.UnstakeValidatorsThatAreReady(); err != nil {
		return err
	}
	if err := u.UnstakeFishermansThatAreReady(); err != nil {
		return err
	}
	if err := u.UnstakeServiceNodesThatAreReady(); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetAppHash() (appHash []byte, err types.Error) {
	appHash, er := u.Context.AppHash()
	if er != nil {
		return nil, types.ErrAppHash(er)
	}
	return appHash, nil
}

func (u *UtilityContext) GetBlockHash(height int64) (blockHash []byte, err types.Error) {
	store := u.Store()
	hash, er := store.GetBlockHash(uint64(height))
	if er != nil {
		return nil, types.ErrGetBlockHash(er)
	}
	return hash, nil
}
