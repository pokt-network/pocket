package utility

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	sharedTypes "github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility/types"
)

var _ modules.UtilityModule = &UtilityModule{}

type UtilityModule struct {
	pocketBusMod modules.Bus

	Mempool sharedTypes.Mempool
}

type UtilityContext struct {
	LatestHeight int64
	Mempool      sharedTypes.Mempool
	Context      *Context
}

func Create(_ *config.Config) (modules.UtilityModule, error) {
	return &UtilityModule{
		Mempool: sharedTypes.NewMempool(1000, 1000),
	}, nil
}

func (p *UtilityModule) Start() error {
	return nil
}

func (p *UtilityModule) Stop() error {
	return nil
}

func (m *UtilityModule) SetBus(pocketBus modules.Bus) {
	m.pocketBusMod = pocketBus
}

func (m *UtilityModule) GetBus() modules.Bus {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (u *UtilityModule) NewContext(height int64) (modules.UtilityContext, error) {
	ctx, err := u.GetBus().GetPersistenceModule().NewContext(height)
	if err != nil {
		return nil, sharedTypes.ErrNewContext(err)
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

func (c *Context) Reset() sharedTypes.Error {
	if err := c.PersistenceContext.Reset(); err != nil {
		return sharedTypes.ErrResetContext(err)
	}
	return nil
}

func (u *UtilityContext) GetLatestHeight() (int64, sharedTypes.Error) {
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

func (u *UtilityContext) RevertLastSavePoint() sharedTypes.Error {
	if u.Context.SavePointsM == nil || len(u.Context.SavePointsM) == 0 {
		return sharedTypes.ErrEmptySavePoints()
	}
	var key []byte
	popIndex := len(u.Context.SavePoints) - 1
	key, u.Context.SavePoints = u.Context.SavePoints[popIndex], u.Context.SavePoints[:popIndex]
	delete(u.Context.SavePointsM, hex.EncodeToString(key))
	if err := u.Context.PersistenceContext.RollbackToSavePoint(key); err != nil {
		return sharedTypes.ErrRollbackSavePoint(err)
	}
	return nil
}

func (u *UtilityContext) NewSavePoint(transactionHash []byte) sharedTypes.Error {
	if u.Context.SavePointsM == nil || len(u.Context.SavePointsM) == 0 {
		u.Context.SavePoints = make([][]byte, 0)
		u.Context.SavePointsM = make(map[string]struct{})
	}
	if err := u.Context.PersistenceContext.NewSavePoint(transactionHash); err != nil {
		return sharedTypes.ErrNewSavePoint(err)
	}
	txHash := hex.EncodeToString(transactionHash)
	if _, ok := u.Context.SavePointsM[txHash]; ok {
		return sharedTypes.ErrDuplicateSavePoint()
	}
	u.Context.SavePoints = append(u.Context.SavePoints, transactionHash)
	u.Context.SavePointsM[txHash] = struct{}{}
	return nil
}
