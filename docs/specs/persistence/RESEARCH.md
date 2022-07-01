# Pocket 1.0 Persistence Module: Research Document: A framework to support the specification of the Pocket 1.0 Persistence Module.

<p align="center">
    Luis Correa de León<br> 
    @luyzdeleon<br>
    Version 1.0.1
</p>

# Introduction

Pocket Network aims to provide decentralized access to Web3 data via a blockchain network composed of multiple actors. This document aims to provide a decision making framework to support the specification of a Persistence Module that will satisfy the storage, retrieval, update and query needs of every actor in the Pocket Network by analyzing the Pocket Network persistence needs. 

By persistence we define any set of data that it’s necessary to a Pocket Network actor continuous participation in the network. The principal philosophy behind this effort is to replicate already existing production level architectures and technologies that have scaled persistent data services to billions of users throughout the world. 


# Scope

The persistence module of Pocket Network 1.0 touches on 3 different key areas: System Deployment, Middleware Persistence Logic and Blockchain State Validation Strategy. 


## System Deployment

The status quo in blockchain systems architecture is to utilize file system databases, where database engines are usually run in-process, this allows the systems to be self-contained, but other than that, is not the most scalable architecture. To allow for a separation of “middleware” and “database engine”, the Persistence Module should allow a “Client-Server” architecture. The advantages of the “Client-Server” architecture is that it separates the database engine logic from the system, and creates the flexibility that allows the system administrators to deploy to the same machine, to separate machines, or any other build in-between. 

This also conforms to the [Cattle Approach](https://medium.com/galvanize/pets-and-cattle-infrastructure-for-software-as-a-service-saas-7d386ec56c0c) to infrastructure, where treating infrastructure as Cattle and not Pets, you are able to leverage multiple properties that are desirable for production-grade infrastructure which is what we’re trying to achieve with Pocket Network’s node operations. Find below a comparative list of desired properties between both approaches:


<table>
  <tr>
   <td>Property
   </td>
   <td>In-Process
   </td>
   <td>Client-Server
   </td>
  </tr>
  <tr>
   <td><strong>Portability</strong>: Is the persistent data portable?
   </td>
   <td>TRUE
   </td>
   <td>TRUE
   </td>
  </tr>
  <tr>
   <td><strong>Individual Scalability</strong>: Can the middleware be independently scaled from the database engine.
   </td>
   <td>FALSE
   </td>
   <td>TRUE
   </td>
  </tr>
  <tr>
   <td><strong>Fault Tolerance</strong>: Can failures be isolated between the middleware and database engine?
   </td>
   <td>FALSE
   </td>
   <td>TRUE
   </td>
  </tr>
  <tr>
   <td><strong>Multi-Process Concurrency</strong>: Can multiple processes access the database engine concurrently?
   </td>
   <td>FALSE
   </td>
   <td>TRUE
   </td>
  </tr>
</table>



## Middleware Persistence Logic

When we talk about middleware, we refer to the business logic that’s being run by the Pocket Network software. We want to make sure the middleware persistence logic optimizes for Pocket Network’s inherent specific needs, and not the status quo in other blockchain networks. By doing this, we believe we can achieve the best possible result, by trading off less important attributes, for others that will rank higher in the Pocket Network “pyramid of persistence needs”. 


### Persistence Schema

To understand Pocket Network persistence needs, we need to understand the main **datasets **that will be persisted to, which are described below:



1. **Block store dataset**: Contains all the blocks, transactions and quorum certificates of the Pocket Network blockchain.
2. **Mempool dataset**: Contains a list of all transactions submitted to the Pocket Network.
3. **State dataset**: Contains the specific Pocket Network state (nodes, apps, params, accounts, etc). The state dataset has the particularity that each copy of the dataset is versioned at each **height **of the **Block store dataset. **The proof of the copy of each version of this dataset is what we will call from now on the **state hash**, in reference to a hashed version that’s included in each height of the **Block store dataset.**
4. **Utility dataset**: Contains all the utility specific data needed for the different actors of the network to achieve their functions. This dataset is local and only affects this individual node.


### Constraints



1. **State dataset **cannot be “completely hashed” at each round, as the size of the state is expected to grow bigger on each height.
2. **State dataset **must leave a trail of “verifications”, meaning every verified version of the database must persist.
3. All **datasets **must be exportable to a standardized format, and importable from that same standardized format.
4. The schema of each **dataset **must be versioned and migratable between versions.
5. All datasets must be verifiable for data integrity and be tolerant and self-healing to corruption.
6. All writes to persistence must be deterministic and “byte perfect consistency”.
7. Idempotent writes and updates to the database: 


### Desired Properties


<table>
  <tr>
   <td>Property
   </td>
   <td>Description
   </td>
   <td>Impacted Resource
   </td>
   <td>Answer to Constraint
   </td>
  </tr>
  <tr>
   <td><strong>State dataset</strong> versioning
   </td>
   <td>The way in which we verify every version of the <strong>State dataset</strong>.
   </td>
   <td>Compute, Memory, Storage
   </td>
   <td>1, 2
   </td>
  </tr>
  <tr>
   <td>“Byte-perfect consistency” data encoding
   </td>
   <td>The encoding mechanism of the persistent data.
   </td>
   <td>Compute and Memory
   </td>
   <td>6
   </td>
  </tr>
  <tr>
   <td>Schema definition mechanism
   </td>
   <td>The mechanism used to define the persistent data schema.
   </td>
   <td>Storage
   </td>
   <td>4
   </td>
  </tr>
  <tr>
   <td>Deterministic Write Mechanism
   </td>
   <td>A Mechanism that allows to roll-back faulty writes that might compromise the data integrity of any given dataset
   </td>
   <td>Compute and Storage
   </td>
   <td>5, 6
   </td>
  </tr>
  <tr>
   <td>Idempotent Dataset Updates
   </td>
   <td>The same update operation to a dataset, applied multiple times, must yield the same dataset state.
   </td>
   <td>Storage
   </td>
   <td>7
   </td>
  </tr>
</table>



## Blockchain State Validation Strategy

In the previous section, we referred to a desired property called “**State dataset** versioning”, to satisfy such property we require a **data representation data structure**, which allows us to represent a particular state of a any given dataset and compare it against copies of that same dataset, this is one of the core pillars of blockchain technology. In [Section 7 of the Bitcoin whitepaper](https://bitcoin.org/bitcoin.pdf) you will find a reference to a data-structure called a [Merkle Tree](https://en.wikipedia.org/wiki/Merkle_tree). This data-structure allows nodes in the Bitcoin network to only need to include the hash of the Root node of the Merkle Tree in the block header, this way the state can be recalculated by any third party and compared against the root. 

This section will compare apples to apples between different implementations and variations of the Merkle Tree, to try and find the one that satisfies the most desirable properties suitable for the Pocket Network implementation.

**A note on datasets: **For the Merkle Tree validation to be accepted, all leaf nodes in the tree must be homogeneous in their structure, so for this research we will consider every homologous collection of data to be a tree by itself.


### Desired Properties


<table>
  <tr>
   <td>Property
   </td>
   <td>Description
   </td>
   <td>Priority
   </td>
  </tr>
  <tr>
   <td>Active Open-Source Golang Implementation
   </td>
   <td>Actively maintained Golang implementation of the data-structure.
   </td>
   <td>1
   </td>
  </tr>
  <tr>
   <td>Insertion and update performance
   </td>
   <td>The theoretical and practical throughput capabilities for the <strong>insert </strong>and <strong>update </strong>operations<strong>.</strong>
   </td>
   <td>2
   </td>
  </tr>
  <tr>
   <td>Query performance
   </td>
   <td>The theoretical and practical throughput capabilities for the <strong>lookup </strong>operations.
   </td>
   <td>2
   </td>
  </tr>
  <tr>
   <td>Integrity Verification Performance
   </td>
   <td>The theoretical and practical throughput capabilities of the <strong>data integrity check </strong>operation.
   </td>
   <td>3
   </td>
  </tr>
  <tr>
   <td>Data deduplication
   </td>
   <td>The rate at which data is de-duplicated across multiple versions of the same dataset.
   </td>
   <td>3
   </td>
  </tr>
</table>



### Comparison between Patricia Merkle Tries, Merkle Bucket Trees and Pattern-Oriented-Split Trees.

A formal experiment was conducted and documented in the [Analysis of Indexing Structures for Immutable Data](https://arxiv.org/pdf/2003.02090.pdf) paper, which after theoretical review, seems to fit within our framework of desired properties for this data structure and goes into great detail comparing these 3 implementations which we will summarize in the comparative table found below. A simple lexicographic (A-F) scoring system aims to summarize each properties values for each implementation, creating a frame of reference upon which they can be easily compared.


<table>
  <tr>
   <td>Property
   </td>
   <td>Description
   </td>
   <td>Patricia Merkle Tries
   </td>
   <td>Merkle Bucket Trees
   </td>
   <td>Pattern-Oriented-Split Trees
   </td>
  </tr>
  <tr>
   <td>Active Open-Source Golang Implementation
   </td>
   <td>Actively maintained Golang implementation of the data-structure.
   </td>
   <td><a href="https://github.com/ethereum/go-ethereum/blob/a5a52371789f2e2a3144c8a842c4238ba92cf301/trie/trie.go">https://github.com/ethereum/go-ethereum/blob/a5a52371789f2e2a3144c8a842c4238ba92cf301/trie/trie.go</a>
   </td>
   <td>N/A (implementation is closed source)
   </td>
   <td>N/A (implementation is closed source)
   </td>
  </tr>
  <tr>
   <td>Insertion and update performance
   </td>
   <td>The theoretical and practical throughput capabilities for the <strong>insert </strong>and <strong>update </strong>operations<strong>.</strong>
   </td>
   <td>B
   </td>
   <td>C
   </td>
   <td>A
   </td>
  </tr>
  <tr>
   <td>Query performance
   </td>
   <td>The theoretical and practical throughput capabilities for the <strong>lookup </strong>operations.
   </td>
   <td>A-
   </td>
   <td>B
   </td>
   <td>A+
   </td>
  </tr>
  <tr>
   <td>Integrity Verification Performance
   </td>
   <td>The theoretical and practical throughput capabilities of the <strong>data integrity check </strong>operation.
   </td>
   <td>B
   </td>
   <td>C
   </td>
   <td>A
   </td>
  </tr>
  <tr>
   <td>Data deduplication
   </td>
   <td>The rate at which data is de-duplicated across multiple versions of the same dataset.
   </td>
   <td>A-
   </td>
   <td>C
   </td>
   <td>A+
   </td>
  </tr>
</table>


As a conclusion, both of the paper and our simple comparison table, in pure empirical benchmarking Pattern-Oriented-Split Trees (specifically the Forkbase implementation) seems to have the most desirable properties, however it’s implementation is closed source and not viable for the Pocket Network Persistence Module, which means the most suitable candidate between these 3 implementations is the **Patricia Merkle Trie**.


### Jellyfish Merkle Tree

The Jellyfish Merkle Tree is a data-structure created for the Diem project (formerly Libra), that aims to find the ideal balance between the complexity of data structure, storage, I/O overhead and computation efficiency, to cater to the needs of the Diem Blockchain. 

**Note on the table below: **There are no available benchmarking experiments on JMT itself, however thanks to this paper [On Performance Stability in LSM-based Storage Systems](https://www.vldb.org/pvldb/vol13/p449-luo.pdf) gives insight on it, given that JMT is theoretically superior to standard LSM implementations.


<table>
  <tr>
   <td>Property
   </td>
   <td>Description
   </td>
   <td>Score
   </td>
  </tr>
  <tr>
   <td>Active Open-Source Golang Implementation
   </td>
   <td>Actively maintained Golang implementation of the data-structure.
   </td>
   <td>Only exists in Rust (<a href="https://github.com/diem/diem/tree/master/storage/jellyfish-merkle">https://github.com/diem/diem/tree/master/storage/jellyfish-merkle</a>)
   </td>
  </tr>
  <tr>
   <td>Insertion and update performance
   </td>
   <td>The theoretical and practical throughput capabilities for the <strong>insert </strong>and <strong>update </strong>operations<strong>.</strong>
   </td>
   <td>WIP
   </td>
  </tr>
  <tr>
   <td>Query performance
   </td>
   <td>The theoretical and practical throughput capabilities for the <strong>lookup </strong>operations.
   </td>
   <td>WIP
   </td>
  </tr>
  <tr>
   <td>Integrity Verification Performance
   </td>
   <td>The theoretical and practical throughput capabilities of the <strong>data integrity check </strong>operation.
   </td>
   <td>WIP
   </td>
  </tr>
  <tr>
   <td>Data deduplication
   </td>
   <td>The rate at which data is de-duplicated across multiple versions of the same dataset.
   </td>
   <td>WIP
   </td>
  </tr>
</table>



# References



[https://bitcoin.org/bitcoin.pdf](https://bitcoin.org/bitcoin.pdf)

[https://medium.com/ontologynetwork/everything-you-need-to-know-about-merkle-trees-82b47da0634a](https://medium.com/ontologynetwork/everything-you-need-to-know-about-merkle-trees-82b47da0634a)

[https://developers.diem.com/papers/jellyfish-merkle-tree/2021-01-14.pdf](https://developers.diem.com/papers/jellyfish-merkle-tree/2021-01-14.pdf)

[https://arxiv.org/pdf/2003.02090.pdf](https://arxiv.org/pdf/2003.02090.pdf)

[https://medium.com/galvanize/pets-and-cattle-infrastructure-for-software-as-a-service-saas-7d386ec56c0c](https://medium.com/galvanize/pets-and-cattle-infrastructure-for-software-as-a-service-saas-7d386ec56c0c)

[https://www.vldb.org/pvldb/vol13/p449-luo.pdf](https://www.vldb.org/pvldb/vol13/p449-luo.pdf)
