package utility

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	utilTypes "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) HandleMessageStakeServiceNode(message *utilTypes.MessageStakeServiceNode) types.Error {
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
	// validate chains
	if len(message.Chains) > maxChains {
		return types.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(utilTypes.ServiceNodeStakePoolName, amount); err != nil {
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

func (u *UtilityContext) HandleMessageEditStakeServiceNode(message *utilTypes.MessageEditStakeServiceNode) types.Error {
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
	if err := u.AddPoolAmount(utilTypes.ServiceNodeStakePoolName, amountToAdd); err != nil {
		return err
	}
	// insert the serviceNode structure
	if err := u.UpdateServiceNode(message.Address, message.ServiceUrl, message.AmountToAdd, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeServiceNode(message *utilTypes.MessageUnstakeServiceNode) types.Error {
	status, err := u.GetServiceNodeStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != utilTypes.StakedStatus {
		return types.ErrInvalidStatus(status, utilTypes.StakedStatus)
	}
	unstakingHeight, err := u.CalculateServiceNodeUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetServiceNodeUnstakingHeightAndStatus(message.Address, unstakingHeight, utilTypes.UnstakingStatus); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeServiceNodesThatAreReady() types.Error {
	ServiceNodesReadyToUnstake, err := u.GetServiceNodesReadyToUnstake()
	if err != nil {
		return err
	}
	for _, serviceNode := range ServiceNodesReadyToUnstake {
		if err := u.SubPoolAmount(utilTypes.ServiceNodeStakePoolName, serviceNode.GetStakeAmount()); err != nil {
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

func (u *UtilityContext) HandleMessagePauseServiceNode(message *utilTypes.MessagePauseServiceNode) types.Error {
	height, err := u.GetServiceNodePauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if height != 0 {
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

func (u *UtilityContext) HandleMessageUnpauseServiceNode(message *utilTypes.MessageUnpauseServiceNode) types.Error {
	pausedHeight, err := u.GetServiceNodePauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == 0 {
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
	if err := u.SetServiceNodePauseHeight(message.Address, utilTypes.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetServiceNodeExists(address []byte) (exists bool, err types.Error) {
	store := u.Store()
	exists, er := store.GetServiceNodeExists(address)
	if er != nil {
		return false, types.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertServiceNode(address, publicKey, output []byte, serviceURL, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.InsertServiceNode(address, publicKey, output, false, utilTypes.StakedStatus, serviceURL, amount, chains, utilTypes.ZeroInt, utilTypes.ZeroInt)
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

func (u *UtilityContext) GetServiceNodesReadyToUnstake() (ServiceNodes []*types.UnstakingActor, err types.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingServiceNodes, er := store.GetServiceNodesReadyToUnstake(latestHeight, utilTypes.UnstakingStatus)
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
	er := store.SetServiceNodesStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, utilTypes.UnstakingStatus)
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetServiceNodeStatus(address []byte) (status int, err types.Error) {
	store := u.Store()
	status, er := store.GetServiceNodeStatus(address)
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) (err types.Error) {
	store := u.Store()
	if er := store.SetServiceNodeUnstakingHeightAndStatus(address, unstakingHeight, status); er != nil {
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetServiceNodePauseHeightIfExists(address []byte) (ServiceNodePauseHeight int64, err types.Error) {
	store := u.Store()
	ServiceNodePauseHeight, er := store.GetServiceNodePauseHeightIfExists(address)
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetPauseHeight(er)
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

func (u *UtilityContext) CalculateServiceNodeUnstakingHeight() (unstakingHeight int64, err types.Error) {
	unstakingBlocks, err := u.GetServiceNodeUnstakingBlocks()
	if err != nil {
		return utilTypes.ZeroInt, err
	}
	unstakingHeight, err = u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return utilTypes.ZeroInt, err
	}
	return
}

func (u *UtilityContext) GetServiceNodesPerSession(height int64) (int, types.Error) {
	store := u.Store()
	i, err := store.GetServiceNodesPerSessionAt(height)
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetServiceNodesPerSessionAt(height, err)
	}
	return i, nil
}

func (u *UtilityContext) GetServiceNodeCount(chain string, height int64) (int, types.Error) {
	store := u.Store()
	i, err := store.GetServiceNodeCount(chain, height)
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetServiceNodeCount(chain, height, err)
	}
	return i, nil
}

func (u *UtilityContext) GetMessageStakeServiceNodeSignerCandidates(msg *utilTypes.MessageStakeServiceNode) (candidates [][]byte, err types.Error) {
	candidates = append(candidates, msg.OutputAddress)
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates = append(candidates, pk.Address())
	return
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeSignerCandidates(msg *utilTypes.MessageEditStakeServiceNode) (candidates [][]byte, err types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeSignerCandidates(msg *utilTypes.MessageUnstakeServiceNode) (candidates [][]byte, err types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeSignerCandidates(msg *utilTypes.MessageUnpauseServiceNode) (candidates [][]byte, err types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessagePauseServiceNodeSignerCandidates(msg *utilTypes.MessagePauseServiceNode) (candidates [][]byte, err types.Error) {
	output, err := u.GetServiceNodeOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetServiceNodeOutputAddress(operator []byte) (output []byte, err types.Error) {
	store := u.Store()
	output, er := store.GetServiceNodeOutputAddress(operator)
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
