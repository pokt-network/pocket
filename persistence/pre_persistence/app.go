package pre_persistence

import (
	"bytes"
	"fmt"
	"math/big"

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

func (m *PrePersistenceContext) GetApp(address []byte) (app *App, err error) {
	app = &App{}
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

func (m *PrePersistenceContext) GetAllApps(height int64) (apps []*App, err error) {
	codec := GetCodec()
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
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		a := App{}
		if err := codec.Unmarshal(bz, &a); err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return
}

func (m *PrePersistenceContext) InsertApplication(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	if exists, _ := m.GetAppExists(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	codec := GetCodec()
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
	bz, err := codec.Marshal(&app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func AddAmounts(amountA, amountB string) (string, error) {
	bigA, err := StringToBigInt(amountA)
	if err != nil {
		return "", err
	}
	bigB, err := StringToBigInt(amountB)
	if err != nil {
		return "", err
	}
	result := new(big.Int).Add(bigA, bigB)
	return BigIntToString(result), nil
}

func (m *PrePersistenceContext) UpdateApplication(address []byte, maxRelaysToAdd string, amountToAdd string, chainsToUpdate []string) error {
	app, _ := m.GetApp(address)
	if app == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	codec := GetCodec()
	db := m.Store()
	key := append(AppPrefixKey, address...)
	// compute new values
	var err error
	app.StakedTokens, err = AddAmounts(app.StakedTokens, amountToAdd)
	if err != nil {
		return err
	}
	app.MaxRelays, err = AddAmounts(app.MaxRelays, maxRelaysToAdd)
	app.Chains = chainsToUpdate
	bz, err := codec.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) DeleteApplication(address []byte) error {
	exists, err := m.GetAppExists(address)
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
	app, err := m.GetApp(address)
	if err != nil {
		return ZeroInt, err
	}
	if app == nil {
		return ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(app.Status), nil
}

func (m *PrePersistenceContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	app, err := m.GetApp(address)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	codec := GetCodec()
	unstakingApps := types.UnstakingActors{}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	app.UnstakingHeight = unstakingHeight
	app.Status = int32(status)
	// marshal
	bz, err := codec.Marshal(app)
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
	unstakingBz, err := codec.Marshal(&unstakingApps)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *PrePersistenceContext) GetAppPauseHeightIfExists(address []byte) (int64, error) {
	app, err := m.GetApp(address)
	if err != nil {
		return ZeroInt, err
	}
	if app == nil {
		return ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(app.PausedHeight), nil
}

// SetAppsStatusAndUnstakingHeightPausedBefore : This unstakes the actors that have reached max pause height
func (m *PrePersistenceContext) SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	codec := GetCodec()
	it := db.NewIterator(&util.Range{
		Start: AppPrefixKey,
		Limit: PrefixEndBytes(AppPrefixKey),
	})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		app := App{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := codec.Unmarshal(bz, &app); err != nil {
			return err
		}
		if app.PausedHeight < uint64(pausedBeforeHeight) {
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
	codec := GetCodec()
	db := m.Store()
	app, err := m.GetApp(address)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	app.Paused = true
	app.PausedHeight = uint64(height)
	bz, err := codec.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(append(AppPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) GetAppOutputAddress(operator []byte) (output []byte, err error) {
	app, err := m.GetApp(operator)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return app.Output, nil
}
