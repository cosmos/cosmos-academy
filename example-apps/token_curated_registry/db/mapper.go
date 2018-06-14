package db

import (
	"github.com/cosmos/cosmos-sdk/x/bank"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-amino"
	"github.com/AdityaSripal/token_curated_registry/types"
)

type BallotMapper struct {
	ListingKey sdk.StoreKey

	CommitKey sdk.StoreKey

	RevealKey sdk.StoreKey

	BallotKey sdk.StoreKey

	Cdc *amino.Codec
}

func NewBallotMapper(listingKey sdk.StoreKey, ballotkey sdk.StoreKey, commitKey sdk.StoreKey, revealKey sdk.StoreKey, _cdc *amino.Codec) BallotMapper {
	return BallotMapper{
		ListingKey: listingKey,
		CommitKey: commitKey,
		RevealKey: revealKey,
		BallotKey: ballotkey,
		Cdc: _cdc,
	}
}

// Will get Ballot using unique identifier. Do not need to specify status
func (bm BallotMapper) GetBallot(ctx sdk.Context, identifier string) types.Ballot {
	store := ctx.KVStore(bm.BallotKey)
	key := []byte(identifier)
	val := store.Get(key)
	if val == nil {
		return types.Ballot{}
	}
	ballot := &types.Ballot{}
	err := bm.Cdc.UnmarshalBinary(val, ballot)
	if err != nil {
		panic(err)
	}
	return *ballot
}

func (bm BallotMapper) AddBallot(ctx sdk.Context, identifier string, owner sdk.Address, applyLen int64, bond int64) sdk.Error {
	store := ctx.KVStore(bm.BallotKey)

	newBallot := types.Ballot{
		Identifier: identifier,
		Owner: owner,
		Bond: bond,
		EndApplyBlockStamp: ctx.BlockHeight() + applyLen,
	}
	// Add ballot with Pending Status
	key := []byte(identifier)
	val, _ := bm.Cdc.MarshalBinary(newBallot)
	store.Set(key, val)
	return nil
}

func (bm BallotMapper) ActivateBallot(ctx sdk.Context, accountKeeper bank.Keeper, owner sdk.Address, challenger sdk.Address, identifier string, commitLen int64, revealLen, minBond int64, challengeBond int64) sdk.Error {
	store := ctx.KVStore(bm.BallotKey)
	ballot := bm.GetBallot(ctx, identifier)

	if ballot.Bond < minBond {
		bm.DeleteBallot(ctx, identifier)
		refund := sdk.Coin{
			Denom: "RegistryCoin",
			Amount: challengeBond,
		}
		_, _, err := accountKeeper.AddCoins(ctx, challenger, []sdk.Coin{refund})
		if err != nil {
			return err
		}
		return nil
	}
	if ballot.Bond != challengeBond {
		return sdk.NewError(2, 115, "Must match candidate's bond")
	}

	ballot.Active = true
	ballot.Challenger = challenger
	ballot.EndCommitBlockStamp = ctx.BlockHeight() + commitLen
	ballot.EndRevealBlockStamp = ballot.EndCommitBlockStamp + revealLen

	newBallot, _ := bm.Cdc.MarshalBinary(ballot)
	key:= []byte(identifier)
	store.Set(key, newBallot)

	return nil
}

func (bm BallotMapper) VoteBallot(ctx sdk.Context, owner sdk.Address, identifier string, vote bool, power int64) sdk.Error {
	ballotStore := ctx.KVStore(bm.BallotKey)

	ballotKey := []byte(identifier)
	bz := ballotStore.Get(ballotKey)
	if bz == nil {
		return sdk.NewError(2, 107, "Ballot does not exist")
	}
	ballot := &types.Ballot{}
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

	return nil
}

func (bm BallotMapper) DeleteBallot(ctx sdk.Context, identifier string) {
	key := []byte(identifier)
	store := ctx.KVStore(bm.BallotKey)
	store.Delete(key)
}

func (bm BallotMapper) AddListing(ctx sdk.Context, identifier string, votes int64) {
	key := []byte(identifier)
	store := ctx.KVStore(bm.ListingKey)

	listing := types.Listing{
		Identifier: identifier,
		Votes: votes,
	}
	val, _ := bm.Cdc.MarshalBinary(listing)

	store.Set(key, val)
}

func (bm BallotMapper) GetListing(ctx sdk.Context, identifier string) types.Listing {
	key := []byte(identifier)
	store := ctx.KVStore(bm.ListingKey)

	bz := store.Get(key)
	if bz == nil {
		return types.Listing{}
	}
	listing := &types.Listing{}
	err := bm.Cdc.UnmarshalBinary(bz, listing)
	if err != nil {
		panic(err)
	}
	
	return *listing
}

func (bm BallotMapper) DeleteListing(ctx sdk.Context, identifier string) {
	key := []byte(identifier)
	store := ctx.KVStore(bm.ListingKey)

	store.Delete(key)
}