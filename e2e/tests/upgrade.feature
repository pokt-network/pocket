Feature: Upgrade Protocol

  Scenario: User can query the current protocol version using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And "validator-001" should be at height "0"
    And the user runs the command "query upgrade"
    Then the user should be able to see standard output containing "<version>"

    Examples:
      | version |
      | 1.0.0   |

  Scenario: ACL Owner Successfully Submits a Protocol Upgrade Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And "validator-001" should be at height "0"
    And the user is an ACL Owner
    When the user submits a major protocol upgrade
    And the network commits the transactions
    When the user runs the command "query upgrade"
    Then the user should be able to see the new version

  Scenario: ACL Owner Submits an Invalid Protocol Upgrade Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And "validator-001" should be at height "0"
    And the user is an ACL Owner
    And the user has an invalid upgrade protocol command
    When the user runs the command "gov upgrade"
    Then the system should validate the command
    And the system should reject the command due to invalid input

  Scenario: ACL Owner Submits a Protocol Upgrade with Too Many Versions Ahead Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And the user is an ACL Owner
    And the user has a upgrade protocol command with too many versions jump
    When the user runs the command "gov upgrade"
    Then the system should validate the command
    And the system should reject the command due to too many versions ahead

  Scenario: Regular User Submits an Upgrade Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    When the user submits a major protocol upgrade
    When the user runs the command "gov upgrade 100.0.0 100000"
    Then the user should be able to see standard output containing "invalid upgrade proposal: sender is not authorized to submit upgrade proposals"
