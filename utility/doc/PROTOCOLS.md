## Protocols

### Session Protocol
`Pocket` implements the V1 Utility Specification's Session Protocol by satisfying the following interface:

```golang
type Session interface {
    NewSession(sessionHeight int64, blockHash string, geoZone GeoZone, relayChain RelayChain, application modules.Actor) (Session, types.Error)
    GetServiceNodes() []modules.Actor // the ServiceNodes providing Web3 access to the Application
    GetFishermen() []modules.Actor    // the Fishermen monitoring the Service Nodes
    GetApplication() modules.Actor    // the Application consuming Web3 access
    GetRelayChain() RelayChain        // the identifier of the web3 Relay Chain
    GetGeoZone() GeoZone              // the geolocation zone where the Application is registered
    GetSessionHeight() int64          // the block height when the Session started
}
```

#### Session Creation Flow

1) Create a session object from the seed data (see #2)
2) Create a key concatenating and hashing the seed data
    - `key = Hash(sessionHeight + blockHash + geoZone + relayChain + appPublicKey)`
3) Get an ordered list of the public keys of serviceNodes who are:
    - actively staked
    - staked within geo-zone
    - staked for relay-chain
4) Pseudo-insert the session `key` string into the list and find the first actor directly below on the list
5) Determine a new seedKey with the following formula: ` key = Hash( key + actor1PublicKey )` where `actor1PublicKey` is the key determined in step 4
6) Repeat steps 4 and 5 until all N serviceNodes are found
7) Do steps 3 - 6 for Fishermen as well

### FAQ

- Q) why do we hash to find a newKey between every actor selection?
- A) pseudo-random selection only works if each iteration is re-randomized or it would be subject to lexicographical proximity bias attacks

- Q) Why do we not use Golang's `rand.Intn` with the key as a seed for random node selection?
- A) A proprietary randomization algorithm makes this approach language & library agnostic, so any client simply has to follow the specifications

- Q) what is `WorldState`?
- A) it represents a queryable view on the internal state of the network at a certain height.

- Q) Do Fishermen stake for a specific RelayChain?
- A) Fishermen are only going to be applicable to Pocket Supported Relay Chains (where the protocol pays out for the relay chain). It is unclear at this time what the limitations and scoping will be for Fishermen RelayChain support.

- Q) What was the reasoning not to allow a list of geozones?
- A) Each session is mono-chain and mono-geo. This is fundamental as it would create even more possible combinations of sessions and increase computational complexity during block production and servicing

### Session Flow

```mermaid
sequenceDiagram
    autonumber
    participant WorldState
    participant Session
    %% The `Qurier` is anyone (app or not) that asks to retrieve session information
    actor Querier
    Querier->>WorldState: Who are my sessionNodes and sessionFish for [app], [relayChain], and [geoZone]
    WorldState->>Session: seedData = height, blockHash, [geoZone], [relayChain], [app]
    Session->>Session: sessionKey = hash(concat(seedData))
    WorldState->>Session: nodeList = Ordered list of public keys of applicable serviceNodes
    Session->>Session: sessionNodes = pseudorandomSelect(sessionKey, nodeList, max)
    WorldState->>Session: fishList = Ordered list of public keys of applicable fishermen
    Session->>Session: sessionFish = pseudorandomSelect(sessionKey, fishList, max)
    Session->>Querier: SessionNodes, SessionFish
```

### Pseudorandom Selection

```mermaid
graph TD
    D[Pseudorandom Selection] -->|Ordered list of actors by pubKey|A
    A[A1, A2, A3, A4] -->|Insert key in Ooder| B[A1, A2, A3, Key, A4]
    B --> |A4 is selected due to order| C{A4}
    C --> |Else| E["Key = Hash(A4) + Key "]
    E --> A
    C --> |IF selection is maxed| F[done]
```

## How to build

Utility Module does not come with its own cmd executables.

Rather, it is purposed to be a dependency (i.e. library) of other modules

## How to use

Utility implements the `UtilityModule` and subsequent interface
[`pocket/shared/modules/utility_module.go`](https://github.com/pokt-network/pocket/shared/modules/utility_module.go).

To use, simply initialize a Utility instance using the factory function like so:

```go
utilityMod, err := utility.Create(config)
```

and use `utilityMod` as desired.

## How to test

```
$ make test_utility_types && make test_utility_module
```

## Code Organization

```bash
utility
├── account.go     # utility context for accounts & pools
├── actor.go       # utility context for apps, fish, nodes, and validators
├── block.go       # utility context for blocks
├── gov.go         # utility context for dao & parameters
├── module.go      # module implementation and interfaces
├── session.go     # utility context for the session protocol
├── transaction.go # utility context for transactions including handlers
├── doc            # contains the documentation and changelog
├── test           # utility unit tests
├── types          # stateless (without relying on persistence) library of utility types
│   ├── proto          # protobuf3 messages that auto-generate into the types directory
│   ├── actor.go
│   ├── error.go
│   ├── mempool.go
│   ├── message.go     # payloads of transactions
│   ├── transaction.go # the finite unit of the block
│   ├── util.go
│   ├── vote.go        # vote structure for double sign transaction
```
