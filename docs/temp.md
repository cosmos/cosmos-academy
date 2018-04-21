# Welcome to the Tendermint Team! 👋

![Gif]()

So you've been hired to help us build software... cool! Welcome! This is a document to help you get oriented. Please contact @jolesbi for any questions or concerns.

Before you dive in to any of the codebases - take some time to watch these videos about the scope and vision of this project.

## Good Intro Videos

* Researcher Sunny Agarwal — [Many Chains, One Ecosystem](https://www..com/watch?v=LApEkXJR_0M)
* CTO Ethan Buchman — [What is Cosmos](https://www.youtube.com/watch?v=QExyiPjC3b8)
* CEO Jae Kwon — [Crypto Economics](https://www.youtube.com/watch?v=8Eex-wQ5yYU)

## A Bunch Of Software

We're building a lot of software under two different GitHub organizations, [Tendermint](https://github.com/tendermint) and [Cosmos](https://github.com/cosmos/). There are many important repos in each of these organizations but below is a short list of some of the ones you may be interacting or hearing about more frequently.

### Tendermint GitHub Organization

#### [Tendermint Core](https://github.com/tendermint/tendermint)

Tendermint Core is Byzantine Fault Tolerant (BFT) middleware that takes a state transition machine - written in any programming language - and securely replicates it on many machines.

#### [ABCI](https://github.com/tendermint/abci)

ABCI is an interface that defines the boundary between the replication engine (the blockchain), and the state machine (the application). By using a socket protocol, we enable a consensus engine running in one process to manage an application state running in another.

#### [Ethermint](https://github.com/tendermint/ethermint)

Ethereum powered by Tendermint consensus

#### [Go-amino](https://github.com/tendermint/go-amino)

Amino is an object encoding specification. Think of it as an object-oriented Protobuf3 with native JSON support and designed for blockchains (deterministic, upgradeable, fast, and compact).

#### [IAVL](https://github.com/tendermint/iavl)

A versioned and immutable AVL+ tree for persistent data.

### Cosmos GitHub Organization

#### [Cosmos SDK](https://github.com/cosmos/cosmos-sdk)

The Cosmos SDK is a framework for building multi-asset Proof-of-Stake (PoS) blockchains on the Cosmos network.

#### [Voyager](https://github.com/cosmos/voyager)

Cosmos Voyager is the official user interface for the Cosmos Network and the Cosmos Hub.

---

#### [Lotion JS](https://github.com/keppel/lotion)

Lotion is a new way to create blockchain apps in JavaScript, which aims to make writing new blockchains fast and fun.
