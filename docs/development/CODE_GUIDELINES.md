# Pocket's Code Development Guidelines <!-- omit in toc -->

_This document is a living document and will be updated as the team learns and grows. It is a supplement to the [code guidelines](./CODE_GUIDELINES.md)_

## Table of Contents <!-- omit in toc -->

- [Comments for Interface Implementation](#comments-for-interface-implementation)
- [Exposing functions for testing purposes](#exposing-functions-for-testing-purposes)

## Comments for Interface Implementation

If there is a `PeerstoreProvider` interface with a `GetStakedPeerstoreAtHeight` function, the interface should contain a comment with a functional explanation:

```go
type PeerstoreProvider interface {

  // GetStakedPeerstoreAtHeight returns a peerstore containing all staked peers
  // at a given height. These peers communicate via the p2p module's staked actor
  // router.
  GetStakedPeerstoreAtHeight(height uint64) (typesP2P.Peerstore, error)
```

And the implementation shold reference it:

```go
// GetStakedPeerstoreAtHeight implements the respective `PeerstoreProvider` interface method.
func (persistencePSP *persistencePeerstoreProvider) GetUnstakedPeerstore() (typesP2P.Peerstore, error) {
  // ...
}
```

## Exposing functions for testing purposes

If possible, move the function into a separate file that has `//go:build test` at the top like so:

![go:build test](https://github.com/pokt-network/pocket/assets/1892194/e7f921c7-6830-4aa6-afca-aef9c6cabbc6)https://github.com/pokt-network/pocket/assets/1892194/e7f921c7-6830-4aa6-afca-aef9c6cabbc6

For further reading, please see [Testutils](https://www.notion.so/pocketnetwork/Testutils-9cba9010e18447248e9daa8a3b87e3f2)
