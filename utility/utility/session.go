package utility

import (
	"github.com/pokt-network/utility-pre-prototype/shared/crypto"
	"github.com/pokt-network/utility-pre-prototype/utility/types"
	"math/big"
)

func (u *UtilityContext) GetSessionServiceNodesAddressAndServiceURL(chain string, appPublicKey []byte) ([]types.SessionNode, types.Error) {
	sessionBlockHeight, err := u.GetLatestSessionHeight()
	if err != nil {
		return nil, err
	}
	blockHash, err := u.GetBlockHash(sessionBlockHeight)
	if err != nil {
		return nil, err
	}
	nodesPerSession, err := u.GetServiceNodesPerSession(sessionBlockHeight)
	if err != nil {
		return nil, err
	}
	max, err := u.GetServiceNodeCount(chain, sessionBlockHeight)
	if err != nil {
		return nil, err
	}
	sessionKey := u.SessionKey(appPublicKey, blockHash, chain)
	// select service nodes
	indices := u.PseudorandomSelection(max, sessionKey, nodesPerSession)
	indices = indices
	// for each indices
	panic("TODO")
}

func (u *UtilityContext) GetLatestSessionHeight() (sessionBlockheight int64, err types.Error) {
	height, err := u.GetLatestHeight()
	if err != nil {
		return types.ZeroInt, err
	}
	// get the blocks per session
	bps, err := u.GetBlocksPerSession()
	if err != nil {
		return types.ZeroInt, err
	}
	blocksPerSession := int64(bps)
	if height%blocksPerSession == 0 {
		sessionBlockheight = height - blocksPerSession + 1
	} else {
		sessionBlockheight = height/blocksPerSession*blocksPerSession + 1
	}
	return
}

func (u *UtilityContext) PseudorandomSelection(max int, hash []byte, amount int) (indices []int64) {
	for i := 0; i < amount; i++ {
		// merkleHash for show and convert back to decimal
		intHash := new(big.Int).SetBytes(hash[:8])
		// mod the selection
		intHash.Mod(intHash, big.NewInt(int64(max)))
		// add to the indices
		indices = append(indices, intHash.Int64())
		// hash the hash
		hash = crypto.SHA3Hash(hash)
	}
	return
}

func (u *UtilityContext) SessionKey(appPubKey []byte, sessionBlockHash []byte, chain string) []byte {
	return crypto.SHA3Hash(append(appPubKey, append(sessionBlockHash, []byte(chain)...)...))
}
