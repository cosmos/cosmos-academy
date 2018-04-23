#
IBC Protocol

**Contents**

* Fundamentals of interoperability
* IBC overview
* Cosmos Hub, relaying IBC messages between chains
* Proofs
* Messages
* Security: handling byzantine failure
* Register chains
* Sending and relaying packets
* Handling validator set update
* Verify block headers
* tx types for IBC
* Packet delivery cycle



## Fundamentals of Interoperability

Any header may be from a malicious chain \(eg. shadowing a real chain id with a fake validator set\), so a subjective decision is required before establishing a connection. This should be performed by on-chain governance to avoid an exploitable position of trust. Establishing a bidirectional root of trust between two blockchains \(A trusts B and B trusts A\) is a necessary and sufficient prerequisite for all other IBC activity.





The Inter Blockchain Communication \(IBC\) Protocol is the the one that allows interoperability between blockchain zones. It is basically packets of information that are communicated from one zone to another by posting Merkle-proofs as evidence that the information was sent and received.

For the receiving chain to check this proof, it must be able keep up with the sender’s block headers. This mechanism is similar to that used by [sidechains](https://blockstream.com/technology/sidechains.pdf), which requires two interacting chains to be aware of one another via a bidirectional stream of proof-of-existence datagrams \(transactions\).

The IBC protocol can naturally be defined using two types of transactions: an `IBCBlockCommitTx` transaction, which allows a blockchain to prove to any observer of its most recent block-hash, and an `IBCPacketTx` transaction, which allows a blockchain to prove to any observer that the given packet was indeed published by the sender’s application, via a Merkle-proof to the recent block-hash.

## IBC connection

Two chains that communicate over IBC need to overlook at each other in some way to prevent double spending.

We define the following variables: Hh is the signed header at height h, Ch are the consensus rules at height h, and P is the unbonding period of this blockchain. Vk,h is the value stored under key k at height h. Note that of all these, only Hh defines a signature and is thus attributable.

To support an IBC connection, two actors must be able to make the following proofs to each other:

given a trusted Hh and Ch and an attributable update message Uh’ it is possible to prove Hh’ where Ch’ = Ch and \(now, Hh\) &lt; P

given a trusted Hh and Ch and an attributable change message Xh’ it is possible to prove Hh’ where Ch’ Ch and \(now, Hh\) &lt; P

given a trusted Hh and a merkle proof Mk,v,h it is possible to prove Vk,h

The merkle proof Mk,v,h is a well-defined concept in the blockchain space, and provides a compact proof that the key value pair \(k, v\) is consistent with a merkle root stored in Hh.

The IBC protocol requires each actor to be a blockchain with complete block finality. All transitions must be provable and attributable to \(at least\) one actor.

Any header may be from a malicious chain \(eg. shadowing a real chain id with a fake validator set\), so a subjective decision is required before establishing a connection. This should be performed by on-chain governance to avoid an exploitable position of trust. Establishing a bidirectional root of trust between two blockchains \(A trusts B and B trusts A\) is a necessary and sufficient prerequisite for all other IBC activity.

## Cosmos Hub: Relaying IBC messages between chains

We define the concept of a **relay process** that connects two chain by querying one for all proofs needed to prove outgoing messages and submit these proofs to the recipient chain.

The relay process must have access to accounts on both chains with sufficient balance to pay for transaction fees but needs no other permissions.

**Receipts**

When we wish to create a transaction that atomically commits or rolls back across two chains, we must look at the receipts from sending the original message. For example, if I want to send tokens from Alice on chain A to Bob on chain B, chain A must decrement Alice’s account if and only if Bob’s account was incremented on chain B.

receipts are stored in a queue with the same key construction as the sending queue, we can generate the same set of proofs for them, and perform a similar sequence of steps to handle a receipt coming back to S for a message previously sent to A:

![](https://lh5.googleusercontent.com/GHnWKSZpIANP0k30OpKtKLYPlHpQ0zscoqDDiJ8nKUzDmImIsrO92kp_QkpHQajKtSMLmPYeRy4Fg8vkX57ozFTbcsv97wOdmypvUibRY_pW8chqeeerVNVhg0ai8w0bLPXy8Sf8 "Receipts\(2\).png")

![](https://lh5.googleusercontent.com/Px-GOQ1fHNCkgKMcEdcWmyiOk3wQTUA0cDxLN-JTHP1p2AYMJkpY5XL8vdOG2zrR1kmBaAO5L3glNZlEudyi7u84vIwAx5pKkJPnFpkgluTcFmRtzU1RMY9b5FSjIbYcZ2N9IKwa "Receipt Error\(1\).png")



VERIFYING HEADERS

Once we have a trusted header with a known validator set, we can quickly validate any new header with the same validator set. To validate a new header, simply verifying that the validator hash has not changed, and that over 2/3 of the voting power in that set has properly signed a commit for that header. We can skip all intervening headers, as we have complete finality \(no forks\) and accountability \(to punish a double-sign\).

This is safe as long as we have a valid signed header by the trusted validator set that is within the unbonding period for staking. In that case, if we were given a false \(forked\) header, we could use this as proof to slash the stake of all the double-signing validators. This demonstrates the importance of attribution and is the same security guarantee of any non-validating full node. Even in the presence of some ultra-powerful malicious actors, this makes the cost of creating a fake proof for a header equal to at least one third of all staked tokens, which should be significantly higher than any gain of a false message.

UPDATING VALIDATORS SET

If the validator hash is different than the trusted one, we must simultaneously both verify that if the change is valid while, as well as use using the new set to validate the header. Since the entire validator set is not provided by default when we give a header and commit votes, this must be provided as extra data to the certifier.

A validator change in Tendermint can be securely verified with the following checks:

First, that the new header, validators, and signatures are internally consistent

We have a new set of validators that matches the hash on the new header

At least 2/3 of the voting power of the new set validates the new header

Second, that the new header is also valid in the eyes of our trust set

Verify at least 2/3 of the voting power of our trusted set, which are also in the new set, properly signed a commit to the new header

In that case, we can update to this header, and update the trusted validator set, with the same guarantees as above \(the ability to slash at least one third of all staked tokens on any false proof\).
