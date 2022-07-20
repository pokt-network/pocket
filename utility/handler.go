package utility

import (
	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"math/big"
)

type ActorStateChanges interface {
	Stake(message *typesUtil.Message) types.Error
	EditStake(message *typesUtil.Message) types.Error
	Unstake(message *typesUtil.Message) types.Error
	Pause(message *typesUtil.Message) types.Error
	Unpause(message *typesUtil.Message) types.Error
	Burn(address []byte, percentage int) types.Error
}

type ActorStore interface {
	// single actor actions (writes)
	Insert(address, publicKey, output []byte, serviceURL, amount string) types.Error
	Update(address []byte, serviceURL, amount string) types.Error
	Delete(address []byte) types.Error
	SetUnstakingHeight(address []byte, unstakingHeight int64) types.Error
	SetPauseHeight(address []byte, height int64) types.Error
	SetStakedTokens(address []byte, tokens *big.Int)
	// single actor actions (reads)
	GetPauseHeight(address []byte)
	GetStakingStatus(address []byte) (int, types.Error)
	GetStakedTokens(address []byte) (*big.Int, types.Error)
	GetOutputAddress(operator []byte) ([]byte, types.Error)
	// multi actor actions
	GetReadyToUnstake() ([]*types.UnstakingActor, types.Error)
	UnstakeMaxPaused(pausedBeforeHeight int64) types.Error
	UnstakeReadyActors() types.Error
	BeginUnstakingActors() types.Error
	// actor helpers
	CalculateUnstakingHeight() (int64, types.Error)
	GetMessageSignerCandidates(msg *typesUtil.MessageStakeValidator) (signers [][]byte, err types.Error)
}

type State interface {
	HandleProposalRewards() types.Error
}
