# Logger <!-- omit in toc -->

- [Configuration](#configuration)
- [Log Types](#log-types)
  - [Levels](#levels)
  - [Fields](#fields)
    - [Int Field](#int-field)
    - [String Field](#string-field)
    - [Map Field](#map-field)
- [Global Logging](#global-logging)
- [Module Logging](#module-logging)
  - [Logger Initialization](#logger-initialization)
- [Submodule / Subcontext Logging](#submodule--subcontext-logging)
- [Accessing Logs](#accessing-logs)
  - [Grafana](#grafana)
  - [Example Queries](#example-queries)

## Configuration

The logger module has the following configuration options found [here](./runtime/confg/../../../../runtime/configs/proto/logger_config.proto):

```json
{
  "logger": {
    "level": "debug",
    "format": "pretty"
  }
}
```

- `level`: log level; one of `debug`, `info`, `warn`, `error`, `fatal`, `panic`
- `format`: log format; one of `pretty`, `json`

NOTE: Additional process wrapper ma change stdout output. For example, `reflex`, used for hot reloading, modifies the log lines. This can be avoided by using the `--decoration` flag.

## Log Types

### Levels

The developer needs to provide the logging level for each log message:

- Debug: `logger.Global.Logger.Debug().Msg(msg)`
- Error with `err`: `logger.Global.Logger.Error().Err(err).Msg(msg)`
- Error without `err`: `logger.Global.Logger.Error().Msg(msg)`
- Fatal:: `logger.Global.Fatal().Err(err).Msg(msg)`

### Fields

Metadata can, and should, be attached to each log level. Using the same key throughout makes the logs easier to parse.

Refer to the [zerolog documentation](https://github.com/rs/zerolog#field-types) for more information on the available field types.

#### Int Field

For example, a single int field can be added like so:

```golang
logger.Global.Logger.Debug().Uint64("height", height).Msg("Block committed")
```

#### String Field

A single string field can be added like so:

```golang
logger.Global.Logger.Debug().String("hash", hash).Msg("Block committed")
```

#### Map Field

Multiple fields can be provided using a map:

```golang
fields := map[string]interface{}{
    "height": height,
    "hash": hash,
}

logger.Global.Logger.Debug.Fields(fields).Msg("Block committed")
```

## Global Logging

The global logger should be used when logging outside a module:

```golang
import (
    ...
    "github.com/pokt-network/pocket/logger"
    ...
)

func DoSomething() {
    logger.Global.Fatal().Msg("Oops, something went wrong!")
    ...
}
```

## Module Logging

Each module should have its own logger to appropriately namespace the logs.

```golang
type sweetModule struct {
    logger    *modules.Logger
}

func (m *sweetModule) DoSomething() {
    m.logger.Fatal().Msg("Something is fishy!")
    ...
}
```

### Logger Initialization

`Global` logger is always available from the `logger` package.

Each module has its own logger to provide an additional layer of granularity.
Please initiate loggers in the `Start` method of the module, like this:

```golang
type sweetModule struct {
    logger    *modules.Logger
}

func (m *sweetModule) Start() error {
    m.logger = logger.Global.CreateLoggerForModule(u.GetModuleName())
    ...
}
```

## Submodule / Subcontext Logging

A common helpful practice is to create a logger that can be easily filtered for within a specific context, such as a specific submodule, a function or a code path.

```golang
m.logger.With().Str("source", "contextName").Logger(),
```

For example:

```golang

func (m *Module) fooFunc() {
  fooLogger := m.logger.With().Str("source", "fooFunc").Logger(),
  // use fooLogger here
}
```

## Accessing Logs

Logs are written to stdout. In LocalNet, Loki is used to capture log output. Logs can then be queried using [LogQL](https://grafana.com/docs/loki/latest/logql/) syntax. Grafana can be used to visualize the logs.

### Grafana

When running LocalNet via `make localnet_up`, Grafana can be accessed at [localhost:42000](https://localhost:42000).

### Example Queries

DOCUMENT: Add common query examples.

<!-- GITHUB_WIKI: logger/readme -->
