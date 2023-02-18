# Kubernetes LocalNet <!-- omit in toc -->

This guide shows how to deploy a LocalNet using [pocket-operator](https://github.com/pokt-network/pocket-operator).

- [Dependencies](#dependencies)
  - [Choosing Kubernetes Distribution](#choosing-kubernetes-distribution)
  - [Enabling Kubernetes For Docker Desktop](#enabling-kubernetes-for-docker-desktop)
- [LocalNet](#localnet)
  - [Starting LocalNet](#starting-localnet)
  - [Viewing Logs](#viewing-logs)
  - [Terminal Logs](#terminal-logs)
  - [Tilt Web UI Logs](#tilt-web-ui-logs)
  - [Scaling Actors](#scaling-actors)
  - [Stopping \& Clean Resources](#stopping--clean-resources)
  - [Interacting w/ LocalNet](#interacting-w-localnet)
    - [Make Targets](#make-targets)
  - [Addresses and keys on LocalNet](#addresses-and-keys-on-localnet)
- [How to change configuration files](#how-to-change-configuration-files)
- [How does it work?](#how-does-it-work)
- [Troubleshooting](#troubleshooting)
  - [Why?](#why)
  - [Force Trigger an Update](#force-trigger-an-update)
  - [Force Restart](#force-restart)
  - [Full Cleanup](#full-cleanup)
- [Code Structure](#code-structure)

## Dependencies

All necessary dependencies, except Kubernetes cluster, are installed automatically when running `make install_cli_deps`. The following dependencies are required:

1. [tilt](https://docs.tilt.dev/install.html)
2. `Kubernetes cluster`: refer to [Choosing Kubernetes Distribution](#choosing-kubernetes-distribution) section for more details.
3. `kubectl`: CLI is required and should be configured to access the cluster. This should happen automatically if using Docker Desktop, Rancher Desktop, k3s, k3d, minikube, etc.
4. [helm](https://helm.sh/docs/intro/install): required to template the yaml manifests for the dependencies (e.g. postgres, grafana). Installation instructions available .

### Choosing Kubernetes Distribution

While any Kubernetes distribution should work, we verified that LocalNet works on:

- [Rancher Desktop](https://rancherdesktop.io/), which is GUI powered by a popular distribution `k3s`.
- [kind](https://kind.sigs.k8s.io/) - official Kubernetes distribution that runs inside docker containers.

Here is a list of alternative set ups that should work: https://docs.tilt.dev/choosing_clusters.html

### Enabling Kubernetes For Docker Desktop

You may need to manually enable Kubernetes if using Docker Desktop:

![Docker desktop kubernetes](https://user-images.githubusercontent.com/1892194/216165581-1372e2b8-c630-4211-8ced-5ec59b129330.png)

## LocalNet

### Starting LocalNet

```bash
make localnet_up
```

This action will create a file called `localnet_config.yaml` in the root of the repo if it doesn't exist. The default configuration can be found in [Tiltfile](Tiltfile#L11).

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

[Watch demo](https://user-images.githubusercontent.com/4950477/216490690-8ac4c16a-25e1-4202-b2e5-03215030c82c.mp4)

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

### Addresses and keys on LocalNet

You can find private keys and addresses for all actors in the [private-keys.yaml](./manifests/private-keys.yaml) file. They have been pre-generated and follow a specific pattern - they start with pre-determined numbers for easier troubleshooting and debugging.

Addresses begin with `YYYXX` number, where `YYY` is a number of an actor and `XX` is [a type of actor](../../shared/core/types/proto/actor.proto#L7).

The current mapping for `XX` is:

- `01` - Application
- `02` - Servicer
- `03` - Fisherman
- `04` - Validator

For example:

- `420043b854e78f2d5f03895bba9ef16972913320` is a validator #420.
- `66603bc4082281b7de23001ffd237da62c66a839` is a fisherperson #666.
- `0010297b55fc9278e4be4f1bcfe52bf9bd0443f8` is a servicer (a servicer) #001.
- `314019dbb7faf8390c1f0cf4976ef1215c90b7e4` is an application #314.

## How to change configuration files

Currently, we provide [a config file](./manifests/configs.yaml) that is shared between all validators and a pocket client. We make use of `pocket` client feature that allows us to override configuration via environment variables. You can check a [validator template](./templates/v1-validator-template.yaml.tpl) as a reference.

## How does it work?

[tilt](https://tilt.dev/) reads the [`Tiltfile`](./Tiltfile), where LocalNet configs are specified. `Tiltfile` is written in [Starlark](https://github.com/bazelbuild/starlark), a dialect of Python.

The k8s manifests that `tilt` submits to the cluster can be found in [this directory](./). Please refer to [code structure](#code-structure) for more details where different parts are located.

Tilt continuously monitors files on local filesystem in [specific directories](Tiltfile#L27), and it rebuilds the binary and distributes it to the pods on every code change. This allows developers to iterate on the code and see the changes immediately (i.e. hot-reloading).

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

## Code Structure

```bash
build/localnet
├── README.md # This file
├── Tiltfile # File outlining tilt process
├── dependencies # Helm charts that install all the dependencies needed to run and observe LocalNet
│   ├── Chart.yaml # Main file of the helm chart, contains metadata
│   ├── dashboards # Directory with all the dashboards that are automatically imported to Grafana
│   │   ├── README.md # README file with instructions on how to add a new dashboard
│   │   └── raintree-telemetry-graphs.json # Raintree Telemetry dashboard
│   ├── requirements.yaml # Specifies dependencies of the chart, this allows us to install all the dependencies with a single command
│   ├── templates # Additional Kubernetes manifests that we need to connect different dependencies together
│   │   ├── _helpers.tpl
│   │   ├── dashboards.yml
│   │   └── datasources.yml
│   └── values.yaml # Configuration values that override the default values of the dependencies, this allows us to connect dependencies together and make them available to our LocalNet services
├── manifests # Static YAML Kubernetes manifests that are consumed by `tilt`
│   ├── cli-client.yaml # Pod that has the latest binary of the pocket client. Makefile targets run CLI in this pod.
│   ├── configs.yaml # Location of the config files (default configs for all validators and a genesis file) that are shared across all actors
│   ├── network.yaml # Networking configuration that is shared between different actors, currently a Service that points to all validators
│   └── private-keys.yaml # Pre-generated private keys with a semantic format for easier development
└── templates # Templates for Kubernetes manifests that are consumed by `tilt`
    ├── v1-validator-template.sh # Shell script that generates Kubenetes manifests for validators, consumed by `tilt`
    └── v1-validator-template.yaml.tpl # Template for a single validator, consumed by `v1-validator-template.sh`
```
