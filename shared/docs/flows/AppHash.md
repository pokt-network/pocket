# AppHash <!-- omit in toc -->

- [Context Initialization](#context-initialization)
- [Block Application](#block-application)
- [Block Commit](#block-commit)

## Context Initialization

This flow shows the process of context initialization between all the modules required to apply a block and compute a state hash during the consensus lifecycle.

The `Hotstuff lifecycle` part refers to the so-called `PreCommit` and `Commit` phases of the protocol.

```mermaid
sequenceDiagram
    %% autonumber
    participant N as Node
    participant C as Consensus
    participant U as Utility
    participant P as Persistence
    participant P2P as P2P

    N-->>C: HandleMessage(anypb.Any)
    critical NewRound Message
        C->>+U: NewContext(height)
        U->>P: NewRWContext(height)
        P->>U: PersistenceRWContext
        U->>U: store context<br>locally
        U->>-C: UtilityContext
        C->>C: store context<br>locally
        Note over C, P: See 'Block Application'
    end

    Note over N, P2P: Hotstuff lifecycle
    N-->>C: HandleMessage(anypb.Any)

    critical Decide Message
        Note over C, P: See 'Block Commit'
    end
```

## Block Application

```mermaid
sequenceDiagram
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

    alt as leader
        C->>+U: GetProposalTransactions(proposer, maxTxBz, [lastVal])
        U->>U: reap mempool
        U->>-C: txs
        Note over C, U: fallthrough to replica behaviour
    else as replica
        C->>+U: ApplyBlock(height, proposer, txs, lastVals)
        loop for each operation in tx
            U->>P: Get*/Set*
            P->>U: result, err_code
            U->>U: validation<br>logic
            U->>P: StoreTransaction(tx)
            P->>P: store tx<br>locally
            P->>U: result, err_code
        end
        U->>+P: UpdateAppHash()
        Note over P: Update State Hash
        P->>-U: stateHash
        U->>-C: stateHash
    end
```

The [V1 Persistence Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/persistence) outlines the use of a **PostgresDB** and **Merkle Trees** to implement the `Update State Hash` component. This is an internal detail which can be done differently depending on the implementation. For the core V1 implementation, see the flows outlined [here](../../../persistence/docs/AppHash.md).

## Block Commit

```mermaid
sequenceDiagram
    %% autonumber
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

    C->>U: CommitContext(quorumCert)
    U->>P: Commit(proposerAddr, quorumCert)
    P->>P: reap stored transactions
    Note over P: Create And Store Block
    P->>U: result, err_code
    U->>P: Release()
    P->>U: result, err_code
    C->>U: Release()
    U->>C: result, err_code
    C->>C: release utilityContext
```

The [V1 Persistence Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/persistence) outlines the use of a **key-value store** to implement the `Create And Store Block` component. This is an internal detail which can be done differently depending on the implementation. For the core V1 implementation, see the flows outlined [here](../../../persistence/docs/AppHash.md).
