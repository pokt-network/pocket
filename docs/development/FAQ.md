# FAQ

A list of common issues & resolutions shared by the V1 contributors

## Avoid redundant files from iCloud backup

* **Issue**: when working on MacOS with iCloud backup turned on, redundant files could be generated in GitHub projects. (e.g. `file.go` and `file 2.go`) Details can be found here in [this link](https://stackoverflow.com/a/62387243).
* **Solution**: adding `.nosync` as suffix to the workspace folder, e.g. `pocket.nosync`. Alternative, working in a folder that iCloud doesn't touch also works.

_NOTE: Consider turning of the `gofmt` in your IDE to prevent unexpected formatting_

## Unable to start LocalNet - permission denied

* **Issue**: when trying to run `make compose_and_watch` on an operating system with SELinux, the command gives the error:

```
Recreating node2.consensus ... done
Recreating node4.consensus ... done
Recreating node1.consensus ... done
Recreating node3.consensus ... done
Attaching to node3.consensus, node1.consensus, node2.consensus, node4.consensus
node2.consensus    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
node1.consensus    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
node3.consensus    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
node1.consensus exited with code 2
node4.consensus    | /bin/sh: can't open 'build/scripts/watch.sh': Permission denied
node2.consensus exited with code 2
node3.consensus exited with code 2
node4.consensus exited with code 2
```

* **Solution**: A temporary fix would be to run

```bash
su -c "setenforce 0"
```

Whereas a permenant approach would be to allow the docker container access to the local repository

```bash
sudo chcon -Rt svirt_sandbox_file_t ./pocket
```

See [this stackoverflow post](https://stackoverflow.com/questions/24288616/permission-denied-on-accessing-host-directory-in-docker) for more details.

## Unable to create a LocalNet client

* **Issue**: When running `make client_start` you get the error message:

```
ERROR: No such service: --build
make: *** [Makefile:125: client_start] Error 1
```

* **Solution**: open `Makefile` and edit the following block of code

```make
.PHONY: client_start
client_start: docker_check ## Run a client daemon which is only used for debugging purposes
    docker-compose -f build/deployments/docker-compose.yaml up -d client --build
```

Move the `--build` profile so that it is before the `client` service, making the lines look like:
```make
.PHONY: client_start
client_start: docker_check ## Run a client daemon which is only used for debugging purposes
    docker-compose -f build/deployments/docker-compose.yaml up -d --build client
```

For more information on this refer to the [docker documentation](https://docs.docker.com/compose/profiles/)
