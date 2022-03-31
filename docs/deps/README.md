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
