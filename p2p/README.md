## The P2P Module

Welcome to the P2P Module architecture guide.

### 1. Introduction: A word or two about the P2P module


We think that it's beneficial to first off set the context properly before we start diving into how this module was architected, so this introduction should help better-understand the decisions behind architecting P2P the way it is.

P2P as a module will primarily deal with sending data, receiving data, broadcasting data to the either specific peers or the entirety of the network.

A peer might be involved in multiple processes at once such as sending to 10 peers, receiving from 20, and broadcasting to everyone else, regardless of where these sends are coming from or what will be happening with the received data.
Thus, it's crucial that the P2P module be architected in a concurrency-friendly way.

Right off the bat, one can see that in a peer should be able to concurrently:
  1- connect to multiple peers.
  2- receive connections from multiple peers.
  3- reads incoming data from connected peers.
  4- send outgoing data to connected peers.

Therefore, we should at least have 3 "types" long running routines:

    - a connections establishment routine (_will be termed listening routine from now onwards_)
    - a read-from-connection routine (_will be termed read-routine from now onwards_)
    - a write-to-connection routine (_will be termed write-routine from now onwards_)

To further help paint a visual image of this, imagine a peer Y being connected to 5 other peers, our peer Y will be having 11 routines running along each other, mainly:

    1- a listening routine that accepts incoming connections
    2- a read-routine and a write-routine for each connected peer (2x5)

Now the main question is, how do we achieve this and how do we write it in code, in a "proper" way?

### 2. Architecture


### 2.1 Separating Concerns

It's of no use to re-introduce you to the concept of "Separation of Concerns" or remind you what benefits will divide-and-conquer bring to the world of architecture, so let's go straight into what we think should be separated.

We think that the operations of a given peer, in regards to itself and its behavior within the network **should be separated** from the operations its having  or performing with the peers its connected to. Meaning, what this given peer wants to do as a singular entity should be separated from what the peer is doing with its active neighbors to either maintain their connection or facilitate connectivity or IO. In clear terms:

 * Operations of a peer:
    - listen for new inbound connections
    - establish new outbound connections
    - store/map/index established connections
    - do something with a connection (send, ack, ping...)

 * Operations of a peer that is having/perfomring with its active neighbours:
    - authenticate the connection in between
    - perform continual reads and writes to the connection
    - handle connectivity issues (timeouts, errors)
    - close the connection

By **"should be separated"** we mean that the two should be overlooked by different components.

The component to manage peer-related operations will be named **Peer** and will live under `p2p/peer.go` whereas the component to manage inter-peer operations will be named **Socket** and will under `p2p/socket.go`.

### 2.2 Concerns breakdown

#### **2.3.1 Peer**

TODO(derrandz):

#### **2.3.2 Socket**

TODO(derrandz):

### 2.3 The Glue