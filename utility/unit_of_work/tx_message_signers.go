package unit_of_work

import (
	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// This file is responsible for retrieving the list of potential signers for a given message.
// This business logic is related to custodial, non-custodial, delegated, and shared stakes.
// RESEARCH(#751, #752): As the work related to delegated and undelegated stake becomes more well-defined,
// the logic in this will likely grow in complexity.

// IMPROVE: Consider return a slice of `crypto.Address` types instead of `[][]byte`
func (u *baseUtilityUnitOfWork) getSignerCandidates(msg typesUtil.Message) ([][]byte, coreTypes.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageSend:
		return u.getMessageSendSignerCandidates(x)
	case *typesUtil.MessageStake:
		return u.getMessageStakeSignerCandidates(x)
	case *typesUtil.MessageUnstake:
		return u.getMessageUnstakeSignerCandidates(x)
	case *typesUtil.MessageUnpause:
		return u.getMessageUnpauseSignerCandidates(x)
	case *typesUtil.MessageChangeParameter:
		return u.getMessageChangeParameterSignerCandidates(x)
	case *ibcTypes.UpdateIBCStore:
		return u.getUpdateIBCStoreSingerCandidates(x)
	case *ibcTypes.PruneIBCStore:
		return u.getPruneIBCStoreSingerCandidates(x)
	default:
		return nil, coreTypes.ErrUnknownMessage(x)
	}
}

func (u *baseUtilityUnitOfWork) getMessageStakeSignerCandidates(msg *typesUtil.MessageStake) ([][]byte, coreTypes.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, coreTypes.ErrNewPublicKeyFromBytes(er)
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, msg.OutputAddress, pk.Address())
	return candidates, nil
}

func (u *baseUtilityUnitOfWork) getMessageEditStakeSignerCandidates(msg *typesUtil.MessageEditStake) ([][]byte, coreTypes.Error) {
	output, err := u.getActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output, msg.Address)
	return candidates, nil
}

func (u *baseUtilityUnitOfWork) getMessageUnstakeSignerCandidates(msg *typesUtil.MessageUnstake) ([][]byte, coreTypes.Error) {
	output, err := u.getActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output, msg.Address)
	return candidates, nil
}

func (u *baseUtilityUnitOfWork) getMessageUnpauseSignerCandidates(msg *typesUtil.MessageUnpause) ([][]byte, coreTypes.Error) {
	output, err := u.getActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output, msg.Address)
	return candidates, nil
}

func (u *baseUtilityUnitOfWork) getMessageSendSignerCandidates(msg *typesUtil.MessageSend) ([][]byte, coreTypes.Error) {
	return [][]byte{msg.FromAddress}, nil
}

func (u *baseUtilityUnitOfWork) getUpdateIBCStoreSingerCandidates(msg *ibcTypes.UpdateIBCStore) ([][]byte, coreTypes.Error) {
	return [][]byte{msg.Signer}, nil
}

func (u *baseUtilityUnitOfWork) getPruneIBCStoreSingerCandidates(msg *ibcTypes.PruneIBCStore) ([][]byte, coreTypes.Error) {
	return [][]byte{msg.Signer}, nil
}
