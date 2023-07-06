# P2P Module <!-- omit in toc -->

This document is meant to be a supplement to the living specification of [1.0 Pocket's P2P Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/p2p) primarily focused on the implementation, and additional details related to the design of the codebase and information related to development.

## Table of Contents <!-- omit in toc -->

- [Definitions](#definitions)
- [Interface & Integration](#interface--integration)
- [Module Architecture](#module-architecture)
  - [Architecture Design Language](#architecture-design-language)
  - [Legends](#legends)
  - [P2P Module / Router Decoupling](#p2p-module--router-decoupling)
  - [Message Propagation & Handling](#message-propagation--handling)
  - [Message Deduplication](#message-deduplication)
  - [Peer Discovery](#peer-discovery)
  - [Code Organization](#code-organization) 
- [Testing](#testing)
  - [Running Unit Tests](#running-unit-tests)
  - [RainTree testing framework](#raintree-testing-framework)
    - [Helpers](#helpers)
    - [Test Generators](#test-generators)

## Definitions

### "gossip"

When used generally, shorthand for "message propogation"; **not to be confused with "gossipsub"**.

### "gossipsub"

A specific ["pubsub"](https://pkg.go.dev/github.com/libp2p/go-libp2p-pubsub) router implementation/protocol which uses information (referred to internally as 'gossip' by LibP2P, **distinct from our definition**) about which peers have seen which messages to facilitate "gossip" (as per our definition).

### "raintree gossip"

A structured "gossip" protocol (and implementation) which uses the raintree algorithm for peer selection. Used between all staked actors to optimize for "gossip" speed.

### "background gossip"

A "gossip" protocol (implementation TBD) which facilitates "gossip" to all P2P participants, including non-staked actors (e.g. full-nodes).

## Interface & Integration

This module aims to implement the interface specified in [`pocket/shared/modules/p2p_module.go`](../shared/modules/p2p_module.go).

_(TODO: diagram legend)_
```mermaid
flowchart TD
    subgraph P2P["P2P Module"]
        L(Host)
        H(Handler)
        B(Bus)
        NM(Routers)
        L --> H
        H <--> NM
    end
    PN((Pocket Network))
    N([Pocket Node])

    B <-.->NM
    B <-.P2P Interface.-> N
    PN -.-> L
    NM -- Send/Broadcast --> PN

    classDef pocket_node fill:#28f,stroke:#016,stroke-width:2px,color:#fff,stroke-dasharray: 5 5
    class N pocket_node

    classDef pocket_network fill:#000fff,stroke:#016,stroke-width:2px,color:#fff,stroke-dasharray: 5 5
    class PN pocket_network
```

`Routers` is where [RainTree](https://github.com/pokt-network/pocket/files/9853354/raintree.pdf) is implemented.
See [`raintree/router.go`](./raintree/router.go) for the specific implementation of RainTree, but please refer to the [specifications](https://github.com/pokt-network/pocket-network-protocol/tree/main/p2p) for more details.

## Module Architecture

_(TODO: move "arch. design lang." & "legends" sections into `shared` to support common usage)_

### Architecture Design Language

The architecture design language expressed in this documentation is based on [UML](https://www.uml-diagrams.org/).
Due to limitations in the current version of mermaid, class diagrams are much more adherant to the UML component specification.
Component diagrams however are much more loosely inspired by their UML counterparts.

Regardless, each architecture diagram should be accompanied by a legend which covers all the design language features used to provide disambiguation. 

References:
- [Class Diagrams](https://www.uml-diagrams.org/class-diagrams-overview.html)
- [Component Diagrams](https://www.uml-diagrams.org/component-diagrams.html)

  _NOTE: mermaid does not support ports, interfaces, ... in component diagrams ("flowcharts)._

### Legends

```mermaid
flowchart
subgraph Legend
    m[[`Method`]]
    c[Component]

    m -- "unconditional usage" --> c
    m -. "conditional usage" .-> c
    m -. "ignored" .-x c
end
```

```mermaid
classDiagram
class ConcreteType {
  +ExportedField
  -unexportedField
  +ExportedMethod(...argTypes) (...returnTypes)
  -unexportedMethod(...argTypes) (...returnTypes)
}

class InterfaceType {
    <<interface>>
    +Method(...argTypes) (...returnTypes)
}

ConcreteType --|> InterfaceType : Interface realization

ConcreteType --> OtherType : Direct usage
ConcreteType --o OtherType : Composition
ConcreteType --* OtherType : Aggregatation
ConcreteType ..*  "(cardinality)" OtherType : Indirect (via interface)
```

#### Interface Realization

> Realization is a specialized abstraction relationship between two sets of model elements, one representing a specification (the supplier) and the other represents an implementation of the latter (the client).

> Realization can be used to model stepwise refinement, optimizations, transformations, templates, model synthesis, framework composition, etc.

_(see: [UML Realization](https://www.uml-diagrams.org/realization.html))_

#### Direct Usage

> Dependency is a directed relationship which is used to show that some UML element or a set of elements requires, needs or depends on other model elements for specification or implementation. Because of this, dependency is called a supplier - client relationship, where supplier provides something to the client, and thus the client is in some sense incomplete while semantically or structurally dependent on the supplier element(s). Modification of the supplier may impact the client elements.

> Usage is a dependency in which one named element (client) requires another named element (supplier) for its full definition or implementation.

_(see: [UML Dependency](https://www.uml-diagrams.org/dependency.html))_

#### Composition

> A "strong" form of aggregation

> If a composite (whole) is deleted, all of its composite parts are "normally" deleted with it.

_(see: [UML Shared composition](https://www.uml-diagrams.org/composition.html))_

#### Aggregation

> A "weak" form of composition

> Shared part could be included in several composites, and if some or all of the composites are deleted, shared part may still exist.

_(see: [UML Shared aggregation](https://www.uml-diagrams.org/aggregation.html))_

#### Cardinality

Cardinality indicates the number or range of simultaneous instances of the classifier at the "cardinality-side" association end that are associated with the classifier at the other end of the given association type.

_(see: [UML Association](https://www.uml-diagrams.org/association.html#association-end))_


### P2P Module / Router Decoupling

The P2P module encapsulates the `RainTreeRouter` and `BackgroundRouter` submodules.
The P2P module internally refers to these as the `stakedActorRouter` and `unstakedActorRouter`, respectively.

Depending on the necessary routing scheme (unicast / broadcast) and whether the peers involved are staked actors, a node will use one or both of these routers.

**Unicast**

| Sender         | Receiver       | Router          | Example Usage                                        |
|----------------|----------------|-----------------|------------------------------------------------------|
| Staked Actor   | Staked Actor   | Raintree only   | Consensus (state sync) messages (to validators only) |
| Unstaked Actor | Staked Actor   | Background only | Consensus (state sync) messages (to validators only) |
| Unstaked Actor | Unstaked Actor | Background only | Consensus (state sync) & Debug (CLI) messages        |

**Broadcast**

| Broadcaster    | Receiver       | Router                | Example Usage                              |
|----------------|----------------|-----------------------|--------------------------------------------|
| Staked Actor   | Staked Actor   | Raintree + Background | Utility tx messages                        |
| Unstaked Actor | Staked Actor   | Background only       | Utility tx messages (gossipsub redundancy) |
| Unstaked Actor | Unstaked Actor | Background only       | Utility tx messages                        |

Both router submodule implementations embed a `UnicastRouter` which enables them to send and receive messages directly to/from a single peer.

**Class Diagram**

```mermaid
classDiagram
    class p2pModule {
        -stakedActorRouter Router
        -unstakedActorRouter Router
        -handlePocketEnvelope([]byte) error
    }

    class P2PModule {
        <<interface>>
        GetAddress() (Address, error)
        HandleEvent(*anypb.Any) error
        Send([]byte, Address) error
        Broadcast([]byte) error
    }
    p2pModule --|> P2PModule

    class RainTreeRouter {
        UnicastRouter
        -handler MessageHandler
        +Broadcast([]byte) error
        -handleRainTreeMsg([]byte) error
    }

    class BackgroundRouter {
        UnicastRouter
        -handler MessageHandler
        +Broadcast([]byte) error
        -handleBackgroundMsg([]byte) error
        -readSubscription(subscription *pubsub.Subscription)
    }

    class UnicastRouter {
        -messageHandler MessageHandler
        -peerHandler PeerHandler
        +Send([]byte, Address) error
        -handleStream(libp2pNetwork.Stream)
        -readStream(libp2pNetwork.Stream)
    }
    RainTreeRouter --* UnicastRouter : (embedded)
    BackgroundRouter --* UnicastRouter : (embedded)

    p2pModule --o "2" Router
    p2pModule ..* RainTreeRouter : (`stakedActorRouter`)
    p2pModule ..* BackgroundRouter : (`unstakedActorRouter`)
    
    class Router {
        <<interface>>
        +Send([]byte, Address) error
        +Broadcast([]byte) error
    }
    BackgroundRouter --|> Router
    RainTreeRouter --|> Router
```

### Message Propagation & Handling

**Unicast**

```mermaid
flowchart
    subgraph lp2p["Local P2P Module (outgoing)"]
        lps[[`Send`]]
        lps -. "(iff local & remote peer are staked)" ..-> lrtu
        lps -. "(if local or remote peer are not staked)" .-> lbgu

        lbgu -- "opens stream\nto target peer" ---> lhost

        lhost[Libp2p Host]

        subgraph lrt[RainTree Router]
            subgraph lRTPS[Raintree Peerstore]
              lStakedPS([staked actors only])
            end
            
            lrtu[UnicastRouter]

            lrtu -- "network address lookup" --> lRTPS
        end

        lrtu -- "opens a stream\nto target peer" ---> lhost

        subgraph lbg[Background Router]
            lbgu[UnicastRouter]
            subgraph lBGPS[Background Peerstore]
              lNetPS([all P2P participants])
            end

            lbgu -- "network address lookup" --> lBGPS
        end
    end

    subgraph rp2p["Remote P2P Module (incoming)"]
        rhost[Libp2p Host]

        subgraph rrt[RainTree Router]
            rrth[[`RainTreeMessage` Handler]]
            rrtu[UnicastRouter]
        end

        subgraph rbg[Background Router]
            rbgh[[`BackgroundMessage` Handler]]
            rbgu[UnicastRouter]
            rbgu --> rbgh
        end

        rp2ph[[`PocketEnvelope` Handler]]
        rbus[bus]
        rhost -. "new stream" .-> rrtu
        rhost -- "new subscription message" --> rbgu
        rrtu --> rrth

        rnd[Nonce Deduper]
        rp2ph -- "deduplicate msg mempool" --> rnd
    end


    rp2ph -. "(iff not duplicate msg)\npublish event" .-> rbus

    rrth --> rp2ph
    rbgh --> rp2ph

    lhost --> rhost
```

**Broadcast**

```mermaid
flowchart
  subgraph lp2p["Local P2P Module (outgoing)"]
    lpb[[`Broadcast`]]
    lpb -. "(iff local & remote peer are staked)" ..-> lrtu
    lpb -- "(always)" --> lbggt
    
    lbggt -- "msg published\n(gossipsub protocol)" ---> lhost
    
    lhost[Libp2p Host]

    subgraph lrt[RainTree Router]
      subgraph lRTPS[Raintree Peerstore]
        lStakedPS([staked actors only])
      end
      
      lrtu[UnicastRouter]
      
      lrtu -- "network address lookup" --> lRTPS
    end
    
    lrtu -- "opens a stream\nto target peer" ---> lhost

    subgraph lbg[Background Router]
      lbggt[Gossipsub Topic]
      subgraph lBGPS[Background Peerstore]
        lNetPS([all P2P participants])
      end
      
      lbggt -- "network address lookup" --> lBGPS
    end
  end

  subgraph rp2p["Remote P2P Module (incoming)"]
    rhost[Libp2p Host]

    subgraph rrt[RainTree Router]
      rrth[[`RainTreeMessage` Handler]]
      rrtu[UnicastRouter]
    end

    subgraph rbg[Background Router]
      rbgh[[`BackgroundMessage` Handler]]
      rbgg[Gossipsub Subscription]
      rbggt[Gossipsub Topic]
      rbgg --> rbgh
      rbgh -- "(background msg\npropagation cont.)" ---> rbggt
    end

    rp2ph[[`PocketEnvelope` Handler]]
    rbus[bus]
    rhost -. "new stream" ..-> rrtu
    rhost -- "new subscription message" --> rbgg
    rbggt -- "(background msg\npropagation cont.)" --> rhost
    rrtu --> rrth
    rrth -. "(iff level > 0)\n(raintree msg\npropagation cont.)" .-> rrtu
    rrtu -- "(raintree msg\npropagation cont.)" --> rhost

    rnd[Nonce Deduper]
    rp2ph -- "deduplicate msg mempool" --> rnd
  end


  rp2ph -. "(iff not duplicate msg)\npublish event" .-> rbus

  rrth --> rp2ph
  rbgh --> rp2ph

  lhost --> rhost
```

### Message Deduplication

Messages MUST be deduplicated before broadcasting their respective event over the bus since it is expected that nodes will receive duplicate messages (for multiple reasons).

The responsibility of deduplication is encapsulated by the P2P module, As such duplicate messages may come from multiple routers in some of these scenarios.

The `NondeDeduper` state is not persisted outside of memory and therefore is cleared during node restarts.

```mermaid
classDiagram
    class RainTreeMessage {
        <<protobuf>>
        +Level uint32
        +Data []byte
    }

    class BackgroundMessage {
        <<protobuf>>
        +Data []byte
    }
    
    class PocketEnvelope {
        <<protobuf>>
        +Content *anypb.Any
        +Nonce uint64
    }

    RainTreeMessage --* PocketEnvelope : serialized as `Data`
    BackgroundMessage --* PocketEnvelope : serialized as `Data`
    
    
    class p2pModule {
        -handlePocketEnvelope([]byte) error
    }

    class P2PModule {
        <<interface>>
        GetAddress() (Address, error)
        HandleEvent(*anypb.Any) error
        Send([]byte, address Address) error
        Broadcast([]byte) error
    }
    p2pModule --|> P2PModule

    class RainTreeRouter {
        UnicastRouter
        -handler MessageHandler
        +Broadcast([]byte) error
        -handleRainTreeMsg([]byte) error
    }

    class NonceDeduper {
        Push(Nonce) error
        Contains(Nonce) bool
    }

  class Bus {
    <<interface>>
    PublishEventToBus(PocketEnvelope)
    GetBusEvent() PocketEnvelope
  }
  p2pModule --> Bus

    class BackgroundRouter {
        UnicastRouter
        -handler MessageHandler
        +Broadcast([]byte) error
        -handleBackgroundMsg([]byte) error
        -readSubscription(subscription *pubsub.Subscription)
    }

    class UnicastRouter {
        -messageHandler MessageHandler
        -peerHandler PeerHandler
        +Send([]byte, address Address) error
        -handleStream(stream libp2pNetwork.Stream)
        -readStream(stream libp2pNetwork.Stream)
    }
    RainTreeRouter --* UnicastRouter : (embedded)
    BackgroundRouter --* UnicastRouter : (embedded)

    p2pModule ..* RainTreeRouter
    RainTreeRouter --o RainTreeMessage
    
    p2pModule ..* BackgroundRouter
    BackgroundRouter --o BackgroundMessage

    p2pModule --o PocketEnvelope
    p2pModule --* NonceDeduper
```

#### Configuration

The size of the `NonceDeduper` queue is configurable via the `P2PConfig.MaxNonces` field.

### Peer Discovery

Peer discovery involves pairing peer IDs to their network addresses (multiaddr).
This pairing always has an associated TTL (time-to-live), near the end of which it must
be refreshed.

In the background gossip overlay network (`backgroundRouter`), peers will re-advertise themselves 7/8th through their TTL.
This refreshes the libp2p peerstore automatically.

In the raintree gossip overlay network (`raintreeRouter`), the libp2p peerstore is **NOT** currently refreshed _(TODO: [#859](https://github.com/pokt-network/network/isues/859))_.

```mermaid
flowchart TD
  subgraph bus
  end
  
  subgraph pers[Persistence Module]
  end

  subgraph cons[Consensus Module]
  end

  cons -- "(staked actor set changed)\npublish event" --> bus
  bus --> rPM
  rPM -- "get staked actors\nat current height" --> pers

  subgraph p2p["P2P Module"]
    host[Libp2p Host]
    host -- "incoming\nraintree message" --> rtu
    host -- "incoming\nbackground message" --> bgu
    host -- "incoming\ntopic message" --> bgr
    host -- "DHT peer discovery" --> rDHT
  
    subgraph rt[RainTree Router]
      subgraph rPS[Raintree Peerstore]
        rStakedPS([staked actors only])
      end

      subgraph rPM[PeerManager]
      end

      rtu[UnicastRouter]
      
      rPM -- "synchronize\n(add/remove)" --> rPS
      rtu -. "(no discovery)" .-x rPS
    end

    subgraph bg[Background Router]
      subgraph rBGPS[Background Peerstore]
        rNetPS([all P2P participants])
      end

      subgraph bgr[GossipSub Topic\nSubscription]
      end

      subgraph rDHT[Kademlia DHT]
      end

      bgu -- "add if new" --> rBGPS
      bgr -- "add if new" --> rBGPS
      rDHT -- "continuous import" --> rBGPS
         
      bgu[UnicastRouter]
    end

  end
```

### Raintree Router Architecture

_NOTE: If you (the reader) feel this needs a diagram, please reach out to the team for additional details._

### Code Organization

```bash
p2p
├── background
│   ├── kad_discovery_baseline_test.go  # Libp2p peer discovery via kademalia baseline example
│   ├── router.go                       # `BackgroundRouter` implementation of `Router` interface
│   └── router_test.go                  # `BackgroundRouter` functional tests
├── bootstrap.go                              # `p2pModule` bootstrap related method(s)
├── CHANGELOG.md
├── config
│   └── config.go
├── event_handler.go
├── module.go                                 # `p2pModule` definition
├── module_raintree_test.go                   # `p2pModule` & `RainTreeRouter` functional tests (routing)
├── module_test.go                            # `p2pModule` & `RainTreeRouter` integration tests
├── peer_test.go                              # `PeerList` unit test(s)
├── protocol
│   └── protocol.go                     # Common, pokt protocol-specific constants
├── providers
│   ├── current_height_provider
│   ├── peerstore_provider
│   └── providers.go
├── raintree
│   ├── peers_manager.go              # `rainTreePeersManager` implementation of `PeersManager` interface
│   ├── peers_manager_test.go
│   ├── peerstore_utils.go            # Raintree routing helpers
│   ├── router.go                     # `RainTreeRouter` implementation of `Router` interface
│   ├── router_test.go                # `RainTreeRouter` functional tests
│   ├── target.go                     # `target` definition
│   ├── testutil.go
│   └── utils_test.go
├── testutil.go
├── transport_encryption_test.go            # Libp2p transport security integration test
├── types
│   ├── background.pb.go
│   ├── errors.go
│   ├── libp2p_mocks.go
│   ├── mocks
│   ├── network_peer.go               # `NetworkPeer` implementation of `Peer` interface
│   ├── peer.go                       # `Peer` interface & `PeerList` definitions
│   ├── peer_manager.go               # `PeerManager` interface & `SortedPeerManager` implementation definitions
│   ├── peerstore.go                  # `Peerstore` interface & `PeerAddrMap` implementation definitions
│   ├── peers_view.go                 # `PeersView` interface & `sortedPeersView` implementation definitions
│   ├── peers_view_test.go
│   ├── proto
│   ├── raintree.pb.go
│   └── router.go                     # `Router` interface definition
├── unicast
│   ├── logging.go
│   ├── router.go
│   └── testutil.go
├── utils
│   ├── host.go                       # Helpers for working with libp2p hosts
│   ├── logging.go                    # Helpers for logging
│   ├── nonce_deduper.go
│   ├── nonce_deduper_test.go
│   ├── peer_conversion.go            # Helpers for converting between "native" and libp2p peer representations
│   ├── url_conversion.go             # Helpers for converting between  "native" and libp2p network address representations
│   └── url_conversion_test.go
└── utils_test.go
```

## Testing

_TODO: The work to add the tooling used to help with unit test generation is being tracked in #314._

### Running Unit Tests

```bash
make test_p2p
```

### RainTree testing framework

The testing framework for RainTree is a work-in-progress and can be found in `module_raintree_test.go`.

The `TestRainTreeCommConfig` struct contains a mapping of `validatorId` to the number of messages it expects to process during a RainTree broadcast:

- `numNetworkReads`: the # of asynchronous reads the node's P2P listener made (i.e. # of messages it received over the network)
- `numNetworkWrites`: the # of asynchronous writes the node's P2P listener made (i.e. # of messages it tried to send over the network)
- NOTE: A `demote` does not go over the network and is therefore not considered a `read`.

#### Helpers

Given a specific `originatorNode` which initiates the broadcast, the `testRainTreeCalls` helper function can be used to configure all the nodes and simulate a broadcast.

#### Test Generators

The [rain-tree-simulator](https://github.com/pokt-network/rain-tree-sim/blob/main/python) library contains an example Golang implementation and a Python implementation of RainTree simulators.

You can read the documentation in the [python simulator](https://github.com/pokt-network/rain-tree-sim/blob/main/python) on how it can be used to generate the unit tests found in `module_raintree_test.go`.

<!-- GITHUB_WIKI: p2p/readme -->
