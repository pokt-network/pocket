Feature: Trustless Relays
    # Happy test case: An Application requests the account balance of a specific address at a specific height from a Servicer staked for the Ethereum RelayChain, and receives a successful response.

  # ADD_IN_THIS_PR: Add a servicer staked for the Ethereum relaychain to the genesis file
  Scenario: Application can send a trustless relay to a relaychain to get an account's balance at a specific height
    Given the application has a valid ethereum relaychain account
    Given the application has a valid ethereum relaychain height
    Given the application has a valid servicer
    # INCOMPLETE: GeoZone	
    When the application sends a get balance relay at a specific height to an Ethereum Servicer
    # Balance: 1,160,126.46817237178258965 ETH  = 0xf5aa94f49d4fd1f8dcd2
    Then the relay response contains 0xf5aa94f49d4fd1f8dcd2
    And the relay response is valid json rpc
    And the relay response has valid id
    # TECHDEBT: replace validator with client
    And the validator should have exited without error

    # ADD_IN_THIS_PR: Sad test case: An Application requests the account balance of a specific address at a specific height from a Servicer staked for the Ethereum RelayChain in the same GeoZone, and the request times out without a response.

    # TODO: add an E2E test for a trustless relay, where the application retrieves the session first, using a new fetch session command
