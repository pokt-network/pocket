package consensus_telemetry

const (
	// How to define metrics in this file?
	//
	// Must define metrics using following template: MetricName_MetricType_Attribute
	// e.g: type= timeseries, name=nodes_alive_counter, attribute=name
	// will result in: NODES_COUNT_COUNTER_TIME_SERIES_NAME="nodes_alive_counter"
	// you can also add: NODES_COUNT_COUNTER_TIME_SERIES_DESCRIPTION="for your description"

	// Time Series Metrics

	// DISCUSS(team): Dmitry has expressed multiple times that this should be a gauge instead of a counter
	// I do not remember the reasoning behind this, but it would be worth revisiting sometime in the future
	CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME        = "consensus_blockchain_height_counter"
	CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_DESCRIPTION = "the counter to track the height of the blockchain"

	// Event Metrics
	CONSENSUS_EVENT_METRICS_NAMESPACE = "event_metrics_namespace_consensus"

	HOTPOKT_MESSAGE_EVENT_METRIC_NAME                         = "hotpokt_message_event_metric"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT                 = "HEIGHT"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_NEW_ROUND         = "STEP_NEW_ROUND"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_PREPARE           = "STEP_PREPARE"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_PRECOMMIT         = "STEP_PRECOMMIT"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_COMMIT            = "STEP_COMMIT"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_DECIDE            = "STEP_DECIDE"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER  = "VALIDATOR_TYPE_LEADER"
	HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_REPLICA = "VALIDATOR_TYPE_REPLICA"
)
