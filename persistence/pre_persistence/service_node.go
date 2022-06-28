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

func (m *PrePersistenceContext) GetServiceNode(address []byte) (sn *typesGenesis.ServiceNode, exists bool, err error) {
	sn = &typesGenesis.ServiceNode{}
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
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
	if err = proto.Unmarshal(bz, sn); err != nil {
		return nil, true, err
	}
	return sn, true, nil
}

func (m *PrePersistenceContext) GetAllServiceNodes(height int64) (sns []*typesGenesis.ServiceNode, err error) {
	codec := types.GetCodec()
	sns = make([]*typesGenesis.ServiceNode, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: ServiceNodePrefixKey,
			Limit: PrefixEndBytes(ServiceNodePrefixKey),
		})
	} else {
		key := HeightKey(height, ServiceNodePrefixKey)
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
		sn := typesGenesis.ServiceNode{}
		if err := codec.Unmarshal(bz, &sn); err != nil {
			return nil, err
		}
		sns = append(sns, &sn)
	}
	return
}

func (m *PrePersistenceContext) InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	if _, exists, _ := m.GetServiceNode(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	codec := types.GetCodec()
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	sn := typesGenesis.ServiceNode{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pausedHeight),
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := codec.Marshal(&sn)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) UpdateServiceNode(address []byte, serviceURL string, amountToAdd string, chains []string) error {
	sn, exists, _ := m.GetServiceNode(address)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	codec := types.GetCodec()
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	// compute new values
	stakedTokens, err := types.StringToBigInt(sn.StakedTokens)
	if err != nil {
		return err
	}
	stakedTokensToAddI, err := types.StringToBigInt(amountToAdd)
	if err != nil {
		return err
	}
	stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	// update values
	sn.ServiceUrl = serviceURL
	sn.StakedTokens = types.BigIntToString(stakedTokens)
	sn.Chains = chains
	bz, err := codec.Marshal(sn)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) DeleteServiceNode(address []byte) error {
	if exists, _ := m.GetServiceNodeExists(address); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *PrePersistenceContext) GetServiceNodeExists(address []byte) (exists bool, err error) {
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
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

func (m *PrePersistenceContext) GetServiceNodesReadyToUnstake(height int64, status int) (ServiceNodes []*types.UnstakingActor, err error) {
	db := m.Store()
	unstakingKey := append(UnstakingServiceNodePrefixKey, types.Int64ToBytes(height)...)
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
		ServiceNodes = append(ServiceNodes, sn)
	}
	return
}

func (m *PrePersistenceContext) GetServiceNodeStatus(address []byte) (status int, err error) {
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(sn.Status), nil
}

func (m *PrePersistenceContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	codec := types.GetCodec()
	unstakingActors := types.UnstakingActors{}
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	sn.UnstakingHeight = unstakingHeight
	sn.Status = int32(status)
	bz, err := codec.Marshal(sn)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingServiceNodePrefixKey, types.Int64ToBytes(unstakingHeight)...)
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
		Address:       sn.Address,
		StakeAmount:   sn.StakedTokens,
		OutputAddress: sn.Output,
	})
	unstakingBz, err := codec.Marshal(&unstakingActors)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *PrePersistenceContext) GetServiceNodePauseHeightIfExists(address []byte) (int64, error) {
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(sn.PausedHeight), nil
}

// SetServiceNodeStatusAndUnstakingHeightPausedBefore : This unstakes the actors that have reached max pause height
func (m *PrePersistenceContext) SetServiceNodeStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	codec := types.GetCodec()
	it := db.NewIterator(&util.Range{
		Start: ServiceNodePrefixKey,
		Limit: PrefixEndBytes(ServiceNodePrefixKey),
	})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		sn := typesGenesis.ServiceNode{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := codec.Unmarshal(bz, &sn); err != nil {
			return err
		}
		if sn.PausedHeight < uint64(pausedBeforeHeight) {
			sn.UnstakingHeight = unstakingHeight
			sn.Status = int32(status)
			if err := m.SetServiceNodeUnstakingHeightAndStatus(sn.Address, sn.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := codec.Marshal(&sn)
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

func (m *PrePersistenceContext) SetServiceNodePauseHeight(address []byte, height int64) error {
	codec := types.GetCodec()
	db := m.Store()
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	if height == types.HeightNotUsed {
		sn.Paused = false
	} else {
		sn.Paused = true
	}
	sn.PausedHeight = uint64(height)
	bz, err := codec.Marshal(sn)
	if err != nil {
		return err
	}
	return db.Put(append(ServiceNodePrefixKey, address...), bz)
}

func (m *PrePersistenceContext) GetServiceNodesPerSessionAt(height int64) (int, error) {
	params, err := m.GetParams(height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ServiceNodesPerSession), nil
}

func (m *PrePersistenceContext) GetServiceNodeCount(chain string, height int64) (int, error) {
	codec := types.GetCodec()
	var it iterator.Iterator
	count := 0
	if m.Height == height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: ServiceNodePrefixKey,
			Limit: PrefixEndBytes(ServiceNodePrefixKey),
		})
	} else {
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: HeightKey(height, ServiceNodePrefixKey),
			Limit: HeightKey(height, PrefixEndBytes(ServiceNodePrefixKey)),
		})
	}
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		node := typesGenesis.ServiceNode{}
		if err := codec.Unmarshal(bz, &node); err != nil {
			return types.ZeroInt, err
		}
		for _, c := range node.Chains {
			if c == chain {
				count++
				break
			}
		}
	}
	return count, nil
}

func (m *PrePersistenceContext) GetServiceNodeOutputAddress(operator []byte) (output []byte, err error) {
	sn, exists, err := m.GetServiceNode(operator)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return sn.Output, nil
}
