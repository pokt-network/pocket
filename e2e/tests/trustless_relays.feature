Feature: Trustless Relays

  Scenario: User Wants Help Using The Servicer Command 
    When the user runs the command "Servicer help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error


  # Happy test cases
	# An Application requests the account balance of a specific address at a specific height from a Servicer staked for the Ethereum RelayChain, and receives a successful response.
 
  Scenario: Application can send a trustless relay to a relaychain to get an account's balance at a specific height
		# Given the application has a valid ethereum relaychain account
		#    Given the application has a valid ethereum relaychain height
		#    Given the application has a valid servicer for the session
    # INCOMPLETE: GeoZone	
    # ADDPR: specify the servicer 
    # ADDPR: specify the relay method and params
    When the application sends a relay to a servicer
    Then the relay response contains 0xf5aa94f49d4fd1f8dcd2
    And the validator should have exited without error

    # ADDPR: Sad test case: An Application requests the account balance of a specific address at a specific height from a Servicer staked for the Ethereum RelayChain in the same GeoZone, and the request times out without a response.
