# Testing the Build System
The build system we're using, mage, has been introduced with four targets; testing it involves invoking them and looking at their output.

## Version

All the targets that produce a binary should inject version information on it, which can be inspected by running the executable with the `-version` flag.

If built outside a repository, or without the `git` executable in the path, it should report version `UNKNOWN`.

If built in a repository where uncommitted changes are present, the version should report a version with `+dirty`

Version reported for a proper git repository with the `git` executable present in the path should be `0.0.0-branchname/commithash`, where the version number is fixed but the branchname reflects the current branch and the commit hash likewise the current commit's hash.

## Listing Targets

running `mage -l` should show the 4 targets with brief descriptions for each.

## Install

Running `mage install` should make it so that the `pocket` binary lives in `$GOPATH/bin/`.

## Uninstall

Running `mage uninstall` after running `mage install` should remove the `pocket` binary from `$GOPATH/bin/`.

## Build

Running `mage build` should attempt to build the project; if it is successful, the `pocket` binary should be placed in the repository's `bin/` directory.

## Build Race

Running `buildRace` should attempt to build the project; if it is successful, the `pocket` binary should be placed in the repository's `bin/` directory.

The version for a binary built with this target should contain `+race`.