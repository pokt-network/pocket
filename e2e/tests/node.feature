Feature: Node Namespace

  Scenario: User Wants Help Using The Node Command
    Given the user has a node
    When the user runs the command "Node help"
    Then the user should be able to see standard output containing "Available Commands"
    And the node should have exited without error

  Scenario: User Can Stake An Address
    Given the user has a node
    When the user stakes their node with amount 150000000001 uPOKT
    Then the user should be able to see standard output containing ""
    And the node should have exited without error

  Scenario: User Can Unstake An Address
    Given the user has a node
    When the user stakes their node with amount 150000000001 uPOKT
    Then the user should be able to see standard output containing ""
    Then the user should be able to unstake their node
    Then the user should be able to see standard output containing ""
    And the node should have exited without error

  Scenario: User Can Send To An Address
    Given the user has a node
    When the user sends 150000000 uPOKT to another address
    Then the user should be able to see standard output containing ""
    And the node should have exited without error
