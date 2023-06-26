# Pocket V1 DevLog #9 <!-- omit in toc -->

**Date Published**: June 19th, 2023

We have kept the goals and details in this document short, but feel free to reach out to @Olshansk in the [core-dev-chat](https://discord.com/channels/553741558869131266/986789914379186226) for additional details, links & resources.

## Table of Contents <!-- omit in toc -->

- [Iteration 18 Goals \& Results](#iteration-18-goals--results)
  - [V0](#v0)
  - [M1: PoS](#m1-pos)
  - [M2: DoS](#m2-dos)
  - [M3: RoS](#m3-ros)
  - [M7: IBC](#m7-ibc)
- [Contribute to V1 🧑‍💻](#contribute-to-v1-)
  - [Links \& References](#links--references)
- [ScreenShots](#screenshots)
  - [Iteration 18 - Completed](#iteration-18---completed)
  - [Iteration 19 - Planned](#iteration-19---planned)

## Iteration 18 Goals & Results

**Iterate Dates**: May 31st - June 14th, 2023

### V0

- 🟡 **Finish TestNet Rehearsal**
  - Documentation is complete
  - Testing currently blocked by community
  - **Grade**: 5 / 10

### M1: PoS

- 🟡 **Consensus - Finish minimum viable state sync**

  - Draft close to being up for review
  - Identified new blocking tech debt
  - Difficult to test
  - **Grade**: 5 / 10

- 🟢 **P2P - Finish minimum viable gossip**

  - Test tooling & deterministic tests added
  - Provable that it works locally
  - Addressed & scoped out tech debt work for the future iteration
  - Balanced decision making given the circumstances of the codebase
  - **Grade**: 7 / 10

- 🟢 **Persistence - Finish persistence module refactor to support atomic operations**
  - Two PRs related to this work are in review → review is taking longer than anticipate
  - Tree store refactor almost ready to go
  - World state serialization & deserialization
  - Miscellaneous bugs blocking merge
  - **Grade**: 7.5 / 10
  - Excluded from grade: Savepoints & rollbacks
    - PR is being spiked out on top of the PRs above
    - Stacked diff on top of the work above

### M2: DoS

- 🟡 **DevNet workshop**

  - Guidelines/docs on how to use, access and gain visibility into DevNet complete
  - UI for DevNet merged in
  - Workshop required a bit more coordination and will be repeated in the future
  - **Grade**: 5 / 10; (2.5 for workshop and 7.5 for UI & dashboard)

- 🔴 **Metrics Foundation**
  - Identified a preliminary list of metrics we need in the near future across persistence, p2p, consensus, utility
  - Not worked on in this iteration due to v0 maintenance requirments
  - **Grade**: 0 / 10

### M3: RoS

- 🟢 **E2E Trustless Relay - Functioning E2E on LocalNet**
  - Addressing final review comments
  - Server side validation complete
  - Close to being merged
  - Not started: E2E tests
- **Grade**: 8.5 / 10

### M7: IBC

- ⭐ **First iteration of IBC module up for review**

  - Stores & commitment proofs that can be serialized
  - PR to ICS23 for exclusion proofs has been submitted; https://github.com/cosmos/ics23/issues/152
  - Uncovered new learnings about the proper direction for implementation in future iterations
  - **Grade**: 10 / 10

- ⭐ **Merkle sum tree** w/ variable node weights feature in SMT
  - Secondary PR (w/ a wrapper vs code duplication) up for review
  - Both PRs (wrappers & code duplication) passing tests
  - **Grade**: 9.5 / 10

## Contribute to V1 🧑‍💻

### Links & References

- [V1 Specifications](https://github.com/pokt-network/pocket-network-protocol)
- [V1 Repo](https://github.com/pokt-network/pocket)
- [V1 Wiki](https://github.com/pokt-network/pocket/wiki)
- [V1 Project Dashboard](https://github.com/pokt-network/pocket/projects?query=is%3Aopen)

## ScreenShots

### Iteration 18 - Completed

![Iteration18](https://github.com/pokt-network/smt/assets/1892194/86046baa-2b16-4dc4-bd53-993feb4f81c0)

### Iteration 19 - Planned

![Iteration 19](https://github.com/pokt-network/smt/assets/1892194/5254c518-9a43-4a16-bd8f-8f356eba0456)

<!-- GITHUB_WIKI: devlog/2023_06_19 -->
