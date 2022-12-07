# LocalNet on Kubernetes

We are developing our own Kubernetes operator to manage v1 workloads both internally and for the. While the operator is still in development, we can already utilize it to run our local networks locally on Kubernetes. This guide will show you how to do that.

- [LocalNet on Kubernetes](#localnet-on-kubernetes)
    - [Dependencies](#dependencies)
    - [Running the localnet](#running-the-localnet)
    - [Cleaning up](#cleaning-up)
    - [How does it work?](#how-does-it-work)
    - [Troubleshooting](#troubleshooting)


### Dependencies

* [tilt](https://docs.tilt.dev/install.html)
* Kubernetes cluster ([different options available](https://docs.tilt.dev/choosing_clusters.html)), please make sure to run 1.23 version or older, until [this issue is resolved](https://github.com/zalando/postgres-operator/issues/2098) - otherwise Postgres DB is not going to be provisioned.
  * `kubectl` is also required to be installed and configured to access the cluster, but that should happen automatically when you install kubernetes cluster locally.
* pocket kubernetes operator codebase in `../pocket-operator` directory, relative to the pocket v1 codebase.
  * Having this codebase available on your computer allows you to iterate/change the operator code while running the localnet.

### Running the localnet

Start the LocalNet with `make localnet_up` command. This will start tilt. Tilt will prompt to press `space` to open a browser window with the UI where you can see the logs of all services running (including the operator, observability stack, etc.). You can also open the UI by going to `http://localhost:10350/` in your browser.

### Cleaning up

`make localnet_down` will stop the localnet and clean up all the resources, except the postgres operator (in case you have other databases provisioned with it).

### How does it work?

### Troubleshooting

