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


