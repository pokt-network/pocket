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

func (m *PrePersistenceContext) GetFishermanExists(address []byte, height int64) (exists bool, err error) {
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
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

func (m *PrePersistenceContext) GetFisherman(address []byte, height int64) (fish *typesGenesis.Fisherman, exists bool, err error) {
	fish = &typesGenesis.Fisherman{}
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
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
	if err = proto.Unmarshal(bz, fish); err != nil {
		return nil, true, err
	}
	return fish, true, nil
}

func (m *PrePersistenceContext) GetAllFishermen(height int64) (fishermen []*typesGenesis.Fisherman, err error) {
	codec := types.GetCodec()
	fishermen = make([]*typesGenesis.Fisherman, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: FishermanPrefixKey,
			Limit: PrefixEndBytes(FishermanPrefixKey),
		})
	} else {
		key := HeightKey(height, FishermanPrefixKey)
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
		fish := typesGenesis.Fisherman{}
		if err := codec.Unmarshal(bz, &fish); err != nil {
			return nil, err
		}
		fishermen = append(fishermen, &fish)
	}
	return
}

func (m *PrePersistenceContext) GetFishermanStakeAmount(height int64, address []byte) (string, error) {
	fish, _, err := m.GetFisherman(address, height)
	if err != nil {
		return "", err
	}
	return fish.StakedTokens, nil
}

func (m *PrePersistenceContext) SetFishermanStakeAmount(address []byte, stakeAmount string) error {
	codec := types.GetCodec()
	db := m.Store()
	fish, _, err := m.GetFisherman(address, m.Height)
	if err != nil {
		return err
	}
	if fish == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	fish.StakedTokens = stakeAmount
	bz, err := codec.Marshal(fish)
	if err != nil {
		return err
	}
	return db.Put(append(FishermanPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedAmount string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	if _, exists, _ := m.GetFisherman(address, height); exists {
		return fmt.Errorf("already exists in world state")
	}
	codec := types.GetCodec()
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	fish := typesGenesis.Fisherman{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedAmount,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := codec.Marshal(&fish)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) UpdateFisherman(address []byte, serviceURL string, amount string, chains []string) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	fish, exists, _ := m.GetFisherman(address, height)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	codec := types.GetCodec()
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	// compute new values
	//stakedTokens, err := types.StringToBigInt(fish.StakedTokens)
	//if err != nil {
	//	return err
	//}
	stakedTokens, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	//stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	// update values
	fish.ServiceUrl = serviceURL
	fish.StakedTokens = types.BigIntToString(stakedTokens)
	fish.Chains = chains
	bz, err := codec.Marshal(fish)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) DeleteFisherman(address []byte) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	if exists, _ := m.GetFishermanExists(address, height); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *PrePersistenceContext) GetFishermenReadyToUnstake(height int64, status int) (fisherman []*types.UnstakingActor, err error) {
	db := m.Store()
	unstakingKey := append(UnstakingFishermanPrefixKey, types.Int64ToBytes(height)...)
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
	unstakingActors := types.UnstakingActors{}
	if err := proto.Unmarshal(val, &unstakingActors); err != nil {
		return nil, err
	}
	fisherman = append(fisherman, unstakingActors.UnstakingActors...)
	return
}

func (m *PrePersistenceContext) GetFishermanStatus(address []byte, height int64) (status int, err error) {
	fish, exists, err := m.GetFisherman(address, height)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(fish.Status), nil
}

func (m *PrePersistenceContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	fish, exists, err := m.GetFisherman(address, height)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	codec := types.GetCodec()
	unstakingActors := types.UnstakingActors{}
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	fish.UnstakingHeight = unstakingHeight
	fish.Status = int32(status)
	bz, err := codec.Marshal(fish)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingFishermanPrefixKey, types.Int64ToBytes(unstakingHeight)...)
	if found := db.Contains(unstakingKey); found {
		val, err := db.Get(unstakingKey)
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(val, &unstakingActors); err != nil {
			return err
		}
	}
	unstakingActors.UnstakingActors = append(unstakingActors.UnstakingActors, &types.UnstakingActor{
		Address:       fish.Address,
		StakeAmount:   fish.StakedTokens,
		OutputAddress: fish.Output,
	})
	unstakingBz, err := codec.Marshal(&unstakingActors)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *PrePersistenceContext) GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error) {
	fish, exists, err := m.GetFisherman(address, height)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(fish.PausedHeight), nil
}

// SetFishermanStatusAndUnstakingHeightIfPausedBefore : This unstakes the actors that have reached max pause height
func (m *PrePersistenceContext) SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	codec := types.GetCodec()
	it := db.NewIterator(&util.Range{
		Start: FishermanPrefixKey,
		Limit: PrefixEndBytes(FishermanPrefixKey),
	})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		fish := typesGenesis.Fisherman{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := codec.Unmarshal(bz, &fish); err != nil {
			return err
		}
		if fish.PausedHeight < pausedBeforeHeight && fish.PausedHeight != types.HeightNotUsed {
			fish.UnstakingHeight = unstakingHeight
			fish.Status = int32(status)
			if err := m.SetFishermanUnstakingHeightAndStatus(fish.Address, fish.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := codec.Marshal(&fish)
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

func (m *PrePersistenceContext) SetFishermanPauseHeight(address []byte, height int64) error {
	codec := types.GetCodec()
	db := m.Store()
	fish, exists, err := m.GetFisherman(address, height)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	if height != types.HeightNotUsed {
		fish.Paused = true
	} else {
		fish.Paused = false
	}
	fish.PausedHeight = height
	bz, err := codec.Marshal(fish)
	if err != nil {
		return err
	}
	return db.Put(append(FishermanPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error) {
	fish, exists, err := m.GetFisherman(operator, height)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return fish.Output, nil
}
