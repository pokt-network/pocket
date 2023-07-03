# Contribution guide

- For reference, [here is the detailed version](https://github.com/pokt-network/pocket/blob/main/utility/doc/E2E_FEATURE_PATH_TEMPLATE.md#e2e-feature-implementation).
  
- This doc assumes you know exactly what to work on, i.e. you are assigned [an issue with details hashed out](https://github.com/pokt-network/pocket/issues/754).

- This only applies to the `utility` module.

# Steps

1. Open a [POC PR](https://github.com/pokt-network/pocket/blob/main/utility/doc/E2E_FEATURE_PATH_TEMPLATE.md#poc-proof-of-concept).
   
   - This will not be merged, but used to A) get a review on the approach and B) breakup the work into a list of PRs
   - Should include the bare minimum E2E test(s) to test/demo the assigned feature: [a `.feature` file](https://github.com/pokt-network/pocket/pull/869/files#diff-30c6c02e8594cf72662ab975e75a810a5bbd702f274e2eaf160c97ec14f5e642) and [the updated `.go` steps file](https://github.com/pokt-network/pocket/pull/869/files#diff-01dec4121ae8acb7a1f4bb72a6c2104827d2c2d2197eb5f45fa5c032ffba32cd)

2. List [the PRs needed for making the E2E test pass](https://github.com/pokt-network/pocket/pull/869#issuecomment-1618484939)

   - Each PR should be about a single change and the corresponding tests

3. Make sure the reviewers of the POC PR are in agreement wih the proposed PR list

4. Work through the set of proposed PRs to make the E2E test pass
