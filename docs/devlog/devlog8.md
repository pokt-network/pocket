# Pocket V1 DevLog #8 <!-- omit in toc -->

**Date Published**: June 5th, 2023

## Table of Contents <!-- omit in toc -->

- [Iteration 17 Goals \& Results](#iteration-17-goals--results)
  - [M1: PoS](#m1-pos)
  - [M2: DoS](#m2-dos)
  - [M3: RoS](#m3-ros)
  - [M\*: North Start](#m-north-start)
- [Demo 💻](#demo-)
- [Contribute to V1 🧑‍💻](#contribute-to-v1-)
  - [Links \& References](#links--references)
- [ScreenShots](#screenshots)
  - [Iteration 17 - Completed](#iteration-17---completed)
  - [Iteration 18 - Planned](#iteration-18---planned)

---

## Iteration 17 Goals & Results

**Iterate Dates**:

### M1: PoS

1. Consensus - Finish minimum viable state sync to sync state between full nodes

- **Score**: 3/10 ± 1
- **Notes**:
  - Worked by @gokutheengineer was picked up
  - Changes are being merged upstream and refactored to work asynchronously

2. P2P - Finish minimum viable gossip to facilitate peer discovery and messages propogation

- **Score**: 5/10 ± 2
- **Notes**:
  - A lot of progress was made by we are hitting some issues on the edges cases
  - Major improvements are being made to the debugging utilities to facilitate investigation

3. Persistence - Finish the atomic store refactor to facilitate rollbacks

- **Score**: 6/10 ± 1
- **Notes**:
  - The largest of the 3 refactor PRs is almost ready for review
  - Local components are separately implemented, but tests are failing and code needs to be cleaned up

### M2: DoS

4. Provide the backend and infra team visibility into DevNet (documentation, dashboarding, tooling, etc...)

- **Score**: 5/10 ± 1
- **Notes**:
  - DevNet Workshop almost complete
  - Work started on a new tool to help explore the V1 state

### M3: RoS

5. E2E Trustless Relay - Kickoff / POC of E2E trustless relay

- **Score**: 7/10 ± 1
- **Notes**:
  - We kicked off the start of the implementation of E2E trustless relay
  - Introduced a new member to the team: Welcome @adshmh!

6. MVT (Minimum Viable TestNet) Feature List

- **Score**: 8/10 ± 1
- **Notes**:
  - We documented the [list of Utility Features](https://github.com/pokt-network/pocket/blob/main/utility/doc/E2E_FEATURE_LIST.md) we plan to have in TestNet & MainNet
  - The approach we will follow to implementing it can be found [here](https://github.com/pokt-network/pocket/blob/main/utility/doc/E2E_FEATURE_PATH_TEMPLATE.md)
  - Bonus: We published [Relay Mining](https://arxiv.org/abs/2305.10672) with the help of @RawthiL from PoktScan on how it will be implemented

### M\*: North Start

7. **Bonus**: IBC & SMT!

- **Notes**:
  - With @h5law FT for the summer, we kicked of IBC implementation
  - We started picking up work on our [Sparse Merkle Tree](https://github.com/pokt-network/smt) implementation

## Demo 💻

There was no demo in this explicit DevLog but check out [this teaser](https://twitter.com/olshansky/status/1661886785662914561) on Twitter from @Olshansk.

![Teaser](https://github.com/pokt-network/pocket/assets/1892194/a2fab136-9337-4926-adf8-c3299c61c1f2)

## Contribute to V1 🧑‍💻

### Links & References

- [V1 Specifications](https://github.com/pokt-network/pocket-network-protocol)
- [V1 Repo](https://github.com/pokt-network/pocket)
- [V1 Wiki](https://github.com/pokt-network/pocket/wiki)
  [V1 Project Dashboard](https://github.com/pokt-network/pocket/projects?query=is%3Aopen)

## ScreenShots

### Iteration 17 - Completed

![Iteration17](https://github.com/pokt-network/pocket/assets/1892194/44763167-7165-4e6e-be9f-456c4103d089)

### Iteration 18 - Planned

![Iteration 18](https://github.com/pokt-network/pocket/assets/1892194/5778c180-e1f9-4e37-9ce3-48d006a290eb)
