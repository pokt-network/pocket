Feature: Trustless Relays

  Scenario: User Wants Help Using The Servicer Command 
    When the user runs the command "Servicer help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error


    # Happy test case: An Application requests the account balance of a specific address at a specific height from a Servicer staked for the Ethereum RelayChain, and receives a successful response.

  # ADDPR: Add a servicer staked for the Ethereum relaychain to the genesis file
  Scenario: Application can send a trustless relay to a relaychain to get an account's balance at a specific height
    Given the application has a valid ethereum relaychain account
    Given the application has a valid ethereum relaychain height
    Given the application has a valid servicer
    # INCOMPLETE: GeoZone	
    When the application sends a relay to a servicer
    # Balance: 1,160,126.46817237178258965 ETH  = 0xf5aa94f49d4fd1f8dcd2
    Then the relay response contains 0xf5aa94f49d4fd1f8dcd2
    And the validator should have exited without error

    # ADDPR: Sad test case: An Application requests the account balance of a specific address at a specific height from a Servicer staked for the Ethereum RelayChain in the same GeoZone, and the request times out without a response.
