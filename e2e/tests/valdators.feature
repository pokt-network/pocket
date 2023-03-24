
Feature: Validators Namespace

  Scenario: User Wants Help Using The Validator Command 
    Given the user has a validator
    When the user runs the command "Validator help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error

  Scenario: User Can Stake Their Wallet
    Given the user has a validator
    When the user stakes their validator with 150000000001 POKT
    Then the user should be able to see standard output containing ""
    And the validator should have exited without error

  Scenario: User Can Unstake An Address
    Given the user has a validator
    When the user stakes their validator with 150000000001 POKT
    Then the user should be able to see standard output containing ""
    Then the user should be able to unstake their wallet
    Then the user should be able to see standard output containing ""
    And the validator should have exited without error