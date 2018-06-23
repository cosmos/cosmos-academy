# token_curated_registry
Token-curated registry built with Cosmos sdk

Design largely inspired by: https://medium.com/@ilovebagels/token-curated-registries-1-0-61a232f8dac7

There are 4 types of Messages in this app. All bonds/deposits are made in RegistryCoin.

DeclareCandidacyMsg: Declare candidacy for a new listing

```go
// DeclareCandidacyMsg is used to propose a new listing to be added to registry.
// Identifier is a unique identifier of the listing
// Users can add details about candidate to convince voters to approve listing
// Deposit is taken and held for entire duration of listing. Awarded to challengers upon successful challenge.
type DeclareCandidacyMsg struct {
	Owner      sdk.Address
	Identifier string
	Details string
	Deposit       sdk.Coin
}
```

One can challenge a listing either during its candidate phase or even after it has been added to registry.
To challenge, a user must match the candidate's deposit by placing a bond. If a listing was added with a
bond smaller than the current minimum bond, it can be removed automatically by challenging with a minimum
bond.

```go
// ChallengeMsg is used to challenge a pending or finalized listing
type ChallengeMsg struct {
	Owner      sdk.Address
	Identifier string
	Bond       sdk.Coin
}
```

If a candidate has been challenged, users can make commitments before the reveal phase starts. A commitment is
a hash of the user's vote and a nonce.

```go
// CommitMsg is used to make a commitment during commit phase on an active challenge to a specific listing identified by Identifier.
type CommitMsg struct {
	Owner      sdk.Address
	Identifier string
	Commitment []byte
}
```

Users can generate commitments by doing the following:

```go
hasher := sha256.New()
vz, _ := cdc.MarshalBinary(vote)
hasher.Sum(vz)

hasher2 := sha256.New()
bz, _ := cdc.MarshalBinary(rand.Int())
nonce = hasher2.Sum(bz)

commitment := hasher.Sum(nonce)
```

Once the reveal phase starts, users can reveal their previously submitted commitments by revealing their vote and nonce.

```go
// RevealMsg is to reveal vote during reveal phase on active challenge to listing identified by Identifier.
type RevealMsg struct {
	Owner      sdk.Address
	Identifier string
	Vote       bool
	Nonce      []byte
	Bond       sdk.Coin
}
```

If the vote and nonce hash to the previously submitted commitment, the ballot gets updated with the user's vote.
The vote is incremented by the Bond amount.


Once the reveal phase ends, the ballot result will be finalized and added/removed from the registry as needed.
The bond posted by the losing side (either the challenger or candidate), gets distributed amongst the winners of the vote.

The counterparty (challenge is counterparty to candidate and vice-versa) gets the bond multiplied by the dispensation pct.

A winning voter gets his bond back along with a reward = (1 - dispensationPct) * bond * (voter.Power / total_power)

Losing voters get their bond back with no reward.
