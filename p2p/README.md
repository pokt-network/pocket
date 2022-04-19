# P2P Module

Welcome to the P2P Module architecture guide.

## 1. Introduction

This introduction should help better-understand the decision-making behind architecting the V1 P2P Module.

P2P -as a module- will primarily deal with sending and receiving data to and from specific peers, and often broadcasting data to the entire network.
A peer might be involved in multiple processes at once (_e.g. sending to 10 peers, receiving from 20, and broadcasting to everyone else_), thus a peer should be able to concurrently:

1. Connect to multiple peers
2. Receive connections from multiple peers
3. Read incoming data from connected peers
4. Send outgoing data to connected peers
5. Broadcast data to all connected peers


Thus, it's crucial that the P2P module be architected in a concurrency-friendly way.

Therefore, The P2P module has been architected such that it allows for 3 "types" of long-running routines:

- A connections' establishment routine (_will be termed **listening routine** from now onwards_)
- A read-from-connection routine (_will be termed **read-routine** from now onwards_)
- A write-to-connection routine (_will be termed **write-routine** from now onwards_)

To further help paint a visual image of this, imagine a peer Y being connected to `5 other peers`. Peer Y will have a total `11 routines` running along each other, mainly:

1. A listening routine that accepts incoming connections (`1 routine`)
2. A read-routine and a write-routine for each connected peer (`2x5 routines`)

## 2. Architecture

### 2.1 General
The p2p module is composed of two main sub-modules, the network module and the socket module in a hierarchical manner such that the network module controls, runs and commands the socket module.

The network module is responsible for:
  - listening for incoming connections
  - dialing outgoing connections
  - broadcasting to the network
  - sending to specific peers
  - redirecting non-p2p incoming data to consensus

The network module delegates connection-management to the socket module by "opening a socket" for each "new dial" or "incoming connection".
When the network module dials a new peer, it is said that the network module "opened an outbound socket" to the peer, similarly, when
the network module receives an incoming connection, it is said that the network module "opened an inbound socket" to the peer.

After "creating a socket" out of the dialed or incoming connection, the network module keeps track of the socket in a `sockets` map. (_more on this later, it wil implement pool-like logic to achieve pooling_)

In turn, the socket module takes care of every established connections' authentication, IO and error handling.

### 2.2 Socket

The socket tructure keeps track of the following attributes:

* A type (_inbound or outbound_)
* An address (_the address of the peer at the other end of the socket_)
* an ID (_the ID of the peer at the other end of the socket_)
* a reference to the TCP socket (_net.Conn_)
* general IO parameters (_timeouts, buffer sizes_)

The lifecycle of a socket is as follows:

1. The network module "opens" a socket to a peer.
2. The socket module "authenticates" the peer.
3. The socket module "opens" the socket.
4. The socket module "handles" the socket by reading and writing to it. (_kicking off the read-routine and write-routine_)
5. On fatal errors, the socket module "closes" the socket.

So in summary, these are the possible states of a socket:

[comment]: <> (Remove this NB when the socket feature is merged as it will add the missing states.)

(_for the moment, intermediery states such as opening, connecting, authenticating, authenticated, connected are not concretely implemented as the time of writing this._)

- **opening**: The socket is being created and authenticated.
- **connecting**: the socket is in the process of connecting to the peer.
- **authenticating**: the socket is in the process of authenticating the peer.
- **authenticated**: the socket has been authenticated and is ready to be used.
- **connected**: the socket has been connected and is ready to be used.
- **open**: the socket is open and ready to be used.
- **errored**: the socket has encountered an error and may or may no longer be usable
- **closing**: the socket is in the process of closing.
- **closed**: the socket has been closed and is no longer usable

[![](https://mermaid.ink/img/pako:eNplkT0LwjAQhv9KuVHs4thBEHUWdHAwDkdzaqBNSnoRRPrfTZv000x37_uQ-_pCbiRBBjUj00Hh02KZvjdCJ_7dVvckTbfJvjA1yaCFuJNPFWmln0GPSeCN1pTz4I15Z-8cv0izynFE5toS64vPpGmpobs-7cwzoVw2Ea2rVWPxyA1DBTUiC7WN_pYy_WCqH6019k8dvujsGdtPFWlYQ0m2RCX9ib4tJ8BvoCQBmQ8lPdAVLEDoxqOukn4vR6nYWMgeWNS0BnRsLh-dQ8bWUQ_FS0eq-QFY0aZD)](https://mermaid-js.github.io/mermaid-live-editor/edit#pako:eNplkT0LwjAQhv9KuVHs4thBEHUWdHAwDkdzaqBNSnoRRPrfTZv000x37_uQ-_pCbiRBBjUj00Hh02KZvjdCJ_7dVvckTbfJvjA1yaCFuJNPFWmln0GPSeCN1pTz4I15Z-8cv0izynFE5toS64vPpGmpobs-7cwzoVw2Ea2rVWPxyA1DBTUiC7WN_pYy_WCqH6019k8dvujsGdtPFWlYQ0m2RCX9ib4tJ8BvoCQBmQ8lPdAVLEDoxqOukn4vR6nYWMgeWNS0BnRsLh-dQ8bWUQ_FS0eq-QFY0aZD)

A socket achieves IO by running two main IO routines:

* A read routine.
* A write routine.

#### 2.2.1 The Socket's Read Routine

The read-routine is a long-running routine that continually waits on new data to be received, and reads it in a buffered manner by relying on the `readChunk` operation.

The socket 'read' routine performs buffered read operations on incoming data using a byte-slice buffer, to which the socket writes the data it gets from the TCP connection using the socket reader (_`io.Reader`_).
This allows the socket to read in "chunks" of data, which is useful for handling large amounts of data.

We can say that the read routine reads in chunks of size `BufferSize` (_This parameter is configurable_).

The `readChunk` operation performs minimal validation against the retrieved chunk to make sure it is of an acceptable size, and whether it's corrupted (_whether the decoding failed_).
If this validation fails, the chunk is rejected and socket is kept open.

The read routine will not close on faulty chunks, but will close on:
 - `io.EOF` error
 - Unexpected errors
 - Graceful shutdown caused by:
   - The closing of the socket's context
   - The runner has stopped (_i.e: the network module_)
   - The socket has been terminated using the `Close` operation.

Here is a flow diagram summarizing the read routien operations:

[![](https://mermaid.ink/img/pako:eNpVkttuwjAMhl_Fys02ib0AF5sGLeIgGCtMaGq5iBoD0UrSOckYKrz70hPrctPE_vL7t5uCpVog67M98fwA6yBR4NdLvD4gRMgFRNpZqXALj49PMLg3cq94BuRTUu3BWE4WxUN9bVBBwyLCFOU3ClB4AqlyZ5-vNTEsictCXyCIN1zabR0O6otd5gPNBcIiVKl2yiJ5tbIoIJGmVi68yY3iAMtWYCMJYcnPmeZi26UqwXExMfCu8CfH1PuGkAg0Qfg6aiXHlZWaXsQrq3OgagT77T-gLDqPI8w1tV0sqsxbPMy0wSY273Y2qg6T0kOEXw7NbS6TTtVpvHTmAEKfFFgN1v-JxgGc_MjK7857ruJocq1uxSZ_1mZdkZJ0SiHdGTBSfTb4tOtt1h5Yjx2RjlwK_yyKMpkwr3DEhPX9VuCOu8wmLFFXj7pccIuhkFYT6-94ZrDHuLN6dVYp61ty2EKB5P6VHRvq-gseTMfY)](https://mermaid-js.github.io/mermaid-live-editor/edit#pako:eNpVkttuwjAMhl_Fys02ib0AF5sGLeIgGCtMaGq5iBoD0UrSOckYKrz70hPrctPE_vL7t5uCpVog67M98fwA6yBR4NdLvD4gRMgFRNpZqXALj49PMLg3cq94BuRTUu3BWE4WxUN9bVBBwyLCFOU3ClB4AqlyZ5-vNTEsictCXyCIN1zabR0O6otd5gPNBcIiVKl2yiJ5tbIoIJGmVi68yY3iAMtWYCMJYcnPmeZi26UqwXExMfCu8CfH1PuGkAg0Qfg6aiXHlZWaXsQrq3OgagT77T-gLDqPI8w1tV0sqsxbPMy0wSY273Y2qg6T0kOEXw7NbS6TTtVpvHTmAEKfFFgN1v-JxgGc_MjK7857ruJocq1uxSZ_1mZdkZJ0SiHdGTBSfTb4tOtt1h5Yjx2RjlwK_yyKMpkwr3DEhPX9VuCOu8wmLFFXj7pccIuhkFYT6-94ZrDHuLN6dVYp61ty2EKB5P6VHRvq-gseTMfY) _A Socket's read routine (_flow diagram_)_

**The Runner's sink**: The socket keeps reference of a Runner interface, which is used to stop the read routine when the runner stops, but also is used to redirect incoming data "upwards" to the runner.
The Runner interface in this case is implemented by the network module, therefore the network module receives the incoming data from the sockets, and decides that to do with it (_is it a broadcast? is it a consensus message?_).

#### 2.2.2 The Socket's Write Routine

The write-routine is a long-running routine that continually **waits** on new data to be sent, and writes it in a _buffered manner_ by relying on the `writeChunk` operation.

The network module will try to send data to a specific peer, and thus write to a specific socket.
The send operation will subsequently call the `writeChunk` operation. This operation does two simple things:

  1. Write the data to the buffer
  2. Signal that new data is available (_this is a blocking operation, meaning that the signaling will block until someone is ready to receive that signal_)

Upon signaling that there is new data to be sent, the write-routine will unblock and stop waiting, and write the buffered data to the underlying TCP connection, and go back to waiting for new data.

[![](https://mermaid.ink/img/pako:eNpVkc1qwzAQhF9l0amF5AV8aGls5-dSShoIrZ3DYm0Sgb0K8opQ7Lx7ZcuBRCcx8w2zy3aqsppUok4OL2fYZSVDeB_F7kywd0YIttaLYTrAfP4Gi5fWnBhrcITa8AlaQSekX2NuMUJpNySFGJiukKHg-y366eD3n7aHrNijkUOUsxh7YCAvYrtYkDDKLv2CyjJTJcYyXAfPTel8DCy7nCvrOeikASMC5Jx19_blCPY_1PawKrZ0sU4icXgChvnWxbL27XkyVqPxW6S1bWnS1qO2ibX31jH03Lp5bH2Shp64tJqphlyDRodTdINUqrB2Q6VKwlfTEX0tpSr5FlB_0SiUayPWqeSIdUszhV7s9x9XKhHn6Q5lBsNlm4m6_QM4nZsP)](https://mermaid-js.github.io/mermaid-live-editor/edit#pako:eNpVkc1qwzAQhF9l0amF5AV8aGls5-dSShoIrZ3DYm0Sgb0K8opQ7Lx7ZcuBRCcx8w2zy3aqsppUok4OL2fYZSVDeB_F7kywd0YIttaLYTrAfP4Gi5fWnBhrcITa8AlaQSekX2NuMUJpNySFGJiukKHg-y366eD3n7aHrNijkUOUsxh7YCAvYrtYkDDKLv2CyjJTJcYyXAfPTel8DCy7nCvrOeikASMC5Jx19_blCPY_1PawKrZ0sU4icXgChvnWxbL27XkyVqPxW6S1bWnS1qO2ibX31jH03Lp5bH2Shp64tJqphlyDRodTdINUqrB2Q6VKwlfTEX0tpSr5FlB_0SiUayPWqeSIdUszhV7s9x9XKhHn6Q5lBsNlm4m6_QM4nZsP)


### 2.3 The Glue
_TODO(derrandz): Write this part._
