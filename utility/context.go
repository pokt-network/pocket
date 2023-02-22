package utility

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/shared/modules"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

type utilityContext struct {
	bus    modules.Bus
	height int64

	persistenceContext modules.PersistenceRWContext
	savePointsSet      map[string]struct{}
	savePointsList     [][]byte

	logger modules.Logger

	// TECHDEBT: Consolidate all these types with the shared Protobuf struct and create a `proposalBlock`
	proposalProposerAddr []byte
	proposalStateHash    string
	proposalBlockTxs     [][]byte
}

func (u *utilityModule) NewContext(height int64) (modules.UtilityContext, error) {
	ctx, err := u.GetBus().GetPersistenceModule().NewRWContext(height)
	if err != nil {
		return nil, typesUtil.ErrNewPersistenceContext(err)
	}
	return &utilityContext{
		bus:                u.GetBus(),
		height:             height,
		logger:             u.logger,
		persistenceContext: ctx,
		savePointsList:     make([][]byte, 0),
		savePointsSet:      make(map[string]struct{}),
	}, nil
}

func (p *utilityContext) SetProposalBlock(blockHash string, proposerAddr []byte, txs [][]byte) error {
	p.proposalProposerAddr = proposerAddr
	p.proposalStateHash = blockHash
	p.proposalBlockTxs = txs
	return nil
}

func (u *utilityContext) Store() modules.PersistenceRWContext {
	return u.persistenceContext
}

func (u *utilityContext) GetPersistenceContext() modules.PersistenceRWContext {
	return u.persistenceContext
}

func (u *utilityContext) Commit(quorumCert []byte) error {
	if err := u.persistenceContext.Commit(u.proposalProposerAddr, quorumCert); err != nil {
		return err
	}
	u.persistenceContext = nil
	return nil
}

func (u *utilityContext) Release() error {
	if u.persistenceContext == nil {
		return nil
	}
	if err := u.persistenceContext.Release(); err != nil {
		return err
	}
	u.persistenceContext = nil
	return nil
}

// TECHDEBT: We should be using the height of the context and shouldn't need to be retrieving
//
//	the height from the store either for "current height" operations.
func (u *utilityContext) getStoreAndHeight() (modules.PersistenceRWContext, int64, error) {
	store := u.Store()
	height, err := store.GetHeight()
	return store, height, err
}

// TODO: This has not been tested or investigated in detail
func (u *utilityContext) revertLastSavePoint() typesUtil.Error {
	if len(u.savePointsSet) == typesUtil.ZeroInt {
		return typesUtil.ErrEmptySavePoints()
	}
	var key []byte
	popIndex := len(u.savePointsList) - 1
	key, u.savePointsList = u.savePointsList[popIndex], u.savePointsList[:popIndex]
	delete(u.savePointsSet, hex.EncodeToString(key))
	if err := u.persistenceContext.RollbackToSavePoint(key); err != nil {
		return typesUtil.ErrRollbackSavePoint(err)
	}
	return nil
}

//nolint:unused // TODO: This has not been tested or investigated in detail
func (u *utilityContext) newSavePoint(txHashBz []byte) typesUtil.Error {
	if err := u.persistenceContext.NewSavePoint(txHashBz); err != nil {
		return typesUtil.ErrNewSavePoint(err)
	}
	txHash := hex.EncodeToString(txHashBz)
	if _, exists := u.savePointsSet[txHash]; exists {
		return typesUtil.ErrDuplicateSavePoint()
	}
	u.savePointsList = append(u.savePointsList, txHashBz)
	u.savePointsSet[txHash] = struct{}{}
	return nil
}

func (u *utilityContext) getBus() modules.Bus {
	return u.bus
}

func (u *utilityContext) setBus(bus modules.Bus) *utilityContext {
	u.bus = bus
	return u
}

func (c *utilityContext) Reset() typesUtil.Error {
	if err := c.persistenceContext.Release(); err != nil {
		return typesUtil.ErrResetContext(err)
	}
	return nil
}
