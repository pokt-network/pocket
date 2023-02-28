
Feature: Accounts Namespace
  
  Scenario: User Checks Account Help Command
    Given the user has a pocket client 
    When the user runs the command "account"
    Then the user should be able to see standard output containing "Available Commands"
    And the pocket client should have exited without error

  Scenario: User can be a validator
    Given the user has a validator
    When the user runs the validator command "help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error
