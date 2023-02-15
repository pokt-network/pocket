# Pocket V1 DevLog Call #2 Notes <!-- omit in toc -->

- **Date and Time**: Tuesday January 24th, 2023 18:00 UTC
- **Location**: [Discord](https://discord.gg/pokt)
- **Duration**: 60 minutes
- [Recording](https://drive.google.com/drive/u/1/folders/1Ts6FHy3fcPjqjKl8grpd93L7DB1-N-LA)
- [Feedback and Discussion Form](https://app.sli.do/event/eF13JYg93rGq4pGLRnHLF5)

---

## Agenda <!-- omit in toc -->

- [Current Iteration 🗓️](#current-iteration-️)
- [Iteration Goals 🎯](#iteration-goals-)
- [Iteration Results ✅](#iteration-results-)
- [External Contributions ⭐](#external-contributions-)
- [Upcoming Iteration 🗓️](#upcoming-iteration-️)
- [Feedback and Open Discussion 💡](#feedback-and-open-discussion-)
  - [Q: Do we have a distributed tracing framework for collecting metrics?](#q-do-we-have-a-distributed-tracing-framework-for-collecting-metrics)
- [Contribute to V1 🧑‍💻](#contribute-to-v1-)
- [About Pocket Network 💙](#about-pocket-network-)

---

### Current Iteration 🗓️

- Duration: January 11 - 24
- [Backlog](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+8%22)

### Iteration Goals 🎯

- M1: PoS
  - P2P: Develop a simple Peer Discovery mechanism for LocalNet
  - State Sync: Build a server to advertise blocks
- M2: DoS
  - Deploy 4 Node LocalNet in a remote environment and document
  - Design MVP DevNet telemetry dashboard
- M3: RoS
  - Update Utility Spec to define permissionless application and gateway behavior

### Iteration Results ✅

- Completed
  - https://github.com/pokt-network/pocket/pull/450
  - https://github.com/pokt-network/pocket/issues/456
  - https://github.com/pokt-network/pocket/pull/451
  - https://github.com/pokt-network/pocket/pull/449
  - https://github.com/pokt-network/pocket/pull/448
  - https://github.com/pokt-network/pocket-network-protocol/pull/27
  - https://github.com/pokt-network/pocket-network-protocol/pull/26
  - https://github.com/pokt-network/pocket/pull/445
  - https://github.com/pokt-network/pocket/pull/444
  - https://github.com/pokt-network/pocket/pull/439
  - https://github.com/pokt-network/pocket/pull/427
  - https://github.com/pokt-network/pocket/issues/268
- In Review
  - https://github.com/pokt-network/pocket/issues/195
  - https://github.com/pokt-network/pocket-operator/issues/10
  - https://github.com/pokt-network/pocket/issues/388
  - https://github.com/pokt-network/pocket-network-protocol/pull/25
- In Progress
  - https://github.com/pokt-network/pocket/issues/409
  - https://github.com/pokt-network/pocket/issues/416
- Refactored
  - https://github.com/pokt-network/pocket-core/issues/1511
    - Requires deeper investigation into data returned from exporter
  - https://github.com/pokt-network/pocket/issues/307
    - Blocked by internal infrastructure migration

### External Contributions ⭐

- https://github.com/pokt-network/pocket/pull/446
- https://github.com/pokt-network/pocket/pull/442
- https://github.com/pokt-network/pocket/pull/407

---

### Upcoming Iteration 🗓️

- Duration: January 25 - February 7
- [Backlog Candidates](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+9%22)

---

### Feedback and Open Discussion 💡

[Feedback and Discussion Form](https://app.sli.do/event/eF13JYg93rGq4pGLRnHLF5)

#### Q: Do we have a distributed tracing framework for collecting metrics?

A: The logging module is being productionized so that it can be integrated with observability tooling like Prometheus. There currently is no distributed tracing of requests between protocol actors. Once the applied logging module has been merged (https://github.com/pokt-network/pocket/pull/420), distributed tracing (https://github.com/pokt-network/pocket/issues/143) can begin development.

---

### Contribute to V1 🧑‍💻

V1 is an open source project that is open to external contributors. Find information about onboarding to the project, browse available bounties, or look for open issues in the linked resources below. For any questions about contributing, contact @jessicadaugherty

- [Configure Development Environment](https://github.com/pokt-network/pocket/blob/main/docs/development/README.md)
- [Available Developer Bounties](https://app.dework.xyz/pokt-network/v1-protocol)
- [V1 Project Board](https://github.com/orgs/pokt-network/projects/142/views/12)
- [V1 Roadmap](https://github.com/pokt-network/pocket/blob/main/docs/roadmap/README.md#m1-pocket-pos-proof-of-stake)

### About Pocket Network 💙

Pocket Network is a blockchain data platform, built for applications, that uses cost-efficient economics to coordinate and distribute data at scale.

- [Website](https://pokt.network)
- [Documentation](https://docs.pokt.network)
- [Discord](https://discord.gg/pokt)
- [Twitter](https://twitter.com/POKTnetwork)

<!-- GITHUB_WIKI: devlog/2023_01_24 -->
