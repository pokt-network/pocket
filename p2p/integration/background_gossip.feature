Feature: Background Gossip Broadcast

  Scenario Outline: Complete broadcast
    Given a fully connected network of <size> peers
    When a node broadcasts a test message via its background router
    Then <size> number of nodes should receive the test message

    Examples:
      | size |
      | 4    |
      | 6    |
      | 12   |
#      | 100  |
#      | 1024 |

  Scenario Outline: Partial broadcast
    Given a faulty network of <size> peers
    And <faulty> number of faulty peers
    When a node broadcasts a test message via its background router
    Then <received> number of nodes should receive the test message

    Examples:
      | size | faulty | received |
      | 4    | 2      | 2        |
      | 6    | 3      | 3        |
      | 12   | 8      | 4        |
#      | 100  | 75     | 25       |
#      | 1024 | 1000   | 24       |

  Scenario Outline: Broadcast during churn
    Given a fully connected network of <size> peers
    When <join> number of nodes join the network
    And <leave> number of nodes leave the network
    And a node broadcasts a test message via its background router
    Then <received> number of nodes should receive the test message

    Examples:
      | size | leave | join | received |
      | 4    | 2     | 2    | 4        |
      | 4    | 3     | 4    | 5        |
      | 12   | 6     | 6    | 12       |
#      | 100  | 50    | 55   | 105      |
#      | 1024 | 1000  | 1200 | 1224     |
