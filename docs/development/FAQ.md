# FAQ

A list of common issues & resolutions shared by the V1 contributors

## Avoid redundant files from iCloud backup

- **Issue**: when working on MacOS with iCloud backup turned on, redundant files could be generated in GitHub projects. (e.g. `file.go` and `file 2.go`) Details can be found here in [this link](https://stackoverflow.com/a/62387243).
- **Solution**: adding `.nosync` as suffix to the workspace folder, e.g. `pocket.nosync`. Alternative, working in a folder that iCloud doesn't touch also works.

_NOTE: Consider turning off the `gofmt` in your IDE to prevent unexpected formatting_

## Unable to start LocalNet - permission denied

- **Issue**: when trying to run `make compose_and_watch` on an operating system with SELinux, the command gives the error:

```
Recreating validator2 ... done
Recreating validator4 ... done
Recreating validator1 ... done
Recreating validator3 ... done
Attaching to validator3, validator1, validator2, validator4
validator2    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
validator1    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
validator3    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
validator1 exited with code 2
validator4    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
validator2 exited with code 2
validator3 exited with code 2
validator4 exited with code 2
```

- **Solution**: A temporary fix would be to run

```bash
su -c "setenforce 0"
```

Whereas a permanent approach would be to allow the docker container access to the local repository

```bash
sudo chcon -Rt svirt_sandbox_file_t ./pocket
```

See [this stackoverflow post](https://stackoverflow.com/questions/24288616/permission-denied-on-accessing-host-directory-in-docker) for more details.

<!-- GITHUB_WIKI: guides/development/FAQ -->
