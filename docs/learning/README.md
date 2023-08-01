# Learning Pocket <!-- omit in toc -->

**IMPORTANT NOTE TO THE READER**: _This document aims to be the `"single source that links out to everything"`, so it is unlikely you will ever have the time to read everything in here. It is intended to be a reference for you to come back to when you need to learn something new. PLEASE update it with additional resources you find helpful along the way. You can ask [GPokT](https://gpoktn.streamlit.app/) for help if you need it._

- [🏁 Where to Start?](#-where-to-start)
- [🤫 Secret Zone](#-secret-zone)
- [🏗️ Technical Foundation](#️-technical-foundation)
  - [Github Development](#github-development)
  - [Golang](#golang)
  - [Mermaid](#mermaid)
- [📚 Technical References](#-technical-references)
  - [Pocket Specific](#pocket-specific)
  - [Consensus](#consensus)
  - [🌲 Trees](#-trees)
  - [Persistence](#persistence)
  - [Ethereum](#ethereum)
  - [Rollups \& Sequencers Sequencers](#rollups--sequencers-sequencers)
  - [Cryptography](#cryptography)
  - [P2P](#p2p)
  - [Blogs](#blogs)
    - [Specific Article Recommendations](#specific-article-recommendations)
    - [General Subscription Recommendations](#general-subscription-recommendations)
- [❌ Non-suggested reads](#-non-suggested-reads)

## 🏁 Where to Start?

_IMPORTANT: If you are reading this, understand that if something looks incomplete, confusing or wrong, it most likely is. Don't be afraid to openly ask questions & submit a PR to change it!_

This is a general set of steps we have found to help new core team members onboard to Protocol development.

1. Watch some of our V1 video presentation to help you understand the core building components Pocket V1 (curtesy of @Olshansk) _(1.5 hours)_

   - [2023 ETH Denver presentation](https://www.youtube.com/watch?v=7rQ4Awfx79g)
   - [2022 Infracon presentation](https://www.youtube.com/watch?v=NJoZyzQuJVc)
   - [2022 ETH Denver presentation](https://www.youtube.com/watch?v=cjuDDdiMbMQ)

2. Run a **LocalNet** by following the [development guide](https://github.com/pokt-network/pocket/blob/main/docs/development/README.md). _(1-3 hours)_

   - This will get you set up to start contributing
   - Reach out to the core team in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel if you hit any issues
   - _TIP: Once you clone the Github repo, you can `cd pocket` and run `make` from to see all the available commands._

3. Get an understanding of the V1 spec summaries by reading about the [4 modules in our docs](https://docs.pokt.network/home/learn/future). _(2-4 hours)_

   - This will get you understand the spec foundations for Pocket V1
   - [Utility](https://docs.pokt.network/learn/future/utility/)
   - [Consensus](https://docs.pokt.network/learn/future/consensus/)
   - [P2P](https://docs.pokt.network/learn/future/peer-to-peer/)
   - [Persistence](https://docs.pokt.network/learn/future/persistence/)-

4. **Optional**: Go through [Otto's Pocket Guide](https://drive.google.com/drive/folders/1t-t0n7uMyvx-wBraDWBKVRLh532ZfD-c?usp=share_link) presentations to understand how Pocket V0 works.

   - [Volume 1](https://docs.google.com/presentation/d/1ftD1B_HTah1rzcO2yOqVZLsPtmIKKWunoAm7WvDpFPQ): The Claim and Proof lifecycle
   - [Volume 2](https://docs.google.com/presentation/d/1swFg6pzJSKXz9JnWkoQx5NegPnzuDWHnS4s25t7djt8): Block, Chains and Staking
   - [Volume 3](https://docs.google.com/presentation/d/1jGkJN7sWouavU1VgSxheL-UnV_Fdyb2ZIO65cwWZGUM): For The Builders and Hackers I - History Lesson
   - [Volume 4](https://docs.google.com/presentation/d/1D7hAAkMPW6Vo4uNA7PGY3KUZ8bcVF2SOjm4jLJdFwMY): For The Builders and Hackers II - History Lesson Continued

5. **Optional**: If you’re interested, you can view the [_OG v0 Pocket whitepaper_](https://pocket-network-assets.s3-us-west-2.amazonaws.com/pdfs/Pocket-Network-Whitepaper-v0.3.0.pdf). _(3-5 hours)_

   - This can provide both valuable and interesting historical context into Pocket

6. **Optional**: If you’re interested, you can view the [V0 Claim & Proof Lifecycle](https://github.com/pokt-network/pocket-core/blob/staging/doc/specs/reward_protocol.md) specifications. _(2-4 hours)_

   - It is a proprietary algorithm that you are unlikely to ever need to touch/modify, but it is fun & interesting to reason about and understand.

7. Find all of our documentation on the [Github Wiki](https://github.com/pokt-network/pocket/wiki).

   - The wiki mirrors all of the READMEs in our codebase so it's easy to find and navigate!

8. **Eventually**: Over time, **and not all at once**, you can start making your way through and updating the [V1 specifications](https://github.com/pokt-network/pocket-network-protocol). _(15-30 hours)_

   - Treat these specifications as guidelines and not sources of truth. If you are reading this, you will likely modify them at some point.
   - Mastering these concepts won't be easy but will make you an expert on Pocket V1
   - _TIP: We hope to publish it with V1 benchmarks on arxiv one day, so this is your chance to contribute 🙂_
   - **Optional**: If you're a core team member or heavily involved in the project, reach out to the team about getting access to the V1 specification research documents.
    <!-- For internal use only. If you're external and are reading this, reach out to the team.
       These decks from October 2021 might also help:
         - [Utility](https://docs.google.com/presentation/d/1NU0PnegtBm5ioLu0VQMiluWT4usHnavDKrGvS3p8QdM/edit)
         - [Persistence](https://docs.google.com/presentation/d/1qDA-pRMT1KV9byUAU49bvd_5seaILPAh6i3vA7j5l8o/edit)
         - [P2P](https://docs.google.com/presentation/d/1CLeAcGJbM_iP76vnCoHreU1chB9vFWIYWAwQHa-MPbc/edit)
         - [Consensus](https://docs.google.com/presentation/d/18CtSxxLLHY1N7HEJtja633mVF1_a9blaE2fe2-WgGAo/edit)
   -->

9. Start getting acquainted with the code structure by looking at the [docs on the shared architecture](https://github.com/pokt-network/pocket/tree/main/shared). _(1 hour)_

- This will help you understand the code architecture of Pocket V1

10. Get acquainted with our [GitHub Wiki](https://github.com/pokt-network/pocket/wiki) which automatically mirrors all of the READMEs we have in the codebase. _(5 mins)_

11. View our [V1 Roadmap](https://github.com/pokt-network/pocket/blob/main/docs/roadmap/README.md). _(10 mins)_

    - This will give you insight into our development & release timelines

12. Get a sense of all the open issues and tickets [in out Github project](https://github.com/orgs/pokt-network/projects/142/views/12). _(1 hour)_
13. If you don't already have a starter task, pick one from [Dework](https://app.dework.xyz/pokt-network), our [open issues](https://github.com/pokt-network/pocket/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc) or ask the team in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel. _(1 hour)_

14. Jump on a call and pair code! _(∞ and beyond)_

- If you need a walk-through of the code and some pointers before getting started, **jump on a call!**
- If you need to take your time to understand the problem and the code first, do so, and then **jump on a call!**
- If you don't need help, **when you’re about 33% of the way done**, show your draft work and get some feedback, so **jump on a call!**
- Just leave a message in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel and someone from the core team will respond.

## 🤫 Secret Zone

We aim to make all our documentation and resources public, but it takes more time and effort to maintain and share. The resources in this section are internal to the Pocket Team, but if you believe it'll bring you value in contributing to the project, reach out to some of the maintainer.

- [What's the hook 🪝](https://www.notion.so/pocketnetwork/What-s-the-Hook-18448627cde04d34bf4c5be96bd37cd3?pvs=4)(_45minutes_)
- [Protocol Hour Notes](https://www.notion.so/pocketnetwork/9e0590bdbd754163a5e7905f2246892c?v=ad14b5c982384c99944b153692c18f33&pvs=4)(_♾️_)

## 🏗️ Technical Foundation

### Github Development

If you're not familiar with the Github workflow, you can reference the [First Contributions](https://github.com/firstcontributions/first-contributions) repository.

### Golang

A great starting point to learn both the basics, idioms and some advanced parts of go is [The Way To Go](https://www.educative.io/courses/the-way-to-go) course on educative.io.

Afterwards, two great references you can constantly refer to are:

- [Effective Go](https://go.dev/doc/effective_go) by the official Gopher community
- [Practical Go](https://dave.cheney.net/practical-go) by Dave Cheney

### Mermaid

We used [Mermaid](https://mermaid.js.org/#/) as our [text-to-diagram](https://text-to-diagram.com/) framework to embed visuals alongside all of our visualisations. It makes the documentation easier to understand and maintain. As explained in [this comment](https://github.com/pokt-network/pocket/issues/335#issuecomment-1352064588), you can use [mermaid.live](https://mermaid.live) to work on them in your browser or install an extension in your editor of choice; which should probably be VSCode 🤓

## 📚 Technical References

_NOTE: We're trying not to make this a link dump, so please only add more references if it was actually helpful in clarifying your understanding. Don't treat these as must reads but as a signal for good sources to learn. If there was something that really helped clarify your understanding, please do include it!_

This is a general set of technical links and recommended reading our team has found useful to review and study for core technical concepts.

### Pocket Specific

- [Pocket V0 Whitepaper](https://pocket-network-assets.s3-us-west-2.amazonaws.com/pdfs/Pocket-Network-Whitepaper-v0.3.0.pdf)
  - This is the OG Pocket whitepaper if you want to go down memory lane
- [Pocket easy-to-learn Documents](https://docs.pokt.network/learn/)
  - This is the best starting point for anyone who is non-technical or very new to the project
- [Pocket Network V1 Specifications](https://github.com/pokt-network/pocket-network-protocol)
- [Relay Mining: Verifiable Multi-Tenant Distributed Rate Limiting](https://arxiv.org/abs/2305.10672)

- First Pocket Network V1 Presentations: The original presentations for Pocket V1 presented at the Mexico 2021 offsite

  - [Utility](https://docs.google.com/presentation/d/1NU0PnegtBm5ioLu0VQMiluWT4usHnavDKrGvS3p8QdM/edit?usp=sharing)
  - [Persistence](https://docs.google.com/presentation/d/1qDA-pRMT1KV9byUAU49bvd_5seaILPAh6i3vA7j5l8o/edit?usp=sharing)
  - [P2P](https://docs.google.com/presentation/d/1CLeAcGJbM_iP76vnCoHreU1chB9vFWIYWAwQHa-MPbc/edit?usp=sharing)
  - [Consensus](https://docs.google.com/presentation/d/18CtSxxLLHY1N7HEJtja633mVF1_a9blaE2fe2-WgGAo/edit?usp=sharing)

- [Pocket YouTube Channel](https://www.youtube.com/c/PocketNetwork/videos)
  - Contains everything from Infracon presentations, to contributor hour calls, etc...

### Consensus

- [Hotstuff whitepaper](https://arxiv.org/abs/1803.05069)
  - The original hotstuff whitepaper does a great job at explaining the algorithm on which HotPOKT is built
- [Attacks on BFT Algorithms](https://arxiv.org/pdf/1904.04098.pdf)
  - Covers various attacks on different BFT algorithms
- [Lessons from hotstuff](https://arxiv.org/pdf/2305.13556.pdf)
  - A 2023 review and lookback on HotStuff consensus
- [LazyLedger: A Distributed Data Availability Ledger With Client-Side Smart Contracts](https://arxiv.org/abs/1905.09274)
  - The original Celestia whitepaper focused on data availability sampling and light clients

### 🌲 Trees

- [Jellyfish Merkle Tree](https://developers.diem.com/papers/jellyfish-merkle-tree/2021-01-14.pdf)
  - An easy-to-read paper on JMT's that contains a good amount of background of how Merkle Trees work
- Verkle Trees
  - [Verkle Tree Whitepaper](https://math.mit.edu/research/highschool/primes/materials/2018/Kuszmaul.pdf)
    - The Verkle Tree whitepaper provides a good background on Merkle trees and some details on polynomial commitments
  - [Vitalik's Verkle Tree Review](https://vitalik.ca/general/2021/06/18/verkle.html)
    - Vitalik's analysis dives deeper into the math behind Verkle Trees with an alternative
  - [Verkle Trees EIP](https://notes.ethereum.org/@vbuterin/verkle_tree_eip)
    - A few notes from Vitalik on how Verkle Trees could be implemented in Ethereum
  - [Verkle trie for Eth1 state](https://dankradfeist.de/ethereum/2021/06/18/verkle-trie-for-eth1.html)
    - A few notes on how Verkle Trees can be use to make Ethereum Stateless
  - [](https://github.com/gballet/go-verkle)
- [Cosmos Discussion about Storage and IAVL](https://github.com/cosmos/cosmos-sdk/issues/7100)
  - This is a Github discussion between various Cosmos contributors of why and how to deprecate IAVL and goes into an intensive discussion of Merkle Tree alternatives
- [State commitment and storage review](https://paper.dropbox.com/doc/State-commitments-and-storage-review--Box9ruOvLDPaPc6ykc5XDnVmAg-wKl2RINZWD9I0DUmZIFwQ)
  - This research report was a result of the discussion above and goes in depth into state commitments and storage alternatives
- [Plasma Core Merkle Sum Tree](https://plasma-core.readthedocs.io/en/latest/specs/sum-tree.html)
  - A good reference to understand some of the underlying cryptography in V0's proof/claim lifecycle 9934927 (Add a couple more helpful links)
- Vitalik
  - [Vitalik's Blog on Polynomial Commitments](https://vitalik.ca/general/2021/01/26/snarks.html#polynomial-commitments)
    - A technical and gentle introduction to polynomial commitments by Vitalik from 2021
  - [Vitalik's Blog on Merkle Proofs](https://vitalik.ca/general/2021/06/18/verkle.html#merkle-proofs)
    - A technical and gentle introduction to Merkle Proofs by Vitalik from 2021
- [Indexed Merkle Trees](https://docs.aztec.network/aztec/protocol/trees/indexed-merkle-tree/)
  - Some problems identified by the Aztec team on using sparse Merkle trees for nullifier and their solution
- [Why you should probably never sort your Merkle tree's leaves](https://alinush.github.io/2023/02/05/Why-you-should-probably-never-sort-your-Merkle-trees-leaves.html)
  - Vulnerabilities and performance issues that come with sorting Merkle tree leaves
- Discussions
  - [Cosmos complete list of all Tree/State Related Discussions](https://github.com/cosmos/cosmos-sdk/issues/9816)
  - [Near discussion of more efficient state representation](https://github.com/near/nearcore/discussions/4815)

### Persistence

- Tendermint Discussion around a rollback tool for state
  - [Should we implement the rewind feature in tendermint?](https://github.com/tendermint/tendermint/issues/3845)
  - [Add command to roll-back a single block](https://github.com/tendermint/tendermint/issues/3196)

### Ethereum

- [Paths toward single-slot finality](https://notes.ethereum.org/@vbuterin/single_slot_finality)
  - An ethereum-pov explanation on the difficulty of having large validator networks
- [A note on data availability and erasure coding](https://github.com/ethereum/research/wiki/A-note-on-data-availability-and-erasure-coding)
  - A read 2018 that kicked off discussions related to Data Availability, leading to what Celestia is today as well as related work in the Ethereum ecosystem
- [EigenLyaer](https://2039955362-files.gitbook.io/~/files/v0/b/gitbook-x-prod.appspot.com/o/spaces%2FPy2Kmkwju3mPSo9jrKKt%2Fuploads%2F9tExk4U2OdiRKGEsUWqW%2FEigenLayer_WhitePaper.pdf?alt=media&token=c20ac4bd-badd-4826-9fb6-492923741c9e); the re-staking protocol

### Rollups & Sequencers Sequencers

- Espresso Systems
- https://hackmd.io/@EspressoSystems/EspressoSequencer
  https://hackmd.io/@EspressoSystems/SharedSequencing
- [RollUps aren't real](https://joncharbonneau.substack.com/p/rollups-arent-real?utm_source=profile&utm_medium=reader2)

### Cryptography

- [Anoma Whitepaper](https://github.com/anoma/whitepaper/blob/main/whitepaper.pdf)
  - This whitepaper is a bit dense but introduces a great way of thinking about the building blocks of decentralized applications and blockchains focused on security and intent. It always provides a great historical background on both Bitcoin and Ethereum.
  - This is the whitepaper is the foundation of "Intenet Driven Design" which has started to become a popular in 2023
- [Threshold signatures presentation](https://docs.google.com/presentation/d/1G4XGqrBLwqMyDQce_xpPQUEMOK4lFrneuvGYU3MVDsI/edit#slide=id.g1246936523c_0_26)
  - A great presentation by Alin Tomescu (founding engineer) that builds intuition around threshold signatures, signature aggregation, etc
- [ECDSA is not that bad: two-party signing without Schnorr or BLS](https://medium.com/cryptoadvance/ecdsa-is-not-that-bad-two-party-signing-without-schnorr-or-bls-1941806ec36f)
  - A gentle introduction to BLS aggregation
- [How Schnorr signatures may improve Bitcoin](https://medium.com/cryptoadvance/how-schnorr-signatures-may-improve-bitcoin-91655bcb4744)
  - A gentle introduction to Schnorr signatures

### P2P

- [Eclipsing Ethereum Peers with False Friends](https://arxiv.org/pdf/1908.10141.pdf)
  - A detailed explanation of how Kademlia (DHT for P2P networks) works, accompanying a deep dive into peer management and peer discovery in Geth, with the goal of outlining several attack vectors and their countermeasures.
- [A Brief Overview of Kademlia, and its use in various decentralized platforms](https://medium.com/coinmonks/a-brief-overview-of-kademlia-and-its-use-in-various-decentralized-platforms-da08a7f72b8f)
  - Short article from 2019 about some kademlia extensions Storj implemented (published in 2007) to prevent Sybil and Eclipse attacks.

### Blogs

#### Specific Article Recommendations

- [Three Quarks](https://davidphelps.substack.com/)
  - [The Case for Modular Maxis](https://davidphelps.substack.com/p/the-case-for-modular-maxis)
- [Aptos engineering blog](https://aptoslabs.medium.com/)
  - [The Evolution of State Sync](https://medium.com/aptoslabs/the-evolution-of-state-sync-the-path-to-100k-transactions-per-second-with-sub-second-latency-at-52e25a2c6f10)
- [Olshansky's blog](https://olshansky.substack.com/)
  - [5P;1R - Celestia (LazyLedger) White Paper](https://olshansky.substack.com/p/5p1r-celestia-lazyledger-white-paper)
  - [5P;1R - Ethereum's Modified Merkle Patricia Trie](https://olshansky.substack.com/p/5p1r-ethereums-modified-merkle-patricia)
  - [5P;1R - Bitcoin's Elliptic Curve Cryptography](https://olshansky.substack.com/p/5p1r-bitcoins-elliptic-curve-cryptography)
  - [5P;1R - Jellyfish Merkle Tree](https://olshansky.substack.com/p/5p1r-jellyfish-merkle-tree)

#### General Subscription Recommendations

- [Pocket Network Blog](https://www.blog.pokt.network/)
- [OG Pocket Network Blog](https://medium.com/@POKTnetwork)
- [Mike's blog](https://morourke.org/); yours truly
- [Vitalik's blog](https://vitalik.ca/categories/blockchains.html); hopefully Vitalik needs no introduction if you're reading this
- [Alin Tomescu](https://alinush.github.io/); Alin the is one of the founding engineers at Aptos Labs
- [Joachim Neu](https://www.jneu.net/); focus on the articles under the **Technical reports** section
- [Decentralized Thoughts by Ittai Ibrahim](about-ittai); The best blog about everything consensus blockchain related by one of the leading researchers in the field (the RSS feed is available [here](https://decentralizedthoughts.github.io/feed.xml))
- [Polymer Labs](https://polymerlabs.medium.com/); polymer has very easy to digest and easy to read blogs on IBC and interoperability

## ❌ Non-suggested reads

The papers in this list were read by our team and would not be recommended to become more productive to contributing to Pocket.

We do not consider them bad, but time is limited so it is important to focus on what will bring the most learning value.

- [Blockchains Meet Distributed Hash Tables: Decoupling Validation from State Storage](https://arxiv.org/abs/1904.01935)
  - An "extended abstract" of how Authenticated Data Structures (i.e. Merkle Trees) could be _"sharded"_ across nodes using Distributed Hash Tables (DHTs) to reduce the state required to be maintained and synched by each node.

<!-- GITHUB_WIKI: guides/learning/readme -->
