package modules

//go:generate mockgen -destination=./mocks/ibc_client_module_mock.go github.com/pokt-network/pocket/shared/modules ClientManager

import (
	"google.golang.org/protobuf/proto"
)

type ClientStatus string

const (
	ClientManagerModuleName = "client_manager"

	// Client Status types
	ActiveStatus       ClientStatus = "active"
	ExpiredStatus      ClientStatus = "expired"
	FrozenStatus       ClientStatus = "frozen"
	UnauthorizedStatus ClientStatus = "unauthorized"
	UnknownStatus      ClientStatus = "unknown"
)

type ClientManagerOption func(ClientManager)

type clientManagerFactory = FactoryWithOptions[ClientManager, ClientManagerOption]

// ClientManager is the interface that defines the methods needed to interact with an
// IBC light client it manages the different lifecycle methods for the different clients
// https://github.com/cosmos/ibc/tree/main/spec/core/ics-002-client-semantics
type ClientManager interface {
	Submodule
	clientManagerFactory

	// === Client Lifecycle Management ===

	// CreateClient creates a new client with the given client state and initial consensus state
	// and initialises its unique identifier in the IBC store
	CreateClient(ClientState, ConsensusState) (string, error)

	// UpdateClient updates an existing client with the given ClientMessage, given that
	// the ClientMessage can be verified using the existing ClientState and ConsensusState
	UpdateClient(identifier string, clientMessage ClientMessage) error

	// UpgradeClient upgrades an existing client with the given identifier using the
	// ClientState and ConsenusState provided. It can only do so if the new client
	// was committed to by the old client at the specified upgrade height
	UpgradeClient(
		identifier string,
		clientState ClientState, consensusState ConsensusState,
		proofUpgradeClient, proofUpgradeConsState []byte,
	) error

	// === Client Queries ===

	// GetConsensusState returns the ConsensusState at the given height for the given client
	GetConsensusState(identifier string, height Height) (ConsensusState, error)

	// GetClientState returns the ClientState for the given client
	GetClientState(identifier string) (ClientState, error)

	// GetHostConsensusState returns the ConsensusState at the given height for the host chain
	GetHostConsensusState(height Height) (ConsensusState, error)

	// GetHostClientState returns the ClientState at the provided height for the host chain
	GetHostClientState(height Height) (ClientState, error)

	// GetCurrentHeight returns the current IBC client height of the network
	GetCurrentHeight() (Height, error)

	// VerifyHostClientState verifies the client state for a client running on a
	// counterparty chain is valid, checking against the current host client state
	VerifyHostClientState(ClientState) error
}

// ClientState is an interface that defines the methods required by a clients
// implementation of their own client state object
//
// ClientState is an opaque data structure defined by a client type. It may keep
// arbitrary internal state to track verified roots and past misbehaviours.
type ClientState interface {
	proto.Message

	GetData() []byte
	GetWasmChecksum() []byte
	ClientType() string
	GetLatestHeight() Height
	Validate() error

	// Status returns the status of the client. Only Active clients are allowed
	// to process packets.
	Status(clientStore ProvableStore) ClientStatus

	// GetTimestampAtHeight must return the timestamp for the consensus state
	// associated with the provided height.
	GetTimestampAtHeight(clientStore ProvableStore, height Height) (uint64, error)

	// Initialise is called upon client creation, it allows the client to perform
	// validation on the initial consensus state and set the client state,
	// consensus state and any client-specific metadata necessary for correct
	// light client operation in the provided client store.
	Initialise(clientStore ProvableStore, consensusState ConsensusState) error

	// VerifyMembership is a generic proof verification method which verifies a
	// proof of the existence of a value at a given CommitmentPath at the
	// specified height. The path is expected to be the full CommitmentPath
	VerifyMembership(
		clientStore ProvableStore,
		height Height,
		delayTimePeriod, delayBlockPeriod uint64,
		proof, path, value []byte,
	) error

	// VerifyNonMembership is a generic proof verification method which verifies
	// the absence of a given CommitmentPath at a specified height. The path is
	// expected to be the full CommitmentPath
	VerifyNonMembership(
		clientStore ProvableStore,
		height Height,
		delayTimePeriod, delayBlockPeriod uint64,
		proof, path []byte,
	) error

	// VerifyClientMessage verifies a ClientMessage. A ClientMessage could be a
	// Header, Misbehaviour, or batch update. It must handle each type of
	// ClientMessage appropriately. Calls to CheckForMisbehaviour, UpdateState,
	// and UpdateStateOnMisbehaviour will assume that the content of the
	// ClientMessage has been verified and can be trusted. An error should be
	// returned if the ClientMessage fails to verify.
	VerifyClientMessage(clientStore ProvableStore, clientMsg ClientMessage) error

	// Checks for evidence of a misbehaviour in Header or Misbehaviour type.
	// It assumes the ClientMessage has already been verified.
	CheckForMisbehaviour(clientStore ProvableStore, clientMsg ClientMessage) bool

	// UpdateStateOnMisbehaviour should perform appropriate state changes on a
	// client state given that misbehaviour has been detected and verified
	UpdateStateOnMisbehaviour(clientStore ProvableStore, clientMsg ClientMessage) error

	// UpdateState updates and stores as necessary any associated information
	// for an IBC client, such as the ClientState and corresponding ConsensusState.
	// Upon successful update, a consensus height is returned.
	// It assumes the ClientMessage has already been verified.
	UpdateState(clientStore ProvableStore, clientMsg ClientMessage) (Height, error)

	// Upgrade functions
	// NOTE: proof heights are not included as upgrade to a new revision is expected to pass only on the last
	// height committed by the current revision. Clients are responsible for ensuring that the planned last
	// height of the current revision is somehow encoded in the proof verification process.
	// This is to ensure that no premature upgrades occur, since upgrade plans committed to by the counterparty
	// may be cancelled or modified before the last planned height.
	// If the upgrade is verified, the upgraded client and consensus states must be set in the client store.
	VerifyUpgradeAndUpdateState(
		clientStore ProvableStore,
		newClient ClientState,
		newConsState ConsensusState,
		proofUpgradeClient,
		proofUpgradeConsState []byte,
	) error
}

// ConsensusState is an interface that defines the methods required by a clients
// implementation of their own consensus state object
//
// ConsensusState is an opaque data structure defined by a client type, used by the
// validity predicate to verify new commits & state roots. Likely the structure will
// contain the last commit produced by the consensus process, including signatures
// and validator set metadata.
type ConsensusState interface {
	proto.Message

	GetData() []byte
	ClientType() string
	GetTimestamp() uint64
	ValidateBasic() error
}

// ClientMessage is an interface that defines the methods required by a clients
// implementation of their own client message object
//
// A ClientMessage is an opaque data structure defined by a client type which
// provides information to update the client. ClientMessages can be submitted
// to an associated client to add new ConsensusState(s) and/or update the
// ClientState. They likely contain a height, a proof, a commitment root, and
// possibly updates to the validity predicate.
type ClientMessage interface {
	proto.Message

	GetData() []byte
	ClientType() string
	ValidateBasic() error
}

// Height is an interface that defines the methods required by a clients
// implementation of their own height object
//
// Heights usually have two components: revision number and revision height.
type Height interface {
	IsZero() bool
	LT(Height) bool
	LTE(Height) bool
	EQ(Height) bool
	GT(Height) bool
	GTE(Height) bool
	Increment() Height
	Decrement() Height
	GetRevisionNumber() uint64
	GetRevisionHeight() uint64
	ToString() string // must define a determinstic `String()` method not the generated protobuf method
}

func (s ClientStatus) String() string {
	return string(s)
}
