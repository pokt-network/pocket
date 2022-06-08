# Build System

The build system we're using, mage, has been introduced with four targets; testing it involves invoking them and looking at their output.

## Versioning

All the targets that produce a binary should inject version information on it, which can be inspected by running the executable with the `-version` flag.

If built outside a repository, or without the `git` executable in the path, it should report the version `UNKNOWN`.

If built in a repository where uncommitted changes are present, the version should report a version with `+dirty`

Version reported for a proper git repository with the `git` executable present in the path should be `0.0.0-branch_name/commit_hash`, where the version number is fixed but the branch name reflects the current branch and the commit hash likewise the current commit's hash.

## Magefile build system

Once the V1 implementation reaches the stage of testable binaries, we are looking to use [Mage](https://magefile.org/) which is being tracked in [pocket/issues/43](https://github.com/pokt-network/pocket/issues/43) that'll inject a version with the `-version` flag.
