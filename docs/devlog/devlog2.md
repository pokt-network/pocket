### Pocket V1 DevLog Call #1 Notes <!-- omit in toc -->

##### Date and Time: Tuesday February 7, 2023 18:00 UTC

##### Location: [Discord](https://discord.gg/pokt)

##### Duration: 60 minutes

##### [Recording](https://drive.google.com/drive/u/1/folders/1Ts6FHy3fcPjqjKl8grpd93L7DB1-N-LA)

##### [Feedback and Discussion Form](https://app.sli.do/event/eF13JYg93rGq4pGLRnHLF5)

---

### Agenda <!-- omit in toc -->

1. [Current Iteration](#current-iteration)
2. [Demo](#demo)
3. [Upcoming Iteration](#upcoming-iteration)
4. [Feedback and Open Discussion](#feedback-and-open-discussion)

---

### Current Iteration 🗓️

- Duration: January 11 - 24
- [Backlog](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+9%22)

#### Iteration Goals 🎯

- M1: PoS
  - Cont. developing a simple peer discovery mechanism for localnet
  - Begin full peer discovery development with LibP2P integration
  - Finalize the server to advertise block and identify state entrypoints
- M2: DoS
  - Merge K8 Localnet and Applied Logging Module
- M3: RoS
  - Address utility foundation issues to unblock M3 implementation

#### Iteration Results ✅

- Completed
  - https://github.com/pokt-network/pocket/pull/354
  - https://github.com/pokt-network/pocket/pull/437
  - https://github.com/pokt-network/pocket-network-protocol/pull/28
  - https://github.com/pokt-network/pocket/pull/464
  - https://github.com/pokt-network/pocket-network-protocol/pull/15
  - https://github.com/pokt-network/pocket/pull/468
  - https://github.com/pokt-network/pocket/pull/481
  - https://github.com/pokt-network/pocket/pull/486
- In Review
  - https://github.com/pokt-network/pocket/issues/409
- In Progress
  - https://github.com/pokt-network/pocket/issues/351
  - https://github.com/pokt-network/pocket/issues/347
  - https://github.com/pokt-network/pocket/issues/325
  - https://github.com/pokt-network/pocket/issues/416
  - https://github.com/pokt-network/pocket/issues/473
  - https://github.com/pokt-network/pocket/issues/475

#### External Contributions ⭐

- https://github.com/pokt-network/pocket/pull/472
- https://github.com/pokt-network/pocket/pull/419
- https://github.com/pokt-network/pocket/pull/453
- https://github.com/pokt-network/pocket/pull/459
- https://github.com/pokt-network/pocket/pull/465
- https://github.com/pokt-network/pocket/pull/474
- https://github.com/pokt-network/pocket/pull/483
- https://github.com/pokt-network/pocket/pull/487

---

### Demo 💻

- Deploy a V1 Localnet using K8s
- Present infrastructure that can be used to deploy remote clusters
- Demonstrate tooling that can be used to scale a cluster
- Demo a mature logging framework integrated with Grafana

Try it out ➡️ [Demo Guide](https://github.com/pokt-network/pocket/blob/main/docs/demos/iteration_9_localnet_infra.md)

---

### Upcoming Iteration 🗓️

- Duration: January 25 - February 7
- [Backlog Candidates](https://github.com/orgs/pokt-network/projects/142/views/12?layout=table&filterQuery=iteration%3A%22Iteration+10%22)

---

### Feedback and Open Discussion 💡

[Feedback and Discussion Form](https://app.sli.do/event/2LFSdaBzJ4FPYANPFcGxC7/live/questions)

##### Q: Can Localnet scale to 999 validators?
A: Containers will be created but the validators will not be participating in consensus until peer discovery is in place and those validators are staked.

##### Q: Where there other projects that inspired how to design the testing and deployment suite?
A: A lot of learning from V0 and a commitment to visibility that won't only allow us to iterate faster but all contributors (reduce dependencies on the core team to make changes to the protocol). We looked at a lot of different projects when we started but most projects don't have these tools in place. In general the Tendermint and Cosmos ecosystem is going that direction, but it's more difficult because they are dealing with legacy code rather than the greenfield of V1. One of the best projects heading this direction is Aptos. They have very mature infrastructure and test suite. 

##### Q: Other projects, like Cosmos, allow people to configure different backends for key storage. Is it worth surveying the community to see what backends the keybase should integrate with (e.g. Leveldb or Pass)?
A: The Keybase currently users badger.db to store locally in the file system. When it comes to key management, enabling anyone to integrate their own key manager into the codebase and identifying maybe one major integration. Please keep in mind that this is not useful for localnet but will be useful once we have a remote deployment.

##### Q: In V0, the Private Key is a plain text in the server but not in the Keybase. In V1, is this all inside of the Keybase or will there be plain text PKs?
A: The Keybase contains all the PK pairs associated with wallets only. Storing in the same way as V0 (base64) to allow for complete interoperability. 

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
