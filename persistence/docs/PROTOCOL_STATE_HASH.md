DO_NOT_REVIEW_THIS_YET

# AppHash <!-- omit in toc -->

This document describes the persistence module internal implementation of how the state hash is updated. Specifically, what happens once the `UpdateStateHash` function in [persistence module interface](../../shared/modules/persistence_module.go) is called.

## Update State Hash

This flow shows the interaction between the PostgresDB and MerkleTrees to compute the state hash.

```mermaid
sequenceDiagram
    participant P as Persistence Module
    participant PP as Persistence (SQLDatabase)
    participant PM as Persistence (MerkleTree)

    loop for each protocol actor type
        P->>+PP: GetActorsUpdated(height)
        PP->>-P: actors
        loop for each state tree
            P->>+PM: Update(addr, serialized(actor))
            PM->>-P: result, err_code
        end
        P->>+PM: GetRoot()
        PM->>-P: rootHash
    end

    P->>P: stateHash = hash(aggregated(rootHashes))
    activate P
    deactivate P
```

## Store Block

This flow shows the interaction between the PostgresDB and Key-Value Store to compute the state hash.

```mermaid
sequenceDiagram
    %% autonumber
    participant P as Persistence
    participant PP as Persistence (PostgresDB)
    participant PK as Persistence (Key-Value Store)

    activate P
    P->>P: reap stored transactions
    P->>P: prepare, serialize <br> & store block
    deactivate P

    P->>+PP: insertBlock(height, serialized(block))
    PP->>-P: result, err_code
    P->>+PK: Put(height, serialized(block))
    PK->>-P: result, err_code
```
