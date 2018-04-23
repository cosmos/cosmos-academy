# Tendermint BFT Consensus Algorithm

Tendermint is consistent PoS [BFT](https://en.wikipedia.org/wiki/Byzantine_fault_tolerance) \(Proof of Stake Bizantine Fault Tolerant\) consensus algorithm, meaning that it can tolerate up to 1/3 faulty nodes.

#### Actors in the Protocol

The participants in the protocol are called [**validators**](https://cosmos.network/staking/validators), who take turns in in proposing blocks of transactions and vote on them. These turns are assigned to the validators using [round-robin](https://en.wikipedia.org/wiki/Round-robin_scheduling) scheduling according to their total amount of stake.

Another key actor in the protocol are the [**delegators**](https://cosmos.network/staking/delegators). Delegators are token holders that do not want to run validator operations \(ie. invest in the necessary equipment to be able to participate in the consensus protocol\) and, as the name suggests, they delegate their stake tokens to validators, who charge a commission on the corresponding fees obtained by the delegated tokens.

#### Consensus Overview

One block per height is committed to the chain. Tendermint BFT works using three steps: **proposing**, **pre-voting** and **pre-committing.** The sequence `(Propose -> Pre-vote -> Pre-commit)`is called a **round**_. _

The ideal scenario is `NewHeight -> (Propose -> Pre-vote -> Pre-commit)+ -> Commit -> NewHeight ->...` and works in the following way:

1\) **Propose step**

* A validator node from the validator set is chosen as block proposer for a given height.

* She picks up transactions and packs them in a block. The proposed block is broadcasted to the rest of nodes.

2\) **Pre-Vote step**

* Each node casts a _**pre-vote**_ and listens until +2/3 pre-votes from validators' nodes have been submitted. A block can be either pre-voted as_ pre-vote_ \(ie. valid block\) or `nil` \(when it's invalid or timeout reached\). We generally call a _**Polka**_ when +2/3 of validator pre-vote for the same block. Also, in the perspective of a validator, if the she voted for the block that is referred in the Polka, she now has what's called a _**proof-of-lock-change**_ or **PoLC** for short in that particular height and round
`(H, R)`.

3\) **Pre-Commit step**

* Once a Polka is reached, validators submit a _**pre-commit**_ block, otherwise they _pre-commit_ `nil`. After they are broadcasted, they wait for +2/3 pre-commits from their peers.
* Finally, the proposed block is committed when validators submit more than +2/3 pre-commits \(aka _**Commit**_\) and a new block height is reached with a new selected block proposer. If not, the network performs a new round and the process starts from the beginning \(1\).

![](/assets/Screen Shot 2018-03-22 at 11.32.07 AM.png)

One important thing to consider is that these pre-votes are included in the next block as proof that the previous block was committed - they cannot be included in the current block, as that block has already been created.

#### Locking conditions during consensus voting

// Why locks are necessary

**Pre-Vote step**

First, if the validator is locked on a block since `LastLockRound` but now has a `PoLC` for something else at round `PoLC - Round` where `LastLockRound < PoLC - R < R`, then it unlocks.

If the validator is still locked on a block, it _pre-votes_ that block.

**Pre-commit**

If the validator has a `PoLC` at `(H, R)` for a particular block B, it \(re\)locks \(or changes lock to\) and _pre-commits_ B and sets `LastLockRound = R`.

Else if the validator has a `PoLC` at `(H, R)` for `nil`, it unlocks and _pre-commits_ `nil`.

Else, it keeps the lock unchanged and _pre-commits_ `nil`.

A pre-commit for `nil` means “I didn’t see a `PoLC` for this round, but I did get +2/3 pre-votes and waited a bit”.

#### New Rounds

When the consensus fails to commit the proposed block, a new round is required. Some examples for why this may happen are the following:

* The designated proposer was **not online**.
* The block proposed by the designated proposer was **not valid**.
* The block proposed by the designated proposer did not propagate in time \(i.e **timeout**\).
* The block proposed was valid, but +2/3 of pre-votes for the proposed block were not received in time for enough validator nodes by the time they reached the Precommit step. Even though +2/3 of pre-votes are necessary to progress to the next step, at least one validator may have voted _nil_ or maliciously voted for something else.
* The block proposed was valid, and +2/3 of pre-votes were received for enough nodes, but +2/3 of pre-commits for the proposed block were not received for enough validator nodes.

_Note: for a deeper look at _[_proofs_](https://tendermint.readthedocs.io/en/master/specification/byzantine-consensus-algorithm.html#proofs)_ and censorship attacks in Tendermint's BFT consensus please refer to the following _[_Specification_](https://tendermint.readthedocs.io/en/master/specification.html)_ section of the documentation._

####
