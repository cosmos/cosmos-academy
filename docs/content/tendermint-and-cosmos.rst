Tendermint & Cosmos
===================

We are building a lot of software under two different GitHub organizations:

- `Tendermint <https://github.com/tendermint>`__ and,
- `Cosmos <https://github.com/cosmos>`__

Repositories are classified as either "Core" or "Secondary" where the former generally produce binaries or code consumable by users and the latter provide building blocks for the former. Note that the status of "Secondary" repos is subject to change (i.e., they into "Core" repos) as we consolidate over time.

Read The Docs
-------------

Three core repositories have their documentation hosted on individual `Read The Docs <https://readthedocs.org/>`__ sites. This allows the docs to be versioned alongside the code itself. As well, each of these projects is sufficiently independent to merit their own documentation.

- http://tendermint.readthedocs.io/en/master/
- http://ethermint.readthedocs.io/en/master/
- http://cosmos-sdk.readthedocs.io/en/master/

Docs are built from the ``docs/`` directory in a project's respective repository. From within that directory you can run ``make html`` to build the docs then ``open _build/html/index.html`` to browse them locally.

Tendermint
----------

Core Repositories
~~~~~~~~~~~~~~~~~

Tendermint
^^^^^^^^^^

Tendermint Core is Byzantine Fault Tolerant (BFT) middleware that takes a state transition machine - written in any programming language - and securely replicates it on many machines.

- `GitHub <https://github.com/tendermint/tendermint>`__
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

Secondary Repositories
~~~~~~~~~~~~~~~~~~~~~~

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
- `Specification <https://github.com/cosmos/cosmos-sdk/tree/master/docs/spec>`__ see issue #33 about spec

Voyager
^^^^^^^

Cosmos Voyager is the official user interface for the Cosmos Network and the Cosmos Hub.

- `GitHub <https://github.com/cosmos/voyager>`__


Lotion JS
^^^^^^^^^

Lotion is a new way to create blockchain apps in JavaScript, which aims to make writing new blockchains fast and fun.

- `GitHub <https://github.com/keppel/lotion>`__
