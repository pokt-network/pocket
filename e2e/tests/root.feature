Feature: Root Namespace

  Scenario: User Needs Help
    Given the user has a validator
    When the user runs the command "help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error