# Pocket V1 DevLog Call #4 Notes <!-- omit in toc -->

- **Date and Time**: <!-- UPDATE_ME -->
- **Location**: [Discord](https://discord.gg/pokt)
- **Duration**: 45 minutes
- [Recording](https://drive.google.com/drive/u/1/folders/1Ts6FHy3fcPjqjKl8grpd93L7DB1-N-LA)
- [Feedback and Discussion Form](https://app.sli.do/event/mPbjhr5FiFo3c7x4QafJR9)

---

## Agenda <!-- omit in toc -->

- [Current Iteration 🗓️](#current-iteration-️)
- [Iteration Goals 🎯](#iteration-goals-)
- [Iteration Results ✅](#iteration-results-)
- [External Contributions ⭐](#external-contributions-)
- [Upcoming Iteration 🗓️](#upcoming-iteration-️)
- [Feedback and Open Discussion 💡](#feedback-and-open-discussion-)
  - [Q: Are these keys also hierarchical in the original sense of bit-44 such as that any given key you can derive valid keys deterministically but not going back up? Can you also re-derive deeper ancestor keys from midpoint between the root and far off ancestor?](#q-are-these-keys-also-hierarchical-in-the-original-sense-of-bit-44-such-as-that-any-given-key-you-can-derive-valid-keys-deterministically-but-not-going-back-up-can-you-also-re-derive-deeper-ancestor-keys-from-midpoint-between-the-root-and-far-off-ancestor)
  - [Contribute to V1 🧑‍💻](#contribute-to-v1-)
- [About Pocket Network 💙](#about-pocket-network-)

---

## Current Iteration 🗓️

- Duration: February 22 to March 7 2023
- [Backlog](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+11%22)

## Iteration Goals 🎯

- M1: PoS
  - Unblock the next steps of peer discovery by fully merging the initial LibP2P integration
  - Cont. state sync MVP
- M2: DoS
  - Cont. deploy localnet in a remote environment
- M3: RoS
  - Cont. Utility module foundations to unblock interface design

## Iteration Results ✅

- Completed
  - https://github.com/pokt-network/pocket/issues/504
  - https://github.com/pokt-network/pocket/issues/347 (please see linked PRs for more details!)
- In Review
  - https://github.com/pokt-network/pocket/issues/351
  - https://github.com/pokt-network/pocket/pull/558
- In Progress
  - https://github.com/pokt-network/pocket/issues/307
  - https://github.com/pokt-network/pocket/issues/352
  - https://github.com/pokt-network/pocket/issues/473
  - https://github.com/pokt-network/pocket/issues/508 (Save Points foundational work)
  -

## External Contributions ⭐

- Data provided from [POKTScan](https://poktscan.com) https://github.com/pokt-network/pocket-core/issues/1523
- https://github.com/pokt-network/pocket/pull/526
- https://github.com/pokt-network/pocket/pull/529
- https://github.com/pokt-network/pocket/issues/489
- https://github.com/pokt-network/pocket/pull/510

---

## Upcoming Iteration 🗓️

- Duration: March 8 to March 21 2023
- [Backlog Candidates](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+12%22)

---

## Feedback and Open Discussion 💡

[Feedback and Discussion Form](https://app.sli.do/event/mPbjhr5FiFo3c7x4QafJR9)

### Q: Are these keys also hierarchical in the original sense of bit-44 such as that any given key you can derive valid keys deterministically but not going back up? Can you also re-derive deeper ancestor keys from midpoint between the root and far off ancestor?

A: The keys are generated from the path so for any given master key it will generate the key up to the index. Given the index child key, there is no way to go backward to the parent key as far as generation is concerned (following bit-44 hierarchy). Currently each time you generate a key from it's from a hardcoded path, so if we wanted to generate up to a certain index and hash from that point to shorten the amount of time that is definitely possibly by enabling new user paths.

---

### Contribute to V1 🧑‍💻

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
