package utility

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) HandleMessageStakeApp(message *typesUtil.MessageStakeApp) types.Error {
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
	if err := u.AddPoolAmount(typesUtil.AppStakePoolName, amount); err != nil {
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
	if err := u.InsertApp(publicKey.Address(), message.PublicKey, message.OutputAddress, maxRelays, message.Amount, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeApp(message *typesUtil.MessageEditStakeApp) types.Error {
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
	if err := u.AddPoolAmount(typesUtil.AppStakePoolName, amountToAdd); err != nil {
		return err
	}
	// calculate maximum relays from stake amount
	maxRelaysToAdd, err := u.CalculateAppRelays(message.AmountToAdd)
	if err != nil {
		return err
	}
	// insert the app structure
	if err := u.UpdateApp(message.Address, maxRelaysToAdd, message.AmountToAdd, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeApp(message *typesUtil.MessageUnstakeApp) types.Error {
	status, err := u.GetAppStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != typesUtil.StakedStatus {
		return types.ErrInvalidStatus(status, typesUtil.StakedStatus)
	}
	unstakingHeight, err := u.CalculateAppUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetAppUnstakingHeightAndStatus(message.Address, unstakingHeight); err != nil {
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
		if err := u.SubPoolAmount(typesUtil.AppStakePoolName, app.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.AddAccountAmountString(app.GetOutputAddress(), app.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.DeleteApp(app.GetAddress()); err != nil {
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

func (u *UtilityContext) HandleMessagePauseApp(message *typesUtil.MessagePauseApp) types.Error {
	height, err := u.GetAppPauseHeightIfExists(message.Address)
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
	if err := u.SetAppPauseHeight(message.Address, height); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnpauseApp(message *typesUtil.MessageUnpauseApp) types.Error {
	pausedHeight, err := u.GetAppPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == typesUtil.HeightNotUsed {
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
	if err := u.SetAppPauseHeight(message.Address, typesUtil.HeightNotUsed); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) CalculateAppRelays(stakedTokens string) (string, types.Error) {
	tokens, err := types.StringToBigInt(stakedTokens)
	if err != nil {
		return typesUtil.EmptyString, err
	}
	// The constant integer adjustment that the DAO may use to move the stake. The DAO may manually
	// adjust an application's MaxRelays at the time of staking to correct for short-term fluctuations
	// in the price of POKT, which may not be reflected in ParticipationRate
	// When this parameter is set to 0, no adjustment is being made.
	stabilityAdjustment, err := u.GetStabilityAdjustment()
	if err != nil {
		return typesUtil.EmptyString, err
	}
	baseRate, err := u.GetBaselineAppStakeRate()
	if err != nil {
		return typesUtil.EmptyString, err
	}
	// convert tokens to float64
	tokensFloat64 := big.NewFloat(float64(tokens.Int64()))
	// get the percentage of the baseline stake rate (can be over 100%)
	basePercentage := big.NewFloat(float64(baseRate) / float64(100))
	// multiply the two
	// TODO (team) evaluate whether or not we should use micro denomination or not
	baselineThroughput := basePercentage.Mul(basePercentage, tokensFloat64)
	// adjust for uPOKT
	baselineThroughput.Quo(baselineThroughput, big.NewFloat(typesUtil.MillionInt))
	// add staking adjustment (can be negative)
	adjusted := baselineThroughput.Add(baselineThroughput, big.NewFloat(float64(stabilityAdjustment)))
	// truncate the integer
	result, _ := adjusted.Int(nil)
	// bounding Max Amount of relays to maxint64
	max := big.NewInt(math.MaxInt64)
	if i := result.Cmp(max); i < -1 {
		result = max
	}
	return types.BigIntToString(result), nil
}

func (u *UtilityContext) GetAppExists(address []byte) (bool, types.Error) {
	store := u.Store()
	exists, er := store.GetAppExists(address)
	if er != nil {
		return false, types.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertApp(address, publicKey, output []byte, maxRelays, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.InsertApp(address, publicKey, output, false, typesUtil.StakedStatus, maxRelays, amount, chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

// TODO (Team) re-evaluate whether the delta should be here or the updated value
func (u *UtilityContext) UpdateApp(address []byte, maxRelays, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.UpdateApp(address, maxRelays, amount, chains)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) DeleteApp(address []byte) types.Error {
	store := u.Store()
	if err := store.DeleteApp(address); err != nil {
		return types.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) GetAppsReadyToUnstake() ([]*types.UnstakingActor, types.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingApps, er := store.GetAppsReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
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
	er := store.SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetAppStatus(address []byte) (int, types.Error) {
	store := u.Store()
	status, er := store.GetAppStatus(address)
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64) types.Error {
	store := u.Store()
	if er := store.SetAppUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus); er != nil { // TODO (Andrew) remove unstaking status from prepersistence
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetAppPauseHeightIfExists(address []byte) (int64, types.Error) {
	store := u.Store()
	appPauseHeight, er := store.GetAppPauseHeightIfExists(address)
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetPauseHeight(er)
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

func (u *UtilityContext) CalculateAppUnstakingHeight() (int64, types.Error) {
	unstakingBlocks, err := u.GetAppUnstakingBlocks()
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	unstakingHeight, err := u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	return unstakingHeight, nil
}

func (u *UtilityContext) GetMessageStakeAppSignerCandidates(msg *typesUtil.MessageStakeApp) ([][]byte, types.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, msg.OutputAddress)
	candidates = append(candidates, pk.Address())
	return candidates, nil
}

func (u *UtilityContext) GetMessageEditStakeAppSignerCandidates(msg *typesUtil.MessageEditStakeApp) ([][]byte, types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnstakeAppSignerCandidates(msg *typesUtil.MessageUnstakeApp) ([][]byte, types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnpauseAppSignerCandidates(msg *typesUtil.MessageUnpauseApp) ([][]byte, types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessagePauseAppSignerCandidates(msg *typesUtil.MessagePauseApp) ([][]byte, types.Error) {
	output, err := u.GetAppOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetAppOutputAddress(operator []byte) ([]byte, types.Error) {
	store := u.Store()
	output, er := store.GetAppOutputAddress(operator)
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
