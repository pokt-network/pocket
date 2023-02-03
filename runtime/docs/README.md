# RuntimeMgr

This document outlines the purpose of this module, its components and how they all interact with the other modules.

## Contents

- [RuntimeMgr](#runtimemgr)
  - [Contents](#contents)
    - [Overview](#overview)
    - [Components](#components)

### Overview

The `RuntimeMgr`'s purpose is to abstract the runtime so that it's easier to test and reason about various configuration scenarios.

It works like a black-box that takes the current environment/machine and therefore the configuration files, flags supplied to the binary, etc. and returns a structure that can be queried for settings that are relevant for the functioning of the modules and the system as a whole.

### Components

This module includes the following components:

- **Config**

  As the name says, it includes, in the form of properties, module specific configurations.

  It also has a `Base` configuration that is supposed to contain more cross-functional settings that cannot really find place in module-specific "subconfigs" (as another way to define module-specific configurations).

  Configuration can be supplied via JSON file but also via environment variables ([12 factor app](https://12factor.net/)).

  The naming convention is as follow:

  `POCKET_[module][configuration key]`

  So, for example, if you want to override the default RPC port we would use:

  > POCKET_RPC_PORT=yourport

- **Genesis**

  The genesis represents the initial state of the blockchain.

  This allows the binary to start with a specific initial state.
  

- **Clock**

  Clock is a drop-in replacement for some of the features offered by the `time` package, it acts as an injectable clock implementation used to provide time manipulation while testing.

  By default, the **real** clock is used and while testing it's possible to override it by using the "option" `WithClock(...)`

<!-- GITHUB_WIKI: runtime/readme -->
