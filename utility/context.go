package utility

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

type UtilityContext struct {
	LatestHeight int64
	Mempool      typesUtil.Mempool
	Context      *Context // IMPROVE: Rename to `persistenceContext` or `storeContext` or `reversibleContext`?
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
	return &UtilityContext{
		LatestHeight: height,
		Mempool:      u.Mempool,
		Context: &Context{
			PersistenceRWContext: ctx,
			SavePoints:           make([][]byte, 0),
			SavePointsM:          make(map[string]struct{}),
		},
	}, nil
}

func (u *UtilityContext) Store() *Context {
	return u.Context
}

func (u *UtilityContext) GetPersistenceContext() modules.PersistenceRWContext {
	return u.Context.PersistenceRWContext
}

func (u *UtilityContext) Commit(quorumCert []byte) error {
	if err := u.Context.PersistenceRWContext.Commit(quorumCert); err != nil {
		return err
	}
	u.Context = nil
	return nil
}

func (u *UtilityContext) Release() error {
	if u.Context == nil {
		return nil
	}
	if err := u.Context.Release(); err != nil {
		return err
	}
	u.Context = nil
	return nil
}

func (u *UtilityContext) GetLatestBlockHeight() (int64, typesUtil.Error) {
	height, er := u.Store().GetHeight()
	if er != nil {
		return 0, typesUtil.ErrGetHeight(er)
	}
	return height, nil
}

func (u *UtilityContext) GetStoreAndHeight() (*Context, int64, typesUtil.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return nil, 0, typesUtil.ErrGetHeight(er)
	}
	return store, height, nil
}

func (u *UtilityContext) Codec() codec.Codec {
	return codec.GetCodec()
}

func (u *UtilityContext) RevertLastSavePoint() typesUtil.Error {
	if len(u.Context.SavePointsM) == typesUtil.ZeroInt {
		return typesUtil.ErrEmptySavePoints()
	}
	var key []byte
	popIndex := len(u.Context.SavePoints) - 1
	key, u.Context.SavePoints = u.Context.SavePoints[popIndex], u.Context.SavePoints[:popIndex]
	delete(u.Context.SavePointsM, hex.EncodeToString(key))
	if err := u.Context.PersistenceRWContext.RollbackToSavePoint(key); err != nil {
		return typesUtil.ErrRollbackSavePoint(err)
	}
	return nil
}

func (u *UtilityContext) NewSavePoint(transactionHash []byte) typesUtil.Error {
	if err := u.Context.PersistenceRWContext.NewSavePoint(transactionHash); err != nil {
		return typesUtil.ErrNewSavePoint(err)
	}
	txHash := hex.EncodeToString(transactionHash)
	if _, exists := u.Context.SavePointsM[txHash]; exists {
		return typesUtil.ErrDuplicateSavePoint()
	}
	u.Context.SavePoints = append(u.Context.SavePoints, transactionHash)
	u.Context.SavePointsM[txHash] = struct{}{}
	return nil
}

func (c *Context) Reset() typesUtil.Error {
	if err := c.PersistenceRWContext.Release(); err != nil {
		return typesUtil.ErrResetContext(err)
	}
	return nil
}
