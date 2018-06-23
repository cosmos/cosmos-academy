package auth

import (
	"crypto/sha256"
	"github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/db"
	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	"github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"
	"testing"
)

func TestCandidacyHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	msg := tcr.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	keeper := db.NewBallotKeeper(listKey, ballotKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handler
	handler := NewCandidacyHandler(accountKeeper, keeper, 100, 10)

	res := handler(ctx, msg)

	// Currently testing against ABCICode type. Hope to transition to Error code.
	assert.Equal(t, sdk.ABCICodeType(0x1000a), res.Code, "Did not pass handler")

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	res = handler(ctx, msg)

	// Check account deducted by bond
	expected := accountKeeper.HasCoins(ctx, addr, []sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 50,
	}})

	assert.Equal(t, expected, true, "Account balances not deducted correctly")

	assert.Equal(t, sdk.Result{}, res, "Handler did not pass")

	// Check that adding candidate twice fails
	res = handler(ctx, msg)

	assert.True(t, keeper.ProposalQueueContains(ctx, "Unique registry listing"), "Proposal queue does not contain listing")
	assert.Equal(t, int(ctx.BlockHeight() + 10), keeper.ProposalQueueGetPriority(ctx, "Unique registry listing"), "Proposal added with incorrect priority")

	assert.Equal(t, sdk.ABCICodeType(0x1000a), res.Code, "Candidate allowed to be added twice")
}

func TestChallengeHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	challenger := utils.GenerateAddress()

	challengeMsg := tcr.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})

	msg := tcr.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	keeper := db.NewBallotKeeper(listKey, ballotKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, keeper, 100, 10)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	handler := NewChallengeHandler(accountKeeper, keeper, 10, 10, 100)

	res := handler(ctx, challengeMsg)

	// Challenge with no balance fails
	assert.Equal(t, sdk.ABCICodeType(0x1000a), res.Code, "Allowed to challenge without bond")

	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	assert.Equal(t, int(ctx.BlockHeight() + 10), keeper.ProposalQueueGetPriority(ctx, "Unique registry listing"), "Priority in queue before challenge is wrong")

	res = handler(ctx, challengeMsg)

	// Valid challengeMsg changes state correcty
	ballot := keeper.GetBallot(ctx, "Unique registry listing")
	assert.Equal(t, true, ballot.Active, "Ballot correctly Activated")
	assert.Equal(t, int64(10), ballot.EndCommitBlockStamp, "Ballot commitstamp wrong")
	assert.Equal(t, int64(20), ballot.EndApplyBlockStamp, "Ballot revealstamp wrong")

	assert.Equal(t, int(ctx.BlockHeight() + 20), keeper.ProposalQueueGetPriority(ctx, "Unique registry listing"), "Priority not updated correctly")

	assert.Equal(t, sdk.Result{}, res, "Handler did not pass")

	// Cannot challenge same candidate twice
	res = handler(ctx, challengeMsg)
	assert.Equal(t, sdk.ABCICodeType(0x1000a), res.Code, "Allowed ballot to be challenged twice")
}

func TestCommitHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	challenger := utils.GenerateAddress()
	committer := utils.GenerateAddress()

	challengeMsg := tcr.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})

	msg := tcr.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	keeper := db.NewBallotKeeper(listKey, ballotKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, keeper, 100, 10)
	challengeHandler := NewChallengeHandler(accountKeeper, keeper, 10, 10, 100)
	commitHandler := NewCommitHandler(cdc, keeper)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	// fund challenger
	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	commitMsg := tcr.NewCommitMsg(committer, "Unique registry listing", []byte("My commitment"))

	// Check that you cannot commit before challenge
	res := commitHandler(ctx, commitMsg)
	assert.Equal(t, sdk.ABCICodeType(0x20068), res.Code, "Allowed commitment before commit phase")

	challengeHandler(ctx, challengeMsg)

	res = commitHandler(ctx, commitMsg)

	// Check commit store updated
	commitment := keeper.GetCommitment(ctx, commitMsg.Owner, commitMsg.Identifier)
	assert.Equal(t, commitMsg.Commitment, commitment, "Commitment not set correctly")

	assert.Equal(t, sdk.Result{}, res, "Valid commitment msg did not pass")

}

func TestRevealHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	challenger := utils.GenerateAddress()
	voter := utils.GenerateAddress()

	challengeMsg := tcr.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})

	msg := tcr.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	keeper := db.NewBallotKeeper(listKey, ballotKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, keeper, 100, 10)
	challengeHandler := NewChallengeHandler(accountKeeper, keeper, 10, 10, 100)
	commitHandler := NewCommitHandler(cdc, keeper)
	revealHandler := NewRevealHandler(accountKeeper, keeper)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	// fund challenger
	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	voterAcc := auth.NewBaseAccountWithAddress(voter)
	voterAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 400,
	}})
	accountMapper.SetAccount(ctx, &voterAcc)

	challengeHandler(ctx, challengeMsg)

	// Create commitment
	hasher := sha256.New()
	vote, _ := cdc.MarshalBinary(true)
	hasher.Sum(vote)
	commitment := hasher.Sum([]byte("My secret nonce"))

	// Make commitment
	commitMsg := tcr.NewCommitMsg(voter, "Unique registry listing", commitment)
	commitHandler(ctx, commitMsg)

	// Create reveal msg's
	revealMsg := tcr.NewRevealMsg(voter, "Unique registry listing", true, []byte("My secret nonce"), sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})
	fakeMsg := tcr.NewRevealMsg(voter, "Unique registry listing", false, []byte("I want to change my vote"), sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})

	// Revealing before reveal phase fails
	res := revealHandler(ctx, revealMsg)
	assert.Equal(t, sdk.ABCICodeType(0x20068), res.Code, "Allowed reveal msg to pass before reveal phase")

	// Fast forward block height
	ctx = ctx.WithBlockHeight(11)

	// Revealing incorrect commitment fails
	res = revealHandler(ctx, fakeMsg)
	assert.Equal(t, sdk.ABCICodeType(0x20069), res.Code, "Allowed invalid reveal to pass")

	// Check ballot votes have not changed after invalid reveals
	ballot := keeper.GetBallot(ctx, "Unique registry listing")
	assert.Equal(t, int64(0), ballot.Approve, "Ballot votes changed after invalid reveal")
	assert.Equal(t, int64(0), ballot.Deny, "Ballot votes changed after invalid reveal")

	// Valid reveal passes
	res = revealHandler(ctx, revealMsg)
	ballot = keeper.GetBallot(ctx, "Unique registry listing")

	savedVote := keeper.GetVote(ctx, revealMsg.Owner, "Unique registry listing")
	expectedVote := tcr.Vote{true, 100}

	assert.Equal(t, int64(100), ballot.Approve, "Ballot votes did not increment correctly")
	assert.Equal(t, int64(0), ballot.Deny, "Deny votes is incorrect")
	assert.Equal(t, expectedVote, savedVote, "Vote not saved in ballotStore correctly")
	assert.Equal(t, sdk.Result{}, res, "Reveal handling did not pass")

	// Check that revealing (voting) twice fails
	res = revealHandler(ctx, revealMsg)
	ballot = keeper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, int64(100), ballot.Approve, "Allowed user to vote twice")
	assert.Equal(t, sdk.ABCICodeType(0x20069), res.Code, "Handler did not fail as expected when voting twice")
}
