# E2E Feature Path <!-- omit in toc -->

_IMPROVE(olshansky): Once we've completed the entire process at least once, we'll add links to each step._

- [Introduction \& Goals](#introduction--goals)
- [Developer Journey](#developer-journey)
- [E2E Feature Specification](#e2e-feature-specification)
  - [Spot Feature](#spot-feature)
  - [Spike Feature](#spike-feature)
  - [Scope Feature](#scope-feature)
    - [1. GitHub Ticket](#1-github-ticket)
    - [2. Origin Document](#2-origin-document)
- [E2E Feature Implementation](#e2e-feature-implementation)
  - [POC: Proof of Concept](#poc-proof-of-concept)
  - [MVC: Minimum Viable Change](#mvc-minimum-viable-change)
  - [PROD: Production](#prod-production)

## Introduction & Goals

The [Pocket Network Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/utility) implementation is driven by various [milestones](https://github.com/pokt-network/pocket/milestones) and protocol/module/component specific tasks. Each feature crosses the boundaries of business logic, data types, and interfaces for different components. Due to the complex nature of implementation, we've designed a streamlined "developer journey".

**The goal** of this document is to outline a well-defined process for incorporating an end-to-end feature path. This makes each feature/task easier to scope, reason about, design, and implement.

## Developer Journey

```mermaid
%%{init: { 'logLevel': 'debug', 'theme': 'base' } }%%
  timeline
    title E2E Feature Path Developer Journey

    section E2E Feature Specification
      Spot Feature:
        User research:
        Product management:
        Protocol intuition/experience:
        Addition/removal/selection of feature from the list
      Spike Feature:
        Research details:
        Identify pointers:
        Note dependencies:
        Find blockers
      Scope Feature:
        Define requirements:
        Document E2E implementation

    section E2E Feature Implementation
        POC:
          POC Spike:
          Explore:
          Hack:
          Have fun!
        MVC:
          Data Structures:
          Interfaces:
          Implementation:
          Unit Tests:
          E2E Tests:
          Documentation
        PROD:
          Identify workarounds & hacks:
          Document future work:
          Ideate!
```

## E2E Feature Specification

### Spot Feature

Choose or add a feature from the Utility E2E feature list [here](./E2E_FEATURE_LIST.md).

### Spike Feature

Create a SPIKE GitHub issue, like [this](<[http](https://github.com/pokt-network/pocket/issue/TODO_LINK_TO_ISSUE_ONCE_WE_HAVE_EXAMPLE)>) to scope the feature. This ticket is responsible for creating the ticket that'll track the work.

### Scope Feature

Leverage the results from the SPIKE to create an implementation GitHub issue, like [this](<[http](https://github.com/pokt-network/pocket/issue/TODO_LINK_TO_ISSUE_ONCE_WE_HAVE_EXAMPLE)>) to track the actual implementation.

#### 1. GitHub Ticket

Open a [new issue](https://github.com/pokt-network/pocket/issues/new?assignees=&labels=&projects=&template=issue.md&title=%5BREPLACE+ME%5D+with+a+descriptive+title) and populate its description, respectively, with the following additional elements:

**Objective**: `Implement MVC E2E Feature Path <Letter>.<Number>: <Name>`

**Origin Document**: _Specify the details from the [Origin Document](#origin-document) below_

**Goals**:

- Complete the MVC implementation of the E2E Feature Path outlined in the objective
- Identify future tasks and test requirements to transition the feature to production

**Deliverables**:

**POC**:

- [ ] A POC SPIKE to be discarded, refactored, and/or restructured into multiple PRs

**MVC**:

- [ ] A PR that adds or modifies relevant structures and interfaces; such as [shared/core/types/proto](../../shared/core/types/proto), [shared/modules](../../shared/modules), etc
- [ ] A PR that materializes an MVC of the feature along with unit tests
- [ ] A PR that introduces a new E2E tests with **one or more happy** and **one or more sad** path scenarios as described in the origin document (refer to [e2e/README.md](../../e2e/README.md)); this may require additions to the [cli](https://github.com/pokt-network/pocket/tree/main/app/client)
- [ ] A PR that updates all pertinent documentation

**PROD**:

- [ ] One or more subsequent GitHub issues that track future work including, but not limited to:
  - Enhancing test coverage
  - Adding subsequent features
  - Patching hacks or workarounds
  - Enabling your [imagination](https://github.com/pokt-network/pocket/assets/1892194/6aff9004-8d3b-48e8-b6d5-9b67ac266e3d)!

#### 2. Origin Document

**Purpose:** [Replace this with a single sentence that captures the intended purpose, behaviour and goal of the E2E feature]
**Actors**: Check all of the protocol actors involved in the feature:

- [ ] Validator
- [ ] Application
- [ ] Servicer
- [ ] Fisherman
- [ ] Portal

**Data Structures**:

- A list of the core types (protobufs, structs, etc) that will be used, added or modified in this feature
- Mention or link to specific files if applicable
- See [shared/core/types/proto](../../shared/core/types/proto) as a reference as they will most likely, but not necessarily, be part of that package
- _TIPS:_
  - _This will be non-exhaustive and will likely change during the POC or MVC stages_
  - You can find all other structs by running `make search_structs`
  - You can find all other protobufs by running `make search_protos`

**Interfaces**:

- A list of the interface (go interfaces, placeholder functions, grpc, etc) that will be used, added or modified in this feature
- Mention or link to specific files if applicable
- See [shared/modules](../../shared/modules) as a reference as they will most likely, but not necessarily, be part of that package
- _TIPS:_
  - _This will be non-exhaustive and will likely change during the POC or MVC stages_
  - You can find all other structs by running `make search_interfaces`
  - You can find all other protobufs by running `make search_protos`

**Diagram**:

- One or more mermaid diagrams that will visualize the E2E feature
- _TIPs:_
  - Use multiple diagrams if a single one ends up exceeding 7 or more core elements or steps
  - See if there’s anything in [pokt-network/pocket-network-protocol/tree/main/utility](../../utility/) or [pokt-network/pocket/tree/main/utility/doc](../../utility/doc) that you can use as a starting point

**User Stories as Tests**:

- Use natural language (long-form or bullet points) to define:
  - One (or more) HAPPY E2E path(s) from start to end with all the relevant details
  - One (or more) SAD E2E path(s) from start to end with all the relevant details
  - **[IMPROVE] Guiding template**: A [User | Actor | Source | etc] [performs an action] [where | when | at] [some specific context or state is guaranteed] and [the expected result is...].
    - **Example**: An Application requests the account balance of a specific address at a specific height when there is a Servicer staked for the Ethereum RelayChain in the same GeoZone, and receives a successful response.
- _NOTE: Keep in mind that these tests will be used to:_
  - _Interact with our [CLI](../../app/client/cli/) and [E2E testing framework](../../e2e) but do not design it for that_
  - _Train ChatGPT to expand on the happy and sad E2E test cases_

**Blockers**:

- A list of other E2E feature paths (in the format `<Letter>.<Number>: <Name>`) that:
  - must be implemented prior to this
  - will be mocked or added as placeholders to unblock this work
- Any other blockers, requirements or dependencies this will need and may need to be implemented as part of the feature implementation (e.g. infrastructure needs)

## E2E Feature Implementation

### POC: Proof of Concept

Create a single PR where you _"do everything"_ with the knowledge that it'll be closed without being merged in. This is an opportunity to get your hands dirty, understand the problem more deeply and have some fun. This PR may be split into multiple smaller PRs or just refactored altogether.

_TODO(example): Link to Alessandro's KISS or Bryan's P2P._

### MVC: Minimum Viable Change

This is the 🥩 (or 🥙) of the whole feature. In several PRs, you will implement, get feedback and merge in the feature to the main branch. Use your best judgment on how to split it so it's easier for the reviewer to give feedback without blocking yourself. It'll include some, or all, of the following PRs:

- Updates to data structures or protobufs
- Updates to interfaces
- The core implementation
- Updates to the [CLI](../../app/client/cli/)
- Updates to the [RPC](../../rpc/) service
- Additions / modifications to the [E2E testing framework](../../e2e)
- Additions / modifications to the documentation, code structure and diagrams

### PROD: Production

One or more follow-up GitHub issues that track follow-up work that my include, but not limited to:

- Increase the test coverage
- Adding subsequent features
- Fixing hacks
- Your [imagination](https://github.com/pokt-network/pocket/assets/1892194/6aff9004-8d3b-48e8-b6d5-9b67ac266e3d)!

<!-- GITHUB_WIKI: utility/e2e_feature_path_template -->
