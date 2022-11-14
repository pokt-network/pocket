# Learning Pocket <!-- omit in toc -->

_This is a live document on how to get ramped up on all the knowledge you need to contribute to the Pocket Protocol._

- [🏁 Where to Start?](#-where-to-start)
- [🏗️ Technical Foundation](#️-technical-foundation)
  - [Github Development](#github-development)
  - [Golang](#golang)
- [📚 Technical References](#-technical-references)
  - [Pocket Specific](#pocket-specific)
  - [Consensus](#consensus)
  - [Merkle Trees](#merkle-trees)
  - [Ethereum](#ethereum)
  - [Cryptography](#cryptography)
  - [P2P](#p2p)
  - [Blogs](#blogs)
    - [Specific Article Recommendations](#specific-article-recommendations)
    - [General Subscription Recommendations](#general-subscription-recommendations)
- [❌ Non-suggested reads](#-non-suggested-reads)

## 🏁 Where to Start?

This is a general set of steps we have found to help new core team members onboard to Protocol development.

1. Watch our [2022 Infracon presentation on v1](https://www.youtube.com/watch?v=NJoZyzQuJVc) to get a general idea of how everything works. _(43 mins)_

   - This will help you understand the core building components of a Pocket V1 Node

2. Run a **LocalNet** by following the [development guide](https://github.com/pokt-network/pocket/blob/main/docs/development/README.md). _(1-2 hours)_

   - This will get you set up to start contributing
   - Reach out to the core team in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel if you hit any issues
   - _TIP: Once you clone the Github repo, you can `cd pocket` and run `make` from to see all the available commands._

3. Get an understanding of the the V1 spec summaries by reading about the [4 modules in our docs](https://docs.pokt.network/home/learn/future). _(3-5 hours)_

   - This will get you understand the spec foundations for Pocket V1

4. **Optional**: If you’re interested, you can view the [_OG v0 Pocket whitepaper_](https://pocket-network-assets.s3-us-west-2.amazonaws.com/pdfs/Pocket-Network-Whitepaper-v0.3.0.pdf). _(10-12 hours)_

   - This can provide both valuable and interesting historical context into Pocket

5. Over time, **and not all at once**, you can start making your way through and updating the [V1 specifications](https://github.com/pokt-network/pocket-network-protocol). _(30 hours)_

   - Mastering these concepts won't be easy but will make you an expert on Pocket V1
   - _TIP: We hope to publish it with V1 benchmarks on arxiv one day, so this is your change to contribute 🙂_
   - **Optional**: If you're a core team member or heavily involved in the project, reach out to the team about getting access to the V1 specification research documents.
    <!-- For internal use only. If you're external and are reading this, reach out to the team.
       These decks from October 2021 might also help:
         - [Utility](https://docs.google.com/presentation/d/1NU0PnegtBm5ioLu0VQMiluWT4usHnavDKrGvS3p8QdM/edit)
         - [Persistence](https://docs.google.com/presentation/d/1qDA-pRMT1KV9byUAU49bvd_5seaILPAh6i3vA7j5l8o/edit)
         - [P2P](https://docs.google.com/presentation/d/1CLeAcGJbM_iP76vnCoHreU1chB9vFWIYWAwQHa-MPbc/edit)
         - [Consensus](https://docs.google.com/presentation/d/18CtSxxLLHY1N7HEJtja633mVF1_a9blaE2fe2-WgGAo/edit)
   -->

6. Start getting acquainted with the code structure by looking at the [docs on the shared architecture](https://github.com/pokt-network/pocket/tree/main/shared). _(1 hour)_

   - This will help you understand the code architecture of Pocket V1

7. View our [V1 Roadmap](https://github.com/pokt-network/pocket/blob/main/docs/roadmap/README.md). _(10 mins)_

   - This will give you insight into our development & release timelines

8. Get a sense of all the open issues and tickets [in out Github project](https://github.com/orgs/pokt-network/projects/142/views/12). _(1 hour)_
9. If you don't already have a starter task, pick one from [Dework](https://app.dework.xyz/pokt-network), our [open issues](https://github.com/pokt-network/pocket/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc) or ask the team in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel. _(1 hour)_

10. Jump on a call and pair code! _(∞ and beyond)_

- If you need a walk-through of the code and some pointers before getting started, **jump on a call!**
- If you need to take your time to understand the problem and the code first, do so, and then **jump on a call!**
- If you don't need help, **when you’re about 33% of the way done**, show your draft work and get some feedback, so **jump on a call!**
- Just leave a message in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel and someone from the core team will respond.

## 🏗️ Technical Foundation

### Github Development

If you're not familiar with the Github workflow, you can reference the [First Contributions](https://github.com/firstcontributions/first-contributions) repository.

### Golang

A great starting point to learn both the basics, idioms and some advanced parts of go is [The Way To Go](https://www.educative.io/courses/the-way-to-go) course on educative.io.

Afterwards, two great references you can constantly refer to are:

* [Effective Go](https://go.dev/doc/effective_go) by the official Gopher community
* [Practical Go](https://dave.cheney.net/practical-go) by Dave Cheney

## 📚 Technical References

_We're trying not to make this a link dump, so please only add more references if it was actually helpful in clarifying your understanding._

This is a general set of technical links and recommended reading our team has found useful to review and study for core technical concepts.

### Pocket Specific

- [Pocket V0 Whitepaper](https://pocket-network-assets.s3-us-west-2.amazonaws.com/pdfs/Pocket-Network-Whitepaper-v0.3.0.pdf)
  - This is the OG Pocket whitepaper if you want to go down memory lane
- [Pocket easy-to-learn Documents](https://docs.pokt.network/learn/)
  - This is the best starting point for anyone who is non-technical or very new to the project
- [Pocket Network V1 Specifications](https://github.com/pokt-network/pocket-network-protocol)

- First Pocket Network V1 Presentations: The original presentations for Pocket V1 presented at the Mexico 2021 offsite

  - [Utility](https://docs.google.com/presentation/d/1NU0PnegtBm5ioLu0VQMiluWT4usHnavDKrGvS3p8QdM/edit?usp=sharing)
  - [Persistence](https://docs.google.com/presentation/d/1qDA-pRMT1KV9byUAU49bvd_5seaILPAh6i3vA7j5l8o/edit?usp=sharing)
  - [P2P](https://docs.google.com/presentation/d/1CLeAcGJbM_iP76vnCoHreU1chB9vFWIYWAwQHa-MPbc/edit?usp=sharing)
  - [Consensus](https://docs.google.com/presentation/d/18CtSxxLLHY1N7HEJtja633mVF1_a9blaE2fe2-WgGAo/edit?usp=sharing)

- [Pocket YouTube Channel](https://www.youtube.com/c/PocketNetwork/videos)
  - Contains everything from Infracon presentations, to contributor hour calls, etc...
  -

### Consensus

- [Hotstuff whitepaper](https://arxiv.org/abs/1803.05069)
  - The original hotstuff whitepaper does a great job at explaining the algorithm on which HotPOKT is built
- [Attacks on BFT Algorithms](https://arxiv.org/pdf/1904.04098.pdf)
  - Covers various attacks on different BFT algorithms

### Merkle Trees

- [Jellyfish Merkle Tree](https://developers.diem.com/papers/jellyfish-merkle-tree/2021-01-14.pdf)
  - An easy-to-read paper on JMT's that contains a good amount of background of how Merkle Trees work
- Verkle Trees
  - [Verkle Tree Whitepaper](https://math.mit.edu/research/highschool/primes/materials/2018/Kuszmaul.pdf)
    - The Verkle Tree whitepaper provides a good background on Merkle trees and some details on polynomial commitments
  - [Vitalik's Verkle Tree Review](https://vitalik.ca/general/2021/06/18/verkle.html)
    - Vitalik's analysis dives deeper into the math behind Verkle Trees with an alternative
- [Cosmos Discussion about Storage and IAVL](https://github.com/cosmos/cosmos-sdk/issues/7100)
  - This is a Github discussion between various Cosmos contributors of why and how to deprecate IAVL and goes into an intensive discussion of Merkle Tree alternatives
- [State commitment and storage review](https://paper.dropbox.com/doc/State-commitments-and-storage-review--Box9ruOvLDPaPc6ykc5XDnVmAg-wKl2RINZWD9I0DUmZIFwQ)
  - This research report was a result of the discussion above and goes in depth into state commitments and storage alternatives
- [Plasma Core Merkle Sum Tree](https://plasma-core.readthedocs.io/en/latest/specs/sum-tree.html)
  - A good reference to understand some of the underlying cryptography in V0's proof/claim lifecycle 9934927 (Add a couple more helpful links)

### Ethereum

- [Paths toward single-slot finality](https://notes.ethereum.org/@vbuterin/single_slot_finality)
  - An ethereum-pov explanation on the difficulty of having large validator networks

### Cryptography

- [Threshold signatures presentation](https://docs.google.com/presentation/d/1G4XGqrBLwqMyDQce_xpPQUEMOK4lFrneuvGYU3MVDsI/edit#slide=id.g1246936523c_0_26)
  - A great presentation by Alin Tomescu (founding engineer) that builds intuition around threshold signatures, signature aggregation, etc
- [ECDSA is not that bad: two-party signing without Schnorr or BLS](https://medium.com/cryptoadvance/ecdsa-is-not-that-bad-two-party-signing-without-schnorr-or-bls-1941806ec36f)
  - A gentle introduction to BLS aggregation
- [How Schnorr signatures may improve Bitcoin](https://medium.com/cryptoadvance/how-schnorr-signatures-may-improve-bitcoin-91655bcb4744)
  - A gentle introduction to Schnorr signatures

### P2P

- [Eclipsing Ethereum Peers with False Friends](https://arxiv.org/pdf/1908.10141.pdf)
  - A detailed explanation of how Kademlia (DHT for P2P networks) works, accompanying a deep dive into peer management and peer discovery in Geth, with the goal of outlining several attack vectors and their countermeasures.

### Blogs

#### Specific Article Recommendations

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
- [Vitalik's blog](https://vitalik.ca/categories/blockchains.html)
- [Mike's blog](https://morourke.org/)
- [Joachim Neu](https://www.jneu.net/)
  - See the articles under the **Technical reports** section

## ❌ Non-suggested reads

The papers in this list were read by our team and would not be recommended to become more productive to contributing to Pocket.

We do not consider them bad, but time is limited so it is important to focus on what will bring the most learning value.

- [Blockchains Meet Distributed Hash Tables: Decoupling Validation from State Storage](https://arxiv.org/abs/1904.01935)
  - An "extended abstract" of how Authenticated Data Structures (i.e. Merkle Trees) could be _"sharded"_ across nodes using Distributed Hash Tables (DHTs) to reduce the state required to be maintained and synched by each node.
