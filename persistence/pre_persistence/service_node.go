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

func (m *PrePersistenceContext) GetServiceNode(address []byte, height int64) (sn *typesGenesis.ServiceNode, exists bool, err error) {
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

func (m *PrePersistenceContext) GetServiceNodeStakeAmount(height int64, address []byte) (string, error) {
	sn, _, err := m.GetServiceNode(address, height)
	if err != nil {
		return "", err
	}
	return sn.StakedTokens, nil
}

func (m *PrePersistenceContext) SetServiceNodeStakeAmount(address []byte, stakeAmount string) error {
	codec := types.GetCodec()
	db := m.Store()
	sn, _, err := m.GetServiceNode(address, m.Height)
	if err != nil {
		return err
	}
	if sn == nil {
		return fmt.Errorf("does not exist in world state: %v", address)
	}
	sn.StakedTokens = stakeAmount
	bz, err := codec.Marshal(sn)
	if err != nil {
		return err
	}
	return db.Put(append(ServiceNodePrefixKey, address...), bz)
}

func (m *PrePersistenceContext) InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	if _, exists, _ := m.GetServiceNode(address, height); exists {
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
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := codec.Marshal(&sn)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) UpdateServiceNode(address []byte, serviceURL string, amount string, chains []string) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	sn, exists, _ := m.GetServiceNode(address, height)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	codec := types.GetCodec()
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	// compute new values
	//stakedTokens, err := types.StringToBigInt(sn.StakedTokens)
	//if err != nil {
	//	return err
	//}
	stakedTokens, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	//stakedTokens.Add(stakedTokens, stakedTokensToAddI)
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
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	if exists, _ := m.GetServiceNodeExists(address, height); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *PrePersistenceContext) GetServiceNodeExists(address []byte, height int64) (exists bool, err error) {
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

func (m *PrePersistenceContext) GetServiceNodeStatus(address []byte, height int64) (status int, err error) {
	sn, exists, err := m.GetServiceNode(address, height)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(sn.Status), nil
}

func (m *PrePersistenceContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	height, err := m.GetHeight()
	if err != nil {
		return err
	}
	sn, exists, err := m.GetServiceNode(address, height)
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

func (m *PrePersistenceContext) GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error) {
	sn, exists, err := m.GetServiceNode(address, height)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(sn.PausedHeight), nil
}

// SetServiceNodeStatusAndUnstakingHeightIfPausedBefore : This unstakes the actors that have reached max pause height
func (m *PrePersistenceContext) SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
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
		if sn.PausedHeight < pausedBeforeHeight && sn.PausedHeight != types.HeightNotUsed {
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
	sn, exists, err := m.GetServiceNode(address, height)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	if height != types.HeightNotUsed {
		sn.Paused = false
	} else {
		sn.Paused = true
	}
	sn.PausedHeight = height
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

func (m *PrePersistenceContext) GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error) {
	sn, exists, err := m.GetServiceNode(operator, height)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return sn.Output, nil
}
