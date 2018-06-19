package db

import (
	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-amino"
)

type BallotMapper struct {
	BallotKey sdk.StoreKey

	Cdc *amino.Codec
}

func NewBallotMapper(ballotkey sdk.StoreKey, _cdc *amino.Codec) BallotMapper {
	return BallotMapper{
		BallotKey:  ballotkey,
		Cdc:        _cdc,
	}
}

// Will get Ballot using unique identifier. Do not need to specify status
func (bm BallotMapper) GetBallot(ctx sdk.Context, identifier string) tcr.Ballot {
	store := ctx.KVStore(bm.BallotKey)
	key := []byte(identifier)
	val := store.Get(key)
	if val == nil {
		return tcr.Ballot{}
	}
	ballot := &tcr.Ballot{}
	err := bm.Cdc.UnmarshalBinary(val, ballot)
	if err != nil {
		panic(err)
	}
	return *ballot
}

func (bm BallotMapper) AddBallot(ctx sdk.Context, identifier string, owner sdk.Address, applyLen int64, bond int64) sdk.Error {
	store := ctx.KVStore(bm.BallotKey)

	newBallot := tcr.Ballot{
		Identifier:         identifier,
		Owner:              owner,
		Bond:               bond,
		EndApplyBlockStamp: ctx.BlockHeight() + applyLen,
	}
	// Add ballot with Pending Status
	key := []byte(identifier)
	val, _ := bm.Cdc.MarshalBinary(newBallot)
	store.Set(key, val)
	return nil
}

func (bm BallotMapper) ActivateBallot(ctx sdk.Context, owner sdk.Address, challenger sdk.Address, identifier string, commitLen int64, revealLen, minBond int64, challengeBond int64) sdk.Error {
	store := ctx.KVStore(bm.BallotKey)
	ballot := bm.GetBallot(ctx, identifier)

	if ballot.Bond != challengeBond {
		return tcr.ErrInvalidBallot(2, "Must match candidate's bond")
	}

	ballot.Active = true
	ballot.Challenger = challenger
	ballot.EndCommitBlockStamp = ctx.BlockHeight() + commitLen
	ballot.EndApplyBlockStamp = ballot.EndCommitBlockStamp + revealLen

	newBallot, _ := bm.Cdc.MarshalBinary(ballot)
	key := []byte(identifier)
	store.Set(key, newBallot)

	return nil
}

func (bm BallotMapper) CommitBallot(ctx sdk.Context, owner sdk.Address, identifier string, commitment []byte) {
	commitKey := []byte(identifier)
	commitKey = append(commitKey, []byte("commits")...)
	commitKey = append(commitKey, owner...)

	store := ctx.KVStore(bm.BallotKey)

	store.Set(commitKey, commitment)
}

func (bm BallotMapper) VoteBallot(ctx sdk.Context, owner sdk.Address, identifier string, vote bool, power int64) sdk.Error {
	ballotStore := ctx.KVStore(bm.BallotKey)

	ballotKey := []byte(identifier)

	// Set Vote with key "identifier|votes|<owner_address>"
	voteKey := append(ballotKey, []byte("votes")...)
	voteKey = append(ballotKey, owner...)

	bz := ballotStore.Get(ballotKey)
	if bz == nil {
		return tcr.ErrInvalidBallot(2, "Ballot does not exist")
	}
	ballot := &tcr.Ballot{}
	err := bm.Cdc.UnmarshalBinary(bz, ballot)
	if err != nil {
		panic(err)
	}
	if vote {
		ballot.Approve += power
	} else {
		ballot.Deny += power
	}
	newBallot, _ := bm.Cdc.MarshalBinary(*ballot)
	ballotStore.Set(ballotKey, newBallot)

	voteStruct := tcr.Vote{Choice: vote, Power: power}
	val, _ := bm.Cdc.MarshalBinary(voteStruct)
	ballotStore.Set(voteKey, val)

	return nil
}

func (bm BallotMapper) DeleteBallot(ctx sdk.Context, identifier string) {
	key := []byte(identifier)
	store := ctx.KVStore(bm.BallotKey)
	store.Delete(key)
}

func (bm BallotMapper) GetCommitment(ctx sdk.Context, identifier string, owner sdk.Address) []byte {
	commitKey := []byte(identifier)
	commitKey = append(commitKey, []byte("commits")...)
	commitKey = append(commitKey, owner...)

	store := ctx.KVStore(bm.BallotKey)

	return store.Get(commitKey)
}

func (bm BallotMapper) GetVote(ctx sdk.Context, identifier string, owner sdk.Address) tcr.Vote {
	voteKey := []byte(identifier)
	voteKey = append(voteKey, []byte("votes")...)
	voteKey = append(voteKey, owner...)

	store := ctx.KVStore(bm.BallotKey)

	bz := store.Get(voteKey)
	if bz == nil {
		return tcr.Vote{}
	}

	vote := &tcr.Vote{}
	bm.Cdc.UnmarshalBinary(bz, vote)

	return *vote
}