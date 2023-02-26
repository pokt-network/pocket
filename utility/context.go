package utility

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

var (
	_ modules.IntegratableModule = &utilityContext{}
	_ modules.UtilityContext     = &utilityContext{}
)

type utilityContext struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	height int64

	store          modules.PersistenceRWContext
	savePointsSet  map[string]struct{}
	savePointsList [][]byte

	// TECHDEBT: Consolidate all these types with the shared Protobuf struct and create a `proposalBlock`
	proposalStateHash    string
	proposalProposerAddr []byte
	proposalBlockTxs     [][]byte
}

func (u *utilityModule) NewContext(height int64) (modules.UtilityContext, error) {
	persistenceCtx, err := u.GetBus().GetPersistenceModule().NewRWContext(height)
	if err != nil {
		return nil, typesUtil.ErrNewPersistenceContext(err)
	}
	ctx := &utilityContext{
		logger: u.logger,
		height: height,

		// No save points on start
		store:          persistenceCtx,
		savePointsList: make([][]byte, 0),
		savePointsSet:  make(map[string]struct{}),
	}
	ctx.IntegratableModule.SetBus(u.GetBus())
	return ctx, nil
}

func (p *utilityContext) SetProposalBlock(blockHash string, proposerAddr []byte, txs [][]byte) error {
	p.proposalStateHash = blockHash
	p.proposalProposerAddr = proposerAddr
	p.proposalBlockTxs = txs
	return nil
}

func (u *utilityContext) Commit(quorumCert []byte) error {
	if err := u.store.Commit(u.proposalProposerAddr, quorumCert); err != nil {
		return err
	}
	u.store = nil
	return nil
}

func (u *utilityContext) Release() error {
	if u.store == nil {
		return nil
	}
	if err := u.store.Release(); err != nil {
		return err
	}
	u.store = nil
	return nil
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
	if err := u.store.RollbackToSavePoint(key); err != nil {
		return typesUtil.ErrRollbackSavePoint(err)
	}
	return nil
}

//nolint:unused // TODO: This has not been tested or investigated in detail
func (u *utilityContext) newSavePoint(txHashBz []byte) typesUtil.Error {
	if err := u.store.NewSavePoint(txHashBz); err != nil {
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
