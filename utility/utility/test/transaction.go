package test

import "pocket/utility/utility/types"

func NewTransaction(msg types.Message, fee string) (*types.Transaction, types.Error) {
	any, err := types.UtilityCodec().ToAny(msg)
	if err != nil {
		return nil, types.ErrProtoNewAny(err)
	}
	return &types.Transaction{
		Msg:   any,
		Fee:   fee,
		Nonce: types.RandBigInt().String(),
	}, nil
}
