# TECHDEBT: Validator should eventually be changed to full node or just node.
Feature: Node Namespace

  Scenario: User Wants Help Using The Node Command 
    Given the user has a validator
    When the user runs the command "Node help"
    Then the user should be able to see standard output containing "Available Commands"
    And the validator should have exited without error
