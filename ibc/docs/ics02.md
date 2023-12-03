# ICS-02 Client Semantics <!-- omit in toc -->

- [Definitions](#definitions)
  - ["light client"](#light-client)
- [Overview](#overview)
- [Implementation](#implementation)
  - [Client Manager](#client-manager)
    - [Lifecycle Management](#lifecycle-management)
    - [Client Queries](#client-queries)
- [Types](#types)
- [Provable Stores](#provable-stores)

## Definitions

### "light client"

In the context of IBC a light client differs from a traditional "light client." An IBC light client is simply a state verification algorithm. It does not sync with the network, it does not download headers. Instead the updates/new headers for a client are provided by an IBC relayer.

## Overview

IBC utilises light clients to verify the correctness of the state of a counterparty chain. This allows for an IBC packet to be committed to in the state of the network on a source chain and then validated through the light client on the counterparty chain.

[ICS-02][ics02] defines the interfaces and types through which the host machine can interact with the light clients it manages. This includes: client creation, client updates and upgrades as well as submitting misbehaviour from the chain the client is tracking. In addition to this, ICS-02 also defines numerous interfaces that are used by the different client implementations in order to carry out the previous actions as well as verify the state of the chain they represent via a proof.

## Implementation

[ICS-02][ics02] is implemented according to the specification. However as the Pocket protocol will utilise [ICS-08][ics08] WASM clients for the improvements to client upgradeability; the implementations of the `ClientState`, `ConsensusState` and other interfaces are specific to a WASM client.

The implementation details are explored below, the code for ICS-02 can be found in [ibc/client](../client/)

### Client Manager

The `ClientManager` is the submodule that governs the light client implementations and implements the [ICS-02][ics02] interface. It is defined in [shared/modules/ibc_client_module.go](../../shared/modules/ibc_client_module.go). The `ClientManager` exposed the following methods:

```go
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
GetCurrentHeight() Height

// VerifyHostClientState verifies the client state for a client running on a
// counterparty chain is valid, checking against the current host client state
VerifyHostClientState(ClientState) error
```

#### Lifecycle Management

The `ClientManager` handles the creation, updates and upgrades for a light client. It does so by utilising the following interfaces:

```go
type ClientState interface
type ConsensusState interface
type ClientMessage interface
```

These interfaces are generic but have unique implementations for each client type. As Pocket utilises WASM light clients each implementation contains a `data []byte` field which contains a serialised, opaque data structure for use within the WASM client.

The `data` field is a JSON serialised payload that contains the data required for the client to carry out the desired operation, as well as the operation name to carry out. For example, a verify membership payload is constructed using the following `struct`s:

```go
type (
	verifyMembershipInnerPayload struct {
		Height           modules.Height            `json:"height"`
		DelayTimePeriod  uint64                    `json:"delay_time_period"`
		DelayBlockPeriod uint64                    `json:"delay_block_period"`
		Proof            []byte                    `json:"proof"`
		Path             core_types.CommitmentPath `json:"path"`
		Value            []byte                    `json:"value"`
	}
	verifyMembershipPayload struct {
		VerifyMembership verifyMembershipInnerPayload `json:"verify_membership"`
	}
)
```

By utilising this pattern of JSON payloads the WASM client itself is able to unmarshal the opaque payload into their own internal protobuf definitions for the implementation of the `ClientState` for example. This allows them to have a much simpler implementation and focus solely on the logic around verification and utilising simple storage.

See: [Types](#types) for more information on the interfaces and types used in the ICS-02 implementation

#### Client Queries

[ICS-24](./ics24.md) instructs that a host must allow for the introspection of both its own `ConsensusState` and `ClientState`. This is done through the `ClientManager`'s `GetHostConsensusState` and `GetHostClientState` methods. These are then used by relayers to:

1. Provide light clients running on counterparty chains the `ConsensusState` and `ClientState` objects they need.
2. Verify the state of a light client running on a counterparty chain, against the host chain's current `ClientState`

The other queries used by the `ClientManager` involve querying the [ICS-24](./ics24.md) stores to retrieve the `ClientState` and `ConsensusState` stored objects on a per-client basis.

See [Provable Stores](#provable-stores) for more information on how the `ProvableStore`s are used in ICS-02.

## Types

The [ICS-02 specification][ics02] defines the need for numerous interfaces:

1. `ClientState`
   - `ClientState` is an opaque data structure defined by a client type. It may keep arbitrary internal state to track verified roots and past misbehaviours.
2. `ConsensusState`
   - `ConsensusState` is an opaque data structure defined by a client type, used by the
     validity predicate to verify new commits & state roots. Likely the structure will contain the last commit produced by the consensus process, including signatures and validator set metadata.
3. `ClientMessage`
   - `ClientMessage` is an opaque data structure defined by a client type which provides information to update the client. `ClientMessage`s can be submitted to an associated client to add new `ConsensusState`(s) and/or update the `ClientState`. They likely contain a height, a proof, a commitment root, and possibly updates to the validity predicate.
4. `Height`
   - `Height` is an interface that defines the methods required by a clients implementation of their own height object `Height`s usually have two components: revision number and revision height.

As previously mentioned these interfaces have different implementations for each light client type. This is due to the different light clients representing different networks, consensus types and chains altogether. The implementation of these interfaces can be found in [ibc/client/types/proto/wasm.proto](../client/types/proto/wasm.proto).

The `data` field in these messages represents the opaque data structure that is internal to the WASM client. This is a part of the JSON serialised payload that is passed into the WASM client, and is used to carry out any relevant operations. This enables the WASM client to define its own internal data structures that can unmarshal the JSON payload into its own internal protobuf definitions.

See: [shared/modules/ibc_client_module.go](../../shared/modules/ibc_client_module.go) for the details on the interfaces and their methods.

## Provable Stores

ICS-02 requires a lot of data to be stored in the IBC stores (defined in [ICS-24](./ics24.md)). In order to do this the provable stores must be initialised on a per client ID basis. This means that any operation using the provable store does not require the use of the `clientID`. This is done as follows:

```go
prefixed := path.ApplyPrefix(core_types.CommitmentPrefix(path.KeyClientStorePrefix), identifier)
clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(string(prefixed))
```

This allows the `clientStore` to be used by the WASM clients without them needing to keep track of their unique identifiers.

See: [ibc/client/submodule.go](../client/submodule.go) for more details on how this is used.

[ics02]: https://github.com/cosmos/ibc/blob/main/spec/core/ics-002-client-semantics/README.md
[ics08]: https://github.com/cosmos/ibc/blob/main/spec/client/ics-008-wasm-client/README.md
