package test_artifacts

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file contains *seemingly* necessary implementations of the shared interfaces in order to
// do things like: create a genesis file and perform testing where cross module implementations are needed
// (see call to action in generator.go to try to remove the cross module testing code)

// TODO (Team) convert these to proper mocks
var _ modules.Actor = &MockActor{}
var _ modules.Account = &MockAcc{}

type MockAcc struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

func (m *MockAcc) GetAddress() string {
	return m.Address
}

func (m *MockAcc) GetAmount() string {
	return m.Amount
}

type MockActor struct {
	Address         string        `json:"address"`
	PublicKey       string        `json:"public_key"`
	Chains          []string      `json:"chains"`
	GenericParam    string        `json:"generic_param"`
	StakedAmount    string        `json:"staked_amount"`
	PausedHeight    int64         `json:"paused_height"`
	UnstakingHeight int64         `json:"unstaking_height"`
	Output          string        `json:"output"`
	ActorType       MockActorType `json:"actor_type"`
}

const (
	MockActorType_App  MockActorType = 0
	MockActorType_Node MockActorType = 1
	MockActorType_Fish MockActorType = 2
	MockActorType_Val  MockActorType = 3
)

func (m *MockActor) GetAddress() string {
	return m.Address
}

func (m *MockActor) GetPublicKey() string {
	return m.PublicKey
}

func (m *MockActor) GetChains() []string {
	return m.Chains
}

func (m *MockActor) GetGenericParam() string {
	return m.GenericParam
}

func (m *MockActor) GetStakedAmount() string {
	return m.StakedAmount
}

func (m *MockActor) GetPausedHeight() int64 {
	return m.PausedHeight
}

func (m *MockActor) GetUnstakingHeight() int64 {
	return m.UnstakingHeight
}

func (m *MockActor) GetOutput() string {
	return m.Output
}

func (m *MockActor) GetActorTyp() modules.ActorType {
	return m.ActorType
}

type MockActorType int32

func (m MockActorType) String() string {
	return fmt.Sprintf("%d", m)
}

var _ modules.PersistenceGenesisState = &MockPersistenceGenesisState{}
var _ modules.ConsensusGenesisState = &MockConsensusGenesisState{}

type MockConsensusGenesisState struct {
	GenesisTime   *timestamppb.Timestamp `json:"genesis_time"`
	ChainId       string                 `json:"chain_id"`
	MaxBlockBytes uint64                 `json:"max_block_bytes"`
	Validators    []modules.Actor        `json:"validators"`
}

func (m *MockConsensusGenesisState) GetGenesisTime() *timestamppb.Timestamp {
	return m.GenesisTime
}

func (m *MockConsensusGenesisState) GetChainId() string {
	return m.ChainId
}

func (m *MockConsensusGenesisState) GetMaxBlockBytes() uint64 {
	return m.MaxBlockBytes
}

type MockPersistenceGenesisState struct {
	Accounts     []modules.Account `json:"accounts"`
	Pools        []modules.Account `json:"pools"`
	Validators   []modules.Actor   `json:"validators"`
	Applications []modules.Actor   `json:"applications"`
	ServiceNodes []modules.Actor   `json:"service_nodes"`
	Fishermen    []modules.Actor   `json:"fishermen"`
	Params       modules.Params    `json:"params"`
}

func (m MockPersistenceGenesisState) GetAccs() []modules.Account {
	return m.Accounts
}

func (m MockPersistenceGenesisState) GetAccPools() []modules.Account {
	return m.Pools
}

func (m MockPersistenceGenesisState) GetApps() []modules.Actor {
	return m.Applications
}

func (m MockPersistenceGenesisState) GetVals() []modules.Actor {
	return m.Validators
}

func (m MockPersistenceGenesisState) GetFish() []modules.Actor {
	return m.Fishermen
}

func (m MockPersistenceGenesisState) GetNodes() []modules.Actor {
	return m.ServiceNodes
}

func (m MockPersistenceGenesisState) GetParameters() modules.Params {
	return m.Params
}

var _ modules.ConsensusConfig = &MockConsensusConfig{}
var _ modules.PacemakerConfig = &MockPacemakerConfig{}
var _ modules.PersistenceConfig = &MockPersistenceConfig{}

type MockPersistenceConfig struct {
	PostgresUrl    string `json:"postgres_url"`
	NodeSchema     string `json:"node_schema"`
	BlockStorePath string `json:"block_store_path"`
}

func (m *MockPersistenceConfig) GetPostgresUrl() string {
	return m.PostgresUrl
}

func (m *MockPersistenceConfig) GetNodeSchema() string {
	return m.NodeSchema
}

func (m *MockPersistenceConfig) GetBlockStorePath() string {
	return m.BlockStorePath
}

type MockConsensusConfig struct {
	MaxMempoolBytes uint64               `json:"max_mempool_bytes"`
	PacemakerConfig *MockPacemakerConfig `json:"pacemaker_config"`
	PrivateKey      string               `json:"private_key"`
}

func (m *MockConsensusConfig) GetMaxMempoolBytes() uint64 {
	return m.MaxMempoolBytes
}

func (m *MockConsensusConfig) GetPaceMakerConfig() modules.PacemakerConfig {
	return m.PacemakerConfig
}

type MockPacemakerConfig struct {
	TimeoutMsec               uint64 `json:"timeout_msec"`
	Manual                    bool   `json:"manual"`
	DebugTimeBetweenStepsMsec uint64 `json:"debug_time_between_steps_msec"`
}

func (m *MockPacemakerConfig) SetTimeoutMsec(u uint64) {
	m.TimeoutMsec = u
}

func (m *MockPacemakerConfig) GetTimeoutMsec() uint64 {
	return m.TimeoutMsec
}

func (m *MockPacemakerConfig) GetManual() bool {
	return m.Manual
}

func (m *MockPacemakerConfig) GetDebugTimeBetweenStepsMsec() uint64 {
	return m.DebugTimeBetweenStepsMsec
}

var _ modules.P2PConfig = &MockP2PConfig{}

type MockP2PConfig struct {
	ConsensusPort         uint32 `json:"consensus_port"`
	UseRainTree           bool   `json:"use_rain_tree"`
	IsEmptyConnectionType bool   `json:"is_empty_connection_type"`
	PrivateKey            string `json:"private_key"`
}

func (m *MockP2PConfig) GetConsensusPort() uint32 {
	return m.ConsensusPort
}

func (m *MockP2PConfig) GetUseRainTree() bool {
	return m.UseRainTree
}

func (m *MockP2PConfig) IsEmptyConnType() bool {
	return m.IsEmptyConnectionType
}

var _ modules.TelemetryConfig = &MockTelemetryConfig{}

type MockTelemetryConfig struct {
	Enabled  bool   `json:"enabled"`
	Address  string `json:"address"`
	Endpoint string `json:"endpoint"`
}

func (m *MockTelemetryConfig) GetEnabled() bool {
	return m.Enabled
}

func (m *MockTelemetryConfig) GetAddress() string {
	return m.Address
}

func (m *MockTelemetryConfig) GetEndpoint() string {
	return m.Endpoint
}

type MockUtilityConfig struct{}

var _ modules.RPCConfig = &MockRPCConfig{}

type MockRPCConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Timeout uint64 `json:"timeout"`
}

func (m *MockRPCConfig) GetEnabled() bool {
	return m.Enabled
}

func (m *MockRPCConfig) GetPort() string {
	return m.Port
}

func (m *MockRPCConfig) GetTimeout() uint64 {
	return m.Timeout
}

var _ modules.Params = &MockParams{}

type MockParams struct {
	//@gotags: pokt:"val_type=BIGINT"
	BlocksPerSession int32 `protobuf:"varint,1,opt,name=blocks_per_session,json=blocksPerSession,proto3" json:"blocks_per_session,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppMinimumStake string `protobuf:"bytes,2,opt,name=app_minimum_stake,json=appMinimumStake,proto3" json:"app_minimum_stake,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	AppMaxChains int32 `protobuf:"varint,3,opt,name=app_max_chains,json=appMaxChains,proto3" json:"app_max_chains,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	AppBaselineStakeRate int32 `protobuf:"varint,4,opt,name=app_baseline_stake_rate,json=appBaselineStakeRate,proto3" json:"app_baseline_stake_rate,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	AppStakingAdjustment int32 `protobuf:"varint,5,opt,name=app_staking_adjustment,json=appStakingAdjustment,proto3" json:"app_staking_adjustment,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	AppUnstakingBlocks int32 `protobuf:"varint,6,opt,name=app_unstaking_blocks,json=appUnstakingBlocks,proto3" json:"app_unstaking_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	AppMinimumPauseBlocks int32 `protobuf:"varint,7,opt,name=app_minimum_pause_blocks,json=appMinimumPauseBlocks,proto3" json:"app_minimum_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	AppMaxPauseBlocks int32 `protobuf:"varint,8,opt,name=app_max_pause_blocks,json=appMaxPauseBlocks,proto3" json:"app_max_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ServiceNodeMinimumStake string `protobuf:"bytes,9,opt,name=service_node_minimum_stake,json=serviceNodeMinimumStake,proto3" json:"service_node_minimum_stake,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ServiceNodeMaxChains int32 `protobuf:"varint,10,opt,name=service_node_max_chains,json=serviceNodeMaxChains,proto3" json:"service_node_max_chains,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	ServiceNodeUnstakingBlocks int32 `protobuf:"varint,11,opt,name=service_node_unstaking_blocks,json=serviceNodeUnstakingBlocks,proto3" json:"service_node_unstaking_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ServiceNodeMinimumPauseBlocks int32 `protobuf:"varint,12,opt,name=service_node_minimum_pause_blocks,json=serviceNodeMinimumPauseBlocks,proto3" json:"service_node_minimum_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	ServiceNodeMaxPauseBlocks int32 `protobuf:"varint,13,opt,name=service_node_max_pause_blocks,json=serviceNodeMaxPauseBlocks,proto3" json:"service_node_max_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ServiceNodesPerSession int32 `protobuf:"varint,14,opt,name=service_nodes_per_session,json=serviceNodesPerSession,proto3" json:"service_nodes_per_session,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	FishermanMinimumStake string `protobuf:"bytes,15,opt,name=fisherman_minimum_stake,json=fishermanMinimumStake,proto3" json:"fisherman_minimum_stake,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	FishermanMaxChains int32 `protobuf:"varint,16,opt,name=fisherman_max_chains,json=fishermanMaxChains,proto3" json:"fisherman_max_chains,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	FishermanUnstakingBlocks int32 `protobuf:"varint,17,opt,name=fisherman_unstaking_blocks,json=fishermanUnstakingBlocks,proto3" json:"fisherman_unstaking_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	FishermanMinimumPauseBlocks int32 `protobuf:"varint,18,opt,name=fisherman_minimum_pause_blocks,json=fishermanMinimumPauseBlocks,proto3" json:"fisherman_minimum_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	FishermanMaxPauseBlocks int32 `protobuf:"varint,19,opt,name=fisherman_max_pause_blocks,json=fishermanMaxPauseBlocks,proto3" json:"fisherman_max_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ValidatorMinimumStake string `protobuf:"bytes,20,opt,name=validator_minimum_stake,json=validatorMinimumStake,proto3" json:"validator_minimum_stake,omitempty"`
	//@gotags: pokt:"val_type=BIGINT"
	ValidatorUnstakingBlocks int32 `protobuf:"varint,21,opt,name=validator_unstaking_blocks,json=validatorUnstakingBlocks,proto3" json:"validator_unstaking_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ValidatorMinimumPauseBlocks int32 `protobuf:"varint,22,opt,name=validator_minimum_pause_blocks,json=validatorMinimumPauseBlocks,proto3" json:"validator_minimum_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ValidatorMaxPauseBlocks int32 `protobuf:"varint,23,opt,name=validator_max_pause_blocks,json=validatorMaxPauseBlocks,proto3" json:"validator_max_pause_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ValidatorMaximumMissedBlocks int32 `protobuf:"varint,24,opt,name=validator_maximum_missed_blocks,json=validatorMaximumMissedBlocks,proto3" json:"validator_maximum_missed_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ValidatorMaxEvidenceAgeInBlocks int32 `protobuf:"varint,25,opt,name=validator_max_evidence_age_in_blocks,json=validatorMaxEvidenceAgeInBlocks,proto3" json:"validator_max_evidence_age_in_blocks,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	ProposerPercentageOfFees int32 `protobuf:"varint,26,opt,name=proposer_percentage_of_fees,json=proposerPercentageOfFees,proto3" json:"proposer_percentage_of_fees,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	MissedBlocksBurnPercentage int32 `protobuf:"varint,27,opt,name=missed_blocks_burn_percentage,json=missedBlocksBurnPercentage,proto3" json:"missed_blocks_burn_percentage,omitempty"`
	//@gotags: pokt:"val_type=SMALLINT"
	DoubleSignBurnPercentage int32 `protobuf:"varint,28,opt,name=double_sign_burn_percentage,json=doubleSignBurnPercentage,proto3" json:"double_sign_burn_percentage,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageDoubleSignFee string `protobuf:"bytes,29,opt,name=message_double_sign_fee,json=messageDoubleSignFee,proto3" json:"message_double_sign_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageSendFee string `protobuf:"bytes,30,opt,name=message_send_fee,json=messageSendFee,proto3" json:"message_send_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeFishermanFee string `protobuf:"bytes,31,opt,name=message_stake_fisherman_fee,json=messageStakeFishermanFee,proto3" json:"message_stake_fisherman_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeFishermanFee string `protobuf:"bytes,32,opt,name=message_edit_stake_fisherman_fee,json=messageEditStakeFishermanFee,proto3" json:"message_edit_stake_fisherman_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeFishermanFee string `protobuf:"bytes,33,opt,name=message_unstake_fisherman_fee,json=messageUnstakeFishermanFee,proto3" json:"message_unstake_fisherman_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseFishermanFee string `protobuf:"bytes,34,opt,name=message_pause_fisherman_fee,json=messagePauseFishermanFee,proto3" json:"message_pause_fisherman_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseFishermanFee string `protobuf:"bytes,35,opt,name=message_unpause_fisherman_fee,json=messageUnpauseFishermanFee,proto3" json:"message_unpause_fisherman_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageFishermanPauseServiceNodeFee string `protobuf:"bytes,36,opt,name=message_fisherman_pause_service_node_fee,json=messageFishermanPauseServiceNodeFee,proto3" json:"message_fisherman_pause_service_node_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageTestScoreFee string `protobuf:"bytes,37,opt,name=message_test_score_fee,json=messageTestScoreFee,proto3" json:"message_test_score_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageProveTestScoreFee string `protobuf:"bytes,38,opt,name=message_prove_test_score_fee,json=messageProveTestScoreFee,proto3" json:"message_prove_test_score_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeAppFee string `protobuf:"bytes,39,opt,name=message_stake_app_fee,json=messageStakeAppFee,proto3" json:"message_stake_app_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeAppFee string `protobuf:"bytes,40,opt,name=message_edit_stake_app_fee,json=messageEditStakeAppFee,proto3" json:"message_edit_stake_app_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeAppFee string `protobuf:"bytes,41,opt,name=message_unstake_app_fee,json=messageUnstakeAppFee,proto3" json:"message_unstake_app_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseAppFee string `protobuf:"bytes,42,opt,name=message_pause_app_fee,json=messagePauseAppFee,proto3" json:"message_pause_app_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseAppFee string `protobuf:"bytes,43,opt,name=message_unpause_app_fee,json=messageUnpauseAppFee,proto3" json:"message_unpause_app_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeValidatorFee string `protobuf:"bytes,44,opt,name=message_stake_validator_fee,json=messageStakeValidatorFee,proto3" json:"message_stake_validator_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeValidatorFee string `protobuf:"bytes,45,opt,name=message_edit_stake_validator_fee,json=messageEditStakeValidatorFee,proto3" json:"message_edit_stake_validator_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeValidatorFee string `protobuf:"bytes,46,opt,name=message_unstake_validator_fee,json=messageUnstakeValidatorFee,proto3" json:"message_unstake_validator_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseValidatorFee string `protobuf:"bytes,47,opt,name=message_pause_validator_fee,json=messagePauseValidatorFee,proto3" json:"message_pause_validator_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseValidatorFee string `protobuf:"bytes,48,opt,name=message_unpause_validator_fee,json=messageUnpauseValidatorFee,proto3" json:"message_unpause_validator_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeServiceNodeFee string `protobuf:"bytes,49,opt,name=message_stake_service_node_fee,json=messageStakeServiceNodeFee,proto3" json:"message_stake_service_node_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeServiceNodeFee string `protobuf:"bytes,50,opt,name=message_edit_stake_service_node_fee,json=messageEditStakeServiceNodeFee,proto3" json:"message_edit_stake_service_node_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeServiceNodeFee string `protobuf:"bytes,51,opt,name=message_unstake_service_node_fee,json=messageUnstakeServiceNodeFee,proto3" json:"message_unstake_service_node_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseServiceNodeFee string `protobuf:"bytes,52,opt,name=message_pause_service_node_fee,json=messagePauseServiceNodeFee,proto3" json:"message_pause_service_node_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseServiceNodeFee string `protobuf:"bytes,53,opt,name=message_unpause_service_node_fee,json=messageUnpauseServiceNodeFee,proto3" json:"message_unpause_service_node_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageChangeParameterFee string `protobuf:"bytes,54,opt,name=message_change_parameter_fee,json=messageChangeParameterFee,proto3" json:"message_change_parameter_fee,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AclOwner string `protobuf:"bytes,55,opt,name=acl_owner,json=aclOwner,proto3" json:"acl_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	BlocksPerSessionOwner string `protobuf:"bytes,56,opt,name=blocks_per_session_owner,json=blocksPerSessionOwner,proto3" json:"blocks_per_session_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppMinimumStakeOwner string `protobuf:"bytes,57,opt,name=app_minimum_stake_owner,json=appMinimumStakeOwner,proto3" json:"app_minimum_stake_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppMaxChainsOwner string `protobuf:"bytes,58,opt,name=app_max_chains_owner,json=appMaxChainsOwner,proto3" json:"app_max_chains_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppBaselineStakeRateOwner string `protobuf:"bytes,59,opt,name=app_baseline_stake_rate_owner,json=appBaselineStakeRateOwner,proto3" json:"app_baseline_stake_rate_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppStakingAdjustmentOwner string `protobuf:"bytes,60,opt,name=app_staking_adjustment_owner,json=appStakingAdjustmentOwner,proto3" json:"app_staking_adjustment_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppUnstakingBlocksOwner string `protobuf:"bytes,61,opt,name=app_unstaking_blocks_owner,json=appUnstakingBlocksOwner,proto3" json:"app_unstaking_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppMinimumPauseBlocksOwner string `protobuf:"bytes,62,opt,name=app_minimum_pause_blocks_owner,json=appMinimumPauseBlocksOwner,proto3" json:"app_minimum_pause_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	AppMaxPausedBlocksOwner string `protobuf:"bytes,63,opt,name=app_max_paused_blocks_owner,json=appMaxPausedBlocksOwner,proto3" json:"app_max_paused_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ServiceNodeMinimumStakeOwner string `protobuf:"bytes,64,opt,name=service_node_minimum_stake_owner,json=serviceNodeMinimumStakeOwner,proto3" json:"service_node_minimum_stake_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ServiceNodeMaxChainsOwner string `protobuf:"bytes,65,opt,name=service_node_max_chains_owner,json=serviceNodeMaxChainsOwner,proto3" json:"service_node_max_chains_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ServiceNodeUnstakingBlocksOwner string `protobuf:"bytes,66,opt,name=service_node_unstaking_blocks_owner,json=serviceNodeUnstakingBlocksOwner,proto3" json:"service_node_unstaking_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ServiceNodeMinimumPauseBlocksOwner string `protobuf:"bytes,67,opt,name=service_node_minimum_pause_blocks_owner,json=serviceNodeMinimumPauseBlocksOwner,proto3" json:"service_node_minimum_pause_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ServiceNodeMaxPausedBlocksOwner string `protobuf:"bytes,68,opt,name=service_node_max_paused_blocks_owner,json=serviceNodeMaxPausedBlocksOwner,proto3" json:"service_node_max_paused_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ServiceNodesPerSessionOwner string `protobuf:"bytes,69,opt,name=service_nodes_per_session_owner,json=serviceNodesPerSessionOwner,proto3" json:"service_nodes_per_session_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	FishermanMinimumStakeOwner string `protobuf:"bytes,70,opt,name=fisherman_minimum_stake_owner,json=fishermanMinimumStakeOwner,proto3" json:"fisherman_minimum_stake_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	FishermanMaxChainsOwner string `protobuf:"bytes,71,opt,name=fisherman_max_chains_owner,json=fishermanMaxChainsOwner,proto3" json:"fisherman_max_chains_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	FishermanUnstakingBlocksOwner string `protobuf:"bytes,72,opt,name=fisherman_unstaking_blocks_owner,json=fishermanUnstakingBlocksOwner,proto3" json:"fisherman_unstaking_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	FishermanMinimumPauseBlocksOwner string `protobuf:"bytes,73,opt,name=fisherman_minimum_pause_blocks_owner,json=fishermanMinimumPauseBlocksOwner,proto3" json:"fisherman_minimum_pause_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	FishermanMaxPausedBlocksOwner string `protobuf:"bytes,74,opt,name=fisherman_max_paused_blocks_owner,json=fishermanMaxPausedBlocksOwner,proto3" json:"fisherman_max_paused_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ValidatorMinimumStakeOwner string `protobuf:"bytes,75,opt,name=validator_minimum_stake_owner,json=validatorMinimumStakeOwner,proto3" json:"validator_minimum_stake_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ValidatorUnstakingBlocksOwner string `protobuf:"bytes,76,opt,name=validator_unstaking_blocks_owner,json=validatorUnstakingBlocksOwner,proto3" json:"validator_unstaking_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ValidatorMinimumPauseBlocksOwner string `protobuf:"bytes,77,opt,name=validator_minimum_pause_blocks_owner,json=validatorMinimumPauseBlocksOwner,proto3" json:"validator_minimum_pause_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ValidatorMaxPausedBlocksOwner string `protobuf:"bytes,78,opt,name=validator_max_paused_blocks_owner,json=validatorMaxPausedBlocksOwner,proto3" json:"validator_max_paused_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ValidatorMaximumMissedBlocksOwner string `protobuf:"bytes,79,opt,name=validator_maximum_missed_blocks_owner,json=validatorMaximumMissedBlocksOwner,proto3" json:"validator_maximum_missed_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ValidatorMaxEvidenceAgeInBlocksOwner string `protobuf:"bytes,80,opt,name=validator_max_evidence_age_in_blocks_owner,json=validatorMaxEvidenceAgeInBlocksOwner,proto3" json:"validator_max_evidence_age_in_blocks_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	ProposerPercentageOfFeesOwner string `protobuf:"bytes,81,opt,name=proposer_percentage_of_fees_owner,json=proposerPercentageOfFeesOwner,proto3" json:"proposer_percentage_of_fees_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MissedBlocksBurnPercentageOwner string `protobuf:"bytes,82,opt,name=missed_blocks_burn_percentage_owner,json=missedBlocksBurnPercentageOwner,proto3" json:"missed_blocks_burn_percentage_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	DoubleSignBurnPercentageOwner string `protobuf:"bytes,83,opt,name=double_sign_burn_percentage_owner,json=doubleSignBurnPercentageOwner,proto3" json:"double_sign_burn_percentage_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageDoubleSignFeeOwner string `protobuf:"bytes,84,opt,name=message_double_sign_fee_owner,json=messageDoubleSignFeeOwner,proto3" json:"message_double_sign_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageSendFeeOwner string `protobuf:"bytes,85,opt,name=message_send_fee_owner,json=messageSendFeeOwner,proto3" json:"message_send_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeFishermanFeeOwner string `protobuf:"bytes,86,opt,name=message_stake_fisherman_fee_owner,json=messageStakeFishermanFeeOwner,proto3" json:"message_stake_fisherman_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeFishermanFeeOwner string `protobuf:"bytes,87,opt,name=message_edit_stake_fisherman_fee_owner,json=messageEditStakeFishermanFeeOwner,proto3" json:"message_edit_stake_fisherman_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeFishermanFeeOwner string `protobuf:"bytes,88,opt,name=message_unstake_fisherman_fee_owner,json=messageUnstakeFishermanFeeOwner,proto3" json:"message_unstake_fisherman_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseFishermanFeeOwner string `protobuf:"bytes,89,opt,name=message_pause_fisherman_fee_owner,json=messagePauseFishermanFeeOwner,proto3" json:"message_pause_fisherman_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseFishermanFeeOwner string `protobuf:"bytes,90,opt,name=message_unpause_fisherman_fee_owner,json=messageUnpauseFishermanFeeOwner,proto3" json:"message_unpause_fisherman_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageFishermanPauseServiceNodeFeeOwner string `protobuf:"bytes,91,opt,name=message_fisherman_pause_service_node_fee_owner,json=messageFishermanPauseServiceNodeFeeOwner,proto3" json:"message_fisherman_pause_service_node_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageTestScoreFeeOwner string `protobuf:"bytes,92,opt,name=message_test_score_fee_owner,json=messageTestScoreFeeOwner,proto3" json:"message_test_score_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageProveTestScoreFeeOwner string `protobuf:"bytes,93,opt,name=message_prove_test_score_fee_owner,json=messageProveTestScoreFeeOwner,proto3" json:"message_prove_test_score_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeAppFeeOwner string `protobuf:"bytes,94,opt,name=message_stake_app_fee_owner,json=messageStakeAppFeeOwner,proto3" json:"message_stake_app_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeAppFeeOwner string `protobuf:"bytes,95,opt,name=message_edit_stake_app_fee_owner,json=messageEditStakeAppFeeOwner,proto3" json:"message_edit_stake_app_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeAppFeeOwner string `protobuf:"bytes,96,opt,name=message_unstake_app_fee_owner,json=messageUnstakeAppFeeOwner,proto3" json:"message_unstake_app_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseAppFeeOwner string `protobuf:"bytes,97,opt,name=message_pause_app_fee_owner,json=messagePauseAppFeeOwner,proto3" json:"message_pause_app_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseAppFeeOwner string `protobuf:"bytes,98,opt,name=message_unpause_app_fee_owner,json=messageUnpauseAppFeeOwner,proto3" json:"message_unpause_app_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeValidatorFeeOwner string `protobuf:"bytes,99,opt,name=message_stake_validator_fee_owner,json=messageStakeValidatorFeeOwner,proto3" json:"message_stake_validator_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeValidatorFeeOwner string `protobuf:"bytes,100,opt,name=message_edit_stake_validator_fee_owner,json=messageEditStakeValidatorFeeOwner,proto3" json:"message_edit_stake_validator_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeValidatorFeeOwner string `protobuf:"bytes,101,opt,name=message_unstake_validator_fee_owner,json=messageUnstakeValidatorFeeOwner,proto3" json:"message_unstake_validator_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseValidatorFeeOwner string `protobuf:"bytes,102,opt,name=message_pause_validator_fee_owner,json=messagePauseValidatorFeeOwner,proto3" json:"message_pause_validator_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseValidatorFeeOwner string `protobuf:"bytes,103,opt,name=message_unpause_validator_fee_owner,json=messageUnpauseValidatorFeeOwner,proto3" json:"message_unpause_validator_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageStakeServiceNodeFeeOwner string `protobuf:"bytes,104,opt,name=message_stake_service_node_fee_owner,json=messageStakeServiceNodeFeeOwner,proto3" json:"message_stake_service_node_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageEditStakeServiceNodeFeeOwner string `protobuf:"bytes,105,opt,name=message_edit_stake_service_node_fee_owner,json=messageEditStakeServiceNodeFeeOwner,proto3" json:"message_edit_stake_service_node_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnstakeServiceNodeFeeOwner string `protobuf:"bytes,106,opt,name=message_unstake_service_node_fee_owner,json=messageUnstakeServiceNodeFeeOwner,proto3" json:"message_unstake_service_node_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessagePauseServiceNodeFeeOwner string `protobuf:"bytes,107,opt,name=message_pause_service_node_fee_owner,json=messagePauseServiceNodeFeeOwner,proto3" json:"message_pause_service_node_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageUnpauseServiceNodeFeeOwner string `protobuf:"bytes,108,opt,name=message_unpause_service_node_fee_owner,json=messageUnpauseServiceNodeFeeOwner,proto3" json:"message_unpause_service_node_fee_owner,omitempty"`
	//@gotags: pokt:"val_type=STRING"
	MessageChangeParameterFeeOwner string `protobuf:"bytes,109,opt,name=message_change_parameter_fee_owner,json=messageChangeParameterFeeOwner,proto3" json:"message_change_parameter_fee_owner,omitempty"`
}

func (m *MockParams) GetBlocksPerSession() int32 {
	return m.BlocksPerSession
}

func (m *MockParams) GetAppMinimumStake() string {
	return m.AppMinimumStake
}

func (m *MockParams) GetAppMaxChains() int32 {
	return m.AppMaxChains
}

func (m *MockParams) GetAppBaselineStakeRate() int32 {
	return m.AppBaselineStakeRate
}

func (m *MockParams) GetAppStakingAdjustment() int32 {
	return m.AppStakingAdjustment
}

func (m *MockParams) GetAppUnstakingBlocks() int32 {
	return m.AppUnstakingBlocks
}

func (m *MockParams) GetAppMinimumPauseBlocks() int32 {
	return m.AppMinimumPauseBlocks
}

func (m *MockParams) GetAppMaxPauseBlocks() int32 {
	return m.AppMaxPauseBlocks
}

func (m *MockParams) GetServiceNodeMinimumStake() string {
	return m.ServiceNodeMinimumStake
}

func (m *MockParams) GetServiceNodeMaxChains() int32 {
	return m.ServiceNodeMaxChains
}

func (m *MockParams) GetServiceNodeUnstakingBlocks() int32 {
	return m.ServiceNodeUnstakingBlocks
}

func (m *MockParams) GetServiceNodeMinimumPauseBlocks() int32 {
	return m.ServiceNodeMinimumPauseBlocks
}

func (m *MockParams) GetServiceNodeMaxPauseBlocks() int32 {
	return m.ServiceNodeMaxPauseBlocks
}

func (m *MockParams) GetServiceNodesPerSession() int32 {
	return m.ServiceNodesPerSession
}

func (m *MockParams) GetFishermanMinimumStake() string {
	return m.FishermanMinimumStake
}

func (m *MockParams) GetFishermanMaxChains() int32 {
	return m.FishermanMaxChains
}

func (m *MockParams) GetFishermanUnstakingBlocks() int32 {
	return m.FishermanUnstakingBlocks
}

func (m *MockParams) GetFishermanMinimumPauseBlocks() int32 {
	return m.FishermanMinimumPauseBlocks
}

func (m *MockParams) GetFishermanMaxPauseBlocks() int32 {
	return m.FishermanMaxPauseBlocks
}

func (m *MockParams) GetValidatorMinimumStake() string {
	return m.ValidatorMinimumStake
}

func (m *MockParams) GetValidatorUnstakingBlocks() int32 {
	return m.ValidatorUnstakingBlocks
}

func (m *MockParams) GetValidatorMinimumPauseBlocks() int32 {
	return m.ValidatorMinimumPauseBlocks
}

func (m *MockParams) GetValidatorMaxPauseBlocks() int32 {
	return m.ValidatorMaxPauseBlocks
}

func (m *MockParams) GetValidatorMaximumMissedBlocks() int32 {
	return m.ValidatorMaximumMissedBlocks
}

func (m *MockParams) GetValidatorMaxEvidenceAgeInBlocks() int32 {
	return m.ValidatorMaxEvidenceAgeInBlocks
}

func (m *MockParams) GetProposerPercentageOfFees() int32 {
	return m.ProposerPercentageOfFees
}

func (m *MockParams) GetMissedBlocksBurnPercentage() int32 {
	return m.MissedBlocksBurnPercentage
}

func (m *MockParams) GetDoubleSignBurnPercentage() int32 {
	return m.DoubleSignBurnPercentage
}

func (m *MockParams) GetMessageDoubleSignFee() string {
	return m.MessageDoubleSignFee
}

func (m *MockParams) GetMessageSendFee() string {
	return m.MessageSendFee
}

func (m *MockParams) GetMessageStakeFishermanFee() string {
	return m.MessageStakeFishermanFee
}

func (m *MockParams) GetMessageEditStakeFishermanFee() string {
	return m.MessageEditStakeFishermanFee
}

func (m *MockParams) GetMessageUnstakeFishermanFee() string {
	return m.MessageUnstakeFishermanFee
}

func (m *MockParams) GetMessagePauseFishermanFee() string {
	return m.MessagePauseFishermanFee
}

func (m *MockParams) GetMessageUnpauseFishermanFee() string {
	return m.MessageUnpauseFishermanFee
}

func (m *MockParams) GetMessageFishermanPauseServiceNodeFee() string {
	return m.MessageFishermanPauseServiceNodeFee
}

func (m *MockParams) GetMessageTestScoreFee() string {
	return m.MessageTestScoreFee
}

func (m *MockParams) GetMessageProveTestScoreFee() string {
	return m.MessageProveTestScoreFee
}

func (m *MockParams) GetMessageStakeAppFee() string {
	return m.MessageStakeAppFee
}

func (m *MockParams) GetMessageEditStakeAppFee() string {
	return m.MessageEditStakeAppFee
}

func (m *MockParams) GetMessageUnstakeAppFee() string {
	return m.MessageUnstakeAppFee
}

func (m *MockParams) GetMessagePauseAppFee() string {
	return m.MessagePauseAppFee
}

func (m *MockParams) GetMessageUnpauseAppFee() string {
	return m.MessageUnpauseAppFee
}

func (m *MockParams) GetMessageStakeValidatorFee() string {
	return m.MessageStakeValidatorFee
}

func (m *MockParams) GetMessageEditStakeValidatorFee() string {
	return m.MessageEditStakeValidatorFee
}

func (m *MockParams) GetMessageUnstakeValidatorFee() string {
	return m.MessageUnstakeValidatorFee
}

func (m *MockParams) GetMessagePauseValidatorFee() string {
	return m.MessagePauseValidatorFee
}

func (m *MockParams) GetMessageUnpauseValidatorFee() string {
	return m.MessageUnpauseValidatorFee
}

func (m *MockParams) GetMessageStakeServiceNodeFee() string {
	return m.MessageStakeServiceNodeFee
}

func (m *MockParams) GetMessageEditStakeServiceNodeFee() string {
	return m.MessageEditStakeServiceNodeFee
}

func (m *MockParams) GetMessageUnstakeServiceNodeFee() string {
	return m.MessageUnstakeServiceNodeFee
}

func (m *MockParams) GetMessagePauseServiceNodeFee() string {
	return m.MessagePauseServiceNodeFee
}

func (m *MockParams) GetMessageUnpauseServiceNodeFee() string {
	return m.MessageUnpauseServiceNodeFee
}

func (m *MockParams) GetMessageChangeParameterFee() string {
	return m.MessageChangeParameterFee
}

func (m *MockParams) GetAclOwner() string {
	return m.AclOwner
}

func (m *MockParams) GetBlocksPerSessionOwner() string {
	return m.BlocksPerSessionOwner
}

func (m *MockParams) GetAppMinimumStakeOwner() string {
	return m.AppMinimumStakeOwner
}

func (m *MockParams) GetAppMaxChainsOwner() string {
	return m.AppMaxChainsOwner
}

func (m *MockParams) GetAppBaselineStakeRateOwner() string {
	return m.AppBaselineStakeRateOwner
}

func (m *MockParams) GetAppStakingAdjustmentOwner() string {
	return m.AppStakingAdjustmentOwner
}

func (m *MockParams) GetAppUnstakingBlocksOwner() string {
	return m.AppUnstakingBlocksOwner
}

func (m *MockParams) GetAppMinimumPauseBlocksOwner() string {
	return m.AppMinimumPauseBlocksOwner
}

func (m *MockParams) GetAppMaxPausedBlocksOwner() string {
	return m.AppMaxPausedBlocksOwner
}

func (m *MockParams) GetServiceNodeMinimumStakeOwner() string {
	return m.ServiceNodeMinimumStakeOwner
}

func (m *MockParams) GetServiceNodeMaxChainsOwner() string {
	return m.ServiceNodeMaxChainsOwner
}

func (m *MockParams) GetServiceNodeUnstakingBlocksOwner() string {
	return m.ServiceNodeUnstakingBlocksOwner
}

func (m *MockParams) GetServiceNodeMinimumPauseBlocksOwner() string {
	return m.ServiceNodeMinimumPauseBlocksOwner
}

func (m *MockParams) GetServiceNodeMaxPausedBlocksOwner() string {
	return m.ServiceNodeMaxPausedBlocksOwner
}

func (m *MockParams) GetServiceNodesPerSessionOwner() string {
	return m.ServiceNodesPerSessionOwner
}

func (m *MockParams) GetFishermanMinimumStakeOwner() string {
	return m.FishermanMinimumStakeOwner
}

func (m *MockParams) GetFishermanMaxChainsOwner() string {
	return m.FishermanMaxChainsOwner
}

func (m *MockParams) GetFishermanUnstakingBlocksOwner() string {
	return m.FishermanUnstakingBlocksOwner
}

func (m *MockParams) GetFishermanMinimumPauseBlocksOwner() string {
	return m.FishermanMinimumPauseBlocksOwner
}

func (m *MockParams) GetFishermanMaxPausedBlocksOwner() string {
	return m.FishermanMaxPausedBlocksOwner
}

func (m *MockParams) GetValidatorMinimumStakeOwner() string {
	return m.ValidatorMinimumStakeOwner
}

func (m *MockParams) GetValidatorUnstakingBlocksOwner() string {
	return m.ValidatorUnstakingBlocksOwner
}

func (m *MockParams) GetValidatorMinimumPauseBlocksOwner() string {
	return m.ValidatorMinimumPauseBlocksOwner
}

func (m *MockParams) GetValidatorMaxPausedBlocksOwner() string {
	return m.ValidatorMaxPausedBlocksOwner
}

func (m *MockParams) GetValidatorMaximumMissedBlocksOwner() string {
	return m.ValidatorMaximumMissedBlocksOwner
}

func (m *MockParams) GetValidatorMaxEvidenceAgeInBlocksOwner() string {
	return m.ValidatorMaxEvidenceAgeInBlocksOwner
}

func (m *MockParams) GetProposerPercentageOfFeesOwner() string {
	return m.ProposerPercentageOfFeesOwner
}

func (m *MockParams) GetMissedBlocksBurnPercentageOwner() string {
	return m.MissedBlocksBurnPercentageOwner
}

func (m *MockParams) GetDoubleSignBurnPercentageOwner() string {
	return m.DoubleSignBurnPercentageOwner
}

func (m *MockParams) GetMessageDoubleSignFeeOwner() string {
	return m.MessageDoubleSignFeeOwner
}

func (m *MockParams) GetMessageSendFeeOwner() string {
	return m.MessageSendFeeOwner
}

func (m *MockParams) GetMessageStakeFishermanFeeOwner() string {
	return m.MessageStakeFishermanFeeOwner
}

func (m *MockParams) GetMessageEditStakeFishermanFeeOwner() string {
	return m.MessageEditStakeFishermanFeeOwner
}

func (m *MockParams) GetMessageUnstakeFishermanFeeOwner() string {
	return m.MessageUnstakeFishermanFeeOwner
}

func (m *MockParams) GetMessagePauseFishermanFeeOwner() string {
	return m.MessagePauseFishermanFeeOwner
}

func (m *MockParams) GetMessageUnpauseFishermanFeeOwner() string {
	return m.MessageUnpauseFishermanFeeOwner
}

func (m *MockParams) GetMessageFishermanPauseServiceNodeFeeOwner() string {
	return m.MessageFishermanPauseServiceNodeFeeOwner
}

func (m *MockParams) GetMessageTestScoreFeeOwner() string {
	return m.MessageTestScoreFeeOwner
}

func (m *MockParams) GetMessageProveTestScoreFeeOwner() string {
	return m.MessageProveTestScoreFeeOwner
}

func (m *MockParams) GetMessageStakeAppFeeOwner() string {
	return m.MessageStakeAppFeeOwner
}

func (m *MockParams) GetMessageEditStakeAppFeeOwner() string {
	return m.MessageEditStakeAppFeeOwner
}

func (m *MockParams) GetMessageUnstakeAppFeeOwner() string {
	return m.MessageUnstakeAppFeeOwner
}

func (m *MockParams) GetMessagePauseAppFeeOwner() string {
	return m.MessagePauseAppFeeOwner
}

func (m *MockParams) GetMessageUnpauseAppFeeOwner() string {
	return m.MessageUnpauseAppFeeOwner
}

func (m *MockParams) GetMessageStakeValidatorFeeOwner() string {
	return m.MessageStakeValidatorFeeOwner
}

func (m *MockParams) GetMessageEditStakeValidatorFeeOwner() string {
	return m.MessageEditStakeValidatorFeeOwner
}

func (m *MockParams) GetMessageUnstakeValidatorFeeOwner() string {
	return m.MessageUnstakeValidatorFeeOwner
}

func (m *MockParams) GetMessagePauseValidatorFeeOwner() string {
	return m.MessagePauseValidatorFeeOwner
}

func (m *MockParams) GetMessageUnpauseValidatorFeeOwner() string {
	return m.MessageUnpauseValidatorFeeOwner
}

func (m *MockParams) GetMessageStakeServiceNodeFeeOwner() string {
	return m.MessageStakeServiceNodeFeeOwner
}

func (m *MockParams) GetMessageEditStakeServiceNodeFeeOwner() string {
	return m.MessageEditStakeServiceNodeFeeOwner
}

func (m *MockParams) GetMessageUnstakeServiceNodeFeeOwner() string {
	return m.MessageUnstakeServiceNodeFeeOwner
}

func (m *MockParams) GetMessagePauseServiceNodeFeeOwner() string {
	return m.MessagePauseServiceNodeFeeOwner
}

func (m *MockParams) GetMessageUnpauseServiceNodeFeeOwner() string {
	return m.MessageUnpauseServiceNodeFeeOwner
}

func (m *MockParams) GetMessageChangeParameterFeeOwner() string {
	return m.MessageChangeParameterFeeOwner
}
