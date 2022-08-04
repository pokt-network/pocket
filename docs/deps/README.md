# Table of Contents

1. [Reflex](#reflex)
2. [Go-ECVRF](#go-ecvrf)

# Reflex

We're using it to hot reload the code. It can be installed with

`go install github.com/cespare/reflex@latest`

# Go-ECVRF

As part of Pocket Network's [consensus protocol](https://github.com/pokt-network/pocket-network-protocol/tree/main/consensus), a custom version of [Algorand's Leader Election](https://algorandcom.cdn.prismic.io/algorandcom%2Fa26acb80-b80c-46ff-a1ab-a8121f74f3a3_p51-gilad.pdf) algorithm is implemented. This implementation requires the use of a [Verifiable Random Function](https://en.wikipedia.org/wiki/Verifiable_random_function) (VRF).

VRFs are not part of Golang's [standard crypto library](https://pkg.go.dev/crypto), and the Crypto / Blockchain / Golang communities do not have a _goto_ VRF library at the time of writing. In addition, though implementing and maintaining an in-house VRF library is possible, it is not necessary given that several open-source implementations are available. For example:

1. [Algorand/libsodium](https://github.com/algorand/libsodium/tree/draft-irtf-cfrg-vrf-03)
   - As discussed in [this article](https://medium.com/algorand/algorand-releases-first-open-source-code-of-verifiable-random-function-93c2960abd61), Algorand forked a C++ crypto library called libsodium in 2018. This is a relatively large library with lots of other unneeded dependencies, that will also require C++ <> Go bindings, and was therefore unused.
2. [Coinbase/Kryptology](https://github.com/coinbase/kryptology/tree/master/pkg/verenc)
   - Originally released in December of 2021, this library has a lot of useful components but is also immature, lacks documentation, is difficult to understand, and we have already run into other issues trying to use it for other purposes ([eg1](https://github.com/coinbase/kryptology/issues/30), [eg2](https://github.com/coinbase/kryptology/issues/40)).
3. [yoseplee/VRF](https://github.com/yoseplee/vrf)
   - This is a great repository that also compares, lists and evaluates some of the other VRF libraries available, but was implemented by a single individual and while the explanation is very extensive, it lacks the backing and verification of a larger company.
4. [ProtonMail/Go-ECVRF](https://github.com/ProtonMail/go-ecvrf)

   - This is a very small and light-weight library dedidcated to a VRF implementation in Go, that was released in December of 2021 by ProtonMail: a well-known privacy-focused impartial company with
     great documentation and a very easy to use API
   - This implementation of VRFs uses the [new draft](https://datatracker.ietf.org/doc/draft-irtf-cfrg-vrf/), which introduced several changes to the original implementations, some of which improved the security model.

5. [coniks-sys/coins-go](https://github.com/coniks-sys/coniks-go/tree/master/crypto/vrf)
   - This is a small submodule of a much larger repository that is 5 years old, lacks documentation,
     with unclear code.

Though there are other articles and libraries available, [ProtonMail/Go-ECVRF](https://github.com/ProtonMail/go-ecvrf) was selected at the time of writing due to its recency (i.e., < 3 months old>), simplicity, clear documentation (i.e., easy to use), small size (lack of additional dependency), complete implementation in Go (i.e., no C++), and the backing of a well-known brand (i.e., ProtonMail).

## ProtonMail/Go-ECVRF - Security Notice

The author, [@wussler](https://github.com/wussler), of the `ProtonMail/Go-ECVRF` library pointed out the following security notice:

```
Watch out that we only implemented the TAI (Try-And-Increment) method to encode a value to the curve, that (as the name suggests) is a non-constant time mapping. I can't tell from the PR if you'll be proving secret inputs, but if you do beware that someone timing the operation might be able to infer some info about the secret input itself.
If your security model needs to consider this attack you should consider implementing the ELL2 (Ellgator map), that is not so trivial.
```

This is okay in the case of Pocket's Leader Election Algorithm because the seed that we are proving is not secret at the time that it is used. Specifically, the flow is:

1. Each validator generates VRF keys at some `height N`
2. The network leverages consensus messages to distribute the keys throughout the network in `O(N)`
3. The VRF keys begin to be used for leader election at some `height (N+M)` where `M > 0`
4. The input to the VRF for each `height (N+M')` where `M' ≥ M` will use publicly known information (e.g. appHash, byzValidators, etc..) known at `height (N+M'-1)` and therefore satisfy the security notice above.
