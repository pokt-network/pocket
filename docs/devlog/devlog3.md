# Pocket V1 DevLog Call #3 Notes <!-- omit in toc -->

- **Date and Time**: Thursday February 21, 2023 18:00 UTC
- **Location**: [Discord](https://discord.gg/pokt)
- **Duration**: 60 minutes
- [Recording](https://drive.google.com/drive/u/1/folders/1Ts6FHy3fcPjqjKl8grpd93L7DB1-N-LA)
- [Feedback and Discussion Form](https://app.sli.do/event/aDvLMjXLy1fxG7UwMpC1sk)

---

## Agenda <!-- omit in toc -->

- [Current Iteration 🗓️](#current-iteration-️)
- [Iteration Goals 🎯](#iteration-goals-)
- [Iteration Results ✅](#iteration-results-)
- [External Contributions ⭐](#external-contributions-)
- [Upcoming Iteration 🗓️](#upcoming-iteration-️)
- [Feedback and Open Discussion 💡](#feedback-and-open-discussion-)
  - [Q: How is our approach to development different than how others are doing it? Are we staying ahead of the curve? How do we think about our work and how we prioritize it to make sure we are pioneering RPC?](#q-how-is-our-approach-to-development-different-than-how-others-are-doing-it-are-we-staying-ahead-of-the-curve-how-do-we-think-about-our-work-and-how-we-prioritize-it-to-make-sure-we-are-pioneering-rpc)
  - [Q: What is IBC and where is it integrated with the protocol?](#q-what-is-ibc-and-where-is-it-integrated-with-the-protocol)
  - [Contribute to V1 🧑‍💻](#contribute-to-v1-)
- [About Pocket Network 💙](#about-pocket-network-)

---

## Current Iteration 🗓️

- Duration: February 8 - 21st 2023
- [Backlog](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+10%22)

## Iteration Goals 🎯

- M1: PoS
  - Complete MVP and continue full peer discovery development with LibP2P integration
  - Finalize state entry points for state machine (syncing, synched, etc.)
- M2: DoS
  - Deploy localnet in a remote environment
- M3: RoS
  - Continue Utility module foundations

## Iteration Results ✅

- Completed
  - https://github.com/pokt-network/pocket/issues/416
  - https://github.com/pokt-network/pocket/issues/490
  - https://github.com/pokt-network/pocket/issues/429
  - https://github.com/pokt-network/pocket/issues/499
  - https://github.com/pokt-network/pocket/issues/516
  - https://github.com/pokt-network/pocket/issues/514
  - https://github.com/pokt-network/pocket/issues/475
- In Review
  - https://github.com/pokt-network/pocket/issues/347
- In Progress
  - https://github.com/pokt-network/pocket/issues/352
  - https://github.com/pokt-network/pocket/issues/351
  - https://github.com/pokt-network/pocket-core/issues/1523
  - https://github.com/pokt-network/pocket/issues/307
  - https://github.com/pokt-network/pocket/issues/493

## External Contributions ⭐

- https://github.com/pokt-network/pocket/issues/484

---

## Upcoming Iteration 🗓️

- Duration: February 22 - March 7 2023
- [Backlog Candidates](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+11%22)

---

## Feedback and Open Discussion 💡

[Feedback and Discussion Form](https://app.sli.do/event/aDvLMjXLy1fxG7UwMpC1sk)

### Q: How is our approach to development different than how others are doing it? Are we staying ahead of the curve? How do we think about our work and how we prioritize it to make sure we are pioneering RPC?

A: On the product side for us it's really important for us to always stay in sync with our sales and marketing team. What are users saying? What is the data saying? We try to stay on top of narratives and market trends. Although our roadmap is well defined currently, there are unanswered questions in the specification that we can incorporate more timely market insights into. Peer Discovery with LibP2P is a great example of this. Choosing the right algorithm to work with Raintree was not a trivial decision and happened over 6+ months of discovery across 7+ developers. Additionally, since the team is so close to RPC and has deep experience in the space, you see more proprietary inventions in V1 than you'll see in other projects. From a more technical standpoint, the goal with V1 is building an application specific blockchain for an application specific purpose. We have had many sleepless nights thinking about this decision. We see competitors using Tendermint or building modular blockchains. We have decided to build our own L1 to keep optionality, functionality, and build it in a very developer friendly way. Since encouraging others to build on Pocket Network is a core value, that guides a lot of how we approach building.

### Q: What is IBC and where is it integrated with the protocol?

A: Think of IBC as TCP (networking transport layer in web2) of blockchain to enable requests and responses between different blockchains with custom consensus mechanisms. By conforming a specification on how to transport requests enables several features that allow for cross-chain communication and security. IBC should not touch utility but should live in persistence (exposing state commitments) and consensus (light clients). However, if we get into more complex integrations we may start touching utility but not in the initial integration. TBD if IBC requires a dedicated module.

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

<!-- GITHUB_WIKI: devlog/2023_02_21 -->
