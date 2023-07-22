Feature: Upgrade Protocol

  Scenario: ACL Owner Successfully Submits a Protocol Upgrade Using CLI
    Given the user is an ACL Owner
    When the user runs the command "gov upgrade <version> <height> <chain-id>"
    Then the user should be able to see standard output containing "<stdout>"
    When the user runs the command "gov query upgrade"
    Then the user should be able to see standard output containing "<version>"

    Examples:
      | version | height | chain-id | stdout |
      | 2.0.0   | 100    | test     | 2.0.0  |

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
