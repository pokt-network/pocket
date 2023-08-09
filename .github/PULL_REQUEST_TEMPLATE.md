<!-- REMOVE this comment block after following the instructions
 1. Make the title of the PR is descriptive and follows this format: `[<Module>] <DESCRIPTION>`
 2. Update the _Assigness_, _Labels_, _Projects_, _Milestone_ before submitting the PR for review.
 3. Add label(s) for the purpose (e.g. `persistence`) and, if applicable, priority (e.g. `low`) labels as well.
 4. See our custom action driven labels if you need to trigger a build or interact with an LLM - https://github.com/pokt-network/pocket/blob/main/docs/development/README.md#github-labels
-->

## Description

<!-- REMOVE this comment block after following the instructions
 1. Add a summary of the change including: motivation, reasons, context, dependencies, etc...
 2. If applicable, specify the key files that should be looked at.
 3. If you leave the `reviewpad:summary` block below, it'll autopopulate an AI generated summary. Alternatively, you can leave a `/reviewpad summarize` comment to trigger it manually.
-->

reviewpad:summary

## Issue

Fixes #<issue_number>

## Type of change

Please mark the relevant option(s):

- [ ] New feature, functionality or library
- [ ] Bug fix
- [ ] Code health or cleanup
- [ ] Major breaking change
- [ ] Documentation
- [ ] Other <!-- add details here if it a different type of change -->

## List of changes

<!-- REMOVE this comment block after following the instructions
 List out all the changes made.
 A PR should, preferably, be about a single change and the corresponding tests
-->

- Change #1
- Change #2
- ...

## Testing

- [ ] `make develop_test`; if any code changes were made
- [ ] `make test_e2e` on [k8s LocalNet](https://github.com/pokt-network/pocket/blob/main/build/localnet/README.md); if any code changes were made
- [ ] `e2e-devnet-test` passes tests on [DevNet](https://pocketnetwork.notion.site/How-to-DevNet-ff1598f27efe44c09f34e2aa0051f0dd); if any code was changed
- [ ] [Docker Compose LocalNet](https://github.com/pokt-network/pocket/blob/main/docs/development/README.md); if any major functionality was changed or introduced
- [ ] [k8s LocalNet](https://github.com/pokt-network/pocket/blob/main/build/localnet/README.md); if any infrastructure or configuration changes were made

<!-- REMOVE this comment block after following the instructions
 If you added additional tests or infrastructure, describe it here.
 Bonus points for images and videos or gifs.
-->

## Required Checklist

- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas (README(s), docs, godoc comments, etc...)
- [ ] I have tested my changes using the available tooling
- [ ] I have added, or updated, [`godoc` format comments](https://go.dev/blog/godoc) on touched members (see: [tip.golang.org/doc/comment](https://tip.golang.org/doc/comment))
- [ ] I have added, or updated, [mermaid.js](https://mermaid-js.github.io) diagrams in the corresponding README(s)
<!-- Changelogs are currently turned off
- [ ] I have updated the corresponding CHANGELOG
      -->
