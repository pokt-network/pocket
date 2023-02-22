# Relay Protocol

## Background

The Relay Protocol is a fundamental sequence that makes up the building blocks of Pocket Network's Utility.

In Pocket Network, a `Relay` is a Read / Write API request operation to a 3rd party `RelayChain`.

The Relay Protocol is the servicing lifecycle that poises staked Servicers to be able to complete
Relays on behalf of the network.

### Lifecycle

The foundational lifecycle of the Relay Protocol is:

1. Validate the inbound `Relay`
2. Store / persist the `Relay`
3. Execute the Relay against the `RelayChain`

```mermaid
sequenceDiagram
    title Steps 1 to 3
    autonumber
    actor App
    actor Client
    actor Servicer
    participant Internal State
    participant Internal Storage
    participant External Relay Chain
    App->>Client: Provision(AppAuthToken)
    loop Repeats Throughout Session Duration
        Client->>Client: Sign(Relay)
        Client->>Servicer: Send(Relay)
        Servicer->>Internal State: Validate(Relay)
        Internal State->>Servicer: IsValid(Relay)
        Servicer->>Internal Storage: IfValid(Relay) -> Persist(Relay)
        Servicer->>External Relay Chain: Execute(Relay, RelayChainURL)
        External Relay Chain->>Servicer: RelayResponse = GetResponse(RelayChain)
        Servicer->>Servicer: Sign(RelayResponse)
        Servicer ->> Client: Send(RelayResponse)
    end
```

4. Wait for `Session` end / secret key to be revealed
5. Collect Volume Applicable Relays (based on secret key) from storage
6. Report Volume Applicable Relays to the assigned `Fisherman`

```mermaid
sequenceDiagram
	    title Steps 4-6
	    autonumber
	    actor Servicer
        participant Internal State
        participant Internal Storage
        actor Fisherman
	    loop Repeats Every Session End
	        Servicer->>Internal State: GetSecretKey(sessionHeader)
            Internal State->>Servicer: HashCollision = SecretKey(govParams)
	        Servicer->>Internal Storage: RelaysThatEndWith(HashCollision)
            Internal Storage->>Servicer: VolumeApplicableRelays
            Servicer->>Fisherman: Send(VolumeApplicableRelays)
	    end
```

### Validate the inbound `Relay`

A multi-step validation process to validate a submitted relay by a client before servicing

1. Validate payload, look for empty or 'bad' request data
2. Validate the metadata, look for empty or 'bad' metadata
3. Ensure the `RelayChain` is supported locally (in the servicer's configuration files)
4. Ensure session block height is current
5. Get the `sessionContext` to access values and parameters from world state at that height
6. Get the application object from the `request.AAT()` (using `sessionContext`)
7. Get session node count from that session height (using `sessionContext`)
8. Get max relays per session for the application (using `sessionContext`)
9. Ensure not over serviced (if max relays is exceeded, not compensated for further work)
10. Generate the session from seed data (see [Session Protocol](https://github.com/pokt-network/pocket/blob/main/utility/doc/PROTOCOLS.md))
11. Validate self against the session (is node within session)

```mermaid
graph TD
    A[Relay.Validate] -->B
    B[HasValidPayload] -->|Yes| C
    B -->|No| Z
    C[HasValidMeta] -->|Yes| D
    C -->|No| Z
    D[RelayChainIsSupported]-->|Yes| E
    D -->|No| Z
    E[IsValidSessionHeight]-->|Yes| F
    E -->|No| Z
    F[IsAppOverServiced]-->|No| G
    F -->|Yes| Z
    G[IsValidSession]-->|Yes| X
    G -->|No| Z
    X[Relay Is Valid]
    Z[Reject Invalid Relay]
```

### Store the `Relay`

Store a submitted `Relay` by a client for volume tracking

1. Marshal `Relay` object into codecBytes
2. Calculate the `hashOf(codecBytes)` <needed for volume tracking>
3. Persist `Relay` object, indexing under session

```mermaid
graph TD
    A[Relay.Store] -->|Encode `Relay` object| B
    B[RelaycodecBytes] -->|Calculate Hash for Volume Tracking| C
    C[RelaycodecBytes.Hash] -->|Add Hash to `Relay` and Persist| D
    D[Relay.Persist] --> |Indexing Under Session| E
    E[Key:SessionKey Val: Relays.AddNew]
```

### Execute the `Relay`

Execute a submitted `Relay` against the `RelayChain` by a client after validation

1. Retrieve the `RelayChain` url from the servicer configuration files
2. Execute http request with the `Relay Payload`
3. Format and digitally sign the response using the servicer's private key
4. Send back to client

##### Wait for Session to end / secret key to be revealed

It's important to note, the secret key isn't revealed by the network until the session is over
to prevent volume based bias. The secret key is usually a pseudorandom selection using the block hash as a seed.
_See the [Session Protocol](https://github.com/pokt-network/pocket/blob/main/utility/doc/PROTOCOLS.md) for more details._

### Get volume metric applicable `Relays` from store

1. Pull all `Relays` whose hash collides with the revealed secret key

`SELECT * FROM relay WHERE HashOf(relay) END WITH hashEndWith AND session=relay.Session`

2. This function also signifies deleting the non-volume-applicable `Relays`

### Report volume metric applicable relays to `Fisherman`

1. All volume applicable relays need to be sent to the assigned trusted `Fisherman` (selected by the [Session Protocol](https://github.com/pokt-network/pocket/blob/main/utility/doc/PROTOCOLS.md)) for a proper verification of the volume completed.
2. Send `volumeRelays` to `fishermanServiceURL` through http.

```mermaid
graph TD
    A[Relay.Execute] -->|Lookup Relay.RelayChain URL|B
    B[Chains.JSON -> RelayChainURL] -->|Execute Http Request| C
    C[RelayChain] -->|Return Response| D
    D[RelayResponse.Sign] --> |Send|E
    E[Requester]
```

## Alt Design

### Claim-Proof Lifecycle

An alternative design is a 2-step, claim-proof lifecycle where the individual servicers
build a Merkle sum index tree from all the relays, submits a root and subsequent Merkle proof to the
network via a commit+reveal schema.

- **Pros**: Can report volume metrics directly to the chain in a trustless fashion
- **Cons**: Large chain bloat, non-trivial compute requirement for creation of claim/proof transactions and trees,
  non-trivial compute requirement to process claim / proofs during ApplyBlock()

This algorithm is not yet documented anywhere, so the following links can act as a reference in the interim.

**Documentation:**

- Pocket docs: https://docs.pokt.network/home/v0/protocol/servicing#claim-proof-lifecycle
- Twitter Thread: https://twitter.com/o_rourke/status/1263847357122326530
- Plasma core Merkle Sum Tree: https://plasma-core.readthedocs.io/en/latest/specs/sum-tree.html

**V0 Source Code References:**

- Merkle: [pocketcore/types/merkle.go](https://github.com/pokt-network/pocket-core/blob/staging/x/pocketcore/types/merkle.go)
- Claim: [pocketcore/keeper/claim.go](https://github.com/pokt-network/pocket-core/blob/staging/x/pocketcore/keeper/claim.go)
- Proof: [pocketcore/keeper/proof.go](https://github.com/pokt-network/pocket-core/blob/staging/x/pocketcore/keeper/proof.go)

<!-- GITHUB_WIKI: utility/relay_protocol -->
