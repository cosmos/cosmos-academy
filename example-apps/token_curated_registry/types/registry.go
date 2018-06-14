package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Listing struct {
	Identifier string
	Votes int64
}

// Create new Voter for address on each Listing
type Voter struct {
	Owner sdk.Address
	Identifier string 
}

// Vote revealed during reveal phase
type Vote struct {
	Choice bool
	Power int64
}

type Ballot struct {
	Identifier string
	Owner sdk.Address
	Challenger sdk.Address
	Active bool
	Approve int64
	Deny int64
	Bond int64
	EndApplyBlockStamp int64
	EndCommitBlockStamp int64
	EndRevealBlockStamp int64
}