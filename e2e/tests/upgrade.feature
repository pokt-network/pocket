Feature: Upgrade Protocol

  Scenario: User can query the current protocol version using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And "validator-001" should be at height "0"
    And the user runs the command with no error "query upgrade"
    Then the user should be able to see standard output containing "<version>"

    Examples:
      | version |
      | 1.0.0   |

  Scenario: ACL Owner Successfully Submits a Protocol Upgrade Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And "validator-001" should be at height "0"
    And the user is an ACL Owner
    When the user runs the command with no error "gov upgrade da034209758b78eaea06dd99c07909ab54c99b45 2.0.0 1"
    And the developer runs the command "TriggerView"
    And the developer waits for "1000" milliseconds
    And "validator-001" should be at height "1"
    And the user runs the command with no error "query upgrade"
    Then the user should be able to see standard output containing "2.0.0"

  Scenario: ACL Owner Fails Basic Validation Submitting a Protocol Upgrade Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And "validator-001" should be at height "0"
    And the user is an ACL Owner
    When the user runs the command with error "<cmd>"
    Then the user should be able to see standard error containing "<error>"

    Examples:
      | cmd                                                             | error                                             |
      | gov upgrade da034209758b78eaea06dd99c07909ab54c99b45 2.0.zxcv 1 | CODE: 149, ERROR: the protocol version is invalid |
      | gov upgrade da034209758b78eaea06dd99c07909ab54c99b45 new 1      | CODE: 149, ERROR: the protocol version is invalid |

  Scenario: ACL Owner Fails Consensus Validation Submitting a Protocol Upgrade Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    And "validator-001" should be at height "0"
    And the user is an ACL Owner
    When the user runs the command with no error "<cmd>"
    And the developer runs the command "TriggerView"
    And the developer waits for "1000" milliseconds
    And "validator-001" should be at height "1"
    And the user queries the transaction
    Then the user should be able to see standard output containing "<error>"

    Examples:
      | cmd                                                          | error                                             |
      | gov upgrade da034209758b78eaea06dd99c07909ab54c99b45 3.0.0 1 | CODE: 149, ERROR: the protocol version is invalid |
      | gov upgrade da034209758b78eaea06dd99c07909ab54c99b45 2.0.0 0 | CODE: 149, ERROR: the protocol version is invalid |

  Scenario: Regular User Fails Consensus Validation Submits an Upgrade Using CLI
    Given the network is at genesis
    And the network has "4" actors of type "Validator"
    When the user submits the transaction "gov upgrade 00101f2ff54811e84df2d767c661f57a06349b7e 2.0.0 1"
    And the developer runs the command "TriggerView"
    And the developer waits for "1000" milliseconds
    And "validator-001" should be at height "1"
    And the user queries the transaction
    Then the user should be able to see standard output containing "CODE: 3, ERROR: the signer of the message is not a proper candidate: da034209758b78eaea06dd99c07909ab54c99b45"
