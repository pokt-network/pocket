# Pocket Network 1.0 Peer-To-Peer Module Pre-Planning Specification: Fast, Scalable, Highly Reliable and Optionaly Redundant Binary Tree Gossip

<table>
<tr style="border-width: 0px;">
  <td style="border-width: 0px;">
    <p align="center">
        Hamza Ouaghad
        @derrandz<br>
        Version 1.0.1
    </p>
  </td>

  <td style="border-width: 0px;">
    <p align="center">
        Andrew Nguyen 
        @andrewnguyen22 <br>
        Version 1.0.1
    </p>
  </td>


  <td style="border-width: 0px;">
    <p align="center">
        Otto Vargas
        @oten91<br>
        Version 1.0.1
    </p>
  </td>

  <td style="border-width: 0px;">
    <p align="center">
        Shawn Regan
        @benvan<br>
        Version 1.0.1
    </p>
  </td>
</tr>
</table>

## Overview
Choosing the proper data structure to represent the structure of a network's overlay is the main and the crucial step to achieving a structured overlay, and a detrimental one for building an efficient and performant network.

We chose to go with a structured overlay approach for our p2p network to be able to theorize about the network and manage churn among other things.

In this section, we will explain the structure that will power Pocket Network V1's P2P layer.

**A reminder of our visual toolbox** (_while thanking devp2p spec for the inspiration_):

💡: indicates an implementation avenue or suggestion

✍🏻: A new term we chose to use for convenience purposes

🗝: Key concept that should be paid attention to

## I. Introduction 

During our [research](#), we have been able to identify a few good candidate structures/algorithms for our overlay, out of which we thought Gemini (_previous version of spec is available in (add link))_ would be serve us best, however, having ran a few simulations and performed a few projections, we realized that as scalable as Gemini is, we could still face some challenges before we reach a relatively significant network size (_peer count > 10K_)


All of a sudden, the challenge was not scaling to a billion nodes but rather accommodating for cases when the network is still at its infancy and growing, a 1000 and below.

We first off tried to keep it as simple as possible and tried to go with an if else approach. If below a certain size, use a PRR algorithm, if beyond the threshold, fallback to Gemini.

However we weren't particularly interested in giving ourselves further work to do, we wanted a structure that will serve us best as simply as possible.
After some thought, and after revisiting other candidates we have covered in our research, we took a special appeal to OneHop or Cosntant Hops algorithms (such as Kelips O(1)), but we hoped to extract the good parts only, so that we don't have to deal with the file-lookup-specific gymnastics and what not.

Out of that thought process, we ended up picking what was first an optimization to our gossip algorithm for block proposal rounds, but later turned out to give us all the right answers with very intresting simplicity! It's called: Rain Tree.

## II. Specification

### 1. Structure

The network structure in RainTree is a list of peers sorted by the numerical distance of their IDs. This list undergoes a set of operations/transformations which are primarily functions of the peers actual numerical position. For instance, when dealing with a message propagation, an originator node is designated as the root member of a binary tree whose right and left branches are the 33th% and the 66th% peers of that root's position in the list, and in and of themselves, these rights and lefts pick their own binary branches using the same logic. So in short, the sorted list has been transformed into a binary tree whose root is the originator's node and whose rights and lefts are always the 33th% and the 66th% of the immediate root and so on and so forth.


In such a tree/list, we make use of the concept of right and left branch targets, tree layers and max possible tree layers. We codify these concepts as follows:

- Max possible Layers = log₃ of List Size
- Layer of a peer = Round up of the count of the exponents of 3 in the peers ID
- Right branch target = Node position + targetListSize/3 (roll over if needed)
- Left branch target = Node position + targetListSize/	1.5 (roll over if needed)
- targetListSize = (topLayer + currentLayer) x 0.666 x Size of full list

This approach is a very simple one. No specifically complex classification or routing logic is required, RainTree relies on the fact that a binary tree lookup is in fact one of the most optimal methods as far as searching in a sorted list with random data goes, and it leverages so to achieve efficient communication in a p2p network.

You can learn more about how we've come to these conclusions by reading the original presentation document.

In summary, RainTree is a very fast, scalable and highly reliable optionaly redundant gossip algorithm that relies on the fact that a binary search is optimal for most randomly distributed datasets.

To gossip a message M, RainTree uses a distance-based metric to build a binary tree view of the network and propagates information down the tree. This traversal algorithm follows a tree reconstruction algorithm such that the resulting tree is one of three branches and not two.

![](https://i.ibb.co/q1KLrCs/Screen-Shot-2022-01-27-at-18-41-20.png)

RainTree requires that the peer list / neighbors list be sorted based on the distance metric.

The root of the tree is message originator Node S, where the immediate left and right branches are the %33th and %66th positions in the peer list respectively. Node S is said to have `level`, which is the starting layer for the gossip, and it's referred to as the `top layer`.


To determine Node S' layer, we use the following formula:

```
 Count the exponents of 3 in the list and then round up:

       Toplayer = Round(Log₃(fullListSize))
  
* fullListSize: the size of the peer list of Node S
```

As mentioned before, `S` will follow a tree reconstruction algorithm to achieve full message propagation, such that when it has delivered message `M` to its left and right, `S` "demotes" itself one level down, and picks a new left and right using the same logic. This demotion goes on until the bottom-most layer (layer 0)

![](https://i.ibb.co/PWqrhK3/Screen-Shot-2022-01-27-at-18-43-07.png)

Each node receiving message `M` will follow the same logic..


### 2. Algorithm (Gossip) 

The originator of the message propagates the message as follows:
```

Let S be the source node of message M, whose peer list is of size N. To propagate a message, S does the following:

Determine max possible layers of the network (using current peer list size)
Determine own layer (top layer)
Determine current layer (top layer - 1)
Pick Right and Left branch targets belonging to current layer and send message to them and to self
Go to next layer (decrement current layer)
Repeat 4 and 5 until current layer is 0

Each peer receiving the message M at a given layer L will replay this procedure. We call the process of going to the next layer "demotion Logic" as the originator leaves his actual original layer and "demotes" himself to "lower" layers all the way to layer 0.
```

![](https://i.ibb.co/pdkKLsM/Screen-Shot-2022-01-27-at-18-59-50.png)

### 3. The Redundancy Layer

An optional redudancy layer can be added to RainTree, such that the originator sends message M to the full list on level 0.

![](https://i.ibb.co/6wj9RP5/Screen-Shot-2022-01-27-at-19-04-27.png)

So, the algorithm becomes:

![](https://i.ibb.co/1dsNzwC/Screen-Shot-2022-01-27-at-19-05-16.png)

This redundancy layer insures against non-participation and incomplete lists without the ACK/RESEND overhead, whereas the reliability layer (Daisy Chain clean-up layer) ensures 100% message propagation in all cases. (See next segment)

### 4. The Failure Detection and Recovery Layer (Daisy-Chain Clean Up Layer)

Networks are prone to failure and partitions, so RainTree offers a clean up layer (_a reliability layer_) that ensures that every node has successfully received the message M.

The Daisy Chain Clean Up layer kicks in at level 0, such that nodes receiving messages from level 1 will go asking sequentially every other node of whether they have or want the just-received.

![](https://i.ibb.co/NYGhrMg/Screen-Shot-2022-01-27-at-18-48-33.png)

This message is denoted as a `IGYW` message, and it propagates following this algorithm:

```
a IGYW ( I GOT, YOU WANT?) mechanism is put in use such that a given peer does not fully send a message until the receiving part signals that it did not receive it before.

This is achieved by level 1 nodes, such that once they have received a message, they do the following:

Send IGYW to immediate left neighbor:
If answer is Yes, send full message and go to step 2
If answer is No, go to step 2
If no answer, increment left counter and go to step 1
Send IGYW to immediate right neighbour:
If answer is Yes, send full message
If answer is No, 
If no answer, increment right counter and go to step 2.
```

![](https://i.ibb.co/QJ19Rn5/Screen-Shot-2022-01-27-at-18-58-33.png)

Thus, the full algorithm becomes as follows:

![](https://i.ibb.co/6sbRy6K/Screen-Shot-2022-01-27-at-19-05-56.png)

This process is an ACK/ADJUST/RESEND mechanism, for if no ACK was received, an ADJUST instruction takes place, which right after a RESEND instruction is initiated. 

### 5 Maintenance

As with any DHT-like network, some level of network maintenance (also known as membership maintenance/protocol or churn management) is required to keep the network connected.

RainTree is different in that it's similar to Constant Hop networks, in that its churn management process is minimal to non-existent. RainTree requires every member to have a close-to-full view of the network.


##### 2.1 Join/Discovery

Any new peer should be able to join the network and participate in it seamlessly. To ensure that our Join/Discovery process achieves this, we would like to answer the following requirements:

Any given random peer should be able to discover other peers in the network from their given current perspective of the network (either their existing state their seed adresse(s))
Any given peer can perform basic discovery and can safely fallback to such a procedure in the absence or presence of specialized peers in the network with no problem at all.

To answer these requirements efficiently, we baked the discovery process into the join process.


##### 2.2 Join

When a new peer X joins the network:

It first contacts an existing bootstrap peer(s) E.
Peer(s) E will answer with their peer lists.
Peer X retrieves the lists and performs a raintree propagation of a Join Message with its Address in it denoted as the new joiner.
ACKs can be enforced to keep peers from being filtered from peer X�s peer list due to lack of response.

This way, when a peer joins, it is immediately given at least one peer list it can start working with, and can by itself clean it up using ACKs and timeouts.

##### 2.3 Leave

A peer that wants to leave the network basically just disconnects and relies on the maintenance routine to "discover" and broadcast its unavailability.

#### 3.5 Network Parameters and Scalability


This will scale! If you tripple the node counts, the only increase is ticks=+2

| Nodes             |  Comms  | ACKs     | Ticks  |
|-------------------|---------|----------|--------|
| 27                |   107   |   56     |   11   |
| 81                |   323   |   164    |   13   |
| 243               |   971   |   488    |   15   |
| 729               |  2,915  |   1,460  |   17   |
| 2,187             |  8,747  |   4,376  |   19   |
| 6,561             | 26,243  |  13,124  |   21   |
| 19,683            | 78,731  |  39,358  |   23   |
| 59,049            | 236,195 |  118,100 |   25   |
| 177,147           | 708,587 |  354,296 |   27   |


##### Real life experimentation data

We will be looking to add some interesting results from a scientific simulation of rain tree available in [rain-tree-simulation](https://github.com/pokt-network/rain-tree-sim) repository.

### 6 Transport Protocol And Security

Transport logic and security are key elements in the inner working of the p2p network. Here we try to outline the general properties and specifications that our network should have and comply with. We also detail some possible attacks that we may be susceptible to. 

##### 6.1 Connection Lifecycle

A connection is initiated by the peers
Handshake protocol is initiated and peers exchange secrets to establish a secure encrypted channel.
Messages are then sent on-demand while the connection is alive.
The connection uses a default timeout to ensure that if idle for x amount of time resources are freed and no unnecessary allocations happen.

##### 6.2 Handshake protocol draft

1. Perform Diffie-Hellman handshake:
    1. Peers generate ephemeral Ed25519 public and private keys
    2. Peers sign a nonce message and send it with their public key to the other party
        1. Define nonce to be (*for instance*): `_p2p_pokt_network_handshake_`
    3. The bytes order is important and can be as follows: `[pubkey... , 0, signature...]`
2. Peers convert the public keys they received into Curve25519 public keys,
3. Peers convert their ephemeral Ed25519 private keys into Curve25519 private keys,
4. Peers establish a shared secret by performing ECDH with their private Curve25519 private key and their peers Curve25519 public key.
5. Peers exchange the produced shared key as follows:
    1. Peer A constructs a message of bytes as follows: `[peer.persistentPubkey..., sharedKey...]`
    2. Peer A signs it with its persistent private key and sends it to be
    3. Peer B decrypts and the messages and sends back the same format: `[peer.persistentPubkey..., sharedKey...]`
    4. Peer A upon receiving the response reconstructs the message with peer B�s publickey and the shared secret it produced earlier and verifies it using B�s persistent Publickey
6. Peers use the shared secret as a symmetric key and communicate from then on with messages encrypted/decrypted via. AES 256-bit GCM with a randomly generated 12-byte nonce.

##### 6.3 Connections Pooling

Connection pooling is required to recycle existing connections and properly utilize the available bandwidth.

The network parameters are theorized for a network bandwidth capacity minimum of 500Mbps (*for both upload and download*).

- Max size of a message is 4MB+DataHeaderSize (*a completely full block*)
- DataHeaderSize describes metadata about the data transmitted, primairly:
    - Size: 4 bytes
    - *insert others if needed*
- Max number of inbound connections is 125 (*each connection consuming 4Mbs*)
- Max number of outbound connections is 125 (*each connection consuming 4Mbs*)

We can possibility open these parameters for external configuration to allow for robust servers to utilize their maximum capacity, but stick to a minimum acceptable network capacity such as the one stated above.

This is a basic bounded connection pool for regular operation with persistent peering options.

We intedn to add a specialized bounded pools for application use cases:

- Syncing allowance
- Consensus tasks (validators)
- Others.

##### 6.4 Protocol

We will rely on TCP/IP with a handshake for direct communications and Gossip where as we intend to use UDP for churn management communication.

##### 6.6 Security
Peer connections could be encrypted using AES 256-bit Galois Counter Mode (GCM) with a Curve25519 shared key established by an Elliptic-Curve Diffie-Hellman Handshake.

Very similar to TLS handshakes.

### Messages In the Overlay

Each protocol will define its messages. Take the following starting index of messages per protocol:

- Membership Protocol
    - Ping
    - Pong
    - Join
    - Leave
- Gossip Protocol (*RainTree)*
    - Gossip
    - GossipACK
    - GossipRESEND
- Daisy-Chain Protocol (*Gossip Reliability Protocol*)
    - IGYW
    - IGYW_AFF
    - IGYW_NEG
