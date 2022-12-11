# State Hash <!-- omit in toc -->

This document describes the cross-module communication using the interfaces in [../shared/modules](../shared/modules) to compute a new state hash. See module specific documentation & implementation details inside each module respectively.

- [Context Management](#context-management)
- [Block Application](#block-application)

_NOTE: The diagrams below use some [Hotstuff specific](https://arxiv.org/abs/1803.05069) terminology as described in the [HotPOKT Consensus Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/consensus) but can be adapted to other BFT protocols as well._

<!-- See if there's an answer in this question to add links to notes: https://stackoverflow.com/questions/74103729/adding-hyperlinks-to-notes-in-mermaid-sequence-diagrams -->

## Context Management

The `Utility` and `Persistence` modules maintain a context (i.e. an ephemeral states) driven by the `Consensus` module that can be `released & reverted` (i.e. the block is invalid / no Validator Consensus reached) or can be `committed & persisted` to disk (i.e. the block is finalized).

On every round of every height:

1. The `Consensus` module handles a `NEWROUND` message
2. A new `UtilityContext` is initialized at the current height
3. A new `PersistenceRWContext` is initialized at the current height
4. The [Block Application](#block-application) flow commences

```mermaid
sequenceDiagram
    title Steps 1-4
    participant B as Bus
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

    %% Handle New Message
    B-->>C: HandleMessage(NEWROUND)

    %% NEWROUND

    activate C
    %% Create Contexts
    C->>+U: NewContext(height)
    U->>+P: NewRWContext(height)
    P->>-U: PersistenceContext
    U->>U: store context<br>locally
    activate U
    deactivate U
    U->>-C: UtilityContext
    C->>C: store context<br>locally
    deactivate C

    %% Apply Block
    Note over C, P: 'Block Application'
```

---

_The **Proposer** drives the **Validators** to agreement via the **Consensus Lifecycle** (i.e. HotPOKT)_

---

5. The `Consensus` module handles the `DECIDE` message
6. The `commitQC` is propagated to the `UtilityContext` & `PersistenceContext` on `Commit`
7. The persistence module's internal implementation for ['Store Block'](../../persistence/docs/PROTOCOL_STORE_BLOCK.md) must execute.
8. Both the `UtilityContext` and `PersistenceContext` are released

```mermaid
sequenceDiagram
    title Steps 5-8
    participant B as Bus
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

    %% Handle New Message
    B-->>C: HandleMessage(DECIDE)

    activate C
    %% Commit Context
    C->>+U: Context.Commit(quorumCert)
    U->>+P: Context.Commit(quorumCert)
    P->>P: Internal Implementation
    Note over P: Store Block
    P->>-U: err_code
    U->>C: err_code
    deactivate U

    %% Release Context
    C->>+U: Context.Release()
    U->>+P: Context.Release()
    P->>-U: err_code
    U->>-C: err_code
    deactivate C
```

## Block Application

When applying the block during the `NEWROUND` message shown above, the majority of the flow is similar between the _leader_ and the _replica_ with one of the major differences being a call to the `Utility` module as seen below.

- `ApplyBlock` - Uses the existing set of transactions to validate & propose
- `CreateAndApplyProposalBlock` - Reaps the mempool for a new set of transaction to validate and propose

```mermaid
graph TD
    B[Should I prepare a new block?] --> |Wait for 2/3+ NEWROUND messages| C

    C[Am I the leader?] --> |Yes| D
    C[Am I the leader?] --> |No| Z

    D[Did I get any prepareQCs?] --> |Find highest valid prepareQC| E
    D[Did I get any prepareQCs?] --> |No| Z

    E[Am I ahead of highPrepareQC?] --> |Yes| G
    E[Am I ahead of highPrepareQC?] --> |No| Z

    G[Do I have a lockedQC] --> |No| H
    G[Do I have a lockedQC] --> |Yes| I

    I[Is highPrepareQC.view > lockedQC.view] --> |"No<br>(lockedQC.block)"| Z
    I[Is highPrepareQC.view > lockedQC.view] --> |"Yes<br>(highPrepareQC.block)"| Z

    H[CreateAndApplyProposalBlock]
    Z[ApplyBlock]
```

As either the _leader_ or _replica_, the following steps are followed to apply the proposal transactions in the block.

1.  Retrieve the `PersistenceContext` from the `UtilityContext`
2.  Update the `PersistenceContext` with the proposed block
3.  Call either `ApplyBlock` or `CreateAndApplyProposalBlock` based on the flow above

```mermaid
sequenceDiagram
    title Steps 1-3
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

        %% Retrieve the persistence context
        C->>+U: GetPersistenceContext()
        U->>-C: PersistenceContext

        %% Update the proposal in the persistence context
        C->>+P: SetProposalBlock
        P->>-C: err_code

        %% Apply the block to the local proposal state
        C->>+U: ApplyBlock / CreateAndApplyProposalBlock
        U->>-C: err_code
```

4. Loop over all transactions proposed
5. Check if the transaction has already been applied to the local state
6. Perform the CRUD operation(s) corresponding to each transaction
7. The persistence module's internal implementation for ['Compute State Hash'](../../persistence/docs/PROTOCOL_STATE_HASH.md) must be triggered
8. Validate that the local state hash computed is the same as that proposed

```mermaid
sequenceDiagram
    title Steps 4-8
    participant C as Consensus
    participant U as Utility
    participant P as Persistence

    loop for each tx in txs
        U->>+P: TransactionExists(txHash)
        P->>-U: false (does not exist)
        loop for each operation in tx
            U->>+P: Get*/Set*/Update*/Insert*
            P->>-U: err_code
            U->>U: Validation logic
            activate U
            deactivate U
        end
    end
    %% TODO: Consolidate AppHash and StateHash
    U->>+P: ComputeStateHash()
    P->>P: Internal Implementation
    Note over P: Compute State Hash
    P->>-U: stateHash
    U->>C: stateHash

    %% Validate the computed hash
    C->>C: Compare local hash<br>against proposed hash
```
