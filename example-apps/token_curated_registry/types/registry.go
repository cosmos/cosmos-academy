package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Listing struct {
	Identifier string
	Votes      int64
}

func NewListing(identifier string, votes int64) Listing {
	return Listing{
		Identifier: identifier,
		Votes: votes,
	}
}

// Vote revealed during reveal phase
type Vote struct {
	Choice bool
	Power  int64
}

func NewVote(choice bool, power int64) Vote {
	return Vote{
		Choice: choice,
		Power: power,
	}
}

type Ballot struct {
	Identifier          string
	Details             string
	Owner               sdk.Address
	Challenger          sdk.Address
	Active              bool
	Approve             int64
	Deny                int64
	Bond                int64
	EndApplyBlockStamp  int64
	EndCommitBlockStamp int64
}

func NewBallot(identifier string, details string, owner sdk.Address, endApplyBlock int64) Ballot {
	return Ballot{
		Identifier: identifier,
		Details: details,
		Owner: owner,
		EndApplyBlockStamp: endApplyBlock,
	}
}


