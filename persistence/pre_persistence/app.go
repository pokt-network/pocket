package pre_persistence

import (
	"bytes"
	"fmt"

	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
)

func (m *PrePersistenceContext) GetAppsUpdated(height int64) ([][]byte, error) {
	// Not implemented
	return nil, nil
}

func (m *PrePersistenceContext) GetAppExists(address []byte, height int64) (exists bool, err error) {
	db := m.Store()
	key := append(AppPrefixKey, address...)
	if found := db.Contains(key); !found {
		return false, nil
	}
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}
	if bz == nil {
		return false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return false, nil
	}
	return true, nil
}

func (m *PrePersistenceContext) GetApp(address []byte, height int64) (app *typesGenesis.App, err error) {
	app = &typesGenesis.App{}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	bz, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return nil, nil
	}
	if err = proto.Unmarshal(bz, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (m *PrePersistenceContext) GetAllApps(height int64) (apps []*typesGenesis.App, err error) {
	codec := types.GetCodec()
	apps = make([]*typesGenesis.App, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: AppPrefixKey,
			Limit: PrefixEndBytes(AppPrefixKey),
		})
	} else {
		key := HeightKey(height, AppPrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		a := typesGenesis.App{}
		if err := codec.Unmarshal(bz, &a); err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return
}

func (m *PrePersistenceContext) GetAppStakeAmount(height int64, address []byte) (string, error) {
	app, err := m.GetApp(address, height)
	if err != nil {
		return "", err
	}
	return app.StakedTokens, nil
}

func (m *PrePersistenceContext) SetAppStakeAmount(address []byte, stakeAmount string) error {
	codec := types.GetCodec()
	db := m.Store()
	app, err := m.GetApp(address, m.Height)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	app.StakedTokens = stakeAmount
	bz, err := codec.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(append(AppPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) InsertApp(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedAmount string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	if exists, _ := m.GetAppExists(address, height); exists {
		return fmt.Errorf("already exists in world state")
	}
	codec := types.GetCodec()
	db := m.Store()
	key := append(AppPrefixKey, address...)
	app := typesGenesis.App{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		Chains:          chains,
		MaxRelays:       maxRelays,
		StakedTokens:    stakedAmount,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := codec.Marshal(&app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) UpdateApp(address []byte, maxRelaysToAdd string, amount string, chainsToUpdate []string) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	app, err := m.GetApp(address, height)
	if err != nil {
		return err
	}
	codec := types.GetCodec()
	db := m.Store()
	key := append(AppPrefixKey, address...)
	// compute new values
	//stakedTokens, err := types.StringToBigInt(app.StakedTokens)
	//if err != nil {
	//	return err
	//}
	stakedTokens, err := types.StringToBigInt(amount)
	//if err != nil {
	//	return err
	//}
	//stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	maxRelays, err := types.StringToBigInt(app.MaxRelays)
	if err != nil {
		return err
	}
	maxRelaysToAddI, err := types.StringToBigInt(maxRelaysToAdd)
	if err != nil {
		return err
	}
	maxRelays.Add(maxRelays, maxRelaysToAddI)
	// update values
	app.MaxRelays = types.BigIntToString(maxRelays)
	app.StakedTokens = types.BigIntToString(stakedTokens)
	app.Chains = chainsToUpdate
	// marshal
	bz, err := codec.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) DeleteApp(address []byte) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	exists, err := m.GetAppExists(address, height)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *PrePersistenceContext) GetAppsReadyToUnstake(height int64, _ int) (apps []*types.UnstakingActor, err error) { // TODO delete unused parameter
	db := m.Store()
	unstakingKey := append(UnstakingAppPrefixKey, types.Int64ToBytes(height)...)
	if has := db.Contains(unstakingKey); !has {
		return nil, nil
	}
	val, err := db.Get(unstakingKey)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return make([]*types.UnstakingActor, 0), nil
	}
	unstakingApps := types.UnstakingActors{}
	if err := proto.Unmarshal(val, &unstakingApps); err != nil {
		return nil, err
	}
	for _, app := range unstakingApps.UnstakingActors {
		apps = append(apps, app)
	}
	return
}

func (m *PrePersistenceContext) GetAppStatus(address []byte, height int64) (status int, err error) {
	app, err := m.GetApp(address, height)
	if err != nil {
		return types.ZeroInt, err
	}
	if app == nil {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(app.Status), nil
}

func (m *PrePersistenceContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	app, err := m.GetApp(address, height)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	codec := types.GetCodec()
	unstakingApps := types.UnstakingActors{}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	app.UnstakingHeight = unstakingHeight
	app.Status = int32(status)
	bz, err := codec.Marshal(app)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingAppPrefixKey, types.Int64ToBytes(unstakingHeight)...)
	if found := db.Contains(unstakingKey); found {
		val, err := db.Get(unstakingKey)
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(val, &unstakingApps); err != nil {
			return err
		}
	}
	unstakingApps.UnstakingActors = append(unstakingApps.UnstakingActors, &types.UnstakingActor{
		Address:       app.Address,
		StakeAmount:   app.StakedTokens,
		OutputAddress: app.Output,
	})
	unstakingBz, err := codec.Marshal(&unstakingApps)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *PrePersistenceContext) GetAppPauseHeightIfExists(address []byte, height int64) (int64, error) {
	app, err := m.GetApp(address, height)
	if err != nil {
		return types.ZeroInt, err
	}
	if app == nil {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(app.PausedHeight), nil
}

// SetAppStatusAndUnstakingHeightIfPausedBefore : This unstakes the actors that have reached max pause height
func (m *PrePersistenceContext) SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	codec := types.GetCodec()
	it := db.NewIterator(&util.Range{
		Start: AppPrefixKey,
		Limit: PrefixEndBytes(AppPrefixKey),
	})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		app := typesGenesis.App{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := codec.Unmarshal(bz, &app); err != nil {
			return err
		}
		if app.PausedHeight < pausedBeforeHeight && app.PausedHeight != types.HeightNotUsed {
			app.UnstakingHeight = unstakingHeight
			app.Status = int32(status)
			if err := m.SetAppUnstakingHeightAndStatus(app.Address, app.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := codec.Marshal(&app)
			if err != nil {
				return err
			}
			if err := db.Put(it.Key(), bz); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *PrePersistenceContext) SetAppPauseHeight(address []byte, height int64) error {
	codec := types.GetCodec()
	db := m.Store()
	app, err := m.GetApp(address, height)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	if height != types.HeightNotUsed {
		app.Paused = true
	} else {
		app.Paused = false
	}
	app.PausedHeight = height
	bz, err := codec.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(append(AppPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) GetAppOutputAddress(operator []byte, height int64) (output []byte, err error) {
	app, err := m.GetApp(operator, height)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return app.Output, nil
}
