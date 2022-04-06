package types

import (
	"encoding/hex"
	crypto2 "github.com/pokt-network/pocket/shared/crypto"
)

func (b *Block) ValidateBasic() Error {
	if err := b.BlockHeader.ValidateBasic(); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		if tx == nil {
			return EmptyTransactionErr()
		}
	}
	return nil
}

func (bh *BlockHeader) ValidateBasic() Error {
	if bh.NetworkId == "" {
		return ErrEmptyNetworkID()
	}
	if bh.Time.Seconds == 0 {
		return ErrEmptyTimestamp()
	}
	if int64(bh.NumTxs) > bh.TotalTxs {
		return ErrInvalidTransactionCount()
	}
	if bh.ProposerAddress == nil {
		return ErrEmptyProposer()
	}
	if _, err := crypto2.NewAddressFromBytes(bh.ProposerAddress); err != nil {
		return ErrNewAddressFromBytes(err)
	}
	hashBytes, err := hex.DecodeString(bh.Hash)
	if err != nil {
		return ErrHexDecodeFromString(err)
	}
	hashLen := len(hashBytes)
	if hashLen != crypto2.SHA3HashLen {
		return ErrInvalidHashLength(crypto2.ErrInvalidHashLen(hashLen))
	}
	hashBytes, err = hex.DecodeString(bh.LastBlockHash)
	if err != nil {
		return ErrHexDecodeFromString(err)
	}
	hashLen = len(hashBytes)
	if hashLen != crypto2.SHA3HashLen {
		return ErrInvalidHashLength(crypto2.ErrInvalidHashLen(hashLen))
	}
	if bh.QuorumCertificate == nil {
		return ErrNilQuorumCertificate()
	}
	return nil
}
