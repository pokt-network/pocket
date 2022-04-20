package utility

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) HandleMessageStakeServiceNode(message *typesUtil.MessageStakeServiceNode) types.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return types.ErrNewPublicKeyFromBytes(er)
	}
	// ensure above minimum stake
	minStake, err := u.GetServiceNodeMinimumStake()
	if err != nil {
		return err
	}
	amount, err := types.StringToBigInt(message.Amount)
	if err != nil {
		return err
	}
	if types.BigIntLessThan(amount, minStake) {
		return types.ErrMinimumStake()
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.GetAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmountError()
	}
	maxChains, err := u.GetServiceNodeMaxChains()
	if err != nil {
		return err
	}
	// validate chain count
	if len(message.Chains) > maxChains {
		return types.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(typesUtil.ServiceNodeStakePoolName, amount); err != nil {
		return err
	}
	// ensure ServiceNode doesn't already exist
	exists, err := u.GetServiceNodeExists(publicKey.Address())
	if err != nil {
		return err
	}
	if exists {
		return types.ErrAlreadyExists()
	}
	// insert the ServiceNode structure
	if err := u.InsertServiceNode(publicKey.Address(), message.PublicKey, message.OutputAddress, message.ServiceUrl, message.Amount, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeServiceNode(message *typesUtil.MessageEditStakeServiceNode) types.Error {
	exists, err := u.GetServiceNodeExists(message.Address)
	if err != nil {
		return err
	}
	if !exists {
		return types.ErrNotExists()
	}
	amountToAdd, err := types.StringToBigInt(message.AmountToAdd)
	if err != nil {
		return err
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.GetAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amountToAdd)
	if signerAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmountError()
	}
	maxChains, err := u.GetServiceNodeMaxChains()
	if err != nil {
		return err
	}
	// validate chains
	if len(message.Chains) > maxChains {
		return types.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(typesUtil.ServiceNodeStakePoolName, amountToAdd); err != nil {
		return err
	}
	// insert the serviceNode structure
	if err := u.UpdateServiceNode(message.Address, message.ServiceUrl, message.AmountToAdd, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeServiceNode(message *typesUtil.MessageUnstakeServiceNode) types.Error {
	status, err := u.GetServiceNodeStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != typesUtil.StakedStatus {
		return types.ErrInvalidStatus(status, typesUtil.StakedStatus)
	}
	unstakingHeight, err := u.CalculateServiceNodeUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetServiceNodeUnstakingHeightAndStatus(message.Address, unstakingHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeServiceNodesThatAreReady() types.Error {
	serviceNodesReadyToUnstake, err := u.GetServiceNodesReadyToUnstake()
	if err != nil {
		return err
	}
	for _, serviceNode := range serviceNodesReadyToUnstake {
		if err := u.SubPoolAmount(typesUtil.ServiceNodeStakePoolName, serviceNode.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.AddAccountAmountString(serviceNode.GetOutputAddress(), serviceNode.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.DeleteServiceNode(serviceNode.GetAddress()); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPausedServiceNodes() types.Error {
	maxPausedBlocks, err := u.GetServiceNodeMaxPausedBlocks()
	if err != nil {
		return err
	}
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	beforeHeight := latestHeight - int64(maxPausedBlocks)
	if beforeHeight < 0 { // genesis edge case
		beforeHeight = 0
	}
	if err := u.UnstakeServiceNodesPausedBefore(beforeHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessagePauseServiceNode(message *typesUtil.MessagePauseServiceNode) types.Error {
	height, err := u.GetServiceNodePauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if height != typesUtil.HeightNotUsed {
		return types.ErrAlreadyPaused()
	}
	height, err = u.GetLatestHeight()
	if err != nil {
		return err
	}
	if err := u.SetServiceNodePauseHeight(message.Address, height); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnpauseServiceNode(message *typesUtil.MessageUnpauseServiceNode) types.Error {
	pausedHeight, err := u.GetServiceNodePauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == typesUtil.HeightNotUsed {
		return types.ErrNotPaused()
	}
	minPauseBlocks, err := u.GetServiceNodeMinimumPauseBlocks()
	if err != nil {
		return err
	}
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	if latestHeight < int64(minPauseBlocks)+pausedHeight {
		return types.ErrNotReadyToUnpause()
	}
	if err := u.SetServiceNodePauseHeight(message.Address, typesUtil.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetServiceNodeExists(address []byte) (bool, types.Error) {
	store := u.Store()
	exists, er := store.GetServiceNodeExists(address)
	if er != nil {
		return false, types.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertServiceNode(address, publicKey, output []byte, serviceURL, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.InsertServiceNode(address, publicKey, output, false, typesUtil.StakedStatus, serviceURL, amount, chains, typesUtil.ZeroInt, typesUtil.ZeroInt)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) UpdateServiceNode(address []byte, serviceURL, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.UpdateServiceNode(address, serviceURL, amount, chains)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) DeleteServiceNode(address []byte) types.Error {
	store := u.Store()
	if err := store.DeleteServiceNode(address); err != nil {
		return types.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) GetServiceNodesReadyToUnstake() ([]*types.UnstakingActor, types.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingServiceNodes, er := store.GetServiceNodesReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
	if er != nil {
		return nil, types.ErrGetReadyToUnstake(er)
	}
	return unstakingServiceNodes, nil
}

func (u *UtilityContext) UnstakeServiceNodesPausedBefore(pausedBeforeHeight int64) types.Error {
	store := u.Store()
	unstakingHeight, err := u.CalculateServiceNodeUnstakingHeight()
	if err != nil {
		return err
	}
	er := store.SetServiceNodesStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetServiceNodeStatus(address []byte) (int, types.Error) {
	store := u.Store()
	status, er := store.GetServiceNodeStatus(address)
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64) types.Error {
	store := u.Store()
	if er := store.SetServiceNodeUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus); er != nil {
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetServiceNodePauseHeightIfExists(address []byte) (int64, types.Error) {
	store := u.Store()
	ServiceNodePauseHeight, er := store.GetServiceNodePauseHeightIfExists(address)
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetPauseHeight(er)
	}
	return ServiceNodePauseHeight, nil
}

func (u *UtilityContext) SetServiceNodePauseHeight(address []byte, height int64) types.Error {
	store := u.Store()
	if err := store.SetServiceNodePauseHeight(address, height); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) CalculateServiceNodeUnstakingHeight() (int64, types.Error) {
	unstakingBlocks, err := u.GetServiceNodeUnstakingBlocks()
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	unstakingHeight, err := u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	return unstakingHeight, nil
}

func (u *UtilityContext) GetServiceNodesPerSession(height int64) (int, types.Error) {
	store := u.Store()
	i, err := store.GetServiceNodesPerSessionAt(height)
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetServiceNodesPerSessionAt(height, err)
	}
	return i, nil
}

func (u *UtilityContext) GetServiceNodeCount(chain string, height int64) (int, types.Error) {
	store := u.Store()
	i, err := store.GetServiceNodeCount(chain, height)
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetServiceNodeCount(chain, height, err)
	}
	return i, nil
}

func (u *UtilityContext) GetMessageStakeServiceNodeSignerCandidates(msg *typesUtil.MessageStakeServiceNode) ([][]byte, types.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, msg.OutputAddress)
	candidates = append(candidates, pk.Address())
	return candidates, nil
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeSignerCandidates(msg *typesUtil.MessageEditStakeServiceNode) ([][]byte, types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeSignerCandidates(msg *typesUtil.MessageUnstakeServiceNode) ([][]byte, types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeSignerCandidates(msg *typesUtil.MessageUnpauseServiceNode) ([][]byte, types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessagePauseServiceNodeSignerCandidates(msg *typesUtil.MessagePauseServiceNode) ([][]byte, types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetServiceNodeOutputAddress(operator []byte) ([]byte, types.Error) {
	store := u.Store()
	output, er := store.GetServiceNodeOutputAddress(operator)
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
