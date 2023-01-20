# LocalNet on Kubernetes <!-- omit in toc -->

This guide shows how to deploy a LocalNet using [pocket-operator](https://github.com/pokt-network/pocket-operator).

- [Dependencies](#dependencies)
- [Running the LocalNet](#running-the-localnet)
- [Scaling actors on LocalNet](#scaling-actors-on-localnet)
- [Stopping and cleaning up the resources](#stopping-and-cleaning-up-the-resources)
- [Interaction with the LocalNet](#interaction-with-the-localnet)
- [How does it work?](#how-does-it-work)
- [Troubleshooting](#troubleshooting)
- [How to change configuration files](#how-to-change-configuration-files)

### Dependencies

* [tilt](https://docs.tilt.dev/install.html) - installed automatically on `make install_cli_deps` command.
* Kubernetes cluster ([different options available](https://docs.tilt.dev/choosing_clusters.html)).
  * `kubectl` CLI is required and should be configured to access the cluster. That should happen automatically if you're using Docker Desktop, Rancher Desktop, k3s, k3d, minikube, etc.
  * `helm` - required to template the yaml manifests for the dependencies (such as postgres, grafana). Installation instructions: https://helm.sh/docs/intro/install/.

### Running the LocalNet

```
make localnet_up
```

The developer can then view the logs of services running via:
  - In terminal:
    - `make localnet_logs_validators` - shows prior logs
    - `make localnet_logs_validators_follow` - shows prior logs and follows the new log lines as validators do their work
  - Tilt web UI, either by:
    - Pressing `space` in the terminal where you started `tilt`
    - Going to [localhost:10350](http://localhost:10350/)

![tilt UI](tilt-ui.png)

### Scaling actors on LocalNet

Once you start LocalNet, new file `localnet_config.yaml` is going to get created in the root of the repo. You can interact with numbers in that config file, and as long as `localnet_up` is running, it will automatically scale the network within seconds.

### Stopping and cleaning up the resources

```
make localnet_down
```

The command stops LocalNet and cleans up all the resources, including postgres database.

### Interaction with the LocalNet

As the workloads run in Kubernetes, you can see and modify any resources on your local kubernetes by a tool of your choice (k9s, Lens, VSCode extension, etc.) - just be mindful that tilt will change manifests back eventually.

We provide some usefult make targets:

Open a shell in the pod that has `client` cli available. It gets updated automatically whenever the code changes:
```
make localnet_shell
```

Open a `client debug` cli. It allows to interact with blockchain, e.g. change pace maker mode, reset to genesis, etc. It gets updated automatically whenever the code changes (though you would need to stop/start the binary to execute the new code):
```
make localnet_client_debug
```

### How does it work?

`tilt` reads the `Tiltfile` in the root of the project, where configuration of LocalNet is provided, and starts the services defined there. `Tiltfile` is written in Starlark, which is a dialect of Python.

Kubernetes manifests `tilt` submits to the Kubernetes cluster can be found in [build/localnet directory](../..//build/localnet):
- [dependencies](../../build/localnet/dependencies/) - a helm chart that installs all necessary dependencies to run and observe LocalNet - postgresql, prometheus, grafana, etc.
- 4 validators. The validator binary that runs inside of the container gets updated automatically and process restarted on each code change.
- v1 cli client - this is a binary that can be used to perform operations on testnet, e.g. you can run `make localnet_client_debug` to execute commands such as `ResetToGenesis`, or `TogglePacemakerMode`. This binary is also automatically updated when you make changes to the codebase.

Tilt continuously monitors files on local filesystem, and it rebuilds the binary and distributes it to the pods on every code change. This allows developers to iterate on the code and see the changes immediately.

### Troubleshooting

Sometimes developers might experience issues with running LocalNet on Kubernetes. Issues might be related to the fact different developers run different clusters/OS/environments, and sometimes our laptops can go to sleep and that might not play well with virtual machines, postgres or pocket client.

- Check tilt web UI by pressing a space in the terminal where you started tilt or going to [this page](http://localhost:10350/) in your browser. If you see any errors, you can click "Trigger Update" on a resource that has issues to restart the service or retry a command.
- If triggering an update didn't help, try to run `make localnet_down` and then `make localnet_up` again. This will clean up most of the resources and start the localnet from scratch.
- If `make localnet_down` didn't help, we suggest to rebuild local kubernetes cluster using the tool you're managing your cluster with - it could be Docker Desktop, Rancher Desktop, k3s, k3d, minikube, etc.
- Open an issue in this repo if you're still experiencing issues with running localnet using this guide.

### How to change configuration files

Currently, we provide [a config file](../../build/localnet/configs.yaml) that is shared between all validators and a pocket client. We make use of `pocket` client feature that allows us to override configuration via environment variables. You can look at [one of the validators](../../build/localnet/v1-validator1.yaml) as a reference.
