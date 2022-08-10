# Automatic Builds

We have automation that creates new container images for the Pocket Network v1 Node.

## Tags

Code built from default branch (i.e. `main`) is tagged as `latest`.

Code built from commits in Pull Requests, is tagged as `pr-<number>`, as well as `sha-<7 digit sha>`.


### Extended images with additional tooling

We also supply an extended image with tooling for each container tag to help you troubleshoot or investigate issues. The extended image is called `<tag>-dev`. For example, `latest-dev`, or `pr-123-dev`.
