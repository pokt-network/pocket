package e2e_tests

// node4.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node4.consensus  | [00] [NODE][4] [DEBUG] Triggering next view...
// node4.consensus  | [00] [NODE][4] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_NEWROUND, 0)! Reason: manual trigger
// node4.consensus  | [00] [NODE][4] Broadcasting message for HOTSTUFF_STEP_NEWROUND step
// node3.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node3.consensus  | [00] [NODE][3] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node1.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node1.consensus  | [00] [NODE][1] [DEBUG] Triggering next view...
// node1.consensus  | [00] [NODE][1] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_NEWROUND, 0)! Reason: manual trigger
// node1.consensus  | [00] [NODE][1] Broadcasting message for HOTSTUFF_STEP_NEWROUND step
// node4.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node1.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node1.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node1.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node3.consensus  | [00] [NODE][3] pacemaker catching up the node's (height, step, round) FROM (12, HOTSTUFF_STEP_NEWROUND, 0) TO (12, HOTSTUFF_STEP_NEWROUND, 1)
// node1.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node4.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node3.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node4.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node4.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node4.consensus  | [00] [NODE][4] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node4.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node2.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node1.consensus  | [00] [NODE][1] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node3.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node3.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node2.consensus  | [00] [NODE][2] [DEBUG] Triggering next view...
// node2.consensus  | [00] [NODE][2] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_NEWROUND, 0)! Reason: manual trigger
// node1.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node2.consensus  | [00] [NODE][2] Broadcasting message for HOTSTUFF_STEP_NEWROUND step
// node2.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node2.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node2.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node2.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node2.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node2.consensus  | [00] [NODE][2] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node4.consensus  | [00] [REPLICA][4] 🙇🙇🙇 Elected leader for height/round 12/1: [2] (3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99) 🙇🙇🙇
// node4.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_REPLICA
// node3.consensus  | [00] [REPLICA][3] 🙇🙇🙇 Elected leader for height/round 12/1: [2] (3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99) 🙇🙇🙇
// node3.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_REPLICA
// node1.consensus  | [00] [REPLICA][1] 🙇🙇🙇 Elected leader for height/round 12/1: [2] (3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99) 🙇🙇🙇
// node2.consensus  | [00] [LEADER][2] 👑👑👑 I am the leader for height/round 12/1: [2] (3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99) 👑👑👑
// node1.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_REPLICA
// node2.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_LEADER
// node3.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_LEADER
// node4.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node2.consensus  | [00] [LEADER][2] Waiting for more HOTSTUFF_STEP_NEWROUND messages; byzantine optimistic threshold not met: (1 > 2.67?)
// node2.consensus  | [00] [LEADER][2] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node1.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node3.consensus  | [00] [REPLICA][3] Waiting for more HOTSTUFF_STEP_NEWROUND messages; byzantine optimistic threshold not met: (1 > 2.67?)
// node4.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_LEADER
// node2.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_LEADER
// node2.consensus  | [00] [LEADER][2] Waiting for more HOTSTUFF_STEP_NEWROUND messages; byzantine optimistic threshold not met: (2 > 2.67?)
// node1.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 1 VALIDATOR_TYPE_LEADER
// node3.consensus  | [00] [REPLICA][3] [DEBUG] Triggering next view...
// node4.consensus  | [00] [REPLICA][4] Waiting for more HOTSTUFF_STEP_NEWROUND messages; byzantine optimistic threshold not met: (1 > 2.67?)
// node2.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric height 12
// node1.consensus  | [00] [REPLICA][1] Waiting for more HOTSTUFF_STEP_NEWROUND messages; byzantine optimistic threshold not met: (1 > 2.67?)
// node1.consensus  | [00] [REPLICA][1] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node1.consensus  | [00] [REPLICA][1] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 2.
// node1.consensus  | [00] [REPLICA][1] pacemaker catching up the node's (height, step, round) FROM (12, HOTSTUFF_STEP_PREPARE, 1) TO (12, HOTSTUFF_STEP_NEWROUND, 2)
// node4.consensus  | [00] [REPLICA][4] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node3.consensus  | [00] [REPLICA][3] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_PREPARE, 1)! Reason: manual trigger
// node1.consensus  | [00] [REPLICA][1] 🙇🙇🙇 Elected leader for height/round 12/2: [3] (67eb3f0a50ae459fecf666be0e93176e92441317) 🙇🙇🙇
// node2.consensus  | [00] [LEADER][2] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 2.
// node4.consensus  | [00] [REPLICA][4] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 2.
// node3.consensus  | [00] About to release postgres context at height 12.
// node2.consensus  | [00] [LEADER][2] pacemaker catching up the node's (height, step, round) FROM (12, HOTSTUFF_STEP_NEWROUND, 1) TO (12, HOTSTUFF_STEP_NEWROUND, 2)
// node2.consensus  | [00] [REPLICA][2] 🙇🙇🙇 Elected leader for height/round 12/2: [3] (67eb3f0a50ae459fecf666be0e93176e92441317) 🙇🙇🙇
// node3.consensus  | [00] [NODE][3] Broadcasting message for HOTSTUFF_STEP_NEWROUND step
// node4.consensus  | [00] [REPLICA][4] pacemaker catching up the node's (height, step, round) FROM (12, HOTSTUFF_STEP_PREPARE, 1) TO (12, HOTSTUFF_STEP_NEWROUND, 2)
// node1.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 2 VALIDATOR_TYPE_REPLICA
// node1.consensus  | [00] [REPLICA][1] [WARN] utilityContext expected to be nil but is not. TODO: Investigate why this is and fix it
// node1.consensus  | [00] About to release postgres context at height 12.
// node3.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node2.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 2 VALIDATOR_TYPE_REPLICA
// node1.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 2 VALIDATOR_TYPE_LEADER
// node3.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node4.consensus  | [00] [REPLICA][4] 🙇🙇🙇 Elected leader for height/round 12/2: [3] (67eb3f0a50ae459fecf666be0e93176e92441317) 🙇🙇🙇
// node4.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 2 VALIDATOR_TYPE_REPLICA
// node4.consensus  | [00] [REPLICA][4] [WARN] utilityContext expected to be nil but is not. TODO: Investigate why this is and fix it
// node1.consensus  | [00] [REPLICA][1] Waiting for more HOTSTUFF_STEP_NEWROUND messages; byzantine optimistic threshold not met: (2 > 2.67?)
// node3.consensus  | [00] [EVENT] event_metrics_namespace_p2p raintree_message_event_metric send send
// node4.consensus  | [00] About to release postgres context at height 12.
// node2.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 2 VALIDATOR_TYPE_LEADER
// node3.consensus  | [00] [NODE][3] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node3.consensus  | [00] [NODE][3] [DEBUG] Handling message w/ Height: 12; Type: HOTSTUFF_STEP_NEWROUND; Round: 1.
// node4.consensus  | [00] [EVENT] event_metrics_namespace_consensus hotpokt_message_event_metric HEIGHT 12 HOTSTUFF_STEP_NEWROUND 2 VALIDATOR_TYPE_LEADER
// node2.consensus  | [00] [REPLICA][2] received enough HOTSTUFF_STEP_NEWROUND votes!
// node4.consensus  | [00] [REPLICA][4] Waiting for more HOTSTUFF_STEP_NEWROUND messages; byzantine optimistic threshold not met: (2 > 2.67?)
// node2.consensus  | [00] [REPLICA][2] [WARN] utilityContext expected to be nil but is not. TODO: Investigate why this is and fix it
// node2.consensus  | [00] About to release postgres context at height 12.
// node2.consensus  | [00] [REPLICA][2] Preparing a new block - no highPrepareQC found
// node2.consensus  | [00] [ERROR][REPLICA][2] could not prepare block: node should not call `prepareBlock` if it is not a leader
// node2.consensus  | [00] [REPLICA][2] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_PREPARE, 2)! Reason: failed to prepare & apply block
// node2.consensus  | [00] About to release postgres context at height 12.
// node3.consensus  | [00] [NODE][3] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_NEWROUND, 2)! Reason: timeout
// node4.consensus  | [00] [REPLICA][4] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_PREPARE, 2)! Reason: timeout
// node4.consensus  | [00] About to release postgres context at height 12.
// node1.consensus  | [00] [REPLICA][1] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_PREPARE, 2)! Reason: timeout
// node1.consensus  | [00] About to release postgres context at height 12.
// node2.consensus  | [00] [NODE][2] INTERRUPT at (height, step, round): (12, HOTSTUFF_STEP_NEWROUND, 3)! Reason: timeout
