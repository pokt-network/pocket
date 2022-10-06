# AppHash <!-- omit in toc -->

## Update State Hash

This flow shows the interaction between the PostgresDB and MerkleTrees to compute the state hash.

```mermaid
sequenceDiagram
    participant P as Persistence Module
    participant PP as Persistence (PostgresDB)
    participant PM as Persistence (MerkleTree)

    loop for each protocol actor type
        P->>PP: GetActorsUpdated(height)
        PP->>P: actors
        loop update tree for each actor
            P->>PM: Update(addr, serialized(actor))
            PM->>P: result, err_code
        end
        P->>PM: GetRoot()
        PM->>P: rootHash
    end
    P->>P: stateHash = hash(aggregated(rootHashes))
```

## Store Block

This flow shows the interaction between the PostgresDB and Key-Value Store to compute the state hash.

```mermaid
sequenceDiagram
    %% autonumber
    participant P as Persistence
    participant PP as Persistence (PostgresDB)
    participant PK as Persistence (Key-Value Store)

    P->>P: reap stored transactions
    P->>P: create & serialize<br>`typesPer.Block`
    P->>PP: insertBlock(height, serialized(block))
    PP->>P: result, err_code
    P->>PK: Put(height, serialized(block))
    PK->>P: result, err_code

```
