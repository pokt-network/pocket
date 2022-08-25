# Genesis Module

This document is meant to be the develompent level documentation for the genesis state details related to the design of the codebase and information related to development.

!!! IMPORTANT !!!

This directory was created for the purposes of integration between the four core modules and is
not intended to store all the core shared types in the long-term.

Speak to @andrewnguyen22 or @Olshansk for more details.

!!! IMPORTANT !!!

## Implementation

In order to maintain code agnostic from the inception of the implementation, protobuf3 is utilized for all structures in this package.

It is important to note, that while Pocket V1 strives to not share objects between modules, the genesis module will inevitably overlap
between other modules.

Another architecture worth considering and perhaps is more optimal as the project nears mainnet is allowing each module to create and maintain their own genesis object and config files

### Code Organization

```bash
genesis
├── docs
│   ├── CHANGELOG.md    # Genesis module changelog
│   ├── README.md       # Genesis module README
├── proto
│   ├── account.proto   # account structure
│   ├── actor.proto     # actor structure
│   ├── config.proto    # configuration structure
│   ├── gov.proto       # params structure
│   ├── state.proto     # genesis state structure
├── test_artifacts            # the central point of all testing code (WIP)
│   ├── generator.go    # generate the genesis and config.json for tests and build
│   ├── gov.go          # default testing parameters

```
