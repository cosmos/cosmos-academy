package db

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/tmlibs/log"
	"github.com/AdityaSripal/token_curated_registry/types"
	"github.com/AdityaSripal/token_curated_registry/utils"
)

func TestAddGet(t *testing.T) {
	ms, listKey, ballotKey, commitKey, revealKey, _ := SetupMultiStore()
	cdc := MakeCodec()


	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	mapper := NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	addr := utils.GenerateAddress()
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	ballot := types.Ballot{
		Identifier: "Unique registry listing",
		Owner: addr,
		Bond: 50,
		EndApplyBlockStamp: 5,
	}

	getBallot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, getBallot, ballot, "Ballot received from store does not match expected value")
}

func TestDelete(t *testing.T) {
	ms, listKey, ballotKey, commitKey, revealKey, _ := SetupMultiStore()
	cdc := MakeCodec()


	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	mapper := NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	addr := utils.GenerateAddress()
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	mapper.DeleteBallot(ctx, "Unique registry listing")

	ballot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, types.Ballot{}, ballot, "Ballot was not correctly deleted")
}

func TestActivate(t *testing.T) {
	ms, listKey, ballotKey, commitKey, revealKey, accountKey := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	mapper := NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	addr := utils.GenerateAddress()
	account := auth.NewBaseAccountWithAddress(addr)
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper :=  bank.NewKeeper(accountMapper)

	accountMapper.SetAccount(ctx, &account)
	testaccount := accountMapper.GetAccount(ctx, addr)

	assert.Equal(t, &account, testaccount, "Accounts don't match")


	// Touch and remove case: Bond posted is less than new minBond
	challenger := utils.GenerateAddress()
	mapper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 100)

	delBallot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, types.Ballot{}, delBallot, "Outdated ballot was not deleted")
	
	// Check that challenger is refunded
	coins := accountKeeper.GetCoins(ctx, challenger)
	assert.Equal(t, int64(100), coins.AmountOf("RegistryCoin"), "Challenger did not get refunded after deleted ballot")


	// Test Activating with less than posted bond
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 150)
	err := mapper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 100)

	assert.Equal(t, sdk.CodeType(115), err.Code(), err.Error())

	err = mapper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 200)

	assert.Equal(t, sdk.CodeType(115), err.Code(), err.Error())


	// Test valid activation
	err = mapper.ActivateBallot(ctx, accountKeeper, addr, challenger, "Unique registry listing", 10, 10, 100, 150)
	if err != nil {
		fmt.Println(err.Error())
	}

	ballot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, true, ballot.Active, "Ballot not activated")
}

func TestVote(t *testing.T) {
	ms, listKey, ballotKey, commitKey, revealKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	mapper := NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	addr := utils.GenerateAddress()
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	mapper.VoteBallot(ctx, addr, "Unique registry listing", true, 50)

	ballot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, int64(50), ballot.Approve, "Votes did not increment correctly")
}

func TestAddDeleteList(t *testing.T) {
	ms, listKey, ballotKey, commitKey, revealKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	mapper := NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	mapper.AddListing(ctx, "Unique registry listing", 200)

	listing := mapper.GetListing(ctx, "Unique registry listing")

	expected := types.Listing{
		Identifier: "Unique registry listing",
		Votes: 200,
	}

	assert.Equal(t, expected, listing, "Listing not added correctly")

	mapper.DeleteListing(ctx, "Unique registry listing")

	delListing := mapper.GetListing(ctx, "Unique registry listing")

	assert.Equal(t, types.Listing{}, delListing, "Listing not added correctly")
}


