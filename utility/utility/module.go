package utility

import (
	"encoding/hex"
	"log"
	"pocket/consensus/pkg/config"
	"pocket/shared/context"
	"pocket/shared/modules"

	"pocket/utility/utility/types"
)

type UtilityModule struct {
	modules.UtilityModule
	pocketBusMod modules.PocketBusModule

	Mempool types.Mempool
}

type UtilityContext struct {
	modules.UtilityContextInterface

	LatestHeight int64
	Mempool      types.Mempool
	Context      *Context
}

func Create(config *config.Config) (modules.UtilityModule, error) {
	return &UtilityModule{
		Mempool: types.NewMempool(1000, 1000),
	}, nil
}

func (p *UtilityModule) Start(ctx *context.PocketContext) error {
	panic("Why are you starting the utility module?")
	return nil
}

func (p *UtilityModule) Stop(ctx *context.PocketContext) error {
	return nil
}

func (m *UtilityModule) SetPocketBusMod(pocketBus modules.PocketBusModule) {
	m.pocketBusMod = pocketBus
}

func (m *UtilityModule) GetPocketBusMod() modules.PocketBusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (u *UtilityModule) NewUtilityContextWrapper(height int64) (modules.UtilityContextInterface, error) {
	ctx, err := u.NewUtilityContext(height)
	if err != nil {
		panic(err)
	}
	return ctx, nil
}

func (u *UtilityModule) NewUtilityContext(height int64) (modules.UtilityContextInterface, types.Error) {
	ctx, err := u.GetPocketBusMod().GetPersistenceModule().NewContext(height)


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

func (u *UtilityContext) Codec() types.Codec {
	return types.UtilityCodec()
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
