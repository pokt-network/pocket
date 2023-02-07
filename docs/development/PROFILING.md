# Profiling

This guide is not meant to replace any official documentation, but rather to provide a quick overview of the steps required to profile the node.

## Profiling endpoint feature flag

We pre-configured the node to expose a web server on port `6060` for the collection of profiling data.

Because of the performance overhead, it is disabled by default.
In order to enable it, you can simply add the environment variable `PPROF_ENABLED=true` to the node.

You can do that in several ways, the most obvious spot would be the `docker-compose.yml` file, in the `environment` section of the `node1.consensus` service.

```yaml
environment:
# Uncomment to enable the pprof server
#  - PPROF_ENABLED=true
```

### Example

Let's assume that you want to profile the memory usage of the node during the execution of the `ResetToGenesis` command.

These are the necessary steps:

1. Enable the pprof server in the `docker-compose.yml` file (see above)
2. Collect a "baseline" heap profile of the node by running the following command:

```bash
curl http://localhost:6060/debug/pprof/heap > baseline.heap
```

3. Run the `ResetToGenesis` command from the debug CLI
4. Collect another heap profile of the node by running the following command:

```bash
curl http://localhost:6060/debug/pprof/heap > after_reset.heap
```

5. Compare the two profiles using the `pprof` tool:

```bash
go tool pprof -base=baseline.heap after_reset.heap
```

6. From the `pprof` prompt, you can run the `top` command to see the top 10 (can be any number) memory consumers:

```bash
(pprof) top10
```

7. By repeating the steps above, and comparing different heap profiles, you can identify the memory consumers that are causing the memory usage to grow. It could be a memory leak, or just a normal behavior.

### Further reading

- PProf: https://go.dev/blog/pprof
- Memory leaking scenarios: https://go101.org/article/memory-leaking.html

<!-- GITHUB_WIKI: guides/development/profiling -->
