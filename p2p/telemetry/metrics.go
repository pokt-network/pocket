package p2p_telemetry

// TODO (Team) move to telemetry module

const (
	// Time Series Metrics
	P2P_NODE_STARTED_TIMESERIES_METRIC_NAME        = "p2p_nodes_started_counter"
	P2P_NODE_STARTED_TIMESERIES_METRIC_DESCRIPTION = "the counter to track the number of nodes online"

	// Event Metrics
	P2P_EVENT_METRICS_NAMESPACE = "event_metrics_namespace_p2p"

	BROADCAST_MESSAGE_REDUNDANCY_PER_BLOCK_EVENT_METRIC_NAME = "broadcast_message_redundancy_per_block_event_metric"
	RAINTREE_MESSAGE_EVENT_METRIC_NAME                       = "raintree_message_event_metric"

	// Attributes
	RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL   = "send"
	RAINTREE_MESSAGE_EVENT_METRIC_HEIGHT_LABEL = "height"
	RAINTREE_MESSAGE_EVENT_METRIC_HASH_LABEL   = "hash"
)
