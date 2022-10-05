package types

import (
	"github.com/pokt-network/pocket/shared/codec"
)

func (b *Block) Bytes() ([]byte, error) {
	codec := codec.GetCodec()
	blockProtoBz, err := codec.Marshal(b)
	if err != nil {
		return nil, err
	}
	return blockProtoBz, nil
}
