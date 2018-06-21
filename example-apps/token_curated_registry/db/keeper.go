package db

import (
	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/tendermint/go-amino"
	"reflect"
	"container/heap"
)

type BallotKeeper struct {
	ListingKey sdk.StoreKey

	BallotKey sdk.StoreKey

	Cdc *amino.Codec
}

func NewBallotKeeper(listingKey sdk.StoreKey, ballotkey sdk.StoreKey, _cdc *amino.Codec) BallotKeeper {
	return BallotKeeper{
		ListingKey: listingKey,
		BallotKey:  ballotkey,
		Cdc:        _cdc,
	}
}

// Will get Ballot using unique identifier. Do not need to specify status
func (bk BallotKeeper) GetBallot(ctx sdk.Context, identifier string) tcr.Ballot {
	store := ctx.KVStore(bk.BallotKey)
	key := []byte(identifier)
	val := store.Get(key)
	if val == nil {
		return tcr.Ballot{}
	}
	ballot := &tcr.Ballot{}
	err := bk.Cdc.UnmarshalBinary(val, ballot)
	if err != nil {
		panic(err)
	}
	return *ballot
}

func (bk BallotKeeper) AddBallot(ctx sdk.Context, identifier string, owner sdk.Address, applyLen int64, bond int64) sdk.Error {
	store := ctx.KVStore(bk.BallotKey)

	newBallot := tcr.Ballot{
		Identifier:         identifier,
		Owner:              owner,
		Bond:               bond,
		EndApplyBlockStamp: ctx.BlockHeight() + applyLen,
	}
	// Add ballot with Pending Status
	key := []byte(identifier)
	val, _ := bk.Cdc.MarshalBinary(newBallot)
	store.Set(key, val)
	return nil
}

func (bk BallotKeeper) ActivateBallot(ctx sdk.Context, accountKeeper bank.Keeper, owner sdk.Address, challenger sdk.Address, 
	identifier string, commitLen int64, revealLen, minBond int64, challengeBond int64) sdk.Error {
	store := ctx.KVStore(bk.BallotKey)
	ballot := bk.GetBallot(ctx, identifier)

	if ballot.Bond < minBond {
		bk.DeleteBallot(ctx, identifier)
		refund := sdk.Coin{
			Denom:  "RegistryCoin",
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
	ballot.EndApplyBlockStamp = ballot.EndCommitBlockStamp + revealLen

	newBallot, _ := bk.Cdc.MarshalBinary(ballot)
	key := []byte(identifier)
	store.Set(key, newBallot)

	return nil
}

func (bk BallotKeeper) CommitBallot(ctx sdk.Context, owner sdk.Address, identifier string, commitment []byte) {
	commitKey := []byte(identifier)
	commitKey = append(commitKey, []byte("commits")...)
	commitKey = append(commitKey, owner...)

	ballotStore := ctx.KVStore(bk.BallotKey)

	ballotStore.Set(commitKey, commitment)
}

func (bk BallotKeeper) GetCommitment(ctx sdk.Context, owner sdk.Address, identifier string) []byte {
	commitKey := []byte(identifier)
	commitKey = append(commitKey, []byte("commits")...)
	commitKey = append(commitKey, owner...)

	ballotStore := ctx.KVStore(bk.BallotKey)

	return ballotStore.Get(commitKey)
}

func (bk BallotKeeper) DeleteCommitment(ctx sdk.Context, owner sdk.Address, identifier string) {
	commitKey := []byte(identifier)
	commitKey = append(commitKey, []byte("commits")...)
	commitKey = append(commitKey, owner...)

	ballotStore := ctx.KVStore(bk.BallotKey)

	ballotStore.Delete(commitKey)
}

func (bk BallotKeeper) VoteBallot(ctx sdk.Context, owner sdk.Address, identifier string, vote bool, power int64) sdk.Error {
	ballotStore := ctx.KVStore(bk.BallotKey)

	ballotKey := []byte(identifier)
	bz := ballotStore.Get(ballotKey)
	if bz == nil {
		return sdk.NewError(2, 107, "Ballot does not exist")
	}
	ballot := &tcr.Ballot{}
	err := bk.Cdc.UnmarshalBinary(bz, ballot)
	if err != nil {
		panic(err)
	}
	if vote {
		ballot.Approve += power
	} else {
		ballot.Deny += power
	}
	newBallot, _ := bk.Cdc.MarshalBinary(*ballot)
	ballotStore.Set(ballotKey, newBallot)

	voteStruct := tcr.Vote{Choice: vote, Power: power}
	voteVal, err2 := bk.Cdc.MarshalBinary(voteStruct)

	if err2 != nil {
		panic(err2)
	}

	// Set voteKey to "{identifier}|votes|{owner_address}|
	voteKey := []byte(identifier)
	voteKey = append(voteKey, []byte("votes")...)
	voteKey = append(voteKey, owner...)

	ballotStore.Set(voteKey, voteVal)
	return nil
}

func (bk BallotKeeper) GetVote(ctx sdk.Context, owner sdk.Address, identifier string) tcr.Vote {
	// Set voteKey to "{identifier}|votes|{owner_address}|
	voteKey := []byte(identifier)
	voteKey = append(voteKey, []byte("votes")...)
	voteKey = append(voteKey, owner...)

	vote := &tcr.Vote{}
	ballotStore := ctx.KVStore(bk.BallotKey)

	bz := ballotStore.Get(voteKey)
	if bz == nil {
		return tcr.Vote{}
	}

	err := bk.Cdc.UnmarshalBinary(bz, vote)
	if err != nil {
		panic(err)
	}
	return *vote
}

func (bk BallotKeeper) DeleteVote(ctx sdk.Context, owner sdk.Address, identifier string) {
	voteKey := []byte(identifier)
	voteKey = append(voteKey, []byte("votes")...)
	voteKey = append(voteKey, owner...)

	ballotStore := ctx.KVStore(bk.BallotKey)

	ballotStore.Delete(voteKey)
}

func (bk BallotKeeper) DeleteBallot(ctx sdk.Context, identifier string) {
	key := []byte(identifier)
	store := ctx.KVStore(bk.BallotKey)
	store.Delete(key)
}

func (bk BallotKeeper) AddListing(ctx sdk.Context, identifier string, votes int64) {
	key := []byte(identifier)
	store := ctx.KVStore(bk.ListingKey)

	listing := tcr.Listing{
		Identifier: identifier,
		Votes:      votes,
	}
	val, _ := bk.Cdc.MarshalBinary(listing)

	store.Set(key, val)
}

func (bk BallotKeeper) GetListing(ctx sdk.Context, identifier string) tcr.Listing {
	key := []byte(identifier)
	store := ctx.KVStore(bk.ListingKey)

	bz := store.Get(key)
	if bz == nil {
		return tcr.Listing{}
	}
	listing := &tcr.Listing{}
	err := bk.Cdc.UnmarshalBinary(bz, listing)
	if err != nil {
		panic(err)
	}

	return *listing
}

func (bk BallotKeeper) DeleteListing(ctx sdk.Context, identifier string) {
	key := []byte(identifier)
	store := ctx.KVStore(bk.ListingKey)

	store.Delete(key)
}

// --------------------------------------------------------------------------------------------------
// Queue stored in key: candidateQueue

// getCandidateQueue gets the CandidateQueue from the context
func (bk BallotKeeper) getCandidateQueue(ctx sdk.Context) tcr.PriorityQueue {
	store := ctx.KVStore(bk.BallotKey)
	bpq := store.Get([]byte("candidateQueue"))
	if bpq == nil {
		return tcr.PriorityQueue{}
	}

	candidateQueue := tcr.PriorityQueue{}
	err := bk.Cdc.UnmarshalBinaryBare(bpq, &candidateQueue)
	if err != nil {
		panic(err)
	}

	return candidateQueue
}

// setProposalQueue sets the CandidateQueue to the context
func (bk BallotKeeper) setCandidateQueue(ctx sdk.Context, candidateQueue tcr.PriorityQueue) {
	store := ctx.KVStore(bk.BallotKey)
	bpq, err := bk.Cdc.MarshalBinaryBare(candidateQueue)
	if err != nil {
		panic(err)
	}
	store.Set([]byte("candidateQueue"), bpq)
}

// ProposalQueueHead returns the head of the PriorityQueue
func (bk BallotKeeper) ProposalQueueHead(ctx sdk.Context) tcr.Ballot {
	candidateQueue := bk.getCandidateQueue(ctx)
	if reflect.DeepEqual(candidateQueue, tcr.PriorityQueue{}) {
		return tcr.Ballot{}
	}
	if candidateQueue.Len() == 0 {
		return tcr.Ballot{}
	}
	ballot := bk.GetBallot(ctx, candidateQueue.Peek().Value)
	return ballot
}

// ProposalQueuePop pops the head from the Proposal queue
func (bk BallotKeeper) ProposalQueuePop(ctx sdk.Context) tcr.Ballot {
	candidateQueue := bk.getCandidateQueue(ctx)
	if reflect.DeepEqual(candidateQueue, tcr.PriorityQueue{}) {
		return tcr.Ballot{}
	}
	if candidateQueue.Len() == 0 {
		return tcr.Ballot{}
	}

	ballot := bk.GetBallot(ctx, heap.Pop(&candidateQueue).(*tcr.Item).Value)
	return ballot
}

// ProposalQueuePush pushes a candidate to Priority Queue
func (bk BallotKeeper) ProposalQueuePush(ctx sdk.Context, identifier string, blockNum int64) {
	candidateQueue := bk.getCandidateQueue(ctx)

	item := tcr.Item{Value: identifier, Priority: int(blockNum)}
	heap.Push(&candidateQueue, &item)
	bk.setCandidateQueue(ctx, candidateQueue)
}

// ProposalQueueUpdate updates a candidate with new priority
func (bk BallotKeeper) ProposalQueueUpdate(ctx sdk.Context, identifier string, newBlockNum int64) sdk.Error {
	candidateQueue := bk.getCandidateQueue(ctx)

	err := candidateQueue.Update(identifier, int(newBlockNum))
	if err != nil {
		return err
	}
	bk.setCandidateQueue(ctx, candidateQueue)
	return nil
}

func (bk BallotKeeper) ProposalQueueContains(ctx sdk.Context, identifier string) bool {
	candidateQueue := bk.getCandidateQueue(ctx)

	for _, item := range candidateQueue {
		if item.Value == identifier {
			return true
		}
	}
	return false
}

func (bk BallotKeeper) ProposalQueueGetPriority(ctx sdk.Context, identifier string) int {
	candidateQueue := bk.getCandidateQueue(ctx)

	for _, item := range candidateQueue {
		if item.Value == identifier {
			return item.Priority
		}
	}
	return -1
}