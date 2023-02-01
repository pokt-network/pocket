# Kubernetes LocalNet <!-- omit in toc -->

This guide shows how to deploy a LocalNet using [pocket-operator](https://github.com/pokt-network/pocket-operator).

- [Dependencies](#dependencies)
  - [Enabling Kubernetes](#enabling-kubernetes)
- [LocalNet](#localnet)
  - [Starting LocalNet](#starting-localnet)
  - [Viewing Logs](#viewing-logs)
  - [Terminal Logs](#terminal-logs)
  - [Tilt Web UI Logs](#tilt-web-ui-logs)
  - [Scaling Actors](#scaling-actors)
  - [Stopping \& Clean Resources](#stopping--clean-resources)
  - [Interacting w/ LocalNet](#interacting-w-localnet)
    - [Make Targets](#make-targets)
- [How does it work?](#how-does-it-work)
- [Troubleshooting](#troubleshooting)
  - [Why?](#why)
  - [Force Trigger an Update](#force-trigger-an-update)
  - [Force Restart](#force-restart)
  - [Full Cleanup](#full-cleanup)
  - [Docker Desktop](#docker-desktop)
- [How to change configuration files](#how-to-change-configuration-files)
- [Code Structure](#code-structure)

## Dependencies

1. [tilt](https://docs.tilt.dev/install.html)
   - Note: automatically installed when running `make install_cli_deps`
2. `Kubernetes cluster`: [installation options](https://docs.tilt.dev/choosing_clusters.html)
3. `kubectl`: CLI is required and should be configured to access the cluster. This should happen automatically if using Docker Desktop, Rancher Desktop, k3s, k3d, minikube, etc.
4. `helm`: required to template the yaml manifests for the dependencies (e.g. postgres, grafana). Installation instructions available [here](https://helm.sh/docs/intro/install).

### Enabling Kubernetes

You may need to manually enable Kubernetes if using Docker desktop:

![Docker desktop kubernetes](https://user-images.githubusercontent.com/1892194/216165581-1372e2b8-c630-4211-8ced-5ec59b129330.png)

## LocalNet

### Starting LocalNet

```bash
make localnet_up
```

### Viewing Logs

The developer can then view logs either from a browser or terminal.

### Terminal Logs

- `make localnet_logs_validators` - shows prior logs
- `make localnet_logs_validators_follow` - shows prior logs and follows the new log lines as validators do their work

### Tilt Web UI Logs

- Pressing `space` in terminal where you started `tilt`
- Go to [localhost:10350](http://localhost:10350/)

![tilt UI](https://user-images.githubusercontent.com/1892194/216165833-b9e5a98c-87a8-4355-87c9-0420a8a598bf.png)

### Scaling Actors

When starting a k8s LocalNet, `localnet_config.yaml` is generated (with default configs) in the root of the repo if doesn't already exist.

The config file can be modified to scale the number of actors up/down. As long as `localnet_up` is running, the changes should be automatically applied within seconds.

### Stopping & Clean Resources

```bash
make localnet_down
```

The command stops LocalNet and cleans up all the resources, including the postgres database.

### Interacting w/ LocalNet

As the workloads run in Kubernetes, you can see and modify any resources on your local kubernetes by a tool of your choice (k9s, Lens, VSCode extension, etc), but note that Tilt will change manifests back eventually.

#### Make Targets

Open a shell in the pod that has `client` cli available. It gets updated automatically whenever the code changes:

```bash
make localnet_shell
```

Open a `client debug` cli. It allows to interact with blockchain, e.g. change pace maker mode, reset to genesis, etc. It gets updated automatically whenever the code changes (though you would need to stop/start the binary to execute the new code):

```bash
make localnet_client_debug
```

## How does it work?

[tilt](https://tilt.dev/) reads the [`Tiltfile`](../../Tiltfile), where LocalNet configs are specified. `Tiltfile` is written in [Starlark](https://github.com/bazelbuild/starlark), a dialect of Python.

The k8s manifests that `tilt` submits to the cluster can be found in [this directory](./):

- **[dependencies](./dependencies/)**: a helm chart that installs all necessary dependencies to run and observe LocalNet (postgresql, prometheus, grafana, etc).
- **[4 Validators](./v1-validator-template.sh)**: The validator binary that runs inside of the container gets updated automatically and process restarted on each code change (i.e. hot reloads).
- **[V1 CLI client](./cli-client.yaml)**: This binary that can be used to perform debug operations. Run `make localnet_client_debug` to execute commands such as `ResetToGenesis` or `TogglePacemakerMode`. This binary is also automatically updated when you make changes to the codebase.

Tilt continuously monitors files on local filesystem, and it rebuilds the binary and distributes it to the pods on every code change. This allows developers to iterate on the code and see the changes immediately.

## Troubleshooting

### Why?

Developers might experience issues with running LocalNet on Kubernetes.

Issues might be related to the fact different developers run different clusters/OS/environments.

Machines going to sleep and that might not play well with virtual machines, postgres or pocket client.

### Force Trigger an Update

Visit the tilt web UI by pressing `space` in the shell where you started tilt or by visiting [this webpage](http://localhost:10350/).

If you see any errors, you can click `Trigger Update` on a resource that has issues to restart the service or retry a command.

### Force Restart

If force triggering an update didn't work, try the following:

1. `make localnet_down`
2. `make localnet_up`

### Full Cleanup

If a force restart didn't help, try rebuilding local kubernetes cluster using the tool you're managing your cluster with (e.g. Docker Desktop, Rancher Desktop, k3s, k3d, minikube, etc).

### Docker Desktop

# TODO_IN_THIS_COMMIT

## How to change configuration files

Currently, we provide [a config file](./configs.yaml) that is shared between all validators and a pocket client. We make use of `pocket` client feature that allows us to override configuration via environment variables. You can look at [one of the validators](../../build/localnet/v1-validator1.yaml) as a reference.

## Code Structure

```bash
build/localnet # TODO_IN_THIS_COMMIT
├── README.md # TODO_IN_THIS_COMMIT
├── cli-client.yaml # TODO_IN_THIS_COMMIT
├── configs.yaml # TODO_IN_THIS_COMMIT
├── dependencies # TODO_IN_THIS_COMMIT
│   ├── Chart.yaml # TODO_IN_THIS_COMMIT
│   ├── dashboards # TODO_IN_THIS_COMMIT
│   │   ├── README.md # TODO_IN_THIS_COMMIT
│   │   └── raintree-telemetry-graphs.json # TODO_IN_THIS_COMMIT
│   ├── requirements.yaml # TODO_IN_THIS_COMMIT
│   ├── templates # TODO_IN_THIS_COMMIT
│   │   ├── _helpers.tpl # TODO_IN_THIS_COMMIT
│   │   ├── dashboards.yml # TODO_IN_THIS_COMMIT
│   │   └── datasources.yml # TODO_IN_THIS_COMMIT
│   └── values.yaml # TODO_IN_THIS_COMMIT
├── network.yaml # TODO_IN_THIS_COMMIT
├── private-keys.yaml # TODO_IN_THIS_COMMIT
├── v1-validator-template.sh # TODO_IN_THIS_COMMIT
└── v1-validator-template.yaml # TODO_IN_THIS_COMMIT
```
