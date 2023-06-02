package modules

import (
	ics23 "github.com/cosmos/ics23/go"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const IBCModuleName = "ibc"

type IBCModule interface {
	Module
	NewHost() (IBCHost, error)
}

type IBCHost interface {
	IBCHandler

	// GetStore returns the IBC StoreManager to interact with the IBC hosts state
	GetStore() StoreManager
}

// TODO: Uncomment interface functions as they are defined and potentially change their signatures
// where necessary
// IBCHandler is the interface through which the different IBC sub-modules can be interacted with
// https://github.com/cosmos/ibc/tree/main/spec/core/ics-025-handler-interface
type IBCHandler interface {
	// === Client Lifecycle Management ===
	// https://github.com/cosmos/ibc/tree/main/spec/core/ics-002-client-semantics

	// CreateClient creates a new client with the given client state and initial consensus state
	// and initialises its unique identifier in the IBC store
	// CreateClient(clientState clientState, consensusState consensusState) error

	// UpdateClient updates an existing client with the given ClientMessage, given that
	// the ClientMessage can be verified using the existing ClientState and ConsensusState
	// UpdateClient(identifier Identifier, clientMessage ClientMessage) error

	// QueryConsensusState returns the ConsensusState at the given height for the given client
	// QueryConsensusState(identifier Identifier, height Height) ConsensusState

	// QueryClientState returns the ClientState for the given client
	// QueryClientState(identifier Identifier) ClientState

	// SubmitMisbehaviour submits evidence for a misbehaviour to the client, possibly invalidating
	// previously valid state roots and thus preventing future updates
	// SubmitMisbehaviour(identifier Identifier, clientMessage ClientMessage) error

	// === Connection Lifecycle Management ===
	// https://github.com/cosmos/ibc/tree/main/spec/core/ics-003-connection-semantics

	// ConnOpenInit attempts to initialise a connection to a given counterparty chain (executed on source chain)
	/**
		ConnOpenInit(
			counterpartyPrefix CommitmentPrefix,
			clientIdentifier, counterpartyClientIdentifier Identifier,
			version: string, // Optional: If version is included, the handshake must fail if the version is not the same
			delayPeriodTime, delayPeriodBlocks uint64,
		) error
	**/

	// ConnOpenTry relays a notice of a connection attempt to a counterpaty chain (executed on destination chain)
	/**
		ConnOpenTry(
			counterpartyPrefix CommitmentPrefix,
			counterpartyConnectionIdentifier, counterpartyClientIdentifier, clientIdentifier Identifier,
			clientState ClientState,
			counterpartyVersions []string,
			delayPeriodTime, delayPeriodBlocks uint64,
			proofInit, proofClient, proofConsensus ics23.CommitmentProof,
			proofHeight, consensusHeight Height,
			hostConsensusStateProof bytes,
		) error
	**/

	// ConnOpenAck relays the acceptance of a connection open attempt from counterparty chain (executed on source chain)
	/**
	ConnOpenAck(
			identifier, counterpartyIdentifier Identifier,
			clientState ClientState,
			version string,
			proofTry, proofClient, proofConsensus ics23.CommitmentProof,
			proofHeight, consensusHeight Height,
			hostConsensusStateProof bytes,
		) error
	**/

	// ConnOpenConfirm confirms opening of a connection to the counterparty chain after which the
	// connection is open to both chains (executed on destination chain)
	// ConnOpenConfirm(identifier Identifier, proofAck ics23.CommitmentProof, proofHeight Height) error

	// QueryConnection returns the ConnectionEnd for the given connection identifier
	// QueryConnection(identifier Identifier) (ConnectionEnd, error)

	// QueryClientConnections returns the list of connection identifiers associated with a given client
	// QueryClientConnections(clientIdentifier Identifier) ([]Identifier, error)

	// === Channel Lifecycle Management ===
	// https://github.com/cosmos/ibc/tree/main/spec/core/ics-004-channel-and-packet-semantics

	// ChanOpenInit initialises a channel opening handshake with a counterparty chain (executed on source chain)
	/**
		ChanOpenInit(
			order ChannelOrder,
			connectionHops []Identifier,
			portIdentifier, counterpartyPortIdentifier Identifier,
			version string,
		) (channelIdentifier Identifier, channelCapability CapabilityKey, err Error)
	**/

	// ChanOpenTry attempts to accept the channel opening handshake from a counterparty chain (executed on destination chain)
	/**
		ChanOpenTry(
			order ChannelOrder,
			connectionHops []Identifier,
			portIdentifier, counterpartyPortIdentifier, counterpartyChannelIdentifier Identifier,
			version, counterpartyVersion string,
			proofInit ics23.CommitmentProof,
		) (channelIdentifier Identifier, channelCapability CapabilityKey, err Error)
	**/

	// ChanOpenAck relays acceptance of a channel opening handshake from a counterparty chain (executed on source chain)
	/**
		ChanOpenAck(
			portIdentifier,	channelIdentifier, counterpartyChannelIdentifier Identifier,
			counterpartyVersion string,
			proofTry ics23.CommitmentProof,
			proofHeight Height,
		) error
	**/

	// ChanOpenConfirm acknowledges the acknowledgment of the channel opening hanshake on the counterparty
	// chain after which the channel opening handshake is complete (executed on destination chain)
	// ChanOpenConfirm(portIdentifier, channelIdentifier Identifier, proofAck ics23.CommitmentProof, proofHeight Height) error

	// ChanCloseInit is called to close the ChannelEnd with the given identifier on the host machine
	// ChanCloseInit(portIdentifier, channelIdentifier Identifier) error

	// ChanCloseConfirm is called to close the ChannelEnd on the counterparty chain as the other end is closed
	// ChanCloseConfirm(portIdentifier, channelIdentifier Identifier, proofInit ics23.CommitmentProof, proofHeight Height) error

	// === Packet Relaying ===

	// SendPacket is called to send an IBC packet on the channel with the given identifier
	/**
		SendPacket(
			capability CapabilityKey,
			sourcePort Identifier,
			sourceChannel Identifier,
			timeoutHeight Height,
			timeoutTimestamp uint64,
			data []byte,
		) (sequence uint64, err error)
	**/

	// RecvPacket is called in order to receive an IBC packet on the corresponding channel end
	// on the counterpaty chain
	// RecvPacket(packet OpaquePacket, proof ics23.CommitmentProof, proofHeight Height, relayer string) (Packet, error)

	// AcknowledgePacket is called to acknowledge the receipt of an IBC packet to the corresponding chain
	/**
		AcknowledgePacket(
			packet OpaquePacket,
			acknowledgement []byte,
			proof ics23.CommitmentProof,
			proofHeight Height,
			relayer string,
		) (Packet, error)
	**/

	// TimeoutPacket is called to timeout an IBC packet on the corresponding channel end after the
	// timeout height or timeout timestamp has passed and the packet has not been committed
	/**
		TimeoutPacket(
			packet OpaquePacket,
			proof ics23.CommitmentProof,
			proofHeight Height,
			nextSequenceRecv *uint64,
			relayer string,
		) (Packet, error)
	**/

	// TimeoutOnClose is called to prove to the counterparty chain that the channel end has been
	// closed and that the packet sent over this channel will not be received
	/**
		TimeoutOnClose(
			packet OpaquePacket,
			proof, proofClosed ics23.CommitmentProof,
			proofHeight Height,
			nextSequenceRecv *uint64,
			relayer string,
		) (Packet, error)
	**/
}

// StoreManager is an interface that allows for the interaction with the numerous
// stores used by the IBC module.
type StoreManager interface {
	GetStore(storeKey string) (Store, error)
	GetProvableStore(storeKey string) (ProvableStore, error)
	AddStore(store Store) error
	RemoveStore(storeKey string) error
}

// Store is a simple interface to interact with data in a key-value manner.
type Store interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
	GetStoreKey() string
	IsProvable() bool
	Stop() error
}

// ProvableStore allows for the creation of proofs for the data stored in the store
// which can be verified for authenticity
type ProvableStore interface {
	Store
	Root() *coreTypes.CommitmentRoot
	CreateMembershipProof(key, value []byte) (*ics23.CommitmentProof, error)
	CreateNonMembershipProof(key []byte) (*ics23.CommitmentProof, error)
}
