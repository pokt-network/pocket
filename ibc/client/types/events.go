package types

const (
	// Event topics for the events emitted by the Client submodule
	EventTopicCreateClient       = "create_client"
	EventTopicUpdateClient       = "update_client"
	EventTopicUpgradeClient      = "upgrade_client"
	EventTopicSubmitMisbehaviour = "client_misbehaviour"
)

var (
	// Attribute keys for the events emitted by the Client submodule
	AttributeKeyClientID        = []byte("client_id")
	AttributeKeyClientType      = []byte("client_type")
	AttributeKeyConsensusHeight = []byte("consensus_height")
	AttributeKeyHeader          = []byte("header")
)
