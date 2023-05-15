<!-- REMOVE this comment block after following the instructions
 1. Make the title of the PR is descriptive and follows this format: `[<Module>] <DESCRIPTION>`
 2. Update the _Assigness_, _Labels_, _Projects_, _Milestone_ before submitting the PR for review.
 3. Add label(s) for the purpose (e.g. `persistence`) and, if applicable, priority (e.g. `low`) labels as well.
-->

## Description

<!-- REMOVE this comment block after following the instructions
1. Add a summary of the change including: motivation, reasons, context, dependencies, etc...
2. If applicable, specify the key files that should be looked at.
-->

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
 List out all the changes made
-->

- Change #1
- Change #2
- ...

## Testing

- [ ] `make develop_test`; if any code changes were made
- [ ] `make test_e2e` on [k8s LocalNet](https://github.com/pokt-network/pocket/blob/main/build/localnet/README.md); if any code changes were made
- [ ] Passes E2E tests on [DevNet](https://pocketnetwork.notion.site/How-to-DevNet-ff1598f27efe44c09f34e2aa0051f0dd); if any code was changed
- [ ] [Docker Compose LocalNet](https://github.com/pokt-network/pocket/blob/main/docs/development/README.md); if any major functionality was changed or introduced
- [ ] [k8s LocalNet](https://github.com/pokt-network/pocket/blob/main/build/localnet/README.md); if any infrastructure or configuration changes were made

<!-- REMOVE this comment block after following the instructions
 If you added additional tests or infrastructure, describe it here.
 Bonus points for images and videos or gifs.
-->

## Required Checklist

- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have added, or updated, [`godoc` format comments](https://go.dev/blog/godoc) on touched members (see: [tip.golang.org/doc/comment](https://tip.golang.org/doc/comment))
- [ ] I have tested my changes using the available tooling
- [ ] I have updated the corresponding CHANGELOG

### If Applicable Checklist

- [ ] I have updated the corresponding README(s); local and/or global
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] I have added, or updated, [mermaid.js](https://mermaid-js.github.io) diagrams in the corresponding README(s)
- [ ] I have added, or updated, documentation and [mermaid.js](https://mermaid-js.github.io) diagrams in `shared/docs/*` if I updated `shared/*`README(s)
