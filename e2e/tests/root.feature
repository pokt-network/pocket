Feature: Root Namespace

  Scenario: User Needs Help
    Given the user has a node
    When the user runs the command with no error "help"
    Then the user should be able to see standard output containing "Available Commands"
    And the node should have exited without error
