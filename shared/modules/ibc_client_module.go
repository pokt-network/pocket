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
)

type ClientManagerOption func(ClientManager)

type clientManagerFactory = FactoryWithOptions[ClientManager, ClientManagerOption]

// ClientManager is the interface that defines the methods needed to interact with an IBC light client
// it manages the different lifecycle methods for the different clients
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

	// QueryConsensusState returns the ConsensusState at the given height for the given client
	QueryConsensusState(identifier string, height Height) (ConsensusState, error)

	// QueryClientState returns the ClientState for the given client
	QueryClientState(identifier string) (ClientState, error)

	// SubmitMisbehaviour submits evidence for a misbehaviour to the client, possibly invalidating
	// previously valid state roots and thus preventing future updates
	SubmitMisbehaviour(identifier string, clientMessage ClientMessage) error
}

// ClientState is an interface that defines the methods required by a clients
// implementation of their own client state object
//
// ClientState is an opaque data structure defined by a client type. It may keep
// arbitrary internal state to track verified roots and past misbehaviours.
type ClientState interface {
	proto.Message

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
	UpdateStateOnMisbehaviour(clientStore ProvableStore, clientMsg ClientMessage)

	// UpdateState updates and stores as necessary any associated information
	// for an IBC client, such as the ClientState and corresponding ConsensusState.
	// Upon successful update, a list of consensus heights is returned.
	// It assumes the ClientMessage has already been verified.
	UpdateState(clientStore ProvableStore, clientMsg ClientMessage) []Height
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
