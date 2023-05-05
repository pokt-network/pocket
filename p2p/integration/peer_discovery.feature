Feature: Background Router Peer Discovery
  # TODO_THIS_COMMIT: reword `node`; at this level, it represents (a) P2P
  # module(s) and/or Router implementation(s).

#  TODO: more scenarios
#  Scenario: Client joins network
#  Scenario: Inactive nodes are removed from the peerstore

  Scenario Outline: Fully connected network bootstrapping
    Given a "bootstrap" node
    When <size> number of nodes join the network
    Then the "bootstrap" node should have <size> number of peers in its peerstore
    And the "bootstrap" node should not be included in its own peerstore
    And other nodes should have <size> number of peers in their peerstores
    And other nodes should not be included in their respective peerstores

    Examples:
      | size |
      | 4    |
      | 6    |
      | 12   |
#      | 100  |
#      | 1024 |

  Scenario Outline: Fully connected network churning
    Given a "bootstrap" node
    When <initial> number of nodes join the network
    And  the "bootstrap" node leaves the network
    And <leave> number of nodes leave the network
    And <join> number of nodes join the network
    Then the network should contain <final> number of nodes
    And each node should have <final> number of peers in their respective peerstores
    And each node should not have any nodes which left in their peerstores

    Examples:
      | initial | leave | join | final |
      | 4       | 2     | 2    | 4     |
      | 4       | 3     | 4    | 5     |
      | 12      | 6     | 6    | 12    |
#      | 100  | 50    | 55   | 105      |
#      | 1024 | 1000  | 1200 | 1224     |

#  TODO_THIS_COMMIT: very similar test will exercise libp2p relaying..
  Scenario Outline: Discovery across pre-bootstrap network partitions
    Given a "bootstrap_A" node in partition "A"
    And a "bootstrap_B" node in partition "B"
    And a "bootstrap_C" node in partition "C"
    And <size_a> number of nodes bootstrap in partition "A"
    And <size_b> number of nodes bootstrap in partition "B"
    And <size_c> number of nodes bootstrap in partition "C"
    When a "bridge_AB" node joins partitions "A" and "B"
    Then all nodes in partition "A" should discover all nodes in partition "B"
    And all nodes in partition "B" should discover all nodes in partition "A"
    When a "bridge_BC" node joins partitions "B" and "C"
    Then all nodes in partition "A" should discover all nodes in partition "C"
    And all nodes in partition "B" should discover all nodes in partition "C"
    And all nodes in partition "C" should discover all nodes in partition "A"
    And all nodes in partition "C" should discover all nodes in partition "B"

    Examples:
      | size_a | size_b | size_c |
      | 2      | 2      | 3      |
      | 10     | 5      | 12     |
