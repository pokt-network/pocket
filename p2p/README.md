# P2P Module

_TODO(derrandz): Add more diagrams_

Welcome to the P2P Module architecture guide.

## 1. Introduction: A word or two about the P2P module

We think that it's beneficial to first off set the context properly before we start diving into how this module is architected, so this introduction should help better-understand the decisions behind architecting P2P the way it is.

P2P as a module will primarily deal with:

- sending data
- receiving data
- broadcasting data

To:

- specific peers
- the entire network

A peer might be involved in multiple processes at once (e.g. sending to 10 peers, receiving from 20, and broadcasting to everyone else) regardless of where these sends are coming from or what will be happening with the received data. Thus, it's crucial that the P2P module be architected in a concurrency-friendly way.

One can immediately see that a peer should be able to concurrently:

1. Connect to multiple peers
2. Receive connections from multiple peers
3. Read incoming data from connected peers
4. Send outgoing data to connected peers

Therefore, we should at least have 3 "types" long running routines:

- A connections establishment routine (_will be termed **listening** routine from now onwards_)
- A read-from-connection routine (_will be termed **read-routine** from now onwards_)
- A write-to-connection routine (_will be termed **write-routine** from now onwards_)

To further help paint a visual image of this, imagine a peer Y being connected to `5 other peers`. Peer Y will have a total `11 routines` running along each other, mainly:

1. A listening routine that accepts incoming connections (`1 routine`)
2. A read-routine and a write-routine for each connected peer (`2x5 routines`)

Now the main question is, how do we achieve this and how do we write it in code, in a "proper" way?

## 2. Architecture

The p2p module is composed of two main sub-modules, the network module and the socket module in a hierarchical manner such that the network module controls, runs and commands the socket module.

The network module is responsible for:
  - listening for incoming connections
  - dialing outgoing connections
  - broadcasting to the network
  - sending to specific peers

The network module abstract away the connections it establishes using the socket module. In turn, the socket module takes care of the established connections' authentication, IO and error handling.

The network module keeps all opened sockets (_i.e: established connections_) mapped in a an "active connections pool".
The network module also stores a map of all network participants' addresses.

### **2.1 Socket**

A socket has the following attributes:
* a type (_inbound or outbound_)
* an address (_the address of the peer at the other end of the socket_)
* an ID (_the ID of the peer at the other end of the socket_)
* a reference to the TCP socket (_net.Conn_)
* general IO parameters (_timeouts, buffer sizes_)

A socket has the following possible states:



[![](https://mermaid.ink/img/pako:eNp9kLEKwjAQhl-l3Cjt4tjBRTsLOjgYh9Bca6BJ5HoRpPTdTZtUxII3_Xz38cPdALVTCCX0LBkPWrYkTfHcCpuFuW5uWVHssn3nelSRxTzj4wNthFOa0Qml0rb9oRfS_KFJWRV_84rIUcRzXMmp8S9PJXGxDORgkIzUKhw9TDsBfEeDAsoQFTbSdyxA2DGo_qHCWyql2RGUjex6zEF6dueXraFk8rhI6XfJGt8fuWrI)](https://mermaid-js.github.io/mermaid-live-editor/edit#pako:eNp9kLEKwjAQhl-l3Cjt4tjBRTsLOjgYh9Bca6BJ5HoRpPTdTZtUxII3_Xz38cPdALVTCCX0LBkPWrYkTfHcCpuFuW5uWVHssn3nelSRxTzj4wNthFOa0Qml0rb9oRfS_KFJWRV_84rIUcRzXMmp8S9PJXGxDORgkIzUKhw9TDsBfEeDAsoQFTbSdyxA2DGo_qHCWyql2RGUjex6zEF6dueXraFk8rhI6XfJGt8fuWrI)
_Socket states (_state diagram_)_

A socket achieves IO by running two main IO routines:
* A read routine.
* A write routine.

#### 2.1.1 The Socket's Read Routine

The socket 'read' routine performs buffered read operations in incoming data.
To achieve this, the socket makes available a byte slice representing the buffer, and reads into it using an `io.Reader` off of the underlying TCP connection (_net.Conn_).

The read routine continually waits on new data to be received, and reads it in a buffer manner by relying on the `readChunk` operation. We can say that the read routine reads in chunks of size `BufferSize`. This is configurable.
The `readChunk` operation performs minimal validation against the read chunk to make sure it is of an acceptable size, and whether it's not corrupted (_decoding does not fail_).
If this validation fails, the chunk is rejected and socket is kept open.

The read routine will not close on faulty chunks, but will close on:
 - `io.EOF` error
 - Unexpected errors
 - Graceful shutdown caused by:
   - The closing of the socket's context
   - The runner has stopped (_i.e: the network module_)
   - The socket has been terminated using the `Close` operation.
   - 

Here is a flow diagram summarizing the read routien operations:

[![](https://mermaid.ink/img/pako:eNpVkttuwjAMhl_Fys02ib0AF5sGLeIgGCtMaGq5iBoD0UrSOckYKrz70hPrctPE_vL7t5uCpVog67M98fwA6yBR4NdLvD4gRMgFRNpZqXALj49PMLg3cq94BuRTUu3BWE4WxUN9bVBBwyLCFOU3ClB4AqlyZ5-vNTEsictCXyCIN1zabR0O6otd5gPNBcIiVKl2yiJ5tbIoIJGmVi68yY3iAMtWYCMJYcnPmeZi26UqwXExMfCu8CfH1PuGkAg0Qfg6aiXHlZWaXsQrq3OgagT77T-gLDqPI8w1tV0sqsxbPMy0wSY273Y2qg6T0kOEXw7NbS6TTtVpvHTmAEKfFFgN1v-JxgGc_MjK7857ruJocq1uxSZ_1mZdkZJ0SiHdGTBSfTb4tOtt1h5Yjx2RjlwK_yyKMpkwr3DEhPX9VuCOu8wmLFFXj7pccIuhkFYT6-94ZrDHuLN6dVYp61ty2EKB5P6VHRvq-gseTMfY)](https://mermaid-js.github.io/mermaid-live-editor/edit#pako:eNpVkttuwjAMhl_Fys02ib0AF5sGLeIgGCtMaGq5iBoD0UrSOckYKrz70hPrctPE_vL7t5uCpVog67M98fwA6yBR4NdLvD4gRMgFRNpZqXALj49PMLg3cq94BuRTUu3BWE4WxUN9bVBBwyLCFOU3ClB4AqlyZ5-vNTEsictCXyCIN1zabR0O6otd5gPNBcIiVKl2yiJ5tbIoIJGmVi68yY3iAMtWYCMJYcnPmeZi26UqwXExMfCu8CfH1PuGkAg0Qfg6aiXHlZWaXsQrq3OgagT77T-gLDqPI8w1tV0sqsxbPMy0wSY273Y2qg6T0kOEXw7NbS6TTtVpvHTmAEKfFFgN1v-JxgGc_MjK7857ruJocq1uxSZ_1mZdkZJ0SiHdGTBSfTb4tOtt1h5Yjx2RjlwK_yyKMpkwr3DEhPX9VuCOu8wmLFFXj7pccIuhkFYT6-94ZrDHuLN6dVYp61ty2EKB5P6VHRvq-gseTMfY) _A Socket's read routine (_flow diagram_)_

### 2.3 The Glue

_TODO(derrandz): Write this part._
