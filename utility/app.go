package utility

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	utilTypes "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) HandleMessageStakeApp(message *utilTypes.MessageStakeApp) types.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return types.ErrNewPublicKeyFromBytes(er)
	}
	// ensure above minimum stake
	minStake, err := u.GetAppMinimumStake()
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
	maxChains, err := u.GetAppMaxChains()
	if err != nil {
		return err
	}
	// validate number of chains
	if len(message.Chains) > maxChains {
		return types.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(utilTypes.AppStakePoolName, amount); err != nil {
		return err
	}
	// calculate maximum relays from stake amount
	maxRelays, err := u.CalculateAppRelays(message.Amount)
	if err != nil {
		return err
	}
	// ensure app doesn't already exist
	exists, err := u.GetAppExists(publicKey.Address())
	if err != nil {
		return err
	}
	if exists {
		return types.ErrAlreadyExists()
	}
	// insert the app structure
	if err := u.InsertApplication(publicKey.Address(), message.PublicKey, message.OutputAddress, maxRelays, message.Amount, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeApp(message *utilTypes.MessageEditStakeApp) types.Error {
	exists, err := u.GetAppExists(message.Address)
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
	maxChains, err := u.GetAppMaxChains()
	if err != nil {
		return err
	}
	// validate number of chains
	if len(message.Chains) > maxChains {
		return types.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(utilTypes.AppStakePoolName, amountToAdd); err != nil {
		return err
	}
	// calculate maximum relays from stake amount
	maxRelaysToAdd, err := u.CalculateAppRelays(message.AmountToAdd)
	if err != nil {
		return err
	}
	// insert the app structure
	if err := u.UpdateApplication(message.Address, maxRelaysToAdd, message.AmountToAdd, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeApp(message *utilTypes.MessageUnstakeApp) types.Error {
	status, err := u.GetAppStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != utilTypes.StakedStatus {
		return types.ErrInvalidStatus(status, utilTypes.StakedStatus)
	}
	unstakingHeight, err := u.CalculateAppUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetAppUnstakingHeightAndStatus(message.Address, unstakingHeight, utilTypes.UnstakingStatus); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeAppsThatAreReady() types.Error {
	appsReadyToUnstake, err := u.GetAppsReadyToUnstake()
	if err != nil {
		return err
	}
	for _, app := range appsReadyToUnstake {
		if err := u.SubPoolAmount(utilTypes.AppStakePoolName, app.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.AddAccountAmountString(app.GetOutputAddress(), app.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.DeleteApplication(app.GetAddress()); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPausedApps() types.Error {
	maxPausedBlocks, err := u.GetAppMaxPausedBlocks()
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
	if err := u.UnstakeAppsPausedBefore(beforeHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessagePauseApp(message *utilTypes.MessagePauseApp) types.Error {
	height, err := u.GetAppPauseHeightIfExists(message.Address)
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
	if err := u.SetAppPauseHeight(message.Address, height); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnpauseApp(message *utilTypes.MessageUnpauseApp) types.Error {
	pausedHeight, err := u.GetAppPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == 0 {
		return types.ErrNotPaused()
	}
	minPauseBlocks, err := u.GetAppMinimumPauseBlocks()
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
	if err := u.SetAppPauseHeight(message.Address, utilTypes.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) CalculateAppRelays(stakedTokens string) (string, types.Error) {
	tokens, err := types.StringToBigInt(stakedTokens)
	if err != nil {
		return utilTypes.EmptyString, err
	}
	stakingAdjustment, err := u.GetStakingAdjustment()
	if err != nil {
		return utilTypes.EmptyString, err
	}
	baseRate, err := u.GetBaselineAppStakeRate()
	if err != nil {
		return utilTypes.EmptyString, err
	}
	// convert tokens to float64
	tokensFloat64 := big.NewFloat(float64(tokens.Int64()))
	// get the percentage of the baseline stake rate (can be over 100%)
	basePercentage := big.NewFloat(float64(baseRate) / float64(100))
	// multiply the two
	baselineThroughput := basePercentage.Mul(basePercentage, tokensFloat64)
	// adjust for uPOKT
	baselineThroughput.Quo(baselineThroughput, big.NewFloat(1000000))
	// add staking adjustment (can be negative)
	adjusted := baselineThroughput.Add(baselineThroughput, big.NewFloat(float64(stakingAdjustment)))
	// truncate the integer
	result, _ := adjusted.Int(nil)
	// bounding Max Amount of relays to maxint64
	max := big.NewInt(math.MaxInt64)
	if i := result.Cmp(max); i < -1 {
		result = max
	}
	return types.BigIntToString(result), nil
}

func (u *UtilityContext) GetAppExists(address []byte) (exists bool, err types.Error) {
	store := u.Store()
	exists, er := store.GetAppExists(address)
	if er != nil {
		return false, types.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertApplication(address, publicKey, output []byte, maxRelays, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.InsertApplication(address, publicKey, output, false, utilTypes.StakedStatus, maxRelays, amount, chains, utilTypes.ZeroInt, utilTypes.ZeroInt)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) UpdateApplication(address []byte, maxRelays, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.UpdateApplication(address, maxRelays, amount, chains)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) DeleteApplication(address []byte) types.Error {
	store := u.Store()
	if err := store.DeleteApplication(address); err != nil {
		return types.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) GetAppsReadyToUnstake() (apps []*types.UnstakingActor, err types.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingApps, er := store.GetAppsReadyToUnstake(latestHeight, utilTypes.UnstakingStatus)
	if er != nil {
		return nil, types.ErrGetReadyToUnstake(er)
	}
	return unstakingApps, nil
}

func (u *UtilityContext) UnstakeAppsPausedBefore(pausedBeforeHeight int64) types.Error {
	store := u.Store()
	unstakingHeight, err := u.CalculateAppUnstakingHeight()
	if err != nil {
		return err
	}
	er := store.SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, utilTypes.UnstakingStatus)
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetAppStatus(address []byte) (status int, err types.Error) {
	store := u.Store()
	status, er := store.GetAppStatus(address)
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) (err types.Error) {
	store := u.Store()
	if er := store.SetAppUnstakingHeightAndStatus(address, unstakingHeight, status); er != nil {
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetAppPauseHeightIfExists(address []byte) (appPauseHeight int64, err types.Error) {
	store := u.Store()
	appPauseHeight, er := store.GetAppPauseHeightIfExists(address)
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetPauseHeight(er)
	}
	return appPauseHeight, nil
}

func (u *UtilityContext) SetAppPauseHeight(address []byte, height int64) types.Error {
	store := u.Store()
	if err := store.SetAppPauseHeight(address, height); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) CalculateAppUnstakingHeight() (unstakingHeight int64, err types.Error) {
	unstakingBlocks, err := u.GetAppUnstakingBlocks()
	if err != nil {
		return utilTypes.ZeroInt, err
	}
	unstakingHeight, err = u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return utilTypes.ZeroInt, err
	}
	return
}

func (u *UtilityContext) GetMessageStakeAppSignerCandidates(msg *utilTypes.MessageStakeApp) (candidates [][]byte, err types.Error) {
	candidates = append(candidates, msg.OutputAddress)
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates = append(candidates, pk.Address())
	return
}

func (u *UtilityContext) GetMessageEditStakeAppSignerCandidates(msg *utilTypes.MessageEditStakeApp) (candidates [][]byte, err types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnstakeAppSignerCandidates(msg *utilTypes.MessageUnstakeApp) (candidates [][]byte, err types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnpauseAppSignerCandidates(msg *utilTypes.MessageUnpauseApp) (candidates [][]byte, err types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessagePauseAppSignerCandidates(msg *utilTypes.MessagePauseApp) (candidates [][]byte, err types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetAppOutputAddress(operator []byte) (output []byte, err types.Error) {
	store := u.Store()
	output, er := store.GetAppOutputAddress(operator)
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
