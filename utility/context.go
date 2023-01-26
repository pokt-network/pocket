package utility

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

type utilityContext struct {
	height             int64
	mempool            typesUtil.Mempool
	persistenceContext *Context // IMPROVE: Rename to `persistenceContext` or `storeContext` or `reversibleContext`?

	// TECHDEBT: Consolidate all these types with the shared Protobuf struct and create a `proposalBlock`
	proposalProposerAddr []byte
	proposalStateHash    string
	proposalBlockTxs     [][]byte
}

// IMPROVE: Consider renaming to `persistenceContext` or `storeContext`?
type Context struct {
	// CLEANUP: Since `Context` embeds `PersistenceRWContext`, we don't need to do `u.Context.PersistenceRWContext`, but can call `u.Context` directly
	modules.PersistenceRWContext
	// TODO(#327): `SavePoints`` have not been implemented yet
	SavePointsM map[string]struct{}
	SavePoints  [][]byte
}

func (u *utilityModule) NewContext(height int64) (modules.UtilityContext, error) {
	ctx, err := u.GetBus().GetPersistenceModule().NewRWContext(height)
	if err != nil {
		return nil, typesUtil.ErrNewPersistenceContext(err)
	}
	return &utilityContext{
		height:  height,
		mempool: u.mempool,
		persistenceContext: &Context{
			PersistenceRWContext: ctx,
			SavePoints:           make([][]byte, 0),
			SavePointsM:          make(map[string]struct{}),
		},
	}, nil
}

func (p *utilityContext) SetProposalBlock(blockHash string, proposerAddr []byte, transactions [][]byte) error {
	p.proposalProposerAddr = proposerAddr
	p.proposalStateHash = blockHash
	p.proposalBlockTxs = transactions
	return nil
}

func (u *utilityContext) Store() *Context {
	return u.persistenceContext
}

func (u *utilityContext) GetPersistenceContext() modules.PersistenceRWContext {
	return u.persistenceContext.PersistenceRWContext
}

func (u *utilityContext) Commit(quorumCert []byte) error {
	if err := u.persistenceContext.PersistenceRWContext.Commit(u.proposalProposerAddr, quorumCert); err != nil {
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

func (u *utilityContext) GetLatestBlockHeight() (int64, typesUtil.Error) {
	height, er := u.Store().GetHeight()
	if er != nil {
		return 0, typesUtil.ErrGetHeight(er)
	}
	return height, nil
}

func (u *utilityContext) getStoreAndHeight() (*Context, int64, typesUtil.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return nil, 0, typesUtil.ErrGetHeight(er)
	}
	return store, height, nil
}

func (u *utilityContext) Codec() codec.Codec {
	return codec.GetCodec()
}

func (u *utilityContext) RevertLastSavePoint() typesUtil.Error {
	if len(u.persistenceContext.SavePointsM) == typesUtil.ZeroInt {
		return typesUtil.ErrEmptySavePoints()
	}
	var key []byte
	popIndex := len(u.persistenceContext.SavePoints) - 1
	key, u.persistenceContext.SavePoints = u.persistenceContext.SavePoints[popIndex], u.persistenceContext.SavePoints[:popIndex]
	delete(u.persistenceContext.SavePointsM, hex.EncodeToString(key))
	if err := u.persistenceContext.PersistenceRWContext.RollbackToSavePoint(key); err != nil {
		return typesUtil.ErrRollbackSavePoint(err)
	}
	return nil
}

func (u *utilityContext) NewSavePoint(transactionHash []byte) typesUtil.Error {
	if err := u.persistenceContext.PersistenceRWContext.NewSavePoint(transactionHash); err != nil {
		return typesUtil.ErrNewSavePoint(err)
	}
	txHash := hex.EncodeToString(transactionHash)
	if _, exists := u.persistenceContext.SavePointsM[txHash]; exists {
		return typesUtil.ErrDuplicateSavePoint()
	}
	u.persistenceContext.SavePoints = append(u.persistenceContext.SavePoints, transactionHash)
	u.persistenceContext.SavePointsM[txHash] = struct{}{}
	return nil
}

func (c *Context) Reset() typesUtil.Error {
	if err := c.PersistenceRWContext.Release(); err != nil {
		return typesUtil.ErrResetContext(err)
	}
	return nil
}
