# Tests

## 4 nodes E2E

### Happy paths

- 4 nodes - 5 blocks - equal stake - 0 byzantine nodes => SUCCESS
- 4 nodes - 5 blocks - 2 whales - 2 remoras - 0 byzantine nodes => SUCCESS

### Failure Paths

- 4 nodes - 5 blocks - 2 rounds at every step - 0 byzanetine nodes
- 4 nodes - 5 blocks - cascading failures (worst case) - 0 byzanetine nodes

### Byzantine paths

- 4 nodes - 5 blocks - equal stake - 1 byzantine nodes => SUCESS
  - What if byzantine node is elected as leader?
- 4 nodes - 5 blocks - equal stake - 2 byzantine nodes => FAILURE

# Other Tests

## Threshold signatures

- 4 nodes - 4 CORRECT partial signatures - verify TRUE threshold signature
- 4 nodes - 3 CORRECT partial signatures - verify TRUE threshold signature
- 4 nodes - 2 CORRECT partial signatures - verify FALSE threshold signature
- 4 nodes - 3 CORRECT partial signatures - 1 FAKE partial signature - verify TRUE threshold signature
- 4 nodes - 2 CORRECT partial signatures - 2 FAKE partial signature - verify FALSE threshold signature
