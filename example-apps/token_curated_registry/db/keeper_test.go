package db

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"testing"
	"container/heap"

	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	"github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/utils"
	"github.com/tendermint/tmlibs/log"
)

func TestAddGet(t *testing.T) {
	ms, listKey, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	keeper := NewBallotKeeper(listKey, ballotKey, cdc)

	addr := utils.GenerateAddress()
	keeper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	ballot := tcr.Ballot{
		Identifier:         "Unique registry listing",
		Owner:              addr,
		Bond:               50,
		EndApplyBlockStamp: 5,
	}

	getBallot := keeper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, getBallot, ballot, "Ballot received from store does not match expected value")
}

func TestDelete(t *testing.T) {
	ms, listKey, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	keeper := NewBallotKeeper(listKey, ballotKey, cdc)

	addr := utils.GenerateAddress()
	keeper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	keeper.DeleteBallot(ctx, "Unique registry listing")

	ballot := keeper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, tcr.Ballot{}, ballot, "Ballot was not correctly deleted")
}

func TestActivate(t *testing.T) {
	ms, listKey, ballotKey, accountKey := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	keeper := NewBallotKeeper(listKey, ballotKey, cdc)

	addr := utils.GenerateAddress()
	account := auth.NewBaseAccountWithAddress(addr)
	keeper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	accountMapper.SetAccount(ctx, &account)
	testaccount := accountMapper.GetAccount(ctx, addr)

	assert.Equal(t, &account, testaccount, "Accounts don't match")

	// Touch and remove case: Bond posted is less than new minBond
	challenger := utils.GenerateAddress()
	keeper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 100)

	delBallot := keeper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, tcr.Ballot{}, delBallot, "Outdated ballot was not deleted")

	// Check that challenger is refunded
	coins := accountKeeper.GetCoins(ctx, challenger)
	assert.Equal(t, int64(100), coins.AmountOf("RegistryCoin"), "Challenger did not get refunded after deleted ballot")

	// Test Activating with less than posted bond
	keeper.AddBallot(ctx, "Unique registry listing", addr, 5, 150)
	err := keeper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 100)

	assert.Equal(t, sdk.CodeType(115), err.Code(), err.Error())

	err = keeper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 200)

	assert.Equal(t, sdk.CodeType(115), err.Code(), err.Error())

	// Test valid activation
	err = keeper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 150)
	if err != nil {
		fmt.Println(err.Error())
	}

	ballot := keeper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, true, ballot.Active, "Ballot not activated")
}

func TestCommit(t *testing.T) {
	ms, listKey, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	keeper := NewBallotKeeper(listKey, ballotKey, cdc)

	addr := utils.GenerateAddress()
	keeper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	keeper.CommitBallot(ctx, addr, "Unique registry listing", []byte("my commitment"))

	commitment := keeper.GetCommitment(ctx, addr, "Unique registry listing")

	assert.Equal(t, []byte("my commitment"), commitment, "Commitment not added to ballotStore correctly")
}

func TestVote(t *testing.T) {
	ms, listKey, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	keeper := NewBallotKeeper(listKey, ballotKey, cdc)

	addr := utils.GenerateAddress()
	keeper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	keeper.VoteBallot(ctx, addr, "Unique registry listing", true, 50)

	ballot := keeper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, int64(50), ballot.Approve, "Votes did not increment correctly")

	vote := keeper.GetVote(ctx, addr, "Unique registry listing")

	assert.Equal(t, tcr.Vote{true, 50}, vote, "Vote did not get added to ballotStore correctly")
}

func TestAddDeleteList(t *testing.T) {
	ms, listKey, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	keeper := NewBallotKeeper(listKey, ballotKey, cdc)

	keeper.AddListing(ctx, "Unique registry listing", 200)

	listing := keeper.GetListing(ctx, "Unique registry listing")

	expected := tcr.Listing{
		Identifier: "Unique registry listing",
		Votes:      200,
	}

	assert.Equal(t, expected, listing, "Listing not added correctly")

	keeper.DeleteListing(ctx, "Unique registry listing")

	delListing := keeper.GetListing(ctx, "Unique registry listing")

	assert.Equal(t, tcr.Listing{}, delListing, "Listing not deleted correctly")
}

// ------------------------------------------------------------------------------------------------------------------
// Test CandidateQueue

func TestQueue(t *testing.T) {
	ms, listKey, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	keeper := NewBallotKeeper(listKey, ballotKey, cdc)

	assert.Equal(t, tcr.PriorityQueue{}, keeper.getCandidateQueue(ctx), "Incorrect behavior on init")

	candidateQueue := tcr.PriorityQueue{}
	item := tcr.Item{Value: "a", Priority: 10}
	heap.Push(&candidateQueue, &item)
	keeper.setCandidateQueue(ctx, candidateQueue)

	item2 := tcr.Item{Value: "b", Priority: 5}
	heap.Push(&candidateQueue, &item2)
	keeper.setCandidateQueue(ctx, candidateQueue)

	queue := keeper.getCandidateQueue(ctx)
	assert.Equal(t, item2, queue.Peek(), "Peek does not work")

	assert.Equal(t, item2.Value, heap.Pop(&queue).(*tcr.Item).Value, "Pop does not work")

	assert.Equal(t, item.Value, heap.Pop(&queue).(*tcr.Item).Value, "Pop does not work")
} 


