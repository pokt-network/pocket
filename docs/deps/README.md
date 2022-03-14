# Table of Contents

1. [Mage](#mage)
2. [Reflex](#reflex)
3. [Go-ECVRF](#go-ecvrf)

# Mage

We currently depend on [mage](https://magefile.org/) as a build tool.

We wanted to reduce the possibility of human error in the creation of releases and get information about the source of the binary for troubleshooting scenarios.

To introduce automation in version generation there are several strategies, the simplest of which involves setting the version information at build time through flags sent to the Go toolchain linker.

Generating version information in general, and generating releases in particular, will use algorithms that can vary in complexity and amount of tasks automated.

In order to facilitate usage for the whole team, choosing a tool that is programmed in Go instead of a domain specific language seems sensible.

Usage in this case means:

- Ease of understanding the algorithms used for targets
- Ease of adding targets
- Ease of modifying targets

Hence, mage.

## Risk Assessment

### On the plus side

- Mage is licensed with Apache 2.0, implying that the copyright burdens of using it and even maintain a fork if it came to that are manageable
- Mage is used by Hugo, which is a big project unrelated to the maintainer; this hints at a group of people that will want to keep the project viable besides us (hugo has over 700 contributors)
- Mage 1.0 was released in 2017, and has seen releases often (18 at the time of writing); it is, therefore, relatively mature and seems healthy development activity wise
- Being a build tool, not part of the meat of the project, the impact of risks in the future of the tool are mitigated

### On the minus side

- The maintainer does not include information for support contracts, although they accept donations.

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
5. [coniks-sys/coins-go](https://github.com/coniks-sys/coniks-go/tree/master/crypto/vrf)
   - This is a small submodule of a much larger repository that is 5 years old, lacks documentation,
     with unclear code.

Though there are other articles and libraries available, [ProtonMail/Go-ECVRF](https://github.com/ProtonMail/go-ecvrf) was selected at the time of writing due to its recency (i.e. < 3 months old>), simplicity, clear documentation (i.e. easy to use), small size (lack of additional dependency), complete implementation in Go (i.e. no C++), and the backing of a well known brand (i.e. ProtonMail).
