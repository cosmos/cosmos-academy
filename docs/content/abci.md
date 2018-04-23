# ABCI Protocol

The purpose of ABCI \(**A**pplication **B**lock**C**hain **I**nterface\) is to provide a clean interface between any finite, deterministic state transition machine on one computer and the mechanics of a blockchain-based replication engine across multiple computers \(aka consensus engine, which involves the consensus and networking layers\).

## ABCI Design components

The ABCI design has some distinct components: Message protocol, Server, Client and the Blockchain protocol.

### Message Protocol

Consists in pairs of requests and responses. For each request, a server should respond with the corresponding response, where order of requests is preserved in the order of responses.

### Server

// Complete

ABCI server that supports 3 multiple connections. There are several implementations of the ABCI server in [Go](https://github.com/tendermint/abci/tree/master/server), [JavaScript](https://github.com/tendermint/js-abci), [Python](https://github.com/tendermint/abci/tree/master/example/python3/abci), [C++](https://github.com/mdyring/cpp-tmsp) and [Java](https://github.com/jTendermint/jabci).

### Client

For the client, there are two main use cases: `abci-cli`, which is a testing tool that allows ABCI requests from the command line, and Tendermint Core, which makes requests to the application every time a new transaction is received or a block is committed.

### Blockchain protocol

––––– REWRITE

In ABCI, a transaction is simply an arbitrary length byte-array. It is the application's responsibility to define the transaction codec as they please, and to use it for both `CheckTx` and `DeliverTx`.

Note that there are two distinct means for running transactions, corresponding to stages of ‘awareness’ of the transaction in the network. The first stage is when a transaction is received by a validator from a client into the so-called mempool or transaction pool - this is where we use `CheckTx`. The second is when the transaction is successfully committed on more than 2/3 of validators - where we use `DeliverTx`. In the former case, it may not be necessary to run all the state transitions associated with the transaction, as the transaction may not ultimately be committed until some much later time, when the result of its execution will be different. For instance, an Ethereum ABCI app would check signatures and amounts in `CheckTx`, but would not actually execute any contract code until the `DeliverTx`, so as to avoid executing state transitions that have not been finalized.

–––––––

The first time a new blockchain is started, Tendermint calls`InitChain`. From then on, the Block Execution Sequence that causes the committed state to be updated is as follows:

`BeginBlock -> [DeliverTx] -> EndBlock -> Commit`

where one`DeliverTx`is called for each transaction in the block. Cryptographic commitments to the results of `DeliverTx`, EndBlock, and Commit are included in the header of the next block.

#### ABCI connections

Tendermint opens three connections to the application interface to handle the different message types:

* **Consensus Connection:** used only when a new block is committed, and communicates all information from the block in a series of requests \(`BeginBlock -> [DeliverTx] -> EndBlock -> Commit`\).
* Message types: `InitChain`, `BeginBlock`, `DeliverTx`, `EndBlock` and `Commit`
* **Mempool Connection: **only for `CheckTx` requests. Transactions are run using `CheckTx` in the same order they were received by the validator. If the `CheckTx` returns OK, the transaction is kept in memory and relayed to other peers in the same order it was received. Otherwise, it is discarded.
* Message types: `CheckTx`
* **Info \(Query\) Connection:** to query the local state of the application without engaging consensus. Tendermint Core currently uses the Query connection to filter peers upon connecting, according to IP address or public key.
* Message types:`Info`, `SetOption`, `Query`

There are also two message types that are not categorized: the`Flush`message that is used on every connection, and the`Echo`message, which is only used for debugging.

Note that messages may be sent concurrently across all connections - a typical application will thus maintain a distinct state for each connection. They may be referred to as the `DeliverTx` state, the `CheckTx` state, and the `Commit` state respectively.

––––– REWRITE

// The application should maintain a separate state to support `CheckTx`. This state can be reset to the latest committed state during `Commit`, where Tendermint ensures the mempool is locked and not sending new `CheckTx`. After `Commit`, the mempool will rerun `CheckTx` on all remaining transactions, throwing out any that are no longer valid.

#### Message types

The ABCI consists of 3 primary message types that get delivered from the blockchain engine to the application. The application replies with corresponding response messages.

1. `DeliverTx`: Delivers each transaction from the blockchain. The Application layer checks and validate each transaction received with the `DeliverTx` against the current state, the implemented application protocol, and the cryptographic credentials of the transaction itself \(i.e sender's signature\). Once validated, the tx updates the current state.
2. `CheckTx`: Used to validate mempool transactions before broadcasting or proposing. It performs stateful but light-weight checks of the validity of the transaction \(like checking signatures and account balances\).
3. `Commit`: Used to compute a cryptographic commitment to the current application state. This message returns a Merkle-hash proof that is placed into the next block header. This also simplifies the development of secure lightweight clients, as Merkle-hash proofs can be verified by checking against the block-hash, and the block-hash is signed by a quorum of validators \(by voting power\). The mempool is locked for processing a `Commit` request so that its state can be safely reset during `Commit`. The remaining transactions in the mempool are replayed on the mempool connection \(`CheckTx`\) following a commit.

The additional ABCI messages allow the application to keep track of and change the validator set, and for the application to receive the application information, the block information \(_i.e._ height and the commit votes\), set application option, query for state, etc. For a further look of the ABCI message types you can check the message [schema](https://github.com/tendermint/abci#message-types).

#### Handshake

When the app or Tendermint restarts, they need to sync to a common height. When an ABCI connection is first established, Tendermint will call Info on the Query connection. The response should contain the `LastBlockHeight` and `LastBlockAppHash` - the former is the last block for which the app ran `Commit` successfully, the latter is the response from that `Commit`.

Using this information, Tendermint will determine what needs to be replayed, if anything, against the app, to ensure both Tendermint and the app are synced to the latest block height.
