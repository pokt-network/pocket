Feature: Upgrade Protocol

  Scenario: ACL Owner Successfully Submits a Protocol Upgrade Using CLI
    Given the user is an ACL Owner
    And the user has a valid upgrade protocol command with signer, height, and new version
    When the user runs the command "gov upgrade"
    Then the system should validate the command
    And the system should successfully accept the command
    And the system should apply the protocol upgrade at the specified activation height
    When the user runs the command "query upgrade"
    Then the system should return the updated protocol version

  Scenario: ACL Owner Submits an Invalid Protocol Upgrade Using CLI
    Given the user is an ACL Owner
    And the user has an invalid upgrade protocol command
    When the user runs the command "gov upgrade"
    Then the system should validate the command
    And the system should reject the command due to invalid input

  Scenario: ACL Owner Submits a Protocol Upgrade with Too Many Versions Ahead Using CLI
    Given the user is an ACL Owner
    And the user has a upgrade protocol command with too many versions jump
    When the user runs the command "gov upgrade"
    Then the system should validate the command
    And the system should reject the command due to too many versions ahead
