# State Sync Protocol Design <!-- omit in toc -->

_NOTE: This document makes some assumption of P2P implementation details, so please see [p2p](../../p2p/README.md) for the latest source of truth._

- [Background](#background)
- [State Sync - Peer Metadata](#state-sync---peer-metadata)
- [State Sync - Operation Modes](#state-sync---operation-modes)
  - [Sync Mode](#sync-mode)
  - [Synced Mode](#synced-mode)
  - [Pacemaker Mode](#pacemaker-mode)
  - [Server Mode](#server-mode)
  - [Operation Modes Lifecycle](#operation-modes-lifecycle)
- [State Sync Designs](#state-sync-designs)
  - [Block by Block](#block-by-block)
    - [Synchronous](#synchronous)
    - [Asynchronous](#asynchronous)
  - [Future Work - Designs](#future-work---designs)
- [Research Items](#research-items)
- [Glossary](#glossary)

## Background

State Sync is a protocol within a `Pocket` node that enables the download and maintenance of the latest world state. This protocol enables network actors to participate in network activities (like _Consensus_ or _Web3 Provisioning & Access_) in present time, by ensuring the synchronization of the individual node with the collective network.

## State Sync - Peer Metadata

A node participating in the `State Sync` protocol can act as both a _server_ and/or a _client_ to its `Network Peers`. A pre-requisite of the State Sync protocol is for the `P2P` module to maintain an active set of network peers, along with metadata corresponding to the persistence data they have available.

Illustrative example of Peer Metadata related to State Sync:

```golang
type PeerSyncMetadata interface {
  // ...
  GetPeerID() string   // An ID (e.g. a derivative of a PublicKey) associated with the Peer
  GetMaxHeight() int64 // The maximum height the peer has in its BlockStore
  GetMinHeight() int64 // The minimum height the peer has in the BlockStore
  // ...
}
```

This data can be collected through the `P2P` module during the `Churn Management Protocol`. It can also be abstracted to an `ask-response` cycle where the node continuously asks this meta-information of its active peers.

```mermaid
sequenceDiagram
    autonumber

    actor N as Node
    participant NP as Network Peer(s)

    loop Churn Management
        N->>+NP: Are you alive? If so, what's your Peer Metadata?
        NP->>N: Yup, here's my Peer Metadata. What's yours?
        N->>NP: ACK, I'll ask again in a bit to make sure I'm up to date.
    end
```

The aggregation and consumption of this peer-meta information enables the State Sync protocol by enabling the node to understand the globalized network state by sampling Peer Metadata through its local peer list.
This gives a view into the data availability layer, with details of what data can be consumed from which peer.

```golang
type PeerSyncAggregateMetadata interface {
  // ...
  GetPeerMetadata() []PeerSyncMeta // The current list of Peers and their known metadata
  GetMaxPeerHeight() uint64  // The maximum height associated with all known Peers
  // ...
}
```

Using the `PeerSyncAggregateMetadata`, a Node is able to compare its local `SyncState` against that of the Global Network that is visible to it (i.e. the world state).

## State Sync - Operation Modes

State sync can be viewed as a state machine that transverses various modes the node can be in, including:

- Sync Mode
- Synced Mode
- Pacemaker Mode
- Server Mode

The functionality of the node depends on the mode it is operating it. Note that the modes are not necessarily mutually exclusive. For example, the node can be in `Server Mode` and `Synced Mode` at the same time.

For illustrative purposes below assume:

- `localSyncState` is an object instance complying with the `PeerSyncMetadata` interface for the local node
- `globalSyncMeta` is an object instance complying with the `PeerSyncAggregateMetadata` interface for the global network

### Sync Mode

The Node is in `Sync` Mode if `localSyncState.MaxHeight < GlobalSyncMeta.Height`.

In `Sync` Mode, the Node is catching up to the latest block by making `GetBlock` requests to eligible peers in its address book. A peer can handle a `GetBlock` request if `PeerSyncMetadata.MinHeight` <= `localSyncState.MaxHeight` <= `PeerSyncMetadata.MaxHeight`.

Though it is unspecified whether or not a Node may make `GetBlock` requests in order or in parallel, the cryptographic restraints of block processing require the Node to call `ApplyBlock` sequentially until it is `Synced`.

### Synced Mode

The Node is in `Synced` mode if `localSyncState.Height == globalSyncMeta.MaxHeight`.

In `SyncedMode`, the Node is caught up to the latest block (based on the visible view of the network) and relies on new blocks to be propagated via the P2P network every time the Validators finalize a new block during the consensus lifecycle.

### Pacemaker Mode

The Node is in `Pacemaker` mode if the Node is in `Synced` mode **and** is an active Validator at the current height.

In `Pacemaker` mode, the Node is actively participating in the HotPOKT lifecycle.

### Server Mode

The Node can serve data to other nodes, upon request, if `ServerMode` is enabled. This sub-protocol runs in parallel to the node's own state sync in order to enable other peers to catch up.

### Operation Modes Lifecycle

```mermaid
flowchart TD
    A[StateSync] --> B{Caught up?}
    A[StateSync] --> P{ServerMode <br> enabled?}

    %% Is serving peers?
    P --> |Yes| Q[Serve Mode]
    P --> |No| R[Noop]
    Q --> |Serve blocks<br>upon request| Z[Peers]

    %% Is caught up?
    B --> |Yes| C{Is Validator?}
    B --> |No| D[SyncMode]

    %% Synching
    D --> |Request blocks| Z[Peers]

    %% Is a validator?
    C --> |No| E[Synched Mode]
    C --> |Yes| F(Pacemaker Mode<br>*HotPOKT*)
    E --> |Listen for<br>new blocks| Z[Peers]

    %% Loop back
    Z[Peers] --> |Blocks| A[StateSync]

```

_IMPORTANT: If any blocks processed result in an invalid `AppHash` during `ApplyBlock`, a new `BlockRequest` must be issued until a valid block is found._

## State Sync Designs

### Block by Block

The block-by-block has a node request one block from its peers at a time and apply them as they are received.

#### Synchronous

```mermaid
sequenceDiagram
  actor N as Node
  actor P as Peer(s)

  loop continous
    N ->> P: Request Metadata
    P ->> N: Metadata
    N ->> N: Update local directory
  end

  loop until caught up
    N ->> P: Request Block
    P ->> N: Block
    N ->> N: ApplyBlock
  end
```

#### Asynchronous

```mermaid
sequenceDiagram
  actor N as Node
  actor P as Peer(s)

  loop continous
    N -->> P: Send Metadata Request
    P -->> N: Send Metadata Response
    N ->> N: Update local directory
  end

  loop until caught up
    N -->> P: Send Block Request
    P -->> N: Send Block Response
    N ->> N: ApplyBlock
  end
```

### Future Work - Designs

- `Fast Sync Design` - Sync only the last `N` blocks from a _snapshot_ containing a network state

- `Optimistic Sync Design` - Optimize the State Sync protocol by parallelling requests and computation with pre-fetching and local caching

- `Block Chunk Design` - Update the singe block-by-block to be able to receive and provide multiple blocks per request.

- `Block Stream Design` - Update the Block by Block design to stream blocks via a WebSocket from a single connectable peer.

## Research Items

_TODO(M5): Create issues to track and discuss these work items in the future_

- Investigate how does the persistence layer design of `pruning` Merkle Tree affects `StateSync`.
- Investigate how DB Pruning of the `SQL DB` might affect fast sync.
- Investigate how churn management in relation to `RainTree` could provide opportunities or obstacles with StateSync.

## Glossary

- `ApplyingBlock`: The process of playing block parts and its subsequent transactions against the Node's world state using the Utility Module and Validating the `AppHash` contained in the block against the produced `AppHash` from the local state.
- `BlockRequests`: A message from an active peer, requesting a block to sync the chain to the Global Network State.
- `Churn Management Protocol`: The protocol in Pocket's P2P Module that ensures the most updated and valid Network Peer list possible.
- `Network Peer`: Another node on the network that this node can directly communicate with, without going through a third-party server. Peers may start the connection through an `inbound` or `outbound` initialization to share and transmit data.
- `SyncState`: The local block state of the node vs the global network block state.

<!-- If you are a core team member, you can also read or contribute to additional rough notes [here](https://www.notion.so/pocketnetwork/Would-you-like-a-proof-with-that-8dd05f8ea6274505a3072c9fb4e44898) -->
