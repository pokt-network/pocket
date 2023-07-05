# Pocket V1 DevLog #10 <!-- omit in toc -->

**Date Published**: July 5th, 2023

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

- 🟢 **TestNet Rehearsal**

  - 100% completeness of TestNet Rehearsal
  - **Grade**: 10 / 10
    - Completed

- 🟢 **MainNet Rehearsal**
  - Have 5% of MainNet run the latest beta (w/o protocol upgrade)
  - **Grade**: 10 / 10
    - Completed and found some bugs along the way too when synching from scratch

### M1: PoS

- 🔴 Consensus

  - Attempt #1: Remove State Sync dependency on FSM
  - Attempt #3: finish minimum viable state sync
  - **Grade**: 1 / 10
    - Very little time left to work on this

- 🟢 Persistence

  - MVP of the full commit & rollback DEMO
  - **Grade**: 7 / 10
    - The test which was going to be the demo has fought me more than expected but good progress has been made, there’s a design document ready, and the test harness is there, the mocks and the submodule interactions have been the problem.

- 🟢 P2P
  - Attempt #N: Finishing off and merging in everything related to gossip and background
  - **Grade**: 8.5 / 10

### M2: DoS

- 🔴 **Primary focus: observability**
  - Open question: need to identify issues w/ metric access
  - Streamlining logging: Make structured logging system easily available to new devs w/ documentation part of LocalNet instructions
  - Attach smaller tickets in a separate repo to V2
  - **Grade**: 0 / 10
    - Other infrastructure related maintenance issues took away time from being able to focus on observability

### M3: RoS

- 🟢 **Trustless Relay**

  - Session caching on the client
  - Finish all the PRs in flight (review, merge in)
  - Provide an E2E test that works, blocks CI if it breaks, documented and visible (DEMO)
  - **Grade**: 8 / 10

- 🟡 **Feature Flags**
  - Scope out the work necessary and create an E2E Feature Path github ticket using the template we created
  - **Grade**: 5 / 10
    - Research and design doc made good progress w/ support from bigBoss bus still a lot to do.

### M7: IBC

- 🟢 **SMST**

  - Get it reviewed & merged in
  - Clean up the documentation & merge it in
  - Visualizers: create a visualizer for the tree
  - Present: Finish off the SMT presentation
  - Stretch goal: potentially start storing trustless relays in it
  - **Grade**: 8.5 / 10
    - SMST merged (wrapper around SMT option)
    - Visualiser is accurate but not pretty could do with some more work
    - Presentation went well but definietly could improve on some packed slides
    - Need to work closer with @Arash Deshmeh to get it in prod with M3

- 🟢 **ICS23**

  - Put up the github ticket and PR for reivew to merge in the proof mechanisms
  - Up to cosmos on ETA to review/merge
  - **Grade**: 9 / 10
    - ICS23 merged in our repo using my fork of `cosmos/ics23` as a dependency
    - My explanations on why the exclusion proof is needed can improve as others find it hard to understand
    - Cosmos PR is ready to merge pending review (probably take a while)

- 🟡 **ICS24**

  - Put up event logging for review; stretch goal is to merge
  - **Grade**: 6 / 10
    - ICS-24 stores have made great progress
    - Event logging unfortunately didnt make this fortnight
    - Message onto/off of bus as transactions works well 👍🏻

- 🟡 **Light client spike**
  - Start knowing where to head with research
  - **Grade**: 5 / 10
    - ICS-02 specced out well
    - ICS-08 needs more work into its design
      - Need to learn more about CosmWasm and WasmVM
    - WIP document needs to be converted to ticket epic
      - https://hackmd.io/0WVMarGpSIGqEyzvnygWpw

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

_tl;dr Aim to demo as much of the work from the previous iteration in action_

![Iteration20](![Screenshot 2023-07-05 at 1 35 07 PM](https://github.com/pokt-network/pocket-core/assets/1892194/8ae047ee-f186-4e1a-8ced-14764ec83886))

<!-- GITHUB_WIKI: devlog/2023_07_05 -->
