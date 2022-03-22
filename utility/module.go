package utility

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

var _ modules.UtilityModule = &UtilityModule{}

type UtilityModule struct {
	bus modules.Bus

	Mempool types.Mempool
}

type UtilityContext struct {
	LatestHeight int64
	Mempool      types.Mempool
	Context      *Context
}

func Create(_ *config.Config) (modules.UtilityModule, error) {
	return &UtilityModule{
		Mempool: types.NewMempool(1000, 1000),
	}, nil
}

func (p *UtilityModule) Start() error {
	return nil
}

func (p *UtilityModule) Stop() error {
	return nil
}

func (m *UtilityModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *UtilityModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (u *UtilityModule) NewContext(height int64) (modules.UtilityContext, error) {
	ctx, err := u.GetBus().GetPersistenceModule().NewContext(height)
	if err != nil {
		return nil, types.ErrNewContext(err)
	}
	return &UtilityContext{
		LatestHeight: height,
		Mempool:      u.Mempool,
		Context:      &Context{PersistenceContext: ctx},
	}, nil
}

func (u *UtilityContext) Store() *Context {
	return u.Context
}

func (u *UtilityContext) GetPersistanceContext() modules.PersistenceContext {
	return u.Context.PersistenceContext
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

func (u *UtilityContext) Codec() typesUtil.Codec {
	return typesUtil.UtilityCodec()
}

type Context struct {
	modules.PersistenceContext
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
