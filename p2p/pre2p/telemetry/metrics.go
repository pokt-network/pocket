package pre2p_telemetry

const (
	// How to define metrics in this file?
	//
	// Must define metrics using following template: MetricName_MetricType_Attribute
	// e.g: type= timeseries, name=nodes_alive_counter, attribute=name
	// will result in: NODES_COUNT_COUNTER_TIME_SERIES_NAME="nodes_alive_counter"
	// you can also add: NODES_COUNT_COUNTER_TIME_SERIES_DESCRIPTION="for your description"

	// Time Series Metrics
	P2P_NODE_STARTED_TIMESERIES_METRIC_NAME        = "p2p_nodes_started_counter"
	P2P_NODE_STARTED_TIMESERIES_METRIC_DESCRIPTION = "the counter to track the number of nodes online"

	// Event Metrics
	P2P_EVENT_METRICS_NAMESPACE = "event_metrics_namespace_p2p"

	BROADCAST_MESSAGE_REDUNDANCY_PER_BLOCK_EVENT_METRIC_NAME = "broadcast_message_redundancy_per_block_event_metric"
	RAINTREE_MESSAGE_EVENT_METRIC_NAME                       = "raintree_message_event_metric"
)
