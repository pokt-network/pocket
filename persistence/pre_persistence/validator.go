package pre_persistence

import (
	"bytes"
	"fmt"

	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
)

func (m *PrePersistenceContext) GetValidator(address []byte) (val *Validator, exists bool, err error) {
	val = &Validator{}
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
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
	if err = proto.Unmarshal(bz, val); err != nil {
		return nil, true, err
	}
	return val, true, nil
}

func (m *PrePersistenceContext) GetAllValidators(height int64) (v []*Validator, err error) {
	codec := GetCodec()
	v = make([]*Validator, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: ValidatorPrefixKey,
			Limit: PrefixEndBytes(ValidatorPrefixKey),
		})
	} else {
		key := HeightKey(height, ValidatorPrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		bz := it.Value()
		//if bz == nil {
		//	break
		//}
		valid := it.Valid()
		valid = valid
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		validator := Validator{}
		if err := codec.Unmarshal(bz, &validator); err != nil {
			return nil, err
		}
		v = append(v, &validator)
	}
	return
}

func (m *PrePersistenceContext) GetValidatorExists(address []byte) (exists bool, err error) {
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
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

func (m *PrePersistenceContext) InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error {
	if _, exists, _ := m.GetFisherman(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	codec := GetCodec()
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	val := Validator{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		MissedBlocks:    0,
		PausedHeight:    uint64(pausedHeight),
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := codec.Marshal(&val)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) UpdateValidator(address []byte, serviceURL string, amountToAdd string) error {
	val, exists, _ := m.GetValidator(address)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	codec := GetCodec()
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	// compute new values
	stakedTokens, err := StringToBigInt(val.StakedTokens)
	if err != nil {
		return err
	}
	stakedTokensToAddI, err := StringToBigInt(amountToAdd)
	if err != nil {
		return err
	}
	stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	// update values
	val.ServiceUrl = serviceURL
	val.StakedTokens = BigIntToString(stakedTokens)
	// marshal
	bz, err := codec.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) DeleteValidator(address []byte) error {
	if exists, _ := m.GetValidatorExists(address); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *PrePersistenceContext) GetValidatorsReadyToUnstake(height int64, status int) (fishermen []*types.UnstakingActor, err error) {
	db := m.Store()
	unstakingKey := append(UnstakingValidatorPrefixKey, []byte(fmt.Sprintf("%d", height))...)
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
	for _, sn := range unstakingActors.UnstakingActors {
		fishermen = append(fishermen, sn)
	}
	return
}

func (m *PrePersistenceContext) GetValidatorStatus(address []byte) (status int, err error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return ZeroInt, err
	}
	if !exists {
		return ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(val.Status), nil
}

func (m *PrePersistenceContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	validator, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	codec := GetCodec()
	unstakingActors := types.UnstakingActors{}
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	validator.UnstakingHeight = unstakingHeight
	validator.Status = int32(status)
	// marshal
	bz, err := codec.Marshal(validator)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingValidatorPrefixKey, []byte(fmt.Sprintf("%d", unstakingHeight))...)
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
		Address:       validator.Address,
		StakeAmount:   validator.StakedTokens,
		OutputAddress: validator.Output,
	})
	unstakingBz, err := codec.Marshal(&unstakingActors)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *PrePersistenceContext) GetValidatorPauseHeightIfExists(address []byte) (int64, error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return ZeroInt, err
	}
	if !exists {
		return ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(val.PausedHeight), nil
}

// SetValidatorsStatusAndUnstakingHeightPausedBefore : This unstakes the actors that have reached max pause height
func (m *PrePersistenceContext) SetValidatorsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	codec := GetCodec()
	it := db.NewIterator(&util.Range{
		Start: ValidatorPrefixKey,
		Limit: PrefixEndBytes(ValidatorPrefixKey),
	})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		validator := Validator{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := codec.Unmarshal(bz, &validator); err != nil {
			return err
		}
		if validator.PausedHeight < uint64(pausedBeforeHeight) {
			validator.UnstakingHeight = unstakingHeight
			validator.Status = int32(status)
			if err := m.SetValidatorUnstakingHeightAndStatus(validator.Address, validator.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := codec.Marshal(&validator)
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

func (m *PrePersistenceContext) SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error {
	codec := GetCodec()
	db := m.Store()
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	val.PausedHeight = uint64(pauseHeight)
	val.MissedBlocks = uint32(missedBlocks)
	bz, err := codec.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(append(ValidatorPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) GetValidatorMissedBlocks(address []byte) (int, error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return ZeroInt, err
	}
	if !exists {
		return ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(val.MissedBlocks), nil
}

func (m *PrePersistenceContext) SetValidatorPauseHeight(address []byte, height int64) error {
	codec := GetCodec()
	db := m.Store()
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	val.Paused = true
	val.PausedHeight = uint64(height)
	bz, err := codec.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(append(ValidatorPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) SetValidatorStakedTokens(address []byte, tokens string) error {
	codec := GetCodec()
	db := m.Store()
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	val.StakedTokens = tokens
	bz, err := codec.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(append(ValidatorPrefixKey, address...), bz)
}

func (m *PrePersistenceContext) GetValidatorStakedTokens(address []byte) (tokens string, err error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return EmptyString, err
	}
	if !exists {
		return EmptyString, fmt.Errorf("does not exist in world state")
	}
	return val.StakedTokens, nil
}

func (m *PrePersistenceContext) GetValidatorOutputAddress(operator []byte) (output []byte, err error) {
	val, exists, err := m.GetValidator(operator)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return val.Output, nil
}
