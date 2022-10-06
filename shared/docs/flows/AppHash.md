# AppHash <!-- omit in toc -->

- [Context Initialization](#context-initialization)
- [Block Application](#block-application)
- [Block Commit](#block-commit)

## Context Initialization

```mermaid
sequenceDiagram
    %% autonumber
    participant N as Node
    participant C as Consensus
    participant U as Utility
    participant P as Persistence
    participant PP as Persistence (PostgresDB)
    participant PM as Persistence (MerkleTree)
    participant P2P as P2P

    N-->>C: HandleMessage(anypb.Any)
    critical NewRound Message
        C->>+U: NewContext(height)
        U->>P: NewRWContext(height)
        P->>U: PersistenceRWContext
        U->>U: Store persistenceContext
        U->>-C: UtilityContext
        C->>C: Store utilityContext
        Note over C, PM: See 'Block Application'
    end

    Note over N, P2P: Hotstuff lifecycle
    N-->>C: HandleMessage(anypb.Any)

    critical Decide Message
        Note over C, PM: See 'Block Commit'
    end
```

## Block Application

```mermaid
sequenceDiagram
    participant C as Consensus
    participant U as Utility
    participant P as Persistence
    participant PP as Persistence (PostgresDB)
    participant PM as Persistence (MerkleTree)

    alt as leader
        C->>+U: GetProposalTransactions(proposer, maxTxBz, [lastVal])
        U->>U: reap mempool
        U->>-C: txs
        Note over C, U: Perform replica behaviour
    else as replica
        C->>+U: ApplyBlock(height, proposer, txs, lastVals)
        loop Update DB: for each operation in tx
            U->>P: ReadOp | WriteOp
            P->>PP: ReadOp | WriteOp
            PP->>P: result, err_code
            P->>U: result, err_code
            U->>U: validate
            U->>P: StoreTransaction(tx)
            P->>P: store locally
            P->>U: result, err_code
        end
        U->>+P: UpdateAppHash()
        loop for each protocol actor type
            P->>PP: GetActorsUpdate(height)
            PP->>P: actors
            loop Update Tree: for each actor
                P->>PM: Update(addr, serialized(actor))
                PM->>P: result, err_code
            end
            P->>PM: GetRoot()
            PM->>P: rootHash
        end
        P->>P: computeStateHash(rootHashes)
        P->>-U: stateHash
        U->>-C: hash
    end
```

## Block Commit

```mermaid
sequenceDiagram
    %% autonumber
    participant C as Consensus
    participant U as Utility
    participant P as Persistence
    participant PP as Persistence (PostgresDB)
    participant PK as Persistence (Key-Value Store)
    C->>U: CommitContext(quorumCert)
    U->>P: Commit(proposerAddr, quorumCert)
    P->>P: create typesPer.Block
    P->>PP: insertBlock(block)
    PP->>P: result, err_code
    P->>PK: Put(height, block)
    PK->>P: result, err_code
    P->>P: commit tx
    P->>U: result, err_code
    U->>P: Release()
    P->>U: result, err_code
    C->>U: Release()
    U->>C: result, err_code
    C->>C: release utilityContext
```
