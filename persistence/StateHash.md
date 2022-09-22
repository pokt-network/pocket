# This discussion is aimed at:

1. Defining how we should compute the state hash
2. Identify potential changes needed in the current codebase
3. Propose next steps and actionable on implementation

## Goals:

- Define how the state hash will be computed
- Propose the necessary changes in separate tasks
- Implement each of the necessary pieces

## Non-goals:

- Choice/decision of Merkle Tree Design & Implementation
- Selection of a key-value store engine

## Primitives / non-negotiables:

- We will be using Merkle Trees (and Merkle Proofs) for this design (i.e. not vector commitments)
- We will be using a SQL engine for this (i.e. specifically PostgresSQL)
- We will be using Protobufs (not Flatbuffers, json, yaml or other) for the schema

## Necessary technical context:

### DB Engines

Insert table from here: [Merkle Tree Design & Implementation](https://tikv.org/deep-dive/key-value-engine/b-tree-vs-lsm/#summary)

- Most **Key-Value Store DB** Engines use **LSM-trees** -> good for writes
- Most **SQL DB** Engines use **B-Trees** -> good for reads

_Basically all but there can be exceptions_

### Addressable Merkle Trees

State is stored use an Account Based (non UTXO) based Modle

Insert image from: https://www.horizen.io/blockchain-academy/technology/expert/utxo-vs-account-model/#:~:text=The%20UTXO%20model%20is%20a,constructions%2C%20as%20well%20a%20sharding.

---

### Data Flow

## Basics:

1. Get each actor (flag, param, etc...) updated at a certain height (the context's height)
2. Compute the protobuf (the deterministic schema we use as source of truth)
3. Serialize the data struct
4. Update the corresponding merkle tree
5. Compute a state hash from the aggregated roots of all trees as per pokt-network/pocket-network-protocol@main/persistence#562-state-transition-sequence-diagram

## Q&A

Q: Can the SQL Engine be changed?
A: Yes

Q: Can the SQL Engine be removed altogether?
A: Yes, but hard

Q: Can the protobuf schema change?
A: Yes, but out-of-scope

Q: Can protobufs be replaced?
A: Maybe, but out-of-scope

---

Learnings / Ideas:

- Consolidate `UtilActorType` and `persistence.ActorType`
- `modules.Actors` interface vs `types.Actor` in persistenceGenesis
