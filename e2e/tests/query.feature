
Feature: Query Namespace

  Scenario: User Wants Help Using The Query Command 
    Given the user has a validator
    When the user runs the command "Query help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error

    Scenario: User Wants To See The Block At Current Height
    Given the user has a validator
    When the user runs the command "Query Block"
    Then the user should be able to see standard output containing "state_hash"
    And the validator should have exited without error