# FAQ

A list of common issues & resolutions shared by the V1 contributors

## Avoid redundant files from iCloud backup

* **Issue**: when working on MacOS with iCloud backup turned on, redundant files could be generated in GitHub projects. (e.g. `file.go` and `file 2.go`) Details can be found here in [this link](https://stackoverflow.com/a/62387243).
* **Solution**: adding `.nosync` as suffix to the workspace folder, e.g. `pocket.nosync`. Alternative, working in a folder that iCloud doesn't touch also works.

_NOTE: Consider turning of the `gofmt` in your IDE to prevent unexpected formatting_

<!-- GITHUB_WIKI: guides/development/FAQ -->
