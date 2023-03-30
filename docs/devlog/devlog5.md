# Pocket V1 DevLog Call #5 Notes <!-- omit in toc -->

- **Date and Time**: Tuesday March 21, 2023 16:00 UTC
- **Location**: [Discord](https://discord.gg/pokt)
- **Duration**: 55 minutes
- [Recording](https://drive.google.com/file/d/1eSAj7oUXKQTpiE25Z4mN4xOTnVLAbYQO/view?usp=share_link)

---

## Agenda <!-- omit in toc -->

- [Current Iteration 🗓️](#current-iteration-️)
- [Iteration Goals 🎯](#iteration-goals-)
- [Iteration Results ✅](#iteration-results-)
- [External Contributions ⭐](#external-contributions-)
- [Upcoming Iteration 🗓️](#upcoming-iteration-️)
- [Feedback and Open Discussion 💡](#feedback-and-open-discussion-)
- [Contribute to V1 🧑‍💻](#contribute-to-v1-)
- [About Pocket Network 💙](#about-pocket-network-)

---

## Current Iteration 🗓️

- Duration: March 8 to 21st 2023
- [Backlog](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+12%22)

## Iteration Goals 🎯

- M1: PoS
  -  Finish the state sync MVP
  -  Build a foundation for rollbacks
  -  Get Raintree working with LibP2P
- M2: DoS
  - Cont. deploy localnet in a remote environment
  - Build Node resource dashboard foundations
- M3: RoS
  - Implement the Session Protocol to enable session generation in V1

## Iteration Results ✅

- Completed
  - https://github.com/pokt-network/pocket/pull/586
  - https://github.com/pokt-network/pocket/pull/587
  - https://github.com/pokt-network/pocket/pull/590
  - https://github.com/pokt-network/pocket/pull/591
  - https://github.com/pokt-network/pocket/pull/561
- In Review
  - https://github.com/pokt-network/pocket/pull/596
  - https://github.com/pokt-network/pocket/pull/577
  - https://github.com/pokt-network/pocket/pull/575
  - https://github.com/pokt-network/pocket/pull/589
  - https://github.com/pokt-network/pocket/issues/544
- In Progress
  - https://github.com/pokt-network/pocket/issues/530 🐛
  - https://github.com/pokt-network/pocket/issues/554
  - https://github.com/pokt-network/pocket/issues/508
  - https://github.com/pokt-network/pocket/issues/473
  - https://github.com/pokt-network/pocket/issues/352
  - https://github.com/pokt-network/pocket/issues/307
  - https://github.com/pokt-network/pocket/issues/197
  - https://github.com/pokt-network/pocket/issues/327

## External Contributions ⭐

- https://github.com/pokt-network/pocket/issues/525
- https://github.com/pokt-network/pocket/pull/573
- https://github.com/pokt-network/pocket/issues/194

---

## Upcoming Iteration 🗓️

- Duration: March 8 to March 21 2023
- [Backlog Candidates](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+13%22)

---

## Feedback and Open Discussion 💡

### Q: Does any other blockchain use PostGres as part of their state layer? I was excited to be able to query the state of the blockchain using SQL. 

A: Last we checked the answer was no :) however, if you look at the Cosmos SDK they spent a lot of time looking into it, but it was too difficult to integrate into their legacy infrastructure, similar to Celestia's intention to build their own Merkle Trees but ultimately using IAVL. Newer systems like Aptos or Espresso Systems might have taken a similar approach to us, but we would have to double check. 

### Q: Once we have gone through testnet/localnet, etc. and are running nodes on mainnet, is the intention for them to be run on a Kubernetes cluster, or is this only for devnet purposes?

A: We have no expectations for the node runners to all utilize Kubernetes. Docker images are container images that can run on any container platform and we are going to supply binaries as well. We will also have documentation on how to spin up simple versions of Kubernetes that can run on one machine and Kubernetes operator to help configure and provision in an automated fashion.   

---

## Contribute to V1 🧑‍💻

V1 is an open source project that is open to external contributors. Find information about onboarding to the project, browse available bounties, or look for open issues in the linked resources below. For any questions about contributing, contact @jessicadaugherty

- [Configure Development Environment](https://github.com/pokt-network/pocket/blob/main/docs/development/README.md)
- [Available Developer Bounties](https://app.dework.xyz/pokt-network/v1-protocol)
- [V1 Project Board](https://github.com/orgs/pokt-network/projects/142/views/12)
- [V1 Roadmap](https://github.com/pokt-network/pocket/blob/main/docs/roadmap/README.md#m1-pocket-pos-proof-of-stake)

## About Pocket Network 💙

Pocket Network is a blockchain data platform, built for applications, that uses cost-efficient economics to coordinate and distribute data at scale.

- [Website](https://pokt.network)
- [Documentation](https://docs.pokt.network)
- [Discord](https://discord.gg/pokt)
- [Twitter](https://twitter.com/POKTnetwork)

<!-- GITHUB_WIKI: devlog/2023_03_09 -->
