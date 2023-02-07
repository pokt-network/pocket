# Iteration 9 Demo: LocalNet k8s & Logging Infrastructure <!-- omit in toc -->

_tl;dr Run a LocalNet using k8s, monitor your cluster, develop on top of it, scale it and search logs in Grafana._

## Table of Contents <!-- omit in toc -->

- [Motivation](#motivation)
  - [Demo Goals](#demo-goals)
  - [Why?](#why)
  - [Demo Non-Goals](#demo-non-goals)
- [Demo](#demo)
  - [Pre-Requisites](#pre-requisites)
  - [Instruction](#instruction)
- [Next Steps](#next-steps)

## Motivation

### Demo Goals

1. Deploy a V1 LocalNet using k8s
2. Present infrastructure that can be used to deploy remote clusters
3. Demonstrate tooling that can be used to scale a cluster
4. Demo a mature logging framework integrated with Grafana

### Why?

1. Make it easier to deploy a DevNet and TestNet in the future
2. Collect telemetry and gain visibility into node operations
3. Enable stress/chaos testing in the future
4. Provide tooling & infrastructure for both PNI and external node runners
5. Streamline development & debugging

### Demo Non-Goals

1. Pocket-specific Utility
2. Sending on-chain transactions

## Demo

### Pre-Requisites

1. Basic env setup based on instructions at [docs/development/README.md](../development/README.md)

2. k8s Localnet env setup based on instructions at [build/localnet/README.md](../../build/localnet/README.md)

3. [Optional] Learn about the logging library at [logger/docs/README.md](../../logger/docs/README.md)

### Instruction

1. Start up a k8s LocalNet and press `Space` when prompted

```bash
make localnet_up
```

2. Confirm that the `Validator`s are present and select the recommended UI at the top left; or visit [http://localhost:10350/r/(all)](<http://localhost:10350/r/(all)>) directly

![Tilt UI](https://user-images.githubusercontent.com/1892194/217139866-f3f4e1e1-5ad7-429e-b26b-15953b59cd49.png)

3. Open up the debug client and select `TriggerNextView` a few times when prompted to increase the chain height

```bash
make localnet_client_debug
```

4. Use the `Tilt UI` to select a `Validator` and inspect its logs

![Validator Logs](https://user-images.githubusercontent.com/1892194/217139864-dbbf15f4-7edd-4089-bd95-2d608bc981b6.png)

5. Commit a few blocks via CLI

6. Search for `Line contains` ”Committing block”

![Committed Block](https://user-images.githubusercontent.com/1892194/217139860-369b6f9b-7827-49e2-9a62-e11f801fa931.png)

7. Rather than viewing in Tilt, you can use Grafana to view & filter them as well. Go to [http://localhost:42000](http://localhost:42000/) click on `explore`; or visit [this link](http://localhost:42000/explore?orgId=1&left=%7B%22datasource%22:%22P8E80F9AEF21F6940%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22datasource%22:%7B%22type%22:%22loki%22,%22uid%22:%22P8E80F9AEF21F6940%22%7D%7D%5D,%22range%22:%7B%22from%22:%22now-1h%22,%22to%22:%22now%22%7D%7D) directly.

![Grafana](https://user-images.githubusercontent.com/1892194/217139859-f3ad4d0b-1204-4da8-a73b-c6f079d183a1.png)

8. You can filter by label:

![Grafana UI](https://user-images.githubusercontent.com/1892194/217139856-f6bae565-f52f-4b51-8547-b95e3b5dcf3b.png)

9. Or create more complex filter using by parsing `{{.log}}` as shown below:

![Complex Filter](https://user-images.githubusercontent.com/1892194/217146156-8e4556b9-7ea6-4135-87aa-66f33f919d1c.png)

10. To scale the number of validator, change `count` to `10` in `./localnet_config.yaml` and visit the Tilt UI:

![Validator 10](https://user-images.githubusercontent.com/1892194/217139832-f619a317-0993-4f14-99a5-67d371d124f2.png)

11. Bonus: Verify that changing the code retriggers the cluster to rebuild the image: Hot-Reloading out of the box!

## Next Steps

1. Improved log parsing
2. Automate validator staking when scaling
3. Discover peers (nodes or actors) when connected to the network
4. Sync new nodes to the latest height
