# Learning Pocket

_This is a live document on how to get ramped up on all the knowledge you need to contribute to the Pocket Protocol._

## Where to Start?

This is a general set of steps we have found to help new core team members onboard to Protocol development.

1.  Watch our [2022 Infracon presentation on v1](https://www.youtube.com/watch?v=NJoZyzQuJVc) to get a general idea of how everything works.
2.  Run a LocalNet by following the instructions [here](https://github.com/pokt-network/pocket/blob/main/docs/development/README.md).

    **Pro tip**: Did you know there are a bunch of `make` commands? Run `make` from the root repo to see what’s available.

3.  Get an understanding of the the V1 spec summaries by reading about the 4 modules [here](https://docs.pokt.network/home/learn/future).
4.  If you’re interested in the _OG v0 Pocket whitepaper_, check out [this link](https://pocket-network-assets.s3-us-west-2.amazonaws.com/pdfs/Pocket-Network-Whitepaper-v0.3.0.pdf).
5.  Over time, **and not all at once**, you can start making your way through and updating the V1 specs: https://github.com/pokt-network/pocket-network-protocol.

    **Tip**: We hope to publish it with V1 benchmarks on arxiv one day, so this is your change to contribute 🙂

    <!-- For internal use only. If you're external and are reading this, reach out to the team.
       These decks from October 2021 might also help:
       - [Utility](https://docs.google.com/presentation/d/1NU0PnegtBm5ioLu0VQMiluWT4usHnavDKrGvS3p8QdM/edit)
       - [Persistence](https://docs.google.com/presentation/d/1qDA-pRMT1KV9byUAU49bvd_5seaILPAh6i3vA7j5l8o/edit)
       - [P2p](https://docs.google.com/presentation/d/1CLeAcGJbM_iP76vnCoHreU1chB9vFWIYWAwQHa-MPbc/edit)
       - [Consensus](https://docs.google.com/presentation/d/18CtSxxLLHY1N7HEJtja633mVF1_a9blaE2fe2-WgGAo/edit)
    -->

6.  To start getting acquainted with the code, look at the [README on the shared architecture](https://github.com/pokt-network/pocket/tree/main/shared).

7.  Want to get a sense of all the open issues and tickets? [Check out this page](https://github.com/orgs/pokt-network/projects/142/views/12) or ask for a starter task in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel if you don’t have one.

8.  Jump on a call and pair code!

- If you need a walk-through of the code and some pointers before getting started, **jump on a call!**
- If you need to take your time to understand the problem and the code first, do so, and then **jump on a call!**
- If you don't need help, **when you’re about 33% of the way done**, show your draft work and get some feedback, so **jump on a call!**
- Just leave a message in the [#v1-dev](https://discord.com/channels/553741558869131266/986789914379186226) discord channel and someone from the core team will respond.

## Technical References

This is a general set of technical links and recommended reading our team has found useful to review and study for core technical concepts.

### Pocket Specific

- [Pocket V0 Whitepaper](https://pocket-network-assets.s3-us-west-2.amazonaws.com/pdfs/Pocket-Network-Whitepaper-v0.3.0.pdf)
  - This is the OG Pocket whitepaper if you want to go down memory lane
- [Pocket easy-to-learn Documents](https://docs.pokt.network/learn/)
  - This is a great starting point for anyone who is not technical or who has a very basic foundation
- [Pocket YouTube Channel](https://www.youtube.com/c/PocketNetwork/videos)
  - Contains everything from infracon presentations, to contributor hour calls, etc...
- [Pocket Network V1 Specifications](https://github.com/pokt-network/pocket-network-protocol)

### Consensus

- [Hotstuff whitepaper](https://arxiv.org/abs/1803.05069)
  - The original hotstuff whitepaper does a great job at explaining the algorithm on which HotPOKT is built
- [Attacks on BFT Algorithms](https://arxiv.org/pdf/1904.04098.pdf)

### Merkle Trees

- [Jellyfish Merkle Tree](https://developers.diem.com/papers/jellyfish-merkle-tree/2021-01-14.pdf)
- Verkle Trees
  - [Verkle Tree Whitepaper](https://math.mit.edu/research/highschool/primes/materials/2018/Kuszmaul.pdf)
    -
  - [Vitalik's Verkle Tree Review](https://vitalik.ca/general/2021/06/18/verkle.html)
    -
- [Comos Discussion about Storage and IAVL](https://github.com/cosmos/cosmos-sdk/issues/7100)


### Cryptography

- [Threshold signatures presentation](https://docs.google.com/presentation/d/1G4XGqrBLwqMyDQce_xpPQUEMOK4lFrneuvGYU3MVDsI/edit#slide=id.g1246936523c_0_26)
  - A great presentation by Alin Tomescu (founding engineer) that builds intuition around threshold signatures, signature aggregation, etc
- BLS Signatures: https://en.wikipedia.org/wiki/BLS_digital_signature

### Blogs
- Aptos

### Blogs
- Pocket engineering blog
  - New
  - Old
- Aptos engineering blog
- Mike's blog
- Olshansky's blog