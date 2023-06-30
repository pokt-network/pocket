# ICS-24 Host Requirements <!-- omit in toc -->

- [Overview](#overview)
- [Implementation](#implementation)
  - [Paths and Identifiers](#paths-and-identifiers)
  - [Timestamps](#timestamps)

## Overview

[ICS-24][ics24] details the requirements of the host chain, in order for it to be compatible with IBC. A host is defined as a node on a chain that runs the IBC software. A host has the ability to create connections with counterparty chains, open channels, and ports as well as commit proofs to the consensus state of its own chain for the relayer to submit to another chain. The host is responsible to managing and creating clients and all other aspects of the IBC module.

As token transfers as defined in [ICS-20][ics20] work on a lock and mint pattern, any tokens sent from **chain A** to **chain B** will have a denomination unique to the connection/channel/port combination that the packet was sent over. This means that if a host where to shutdown a connection or channel without warning any tokens yet to be returned to the host chain would be lost. For this reason, only validator nodes are able to become hosts, as they provide the most reliability out of the different node types.

## Implementation

**Note**: The ICS-24 implementation is still a work in progress and is not yet fully implemented.

ICS-24 has numerous sub components that must be implemented in order for the host to be fully functional. These range from type definitions for identifiers, paths and stores as well as the methods to interact with them. Alongside these ICS-24 also defines the Event Logging system which is used to store the packet data and timeouts for the relayers to read, as only the `CommitmentProof` objects are committed to the chain state. In addition to these numerous other features are part of ICS-24 that are closely linked to other ICS components such as consensus state introspection and client state validation.

### Paths and Identifiers

Paths are defined as bytestrings that are used to access the elements in the different stores. They are built with the function `ApplyPrefix()` which takes a store key as a prefix and a path string and will return the key to access an element in the specific store. The logic for paths can be found in [host/keys.go](../host/keys.go) and [host/prefix.go](../host/prefix.go)

Identifiers are bytestrings constrained to specific characters and lengths depending on their usages. They are used to identify: channels, clients, connections and ports. Although the minimum length of the identifiers is much less we use a minimum length of 32 bytes and a maximum length that varies depending on the use case to randomly generate identifiers. This allows for an extremely low chance of collision between identifiers. Identifiers have no significance beyond their use to store different elements in the IBC stores and as such there is no need for non-random identifiers. The logic for identifiers can be found in [host/identifiers.go](../host/identifiers.go).

### Timestamps

The `GetTimestamp()` function returns the current unix timestamp of the host machine and is used to calculate timeout periods for packets

[ics24]: https://github.com/cosmos/ibc/blob/main/spec/core/ics-024-host-requirements/README.md
[ics20]: https://github.com/cosmos/ibc/blob/main/spec/app/ics-020-fungible-token-transfer/README.md
[smt]: https://github.com/pokt-network/smt
