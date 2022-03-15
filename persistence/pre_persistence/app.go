package pre_persistence

import (
	"bytes"
	"fmt"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
)

func (m *PrePersistenceContext) GetAppExists(address []byte) (exists bool, err error) {
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

func (m *PrePersistenceContext) GetApp(address []byte) (app *App, exists bool, err error) {
	app = &App{}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	bz, err := db.Get(key)
	if err != nil {
		return nil, false, err
	}
	if bz == nil {
		return nil, false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return nil, false, nil
	}
	if err = proto.Unmarshal(bz, app); err != nil {
		return nil, true, err
	}
	return app, true, nil
}

func (m *PrePersistenceContext) GetAllApps(height int64) (apps []*App, err error) {
	cdc := Cdc()
	apps = make([]*App, 0)
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
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		a := App{}
		if err := cdc.Unmarshal(bz, &a); err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return
}

func (m *PrePersistenceContext) InsertApplication(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	if _, exists, _ := m.GetApp(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	cdc := Cdc()
	db := m.Store()
	key := append(AppPrefixKey, address...)
	app := App{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		Chains:          chains,
		MaxRelays:       maxRelays,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pausedHeight),
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := cdc.Marshal(&app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) UpdateApplication(address []byte, maxRelaysToAdd string, amountToAdd string, chainsToUpdate []string) error {
	app, exists, _ := m.GetApp(address)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := Cdc()
	db := m.Store()
	key := append(AppPrefixKey, address...)
	// compute new values
	stakedTokens, err := StringToBigInt(app.StakedTokens)
	if err != nil {
		return err
	}
	stakedTokensToAddI, err := StringToBigInt(amountToAdd)
	if err != nil {
		return err
	}
	stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	maxRelays, err := StringToBigInt(app.MaxRelays)
	if err != nil {
		return err
	}
	maxRelaysToAddI, err := StringToBigInt(maxRelaysToAdd)
	if err != nil {
		return err
	}
	maxRelays.Add(maxRelays, maxRelaysToAddI)
	// update values
	app.MaxRelays = BigIntToString(maxRelays)
	app.StakedTokens = BigIntToString(stakedTokens)
	app.Chains = chainsToUpdate
	// marshal
	bz, err := cdc.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) DeleteApplication(address []byte) error {
	if exists, _ := m.GetAppExists(address); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *PrePersistenceContext) GetAppsReadyToUnstake(height int64, status int) (apps []*types.UnstakingActor, err error) { // TODO delete unstaking
	db := m.Store()
	unstakingKey := append(UnstakingAppPrefixKey, []byte(fmt.Sprintf("%d", height))...)
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

func (m *PrePersistenceContext) GetAppStatus(address []byte) (status int, err error) {
	app, exists, err := m.GetApp(address)
	if err != nil {
		return ZeroInt, err
	}
	if !exists {
		return ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(app.Status), nil
}

func (m *PrePersistenceContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	app, exists, err := m.GetApp(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := Cdc()
	unstakingApps := types.UnstakingActors{}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	app.UnstakingHeight = unstakingHeight
	app.Status = int32(status)
	// marshal
	bz, err := cdc.Marshal(app)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingAppPrefixKey, []byte(fmt.Sprintf("%d", unstakingHeight))...)
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
	unstakingBz, err := cdc.Marshal(&unstakingApps)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *PrePersistenceContext) GetAppPauseHeightIfExists(address []byte) (int64, error) {
	app, exists, err := m.GetApp(address)
	if err != nil {
		return ZeroInt, err
	}
	if !exists {
		return ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(app.PausedHeight), nil
}

func (m *PrePersistenceContext) SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	cdc := Cdc()
	it := db.NewIterator(&util.Range{
		Start: AppPrefixKey,
		Limit: PrefixEndBytes(AppPrefixKey),
	})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		app := App{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := cdc.Unmarshal(bz, &app); err != nil {
			return err
		}
		if app.PausedHeight < uint64(pausedBeforeHeight) {
			app.UnstakingHeight = unstakingHeight
			app.Status = int32(status)
			if err := m.SetAppUnstakingHeightAndStatus(app.Address, app.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := cdc.Marshal(&app)
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
	cdc := Cdc()
	db := m.Store()
	app, exists, err := m.GetApp(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	if height == 0 {
		app.Paused = false
	} else {
		app.Paused = true
	}
	app.PausedHeight = uint64(height)
	bz, err := cdc.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(append(AppPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) GetAppOutputAddress(operator []byte) (output []byte, err error) {
	app, exists, err := m.GetApp(operator)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return app.Output, nil
}
