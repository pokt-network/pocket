Feature: State Sync Namespace

  # IMPROVE(#959): Remove time-based waits from tests


  @skip_in_ci
  Scenario: New FullNode does not sync to Blockchain at height 2
        Given the network is at genesis
        And the network has "4" actors of type "Validator"
        When the developer runs the command "ScaleActor full_nodes 1"
        And the developer waits for "3000" milliseconds
        Then "full-node-002" should be unreachable
        When the developer runs the command "TriggerView"
        And the developer waits for "1000" milliseconds
        And the developer runs the command "TriggerView"
        And the developer waits for "1000" milliseconds
        Then "validator-001" should be at height "2"
        And "validator-004" should be at height "2"
        # full_nodes is the key used in `localnet_config.yaml`
        When the developer runs the command "ScaleActor full_nodes 2"
        # IMPROVE: Figure out if there's something better to do then waiting for a node to spin up
        And the developer waits for "40000" milliseconds
        # TODO(#812): The full node should be at height "2" after state sync is implemented
        Then "full-node-002" should be at height "0"