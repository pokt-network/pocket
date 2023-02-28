
Feature: Validators Namespace

  Scenario: User wants help using the Validator command 
    Given the user has a validator
    When the user runs the validator command "Validator help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error
