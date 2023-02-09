# Building & Versioning

- [Building \& Versioning](#building--versioning)
  - [Release Tag Versioning](#release-tag-versioning)
  - [\[WIP\] Build Versioning](#wip-build-versioning)
  - [Container Images](#container-images)
    - [Tags](#tags)
    - [Extended images with additional tooling](#extended-images-with-additional-tooling)
  - [\[TODO\] Magefile build system](#todo-magefile-build-system)

## Release Tag Versioning

We follow Go's [Module Version Numbering](https://go.dev/doc/modules/version-numbers) for software releases along with typical [Software release life cycles](https://en.wikipedia.org/wiki/Software_release_life_cycle).

For example, `v0.0.1-alpha.pre.1` is the tag used for the first milestone merge and `v0.0.1-alpha.1` can be used for the first official alpha release.

## [WIP] Build Versioning

Automatic development / test / production builds are still a work in progress, but we plan to incorporate the following when we reach that point:

- `+dirty` for uncommitted changes
- `-version` flag that can be injected or defaults to `UNKNOWN`
- `branch_name` and a shortened `commit_hash` should be included

For example, `X.Y.Z[-<pre_release_tag][+branch_name][+short_hash][+dirty]` is the name of a potential build we will release in the future.

## Container Images

Our images are hosted on Github's Container Registry (GHCR) and are available at `ghcr.io/poktnetwork/pocket-v1`. You can find the list of latest images [here](https://github.com/pokt-network/pocket/pkgs/container/pocket-v1).

### Tags

Code built from the default branch (i.e. `main`) is tagged as `latest`.

Code built from commits in Pull Requests, is tagged as `pr-<number>`, as well as `sha-<7 digit sha>`.

Once releases are managed, they will be tagged with the version number (e.g. `v0.0.1-alpha.pre.1`).

### Extended images with additional tooling

Extended images with additional tooling are built to aid in troubleshoot and debugging. The extended image is formatted as `<tag>-dev`. For example, `latest-dev`, or `pr-123-dev`.

## [TODO] Magefile build system

Once the V1 implementation reaches the stage of testable binaries, we are looking to use [Mage](https://magefile.org/) which is being tracked in [pocket/issues/43](https://github.com/pokt-network/pocket/issues/43) that'll inject a version with the `-version` flag.

<!-- GITHUB_WIKI: guides/releases/readme -->
