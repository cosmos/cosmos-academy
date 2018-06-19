package db

import (
	"github.com/cosmos/cosmos-sdk/x/bank"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tcr "github.com/AdityaSripal/token_curated_registry/types"
)

type BallotKeeper struct {
	bm BallotMapper
	ack bank.Keeper

	listingKey sdk.StoreKey
}

func NewBallotKeeper(_bm BallotMapper, _ack bank.Keeper, listKey sdk.StoreKey) BallotKeeper {
	return BallotKeeper{
		bm: _bm,
		ack: _ack,
		listingKey: listKey,
	}
}

func (bk BallotKeeper) GetBallot(ctx sdk.Context, identifier string) tcr.Ballot {
	return bk.GetBallot(ctx, identifier)
}

func (bk BallotKeeper) AddBallot(ctx sdk.Context, identifier string, owner sdk.Address, applyLen int64, bond int64) sdk.Error {
	return bk.AddBallot(ctx, identifier, owner, applyLen, bond)
}

func (bk BallotKeeper) ActivateBallot(ctx sdk.Context, owner sdk.Address, challenger sdk.Address, identifier string, commitLen, revealLen, minBond, challengBond int64) sdk.Error {
	ballot := bk.GetBallot(ctx, identifier)
	if ballot.Bond < minBond {
		bk.bm.DeleteBallot(ctx, identifier)
		bk.ack.AddCoins(ctx, owner, []sdk.Coin{{"RegistryCoin", challengBond}})
		return nil
	}
	return bk.bm.ActivateBallot(ctx, owner, challenger, identifier, commitLen, revealLen, minBond, challengBond)
}

func (bk BallotKeeper) CommitBallot(ctx sdk.Context, owner sdk.Address, identifier string, commitment []byte) {
	bk.bm.CommitBallot(ctx, owner, identifier, commitment)
}

func (bk BallotKeeper) VoteBallot(ctx sdk.Context, owner sdk.Address, identifier string, vote bool, power int64) sdk.Error {
	return bk.bm.VoteBallot(ctx, owner, identifier, vote, power)
}

func (bk BallotKeeper) DeleteBallot(ctx sdk.Context, identifier string) {
	bk.bm.DeleteBallot(ctx, identifier)
}

func (bk BallotKeeper) AddListing(ctx sdk.Context, identifier string, votes int64) {
	key := []byte(identifier)
	store := ctx.KVStore(bk.listingKey)

	listing := tcr.Listing{
		Identifier: identifier,
		Votes:      votes,
	}
	val, _ := bk.bm.Cdc.MarshalBinary(listing)

	store.Set(key, val)
}

func (bk BallotKeeper) GetListing(ctx sdk.Context, identifier string) tcr.Listing {
	key := []byte(identifier)
	store := ctx.KVStore(bk.listingKey)

	bz := store.Get(key)
	if bz == nil {
		return tcr.Listing{}
	}
	listing := &tcr.Listing{}
	err := bk.bm.Cdc.UnmarshalBinary(bz, listing)
	if err != nil {
		panic(err)
	}

	return *listing
}

func (bk BallotKeeper) DeleteListing(ctx sdk.Context, identifier string) {
	key := []byte(identifier)
	store := ctx.KVStore(bk.listingKey)

	store.Delete(key)
}
