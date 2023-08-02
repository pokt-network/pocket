Feature: Node Namespace

  Scenario: User Wants Help Using The Node Command
    Given the user has a node
    When the user runs the command "Validator help"
    Then the user should be able to see standard output containing "Available Commands"
    And the node should have exited without error

  Scenario: User Can Stake A Validator
    Given the user has a node
    When the user stakes their validator with amount 150000000001 uPOKT
    Then the user should be able to see standard output containing ""
    And the node should have exited without error

  Scenario: User Can Unstake A Validator
    Given the user has a node
    When the user stakes their validator with amount 150000000001 uPOKT
    Then the user should be able to see standard output containing ""
    Then the user should be able to unstake their validator
    Then the user should be able to see standard output containing ""
    And the node should have exited without error

  Scenario: User Can Send To An Address
    Given the user has a node
    When the user sends 150000000 uPOKT to another address
    Then the user should be able to see standard output containing ""
    And the node should have exited without error
