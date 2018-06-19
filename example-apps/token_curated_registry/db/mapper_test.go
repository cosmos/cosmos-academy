package db

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"testing"

	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	"github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/utils"
	"github.com/tendermint/tmlibs/log"
)

func TestAddGet(t *testing.T) {
	ms, _, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	mapper := NewBallotMapper(ballotKey, cdc)

	addr := utils.GenerateAddress()
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	ballot := tcr.Ballot{
		Identifier:         "Unique registry listing",
		Owner:              addr,
		Bond:               50,
		EndApplyBlockStamp: 5,
	}

	getBallot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, getBallot, ballot, "Ballot received from store does not match expected value")
}

func TestDelete(t *testing.T) {
	ms, _, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	mapper := NewBallotMapper(ballotKey, cdc)

	addr := utils.GenerateAddress()
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	mapper.DeleteBallot(ctx, "Unique registry listing")

	ballot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, tcr.Ballot{}, ballot, "Ballot was not correctly deleted")
}

func TestVote(t *testing.T) {
	ms, _, ballotKey, _ := SetupMultiStore()
	cdc := MakeCodec()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	ctx.WithBlockHeight(10)
	mapper := NewBallotMapper(ballotKey, cdc)

	addr := utils.GenerateAddress()
	mapper.AddBallot(ctx, "Unique registry listing", addr, 5, 50)

	mapper.VoteBallot(ctx, addr, "Unique registry listing", true, 50)

	ballot := mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, int64(50), ballot.Approve, "Votes did not increment correctly")
}
