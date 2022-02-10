package test

import (
	"encoding/hex"
	"github.com/pokt-network/utility-pre-prototype/shared/crypto"
	"github.com/pokt-network/utility-pre-prototype/utility/types"
)

type ValidatorSet []Validator

func (b *Block) ValidateBasic() types.Error {
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

func (bh *BlockHeader) ValidateBasic() types.Error {
	if bh.NetworkID == "" {
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
	if _, err := crypto.NewAddressFromBytes(bh.ProposerAddress); err != nil {
		return types.ErrNewAddressFromBytes(err)
	}
	hashBytes, err := hex.DecodeString(bh.Hash)
	if err != nil {
		return types.ErrHexDecodeFromString(err)
	}
	hashLen := len(hashBytes)
	if hashLen != crypto.SHA3HashLen {
		return types.ErrInvalidHashLength(crypto.ErrInvalidHashLen())
	}
	hashBytes, err = hex.DecodeString(bh.LastBlockHash)
	if err != nil {
		return types.ErrHexDecodeFromString(err)
	}
	hashLen = len(hashBytes)
	if hashLen != crypto.SHA3HashLen {
		return types.ErrInvalidHashLength(crypto.ErrInvalidHashLen())
	}
	return nil
}
