## 1. Configuration

Logger module has the following configuration options:

```json
{
  "logger": {
    "level": "debug",
    "format": "pretty"
  },
}
```

* `level` - the level of logging. Can be one of `debug`, `info`, `warn`, `error`, `fatal`, `panic`
* `format` - the format of the logs. Can be one of `pretty`, `json`

When utilizing additional wrappers to run processes, beware that they can change stdout output. For example, `reflex` we use for hot reloading modifies the log lines. To avoid this, we use the `--decoration` flag.


## 2. Log Types

### 2.1 Levels

The developer needs to provide the logging level for each log output.

* When logging debug: `logger.Global.Logger.Debug().Msg(msg)`
* When logging error with `err`: `logger.Global.Logger.Error().Err(err).Msg(msg)`
* When logging error without `err`: `logger.Global.Logger.Error().Msg(msg)`
* When logging a fatal error: `logger.Global.Fatal().Err(err).Msg(msg)`

### 2.2. Fields

We encourage to provide additional context to the log message by using fields. Please be consistent with the field names, e.g. utilizing "height" key for the height of a block instead of "h" or "block_height" because it makes it easier to search for logs related to that context.

For example:
```golang
logger.Global.Logger.Debug().Uint64("height", height).Msg("Block committed")
```

Refer to the [zerolog documentation](https://github.com/rs/zerolog#field-types) for more information on the available field types.

Fields also can be provided utilizing a map:

```golang
fields := map[string]interface{}{
    "height": height,
    "hash": hash,
}
logger.Global.Logger.Debug.Fields(fields).Msg("Block committed")
```

## 3. Global Logging

When not logging inside a module, use the global logger.

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


## 4. Module Logging

As each module has its own logger, please utilize it instead of the global logger.

```golang
type sweetModule struct {
    logger    modules.Logger
}

func (m *sweetModule) DoSomething() {
    m.Fatal().Msg("Something is fishy!")
    ...
}
```

### 4.1 Logger Initialization

`Global` logger is always available from the `logger` package.

Each module has its own logger to provide an additional layer of granularity.
Please initiate loggers in the `Start` method of the module, like this:

```golang
type sweetModule struct {
    logger    modules.Logger
}

func (m *sweetModule) Start() error {
    m.logger = logger.Global.CreateLoggerForModule(u.GetModuleName())
    ...
}
```

## 5. Accessing Logs

Logs are written to stdout. In our LocalNet we use Loki to capture log output and query logs using [LogQL](https://grafana.com/docs/loki/latest/logql/) syntax. We also use Grafana to visualize the logs.

### 5.1. Grafana

If you run the LocalNet using the `make localnet_up` command, you can access Grafana at http://localhost:42000.

#### 5.2. Example Queries

We will populate this section with useful queries as we go.
