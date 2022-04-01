package genesis

import (
	"encoding/hex"

	crypto2 "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
)

type ValidatorSet []Validator

func (b *Block) ValidateBasic() types.Error { // TODO (Andrew) Consolidate block
	if err := b.BlockHeader.ValidateBasic(); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		if tx == nil {
			return types.EmptyTransactionErr()
		}
	}
	return nil
}

func (bh *BlockHeader) ValidateBasic() types.Error { // TODO (team) move this into shared
	if bh.NetworkId == "" {
		return types.ErrEmptyNetworkID()
	}
	if bh.Time.Seconds == 0 {
		return types.ErrEmptyTimestamp()
	}
	if int64(bh.NumTxs) > bh.TotalTxs {
		return types.ErrInvalidTransactionCount()
	}
	if bh.ProposerAddress == nil {
		return types.ErrEmptyProposer()
	}
	if _, err := crypto2.NewAddressFromBytes(bh.ProposerAddress); err != nil {
		return types.ErrNewAddressFromBytes(err)
	}
	hashBytes, err := hex.DecodeString(bh.Hash)
	if err != nil {
		return types.ErrHexDecodeFromString(err)
	}
	hashLen := len(hashBytes)
	if hashLen != crypto2.SHA3HashLen {
		return types.ErrInvalidHashLength(crypto2.ErrInvalidHashLen(hashLen))
	}
	hashBytes, err = hex.DecodeString(bh.PrevBlockHash)
	if err != nil {
		return types.ErrHexDecodeFromString(err)
	}
	hashLen = len(hashBytes)
	if hashLen != crypto2.SHA3HashLen {
		return types.ErrInvalidHashLength(crypto2.ErrInvalidHashLen(hashLen))
	}
	return nil
}
