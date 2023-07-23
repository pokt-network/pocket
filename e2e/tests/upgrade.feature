Feature: Upgrade Protocol

  Scenario: ACL Owner Successfully Submits a Protocol Upgrade Using CLI
    Given the user is an ACL Owner
    When the user runs the command "gov upgrade <owner> <version> <height>"
    Then the user should be able to see standard output containing "<stdout>"
    When the user runs the command "gov query upgrade"
    Then the user should be able to see standard output containing "<version>"

    Examples:
      | owner                                    | version | height | stdout |
      | da034209758b78eaea06dd99c07909ab54c99b45 | 2.0.0   | 100    | 2.0.0  |

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
