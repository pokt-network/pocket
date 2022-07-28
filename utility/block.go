package utility

import (
	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) ApplyProposalTransactions(latestHeight int64, proposerAddress []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) ([]byte, error) {
	u.LatestHeight = latestHeight
	// begin block lifecycle phase
	if err := u.BeginBlock(lastBlockByzantineValidators); err != nil {
		return nil, err
	}
	// deliver txs lifecycle phase
	for _, transactionProtoBytes := range transactions {
		tx, err := typesUtil.TransactionFromBytes(transactionProtoBytes)
		if err != nil {
			return nil, err
		}
		if err := tx.ValidateBasic(); err != nil {
			return nil, err
		}
		// Validate and apply the transaction to the Postgres database
		if err := u.ApplyTransaction(tx); err != nil {
			return nil, err
		}
		if err := u.GetPersistenceContext().StoreTransaction(transactionProtoBytes); err != nil {
			return nil, err
		}

		// TODO: if found, remove transaction from mempool
		// DISCUSS: What if the context is rolled back or cancelled. Do we add it back to the mempool?
		// if err := u.Mempool.DeleteTransaction(transaction); err != nil {
		// 	return nil, err
		// }
	}
	// end block lifecycle phase
	if err := u.EndBlock(proposerAddress); err != nil {
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
	if err := u.HandleProposalRewards(proposer); err != nil {
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
	if err := u.BeginUnstakingMaxPausedFishermen(); err != nil {
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
	if err := u.UnstakeFishermenThatAreReady(); err != nil {
		return err
	}
	if err := u.UnstakeServiceNodesThatAreReady(); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetAppHash() ([]byte, types.Error) {
	appHash, er := u.Context.AppHash()
	if er != nil {
		return nil, types.ErrAppHash(er)
	}
	return appHash, nil
}

func (u *UtilityContext) GetBlockHash(height int64) ([]byte, types.Error) {
	store := u.Store()
	hash, er := store.GetBlockHash(int64(height))
	if er != nil {
		return nil, types.ErrGetBlockHash(er)
	}
	return hash, nil
}

func (u *UtilityContext) StoreBlock(blockProtoBytes []byte) error {
	store := u.Store()

	// Store in KV Store
	if err := store.StoreBlock(blockProtoBytes); err != nil {
		return err
	}

	// Store in SQL Store
	// OPTIMIZE: Ideally we'd pass in the block proto struct to utility so we don't
	//           have to unmarshal it here, but that's a major design decision for the interfaces.
	codec := u.Codec()
	block := &types.Block{}
	if err := codec.Unmarshal(blockProtoBytes, block); err != nil {
		return types.ErrProtoUnmarshal(err)
	}

	header := block.BlockHeader
	if err := store.InsertBlock(uint64(header.Height), header.Hash, header.ProposerAddress, header.QuorumCertificate, block.Transactions); err != nil {
		return err
	}

	return nil

}
