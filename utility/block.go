package utility

import (
	"math/big"

	typesCons "github.com/pokt-network/pocket/consensus/types" // TODO (andrew) importing consensus and persistence in this file?
	typesGenesis "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"

	typesUtil "github.com/pokt-network/pocket/utility/types"
)

/*
	This 'block' file contains all the lifecycle block operations.

	The ApplyBlock function is the 'main' operation that executes a 'block' object against the state.

	Pocket Network adopt a Tendermint-like lifecycle of BeginBlock -> DeliverTx -> EndBlock in that
	order. Like the name suggests, BeginBlock is an autonomous state operation that executes at the
	beginning of every block DeliverTx individually applies each transaction against the state and
	rolls it back (not yet implemented) if fails. Like BeginBlock, EndBlock is an autonomous state
	operation that executes at the end of every block.
*/

// TODO(andrew): consolidate with `utility/types/actor.go`
var (
	UtilActorTypes = []typesUtil.UtilActorType{
		typesUtil.UtilActorType_App,
		typesUtil.UtilActorType_Node,
		typesUtil.UtilActorType_Fish,
		typesUtil.UtilActorType_Val,
	}
)

func (u *UtilityContext) ApplyBlock(latestHeight int64, proposerAddress []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) ([]byte, error) {
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

func (u *UtilityContext) BeginBlock(previousBlockByzantineValidators [][]byte) typesUtil.Error {
	if err := u.HandleByzantineValidators(previousBlockByzantineValidators); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) EndBlock(proposer []byte) typesUtil.Error {
	// reward the block proposer
	if err := u.HandleProposalRewards(proposer); err != nil {
		return err
	}
	// unstake actors that have been 'unstaking' for the <Actor>UnstakingBlocks
	if err := u.UnstakeActorsThatAreReady(); err != nil {
		return err
	}
	// begin unstaking the actors who have been paused for MaxPauseBlocks
	if err := u.BeginUnstakingMaxPaused(); err != nil {
		return err
	}
	if _, err := u.Context.UpdateAppHash(); err != nil {
		return types.ErrAppHash(err)
	}
	return nil
}

func (u *UtilityContext) GetAppHash() ([]byte, typesUtil.Error) {
	// Get the root hash of the merkle state tree for state consensus integrity
	appHash, er := u.Context.AppHash()
	if er != nil {
		return nil, typesUtil.ErrAppHash(er)
	}
	return appHash, nil
}

// HandleByzantineValidators handles the validators who either didn't sign at all or disagreed with the 2/3+ majority
func (u *UtilityContext) HandleByzantineValidators(lastBlockByzantineValidators [][]byte) typesUtil.Error {
	latestBlockHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	maxMissedBlocks, err := u.GetValidatorMaxMissedBlocks()
	if err != nil {
		return err
	}
	for _, address := range lastBlockByzantineValidators {
		numberOfMissedBlocks, err := u.GetValidatorMissedBlocks(address)
		if err != nil {
			return err
		}
		// increment missed blocks
		numberOfMissedBlocks++
		// handle if over the threshold
		if numberOfMissedBlocks >= maxMissedBlocks {
			// pause the validator and reset missed blocks
			if err = u.PauseValidatorAndSetMissedBlocks(address, latestBlockHeight, int(typesUtil.HeightNotUsed)); err != nil {
				return err
			}
			// burn validator for missing blocks
			burnPercentage, err := u.GetMissedBlocksBurnPercentage()
			if err != nil {
				return err
			}
			if err = u.BurnActor(typesUtil.UtilActorType_Val, burnPercentage, address); err != nil {
				return err
			}
		} else if err := u.SetValidatorMissedBlocks(address, numberOfMissedBlocks); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) UnstakeActorsThatAreReady() (err typesUtil.Error) {
	var er error
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	for _, utilActorType := range typesUtil.ActorTypes {
		var readyToUnstake []modules.IUnstakingActor
		poolName := utilActorType.GetActorPoolName()
		switch utilActorType {
		case typesUtil.UtilActorType_App:
			readyToUnstake, er = store.GetAppsReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
		case typesUtil.UtilActorType_Fish:
			readyToUnstake, er = store.GetFishermenReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
		case typesUtil.UtilActorType_Node:
			readyToUnstake, er = store.GetServiceNodesReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
		case typesUtil.UtilActorType_Val:
			readyToUnstake, er = store.GetValidatorsReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)

		}
		if er != nil {
			return typesUtil.ErrGetReadyToUnstake(er)
		}
		for _, actor := range readyToUnstake {
			if err = u.SubPoolAmount(poolName, actor.GetStakeAmount()); err != nil {
				return err
			}
			if err = u.AddAccountAmountString(actor.GetOutputAddress(), actor.GetStakeAmount()); err != nil {
				return err
			}
			if err = u.DeleteActor(utilActorType, actor.GetAddress()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPaused() (err typesUtil.Error) {
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	for _, UtilActorType := range UtilActorTypes {
		maxPausedBlocks, err := u.GetMaxPausedBlocks(UtilActorType)
		if err != nil {
			return err
		}
		beforeHeight := latestHeight - int64(maxPausedBlocks)
		// genesis edge case
		if beforeHeight < 0 {
			beforeHeight = 0
		}
		if err := u.UnstakeActorPausedBefore(beforeHeight, UtilActorType); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) UnstakeActorPausedBefore(pausedBeforeHeight int64, UtilActorType typesUtil.UtilActorType) (err typesUtil.Error) {
	var er error
	store := u.Store()
	unstakingHeight, err := u.GetUnstakingHeight(UtilActorType)
	if err != nil {
		return err
	}
	switch UtilActorType {
	case typesUtil.UtilActorType_App:
		er = store.SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.UtilActorType_Fish:
		er = store.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.UtilActorType_Node:
		er = store.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.UtilActorType_Val:
		er = store.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	}
	if er != nil {
		return typesUtil.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) HandleProposalRewards(proposer []byte) typesUtil.Error {
	feePoolName := typesGenesis.Pool_Names_FeeCollector.String()
	feesAndRewardsCollected, err := u.GetPoolAmount(feePoolName)
	if err != nil {
		return err
	}
	if err := u.SetPoolAmount(feePoolName, big.NewInt(0)); err != nil {
		return err
	}
	proposerCutPercentage, err := u.GetProposerPercentageOfFees()
	if err != nil {
		return err
	}
	daoCutPercentage := 100 - proposerCutPercentage
	if daoCutPercentage < 0 || daoCutPercentage > 100 {
		return typesUtil.ErrInvalidProposerCutPercentage()
	}
	amountToProposerFloat := new(big.Float).SetInt(feesAndRewardsCollected)
	amountToProposerFloat.Mul(amountToProposerFloat, big.NewFloat(float64(proposerCutPercentage)))
	amountToProposerFloat.Quo(amountToProposerFloat, big.NewFloat(100))
	amountToProposer, _ := amountToProposerFloat.Int(nil)
	amountToDAO := feesAndRewardsCollected.Sub(feesAndRewardsCollected, amountToProposer)
	if err = u.AddAccountAmount(proposer, amountToProposer); err != nil {
		return err
	}
	if err = u.AddPoolAmount(typesGenesis.Pool_Names_DAO.String(), amountToDAO); err != nil {
		return err
	}
	return nil
}

// GetValidatorMissedBlocks gets the total blocks that a validator has not signed a certain window of time denominated by blocks
func (u *UtilityContext) GetValidatorMissedBlocks(address []byte) (int, typesUtil.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetMissedBlocks(er)
	}
	missedBlocks, er := store.GetValidatorMissedBlocks(address, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetMissedBlocks(er)
	}
	return missedBlocks, nil
}

func (u *UtilityContext) PauseValidatorAndSetMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) typesUtil.Error {
	store := u.Store()
	if err := store.SetValidatorPauseHeightAndMissedBlocks(address, pauseHeight, missedBlocks); err != nil {
		return typesUtil.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) typesUtil.Error {
	store := u.Store()
	er := store.SetValidatorMissedBlocks(address, missedBlocks)
	if er != nil {
		return typesUtil.ErrSetMissedBlocks(er)
	}
	return nil
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
	block := &typesCons.Block{}
	if err := codec.Unmarshal(blockProtoBytes, block); err != nil {
		return typesUtil.ErrProtoUnmarshal(err)
	}
	header := block.BlockHeader
	if err := store.InsertBlock(uint64(header.Height), header.Hash, header.ProposerAddress, header.QuorumCertificate); err != nil {
		return err
	}

	return nil
}
