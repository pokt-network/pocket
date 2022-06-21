## Welcome to the Telemetry Module

This module is used to collect telemetry data from the node.

At the moment, we are using two types of metrics:

- Time series metrics: these are metrics that are collected periodically and are stored in a database.
- Event series metrics: these are metrics that are collected when an event occurs and are stored in a database.


At the moment, we are using for each of the two types of metrics respectively:

- Prometheus
- Plain Logs. (_Might be subsituted in the future with an events database_)


## Usage

### Time Series Metrics

If you aren't familiar with time series metrics that Prometheus offers, please check out [Prometheus Metrics](https://prometheus.io/docs/concepts/metric_types/)


We are primarily using:

- Gauges

We use Gauges to keep track of:

- Blockheight
- Nodes Online

#### How to use the time series metrics in your code

In your module, make sure you have access to the bus, then use the metrics you need as follows:
```go

timeseriesTelemetry := module.GetBus().GetTelemetryModule().GetTimeSeriesAgent()
// explore the methods you can use in shared/modules/telemetry_module.go

// To increment a gauge
timeseriesTelemetry.IncGauge("gauge_name", 1)
// etc...
```

### Event Metrics

In the current implementation, we are recording events through logs. For each unique witnessing of a given event, we will emit a log line with specific formatting to stdout.

Using Loki and Grafana, we parse the logs and generate the desired graphs.

In your module, make sure you have access to the bus, then use the metrics you need as follows:
```go

timeseriesTelemetry := module.GetBus().GetTelemetryModule().GetTimeSeriesAgent()
// explore the methods you can use in shared/modules/telemetry_module.go

eventMetricsTelemetry.EmitEvent(
    "namespace",
    "event_name",
    ... // any other fields you want to include
)
```

#### Consuming logs on loki

To test this out, track an event in your code, and then go to your Grafana's local setup's link, and to the the explore page.

Run the following LogQL query:

```
{host="desktop-docker"} |= "[EVENT] your_namespace your_event" | pattern `<datetime> <_> <time> <type> <event_name> <any> <aditional> <whitespaced> <items>` | logfmt
```

You should see a log stream coming out, click a line to explore how you've used the `pattern` keyword in LogQL to parse the log line. Now you can reference your parsed fields as you like. Example:

Counting how many events we've seen by type over 5m:
```
sum by (type) (count_over_time(
    {host="desktop-docker"}
    |= "[EVENT] your_namespace"
    | pattern `<datetime> <_> <time> <type> <event_name> <any> <aditional> <whitespaced> <items>`[5m]
))
```
Counting how many events of a certain type have we seen over 5m:
```
sum (count_over_time(
    {host="desktop-docker"}
    |= "[EVENT] your_namespace your_event"
    | pattern `<datetime> <_> <time> <type> <event_name> <any> <aditional> <whitespaced> <items>`[5m]
))
```

#### Node Configuration

It is necessary to provide a telemetry configuration to your node:

```json
 "use_telemetry": true,
 "telemetry": {
    "address": "0.0.0.0:9000",
    "endpoint": "/metrics"
  }
```

`use_telemetry`: is a boolean json entry defined at the root of the document that tells the node whether to use the telemetry module or use a NOOP version.
`address`: is the prometheus server's address that the telemetry module will listen on.
`endpoint`: the scraping endpoint that prometheus exposes through the telemetry module.
