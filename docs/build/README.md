# Building & Versioning

## Release Tag Versioning

We follow Go's [Module Version Numbering](https://go.dev/doc/modules/version-numbers) for software releases along with typical [Software release life cycles](https://en.wikipedia.org/wiki/Software_release_life_cycle).

For example, `v0.0.1-alpha.pre.1` is the tag used for the first milestone merge and `v0.0.1-alpha.1` can be used for the first official alpha release.

## [WIP] Build Versioning

Automatic development / test / production builds are still a work in progress, but we plan to incorporate the following when we reach that point:

- `+dirty` for uncommited changes
- `-version` flag that can be injected or defaults to `UNKNOWN`
- `branch_name` and a shortened `commit_hash` shold be included

For example, `X.Y.Z[-<pre_release_tag][+branch_name][+short_hash][+dirty]` is the name of a potential build we will release in the future.

## [TODO] Magefile build system

Once the V1 implementation reaches the stage of testable binaries, we are looking to use [Mage](https://magefile.org/) which is being tracked in [pocket/issues/43](https://github.com/pokt-network/pocket/issues/43) that'll inject a version with the `-version` flag.
