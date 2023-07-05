# Pocket V1 DevLog #10 <!-- omit in toc -->

**Date Published**: July 3rd, 2023

We have kept the goals and details in this document short, but feel free to reach out to @Olshansk in the [core-dev-chat](https://discord.com/channels/553741558869131266/986789914379186226) for additional details, links & resources.

## Table of Contents <!-- omit in toc -->

- [Iteration 20 Goals \& Results](#iteration-20-goals--results)
  - [V0](#v0)
  - [M1: PoS](#m1-pos)
  - [M2: DoS](#m2-dos)
  - [M3: RoS](#m3-ros)
  - [M7: IBC](#m7-ibc)
- [Contribute to V1 🧑‍💻](#contribute-to-v1-)
  - [Links \& References](#links--references)
- [ScreenShots](#screenshots)
  - [Iteration 19 - Completed](#iteration-19---completed)
  - [Iteration 20 - Planned](#iteration-20---planned)

## Iteration 20 Goals & Results

**Iterate Dates**: June 15th - June 30th, 2023

### V0

- 🟡 **TestNet Rehearsal**

  - 100% completeness of TestNet Rehearsal
  - **Grade**: 10 / 10

- 🟡 **MainNet Rehearsal**
  - Have 5% of MainNet run the latest beta (w/o protocol upgrade)
  - **Grade**: 10 / 10

### M1: PoS

- 🟡 **Consensus - Attempt #3: finish minimum viable state sync**

  - **Grade**: 1 / 10

- 🟡 **Consensus - Attempt #1: Remove State Sync dependency on FSM**

  - **Grade**: 1 / 10

- 🟡 **Persistence - MVP of the full commit & rollback DEMO**

  - **Grade**: 7 / 10

- 🟡 **P2P - Attempt #N: Finishing off and merging in everything related to gossip and background**
  - **Grade**: / 10

### M2: DoS

- 🟡 **Primary focus: observability**
  - Open question: need to identify issues w/ metric access
  - Streamlining logging: Make structured logging system easily available to new devs w/ documentation part of LocalNet instructions
  - Attach smaller tickets in a separate repo to V2
  - **Grade**: / 10

### M3: RoS

- 🟡 **Trustless Relay**

  - Session caching on the client
  - Finish all the PRs in flight (review, merge in)
  - Provide an E2E test that works, blocks CI if it breaks, documented and visible (DEMO)
  - **Grade**: / 10

- 🟡 **Feature Flags**
  - Scope out the work necessary and create an E2E Feature Path github ticket using the template we created
  - **Grade**: / 10

### M7: IBC

- 🟡 **SMST**

  - Get it reviewed & merged in
  - Clean up the documentation & merge it in
  - Visualizers: create a visualizer for the tree
  - Present: Finish off the SMT presentation
  - Stretch goal: potentially start storing trustless relays in it
  - **Grade**: / 10

- 🟡 **ICS23**

  - Put up the github ticket and PR for review to merge in the proof mechanisms
  - Up to cosmos on ETA to review/merge
  - **Grade**: / 10

- 🟡 **ICS24**

  - Put up event logging for review; stretch goal is to merge
  - **Grade**: / 10

- 🟡 **Light client spike**
  - Start knowing where to head with research
  - **Grade**: / 10

## Contribute to V1 🧑‍💻

### Links & References

- [V1 Specifications](https://github.com/pokt-network/pocket-network-protocol)
- [V1 Repo](https://github.com/pokt-network/pocket)
- [V1 Wiki](https://github.com/pokt-network/pocket/wiki)
- [V1 Project Dashboard](https://github.com/pokt-network/pocket/projects?query=is%3Aopen)

## ScreenShots

Please note that everything that was not `Done` in iteration19 is moving over to iteration20.

### Iteration 19 - Completed

![Iteration19_1](https://github.com/pokt-network/pocket/assets/1892194/93f033e9-a408-49bf-9531-9f84cc1bc254)
![Iteration19_2](https://github.com/pokt-network/pocket/assets/1892194/2c600d90-fe4c-496b-a4e4-66ef0afb4771)

### Iteration 20 - Planned

TODO
![Iteration20]()

<!-- GITHUB_WIKI: devlog/2023_07_03 -->
