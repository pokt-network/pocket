Feature: Trustless Relays

  Scenario: User Wants Help Using The Servicer Command 
    When the user runs the command "Servicer help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error

  Scenario: User can send a trustless relay
    When the user sends a relay to a servicer
    Then the user should be able to see standard output containing "result"
    And the validator should have exited without error
