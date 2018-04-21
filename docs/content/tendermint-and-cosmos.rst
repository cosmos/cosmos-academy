Tendermint & Cosmos
===================

We're building a lot of software under two different GitHub organizations:

- `Tendermint <https://github.com/tendermint>`__ and,
- `Cosmos <https://github.com/cosmos>`__

Repositories are classified as either "Core" or "Secondary" where the former generally produce binaries or code consumable by users and the latter provide building blocks for the former. Note that the status of "Secondary" repos is subject to change (i.e., they into "Core" repos) as we consolidate over time.

- TODO, still missing a bunch of relevant public repos

Tendermint
----------

Core Repositories
~~~~~~~~~~~~~~~~~

blurb 

Tendermint
^^^^^^^^^^

Tendermint Core is Byzantine Fault Tolerant (BFT) middleware that takes a state transition machine - written in any programming language - and securely replicates it on many machines.

- `GitHub <https://github.com/tendermint/tendermint>`__
- `Read The Docs <http://tendermint.readthedocs.io/en/master/>`__
- `Specification <https://github.com/tendermint/tendermint/tree/master/docs/specification/new-spec>`__ see issue #33 about spec

ABCI
^^^^

ABCI is an interface that defines the boundary between the replication engine (the blockchain), and the state machine (the application). By using a socket protocol, we enable a consensus engine running in one process to manage an application state running in another.

- `GitHub <https://github.com/tendermint/abci>`__
- `Specification <https://github.com/tendermint/abci/blob/master/specification.rst>`__ see issue #33 about spec

Ethermint
^^^^^^^^^

Ethereum powered by Tendermint consensus

- `GitHub <https://github.com/tendermint/ethermint>`__
- `Read The Docs <http://ethermint.readthedocs.io/en/master/>`__

Secondary Repositories
~~~~~~~~~~~~~~~~~~~~~~

blurb

Go-amino
^^^^^^^^

Amino is an object encoding specification. Think of it as an object-oriented Protobuf3 with native JSON support and designed for blockchains (deterministic, upgradeable, fast, and compact).

- `GitHub <https://github.com/tendermint/go-amino>`__

IAVL
^^^^

A versioned and immutable AVL+ tree for persistent data.

- `GitHub <https://github.com/tendermint/iavl>`__

Cosmos
------

Core Repositories
~~~~~~~~~~~~~~~~~

Cosmos SDK
^^^^^^^^^^

The Cosmos SDK is a framework for building multi-asset Proof-of-Stake (PoS) blockchains on the Cosmos network.

- `GitHub <https://github.com/cosmos/cosmos-sdk>`__
- `Read The Docs <http://cosmos-sdk.readthedocs.io/en/master>`__
- `Specification <https://github.com/cosmos/cosmos-sdk/tree/master/docs/spec>`__ see issue #33 about spec

Voyager
^^^^^^^

Cosmos Voyager is the official user interface for the Cosmos Network and the Cosmos Hub.

- `GitHub <https://github.com/cosmos/voyager`__


Lotion JS
^^^^^^^^^

Lotion is a new way to create blockchain apps in JavaScript, which aims to make writing new blockchains fast and fun.

- `GitHub <https://github.com/keppel/lotion>`__
