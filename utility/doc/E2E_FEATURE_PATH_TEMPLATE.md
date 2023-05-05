# Creating An Issue

Following the GitHub template we have [here](https://github.com/pokt-network/pocket/blob/main/.github/ISSUE_TEMPLATE/issue.md):

- **Objective**: `Implement MVP E2E Feature Path <Letter>.<Number>: <Name>`
- **Origin Document**: The prepared `Feature Path Template`
- **Goals**:
  - MVP Implementation of the E2E Feature Path in the objective
  - Identification of follow-up work and testing requirements to bring the feature to production
- **Deliverables**:
  - [ ] A POC SPIKE that will be closed and split into multiple PRs
  - [ ] A PR that adds or updates relevant structures and interfaces (e.g. in [shared/core/types/proto](https://github.com/pokt-network/pocket/tree/main/shared/core/types/proto), [shared/modules](https://github.com/pokt-network/pocket/tree/main/shared/modules) or elsewhere)
  - [ ] A PR that implements an MVP of the feature w/ unit tests
  - [ ] A PR that adds a new E2E testing feature with one happy and one sad path scenario as described in the origin document (see [e2e/README.md](https://github.com/pokt-network/pocket/blob/main/e2e/README.md)); may require additions the [cli](https://github.com/pokt-network/pocket/tree/main/app/client)
  - [ ] A PR that updates all relevant documentation
  - [ ] One or more follow-up GitHub issues that track follow-up work that my include, but not limited to:
    - Increase the test coverage
    - Adding subsequent features
    - Fixing hacks

## [Origin Document] Feature Path Template

**Purpose:** A single sentence that captures the intended purpose, behaviour and goal of the E2E feature.

**Actors**: Check all of the protocol actors involved in the feature

- [ ] Validator
- [ ] Application
- [ ] Servicer
- [ ] Fisherman
- [ ] Portal

**Data Structures**

- A list of the core types (protobufs, structs, etc…) that will be used, added or modified in this feature
- Mention or link to specific files if applicable
- See [shared/core/types/proto](https://github.com/pokt-network/pocket/tree/main/shared/core/types/proto) as a reference as they will most likely, but not necessarily, be part of that package
- _TIPS:_
  - _This will be non-exhaustive and will likely change during implementation_
  - You can find all other structs by running this command: `grep -r "type .* struct" --exclude-dir="vendor" --exclude="*.gen.go" --exclude="*.pb.go" .`
  - You can find all other protobufs by running this command: `find . -name "*.proto" -not -path "./vendor/*"`

**Interfaces**

- A list of the interface (go interface, placeholder function, grpc, etc…) that will be used, added or modified in this feature
- Mention or link to specific files if applicable
- See [shared/modules](https://github.com/pokt-network/pocket/tree/main/shared/modules) as a reference as they will most likely, but not necessarily, be part of that package
- _TIPS:_
  - _This will be non-exhaustive and will likely change during implementation_
  - You can find all other structs by running this command: `grep -r "type .* interface" --exclude-dir="vendor" --exclude="*.gen.go" --exclude="*.pb.go" .`
  - You can find all other protobufs by running this command: `find . -name "*.proto" -not -path "./vendor/*"`

**Diagram**

- _One or more mermaid diagrams that will visualize the E2E feature_
- _TIPs:_
  - _Use multiple diagrams if a single one ends up exceeding 7 or more core elements or steps_
  - See if there’s anything in [pokt-network/pocket-network-protocol/tree/main/utility](https://github.com/pokt-network/pocket-network-protocol/tree/main/utility) or [pokt-network/pocket/tree/main/utility/doc](https://github.com/pokt-network/pocket/tree/main/utility/doc) that you can use as a starting point

**User Stories as Tests**

- Use natural language (long-form or bullet points) to define:
  - One (or more) HAPPY e2e path from start to end with all the relevant details
  - One (or more) SAD e2e path from start to end with all the relevant details
  - _TODO(@Bryan White @Daniel Olshansky ) :_ Present a template for how to formulate this to be primarily used in the earlier stages of MVP feature implementation until we have a structure to feed ChatGPT
- _NOTE: Keep in mind that these tests will be used to:_
  - _Interact with our CLI and E2E testing framework but do not design it for that_
  - _Train ChatGPT to expand on the happy and sad E2E test cases_

**Blockers**

- A list of other E2E feature paths (in the format `<Letter>.<Number>: <Name>`) that:
  - must be implemented prior to this
  - will be mocked or be added as placeholders as part of this
- Any other blockers, requirements or dependencies this will need and may need to be implemented as part of the feature implementation
