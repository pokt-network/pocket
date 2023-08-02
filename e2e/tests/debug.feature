Feature: Debug Namespace

    # IMPROVE(#959): Remove time-based waits from tests

    # Since the configuration for consensus is optimistically responsive, we need to be in manual
    # Pacemaker mode and call TriggerView to further the blockchain.
    # 1 second was chosen arbitrarily for the time for block propagation.
    Scenario: 4 Validator blockchain from genesis reaches block 2 when TriggerView is executed twice
        Given the network is at genesis
        And the network has "4" actors of type "Validator"
        When the developer runs the command "TriggerView"
        And the developer waits for "1000" milliseconds
        Then "validator-001" should be at height "1"
        And "validator-004" should be at height "1"
        When the developer runs the command "TriggerView"
        And the developer waits for "1000" milliseconds
        Then "validator-001" should be at height "2"
        And "validator-004" should be at height "2"