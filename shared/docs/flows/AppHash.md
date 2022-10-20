# AppHash <!-- omit in toc -->

- [Context Initialization](#context-initialization)
- [Block Application](#block-application)
- [Block Commit](#block-commit)

<!-- See if there's an answer in this question to add links to notes: https://stackoverflow.com/questions/74103729/adding-hyperlinks-to-notes-in-mermaid-sequence-diagrams -->

## Context Initialization

This flow shows the process of context initialization between all the modules required to apply a block and compute a state hash during the consensus lifecycle.

The `Hotstuff lifecycle` part refers to the so-called `PreCommit` and `Commit` phases of the protocol.

```mermaid
sequenceDiagram
    participant N as Node
    participant C as Consensus
    participant U as Utility
    participant P as Persistence
    participant P2P as P2P

    N-->>C: HandleMessage(anypb.Any)
    critical NewRound Message
        C->>+U: NewContext(height)
        U->>+P: NewRWContext(height)
        P->>-U: PersistenceRWContext
        U->>U: store context<br>locally
        activate U
        deactivate U
        U->>-C: UtilityContext
        C->>C: store context<br>locally
        activate C
        deactivate C
        Note over C, P: See 'Block Application'
    end

    Note over N, P2P: Hotstuff lifecycle
    N-->>C: HandleMessage(anypb.Any)

    critical Decide Message
        Note over C, P: See 'Block Commit'
    end
```

## Block Application

This flow shows how the `leader` and the `replica`s behave in order to apply a `block` and return a `stateHash`.

```mermaid
sequenceDiagram
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

    %% Prepare or get block as leader
    opt if leader
        C->>U: GetProposalTransactions(proposer, maxTxBz, [lastVal])
        activate U
        alt no QC in NewRound
        U->>U: reap mempool <br> & prepare block
        activate U
        deactivate U
    else
        U->>U: find QC <br> & get block
        activate U
        deactivate U
        end
        U-->>C: txs
        deactivate U
    end

    %% Apply block as leader or replica
    C->>+U: ApplyBlock(height, proposer, txs, lastVals)
    loop [for each op in tx] for each tx in txs
        U->>+P: TransactionExists(txHash)
        P->>-U: true | false
        opt if tx is not indexed
            U->>+P: Get*/Set*
            P-->>-U: result, err_code
            U->>U: Validation logic
            activate U
            deactivate U
            U->>+P: StoreTransaction(tx)
            P->>P: Store tx locally
            activate P
            deactivate P
            P-->>-U: result, err_code
        end
    end
    U->>+P: UpdateAppHash()
    P->>P: Update state hash
    activate P
    deactivate P
    P-->>-U: stateHash
    U-->>-C: stateHash
```

The [V1 Persistence Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/persistence) outlines the use of a **PostgresDB** and **Merkle Trees** to implement the `Update State Hash` component. This is an internal detail which can be done differently depending on the implementation. For the core V1 implementation, see the flows outlined [here](../../../persistence/docs/AppHash.md).

## Block Commit

```mermaid
sequenceDiagram
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

    %% Commit Context
    C->>+U: CommitContext(quorumCert)
    U->>+P: Commit(proposerAddr, quorumCert)
    P->>P: See 'Store Block'
    P->>-U: result, err_code
    U->>+P: Release()
    P->>-U: result, err_code
    deactivate U

    %% Release Context
    C->>+U: Release()
    U->>-C: result, err_code
    C->>C: release utilityContext
    activate C
    deactivate C
```

The [V1 Persistence Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/persistence) outlines the use of a **key-value store** to implement the `Create And Store Block` component. This is an internal detail which can be done differently depending on the implementation. For the core V1 implementation, see the flows outlined [here](../../../persistence/docs/AppHash.md).
