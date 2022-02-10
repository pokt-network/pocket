package utility

import (
	"encoding/hex"
	"github.com/pokt-network/utility-pre-prototype/shared/bus"
	"github.com/pokt-network/utility-pre-prototype/utility/types"
)

var _ bus.UtilityModule = &UtilityContext{}

type UtilityModule struct {
	Mempool types.Mempool
	Bus     bus.Bus
}

func NewUtilityModule(bus bus.Bus) (UtilityModule, error) {
	return UtilityModule{
		Mempool: types.NewMempool(1000, 1000),
		Bus:     bus,
	}, nil
}

type UtilityContext struct {
	LatestHeight int64
	Mempool      types.Mempool
	Context      *Context
}

func (u *UtilityModule) NewUtilityContext(height int64) (*UtilityContext, types.Error) {
	context, err := u.Bus.GetPersistenceModule().NewContext(height)
	if err != nil {
		return nil, types.ErrNewContext(err)
	}
	return &UtilityContext{
		LatestHeight: height,
		Mempool:      u.Mempool,
		Context:      &Context{PersistenceContext: context},
	}, nil
}

func (u *UtilityContext) Store() *Context {
	return u.Context
}

func (u *UtilityContext) ReleaseContext() {
	u.Context.Release()
	u.Context = nil
}

func (c *Context) Reset() types.Error {
	if err := c.PersistenceContext.Reset(); err != nil {
		return types.ErrResetContext(err)
	}
	return nil
}

func (u *UtilityContext) GetLatestHeight() (int64, types.Error) {
	return u.LatestHeight, nil
}

func (u *UtilityContext) Codec() types.Codec {
	return types.UtilityCodec()
}

type Context struct {
	bus.PersistenceContext
	SavePointsM map[string]struct{}
	SavePoints  [][]byte
}

func (u *UtilityContext) RevertLastSavePoint() types.Error {
	if u.Context.SavePointsM == nil || len(u.Context.SavePointsM) == 0 {
		return types.ErrEmptySavePoints()
	}
	var key []byte
	popIndex := len(u.Context.SavePoints) - 1
	key, u.Context.SavePoints = u.Context.SavePoints[popIndex], u.Context.SavePoints[:popIndex]
	delete(u.Context.SavePointsM, hex.EncodeToString(key))
	if err := u.Context.PersistenceContext.RollbackToSavePoint(key); err != nil {
		return types.ErrRollbackSavePoint(err)
	}
	return nil
}

func (u *UtilityContext) NewSavePoint(transactionHash []byte) types.Error {
	if u.Context.SavePointsM == nil || len(u.Context.SavePointsM) == 0 {
		u.Context.SavePoints = make([][]byte, 0)
		u.Context.SavePointsM = make(map[string]struct{})
	}
	if err := u.Context.PersistenceContext.NewSavePoint(transactionHash); err != nil {
		return types.ErrNewSavePoint(err)
	}
	txHash := hex.EncodeToString(transactionHash)
	if _, ok := u.Context.SavePointsM[txHash]; ok {
		return types.ErrDuplicateSavePoint()
	}
	u.Context.SavePoints = append(u.Context.SavePoints, transactionHash)
	u.Context.SavePointsM[txHash] = struct{}{}
	return nil
}
