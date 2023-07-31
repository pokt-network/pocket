<div align="center">
  <a href="https://www.pokt.network">
    <img src="https://user-images.githubusercontent.com/2219004/151564884-212c0e40-3bfa-412e-a341-edb54b5f1498.jpeg" alt="Pocket Network logo" width="340"/>
  </a>
</div>

# Pocket <!-- omit in toc -->

The official implementation of the [V1 Pocket Network Protocol Specification](https://github.com/pokt-network/pocket-network-protocol).

\*_Please note that V1 protocol is currently under development and see [pocket-core](https://github.com/pokt-network/pocket-core) for the version that is currently live on mainnet._\*

- [Implementation](#implementation)
- [Overview](#overview)
- [Getting Started](#getting-started)
  - [Guides](#guides)
  - [Architectures](#architectures)
  - [Changelogs](#changelogs)
  - [Project Management Resources](#project-management-resources)
- [Support \& Contact](#support--contact)
  - [GPokT](#gpokt)
- [License](#license)

## Implementation

Official Golang implementation of the Pocket Network v1 Protocol.

<div>
  <a href="https://godoc.org/github.com/pokt-network/pocket"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"/></a>
  <a href="https://goreportcard.com/report/github.com/pokt-network/pocket"><img src="https://goreportcard.com/badge/github.com/pokt-network/pocket"/></a>
  <a href="https://golang.org"><img  src="https://img.shields.io/badge/golang-v1.20-green.svg"/></a>
  <a href="https://github.com/tools/godep" ><img src="https://img.shields.io/badge/godep-dependency-71a3d9.svg"/></a>
</div>

## Overview

<div>
    <a href="https://discord.gg/pokt"><img src="https://img.shields.io/discord/553741558869131266"></a>
    <a  href="https://github.com/pokt-network/pocket/releases"><img src="https://img.shields.io/github/release-pre/pokt-network/pocket.svg"/></a>
    <!-- <a href="https://circleci.com/gh/pokt-network/pocket"><img src="https://circleci.com/gh/pokt-network/pocket.svg?style=svg"/></a> -->
    <a  href="https://github.com/pokt-network/pocket/pulse"><img src="https://img.shields.io/github/contributors/pokt-network/pocket.svg"/></a>
    <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg"/></a>
    <a href="https://github.com/pokt-network/pocket/pulse"><img src="https://img.shields.io/github/last-commit/pokt-network/pocket.svg"/></a>
    <a href="https://github.com/pokt-network/pocket/pulls"><img src="https://img.shields.io/github/issues-pr/pokt-network/pocket.svg"/></a>
    <a href="https://github.com/pokt-network/pocket/releases"><img src="https://img.shields.io/badge/platform-linux%20%7C%20macos-pink.svg"/></a>
    <a href="https://github.com/pokt-network/pocket/issues"><img src="https://img.shields.io/github/issues/pokt-network/pocket.svg"/></a>
    <a href="https://github.com/pokt-network/pocket/issues"><img src="https://img.shields.io/github/issues-closed/pokt-network/pocket.svg"/></a>
</div>

## Getting Started

---

Some relevant links are listed below. Refer to the complete ongoing documentation at the **[Pocket GitHub Wiki](https://github.com/pokt-network/pocket/wiki)**.

If you'd like to contribute to the Pocket V1 Protocol, start by:

1. Get up and running by reading the [Development Guide](docs/development/README.md)
2. Find a task by reading the [Contribution Guide](docs/contributing/README.md)
3. Dive into any of the other guides or modules depending on where your interests lie

<!--
  The list of documents below was created by manually curating the output of the following command:
    find .. -name "*.md" | grep -v -e "vendor" -e "prototype" -e "SUMMARY.md" -e "TASTE.md"
-->

### Guides

- [Development Guide](docs/development/README.md)
- [End-to-end testing Guide](e2e/README.md)
- [Learning Guide](docs/learning/README.md)
- [Contribution Guide](docs/contributing/README.md)
- [Release Guide](docs/build/README.md)
- [Dependencies Guide](docs/deps/README.md)
- [Telemetry Guide](telemetry/README.md)

### Architectures

- [Shared Architecture](shared/README.md)
- [Utility Architecture](utility/doc/README.md)
- [Consensus Architecture](consensus/README.md)
- [Persistence Architecture](persistence/docs/README.md)
- [P2P Architecture](p2p/README.md)
- [APP Architecture](app/client/doc/README.md)
- [RPC Architecture](rpc/doc/README.md)
- [Node binary Architecture](app/pocket/doc/README.md)

### Changelogs

- [APP Changelog](app/client/doc/CHANGELOG.md)
- [Consensus Changelog](consensus/doc/CHANGELOG.md)
- [E2E Changelog](e2e/docs/CHANGELOG.md)
- [Node binary Changelog](app/pocket/doc/CHANGELOG.md)
- [P2P Changelog](p2p/CHANGELOG.md)
- [Persistence Changelog](persistence/docs/CHANGELOG.md)
- [RPC Changelog](rpc/doc/CHANGELOG.md)
- [Shared Changelog](shared/CHANGELOG.md)
- [Telemetry Changelog](telemetry/CHANGELOG.md)
- [Utility Changelog](utility/doc/CHANGELOG.md)

### Project Management Resources

- [V1 Roadmap](https://github.com/pokt-network/pocket/blob/main/docs/roadmap/README.md)
- [V1 Project Board](https://github.com/orgs/pokt-network/projects/142/views/12)

## Support & Contact

<div>
  <a href="https://twitter.com/poktnetwork"><img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"></a>
  <a href="https://t.me/POKTnetwork"><img src="https://img.shields.io/badge/Telegram-blue.svg"></a>
  <a href="https://research.pokt.network"><img src="https://img.shields.io/discourse/https/research.pokt.network/posts.svg"></a>
</div>

### GPokT

You can also use our chatbot, [GPokT](https://gpoktn.streamlit.app), to ask questions about Pocket Network. As of updating this documentation, please note that it may require you to provide your own LLM API token. If the deployed version of GPokT is down, you can deploy your own version by following the instructions [here](https://github.com/pokt-network/gpokt).

---

## License

This project is licensed under the MIT License; see the [LICENSE](LICENSE) file for details.

<!-- GITHUB_WIKI: home/readme -->
