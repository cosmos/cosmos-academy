package auth

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/AdityaSripal/token_curated_registry/types"
	"github.com/AdityaSripal/token_curated_registry/db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/AdityaSripal/token_curated_registry/utils"
	"crypto/sha256"
)

func TestCandidacyHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, commitKey, revealKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	mapper := db.NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handler
	handler := NewCandidacyHandler(accountKeeper, mapper, 100, 10)

	res := handler(ctx, msg)

	// Currently testing against ABCICode type. Hope to transition to Error code.
	assert.Equal(t, sdk.ABCICodeType(0x1000a), res.Code, "Did not pass handler")

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	res = handler(ctx, msg)

	// Check account deducted by bond
	expected := accountKeeper.HasCoins(ctx, addr, []sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 50,
	}})

	assert.Equal(t, expected, true, "Account balances not deducted correctly")
	
	assert.Equal(t, sdk.Result{}, res, "Handler did not pass")

	// Check that adding candidate twice fails
	res = handler(ctx, msg)

	assert.Equal(t, sdk.ABCICodeType(0x1000a), res.Code, "Candidate allowed to be added twice")
}

func TestChallengeHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	challenger := utils.GenerateAddress()

	challengeMsg := types.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, commitKey, revealKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	mapper := db.NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, mapper, 100, 10)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	handler := NewChallengeHandler(accountKeeper, mapper, 10, 10, 100)

	res := handler(ctx, challengeMsg)

	// Challenge with no balance fails
	assert.Equal(t, sdk.ABCICodeType(0x1000a), res.Code, "Allowed to challenge without bond")

	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	res = handler(ctx, challengeMsg)

	// Valid challengeMsg changes state correcty
	ballot := mapper.GetBallot(ctx, "Unique registry listing")
	assert.Equal(t, true, ballot.Active, "Ballot correctly Activated")
	assert.Equal(t, int64(10), ballot.EndCommitBlockStamp, "Ballot commitstamp wrong")
	assert.Equal(t, int64(20), ballot.EndRevealBlockStamp, "Ballot revealstamp wrong")

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

	challengeMsg := types.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, commitKey, revealKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	mapper := db.NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, mapper, 100, 10)
	challengeHandler := NewChallengeHandler(accountKeeper, mapper, 10, 10, 100)
	commitHandler := NewCommitHandler(cdc, ballotKey, commitKey)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	// fund challenger
	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	commitMsg := types.NewCommitMsg(committer, "Unique registry listing", []byte("My commitment"))
	
	// Check that you cannot commit before challenge
	res := commitHandler(ctx, commitMsg)
	assert.Equal(t, sdk.ABCICodeType(0x20070), res.Code, "Allowed commitment before commit phase")

	challengeHandler(ctx, challengeMsg)

	res = commitHandler(ctx, commitMsg)

	// Check commit store updated
	store := ctx.KVStore(commitKey)
	voter := types.Voter{
		Owner: committer,
		Identifier: "Unique registry listing",
	}
	key, _ := cdc.MarshalBinary(voter)
	commitment := store.Get(key)
	assert.Equal(t, commitMsg.Commitment, commitment, "Commitment not set correctly")

	assert.Equal(t, sdk.Result{}, res, "Valid commitment msg did not pass")

}


func TestRevealHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	challenger := utils.GenerateAddress()
	voter := utils.GenerateAddress()

	challengeMsg := types.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, commitKey, revealKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	mapper := db.NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, mapper, 100, 10)
	challengeHandler := NewChallengeHandler(accountKeeper, mapper, 10, 10, 100)
	commitHandler := NewCommitHandler(cdc, ballotKey, commitKey)
	revealHandler := NewRevealHandler(accountKeeper, mapper)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	// fund challenger
	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	voterAcc := auth.NewBaseAccountWithAddress(voter)
	voterAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
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
	commitMsg := types.NewCommitMsg(voter, "Unique registry listing", commitment)
	commitHandler(ctx, commitMsg)

	// Create reveal msg's
	revealMsg := types.NewRevealMsg(voter, "Unique registry listing", true, []byte("My secret nonce"), sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	fakeMsg := types.NewRevealMsg(voter, "Unique registry listing", false, []byte("I want to change my vote"), sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	// Revealing before reveal phase fails
	res := revealHandler(ctx, revealMsg)
	assert.Equal(t, sdk.ABCICodeType(0x20070), res.Code, "Allowed reveal msg to pass before reveal phase")

	// Fast forward block height
	ctx = ctx.WithBlockHeight(11)

	// Revealing incorrect commitment fails
	res = revealHandler(ctx, fakeMsg)
	assert.Equal(t, sdk.ABCICodeType(0x2006a), res.Code, "Allowed invalid reveal to pass")

	// Check ballot votes have not changed after invalid reveals
	ballot := mapper.GetBallot(ctx, "Unique registry listing")
	assert.Equal(t, int64(0), ballot.Approve, "Ballot votes changed after invalid reveal")
	assert.Equal(t, int64(0), ballot.Deny, "Ballot votes changed after invalid reveal")

	// Valid reveal passes
	res = revealHandler(ctx, revealMsg)
	ballot = mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, int64(100), ballot.Approve, "Ballot votes did not increment correctly")
	assert.Equal(t, int64(0), ballot.Deny, "Deny votes is incorrect")
	assert.Equal(t, sdk.Result{}, res, "Reveal handling did not pass")

	// Check that revealing (voting) twice fails
	res = revealHandler(ctx, revealMsg)
	ballot = mapper.GetBallot(ctx, "Unique registry listing")

	assert.Equal(t, int64(100), ballot.Approve, "Allowed user to vote twice")
	assert.Equal(t, sdk.ABCICodeType(0x20080), res.Code, "Handler did not fail as expected when voting twice")
}

func TestApplyHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	challenger := utils.GenerateAddress()
	voter := utils.GenerateAddress()

	challengeMsg := types.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	ms, listKey, ballotKey, commitKey, revealKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	mapper := db.NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, mapper, 100, 10)
	challengeHandler := NewChallengeHandler(accountKeeper, mapper, 10, 10, 100)
	commitHandler := NewCommitHandler(cdc, ballotKey, commitKey)
	revealHandler := NewRevealHandler(accountKeeper, mapper)
	applyHandler := NewApplyHandler(accountKeeper, mapper, listKey, 0.5, 0.5)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	// fund challenger
	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	voterAcc := auth.NewBaseAccountWithAddress(voter)
	voterAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
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
	commitMsg := types.NewCommitMsg(voter, "Unique registry listing", commitment)
	commitHandler(ctx, commitMsg)

	// Create reveal msg's
	revealMsg := types.NewRevealMsg(voter, "Unique registry listing", true, []byte("My secret nonce"), sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	
	// Fast forward block height
	ctx = ctx.WithBlockHeight(11)

	revealHandler(ctx, revealMsg)

	// Create Apply msg
	applyMsg := types.NewApplyMsg(addr, "Unique registry listing")
	
	// Apply before end of reveal phase fails
	res := applyHandler(ctx, applyMsg)
	assert.Equal(t, sdk.ABCICodeType(0x20078), res.Code, "Allowed ballot to be finalized before end of reveal phase")

	// Fast forward block height past reveal phase
	ctx = ctx.WithBlockHeight(21)

	res = applyHandler(ctx, applyMsg)

	// check that applier got reward Bond(100) * dispPct(0.5) = 50. Note applier has 50 coins before apply
	actual := accountKeeper.HasCoins(ctx, addr, []sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	}})

	assert.Equal(t, true, actual, "Applier was not rewarded correctly")

	// check candidate was added to  listing
	ballot := mapper.GetBallot(ctx, "Unique registry listing")
	listing := mapper.GetListing(ctx, "Unique registry listing")
	expected := types.Listing{
		Identifier: "Unique registry listing",
		Votes: ballot.Approve,
	}

	assert.Equal(t, expected, listing, "Listing not added to registry correctly")

	assert.Equal(t, sdk.Result{}, res, "Handler did not pass")


	// Test apply works without challenge
	addr = utils.GenerateAddress()
	// Reset block height for clarity
	ctx = ctx.WithBlockHeight(0)

	// fund account
	account = auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &account)

	msg = types.NewDeclareCandidacyMsg(addr, "Unique registry listing 2", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	declareHandler(ctx, msg)

	// Fast forward past application stage
	ctx = ctx.WithBlockHeight(11)

	applyMsg = types.NewApplyMsg(addr, "Unique registry listing 2")
	res = applyHandler(ctx, applyMsg)

	// Check that listing added to registry
	expected = types.Listing{
		Identifier: "Unique registry listing 2",
		Votes: 0,
	}
	actualList := mapper.GetListing(ctx, "Unique registry listing 2")

	assert.Equal(t, expected, actualList, "Listing not added properly when unchallenged")

	// Check handler passes
	assert.Equal(t, sdk.Result{}, res, "applyHandler does not pass when unchallenged")


	// Check that challenging and removing an already existing listing works
	challenger = utils.GenerateAddress()
	// Reset block height for clarity
	ctx = ctx.WithBlockHeight(0)

	// fund challenger
	challengerAcc = auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	challengeMsg = types.NewChallengeMsg(challenger, "Unique registry listing 2", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	challengeHandler(ctx, challengeMsg)

	// Create commitment
	hasher = sha256.New()
	vote, _ = cdc.MarshalBinary(false)
	hasher.Sum(vote)
	commitment = hasher.Sum([]byte("My secret nonce"))

	commitMsg = types.NewCommitMsg(challenger, "Unique registry listing 2", commitment)
	commitHandler(ctx, commitMsg)

	// Fast forward to reveal stage
	ctx = ctx.WithBlockHeight(11)

	revealMsg = types.NewRevealMsg(challenger, "Unique registry listing 2", false, []byte("My secret nonce"), sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 50,
	})

	// Fast forward to apply stage
	ctx = ctx.WithBlockHeight(21)

	res = applyHandler(ctx, applyMsg)

	expected = types.Listing{}
	actualList = mapper.GetListing(ctx, "Unique registry listing 2")

	// Check that listing is deleted
	assert.Equal(t, expected, actualList, "Listing was not deleted from registry after successful challenge")

	// challenger should receive his original bond(100) as well as dispPct(0.5) of applier bond(100). Total 150. Note current balance of challenger is 0.
	actualBalance := accountKeeper.HasCoins(ctx, challenger, []sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 150,
	}})
	assert.Equal(t, true, actualBalance, "Challenger balance did not update correctly")

	// Check handler passes
	assert.Equal(t, sdk.Result{}, res, "Handler did not pass")
}

func TestClaimRewardHandler(t *testing.T) {
	// setup
	addr := utils.GenerateAddress()
	challenger := utils.GenerateAddress()
	victor1 := utils.GenerateAddress()
	victor2 := utils.GenerateAddress()
	loser := utils.GenerateAddress()

	challengeMsg := types.NewChallengeMsg(challenger, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 200,
	})

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 200,
	})
	ms, listKey, ballotKey, commitKey, revealKey, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	mapper := db.NewBallotMapper(listKey, ballotKey, commitKey, revealKey, cdc)

	accountMapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)

	// set handlers
	declareHandler := NewCandidacyHandler(accountKeeper, mapper, 100, 10)
	challengeHandler := NewChallengeHandler(accountKeeper, mapper, 10, 10, 100)
	commitHandler := NewCommitHandler(cdc, ballotKey, commitKey)
	revealHandler := NewRevealHandler(accountKeeper, mapper)
	applyHandler := NewApplyHandler(accountKeeper, mapper, listKey, 0.5, 0.5)
	claimRewardHandler := NewClaimRewardHandler(cdc, accountKeeper, ballotKey, revealKey, listKey, 0.5)

	// fund account
	account := auth.NewBaseAccountWithAddress(addr)
	account.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 250,
	}})
	accountMapper.SetAccount(ctx, &account)

	declareHandler(ctx, msg)

	// fund challenger
	challengerAcc := auth.NewBaseAccountWithAddress(challenger)
	challengerAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 250,
	}})
	accountMapper.SetAccount(ctx, &challengerAcc)

	// fund victor1
	victorAcc1 := auth.NewBaseAccountWithAddress(victor1)
	victorAcc1.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	}})
	accountMapper.SetAccount(ctx, &victorAcc1)

	// fund victor2
	victorAcc2 := auth.NewBaseAccountWithAddress(victor2)
	victorAcc2.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 300,
	}})
	accountMapper.SetAccount(ctx, &victorAcc2)

	// fund loser
	loserAcc := auth.NewBaseAccountWithAddress(loser)
	loserAcc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	}})
	accountMapper.SetAccount(ctx, &loserAcc)

	challengeHandler(ctx, challengeMsg)

	// Create victor commitment
	hasher := sha256.New()
	victorVote, _ := cdc.MarshalBinary(true)
	hasher.Sum(victorVote)
	victorCommitment1 := hasher.Sum([]byte("Victor1 secret nonce"))

	hasher = sha256.New()
	hasher.Sum(victorVote)
	victorCommitment2 := hasher.Sum([]byte("Victor2 secret nonce"))

	// Create loser commitment
	hasher = sha256.New()
	loserVote, _ := cdc.MarshalBinary(false)
	hasher.Sum(loserVote)
	loserCommitment := hasher.Sum([]byte("Loser secret nonce"))

	// Make commitments
	victorCommitMsg1 := types.NewCommitMsg(victor1, "Unique registry listing", victorCommitment1)
	commitHandler(ctx, victorCommitMsg1)

	victorCommitMsg2 := types.NewCommitMsg(victor2, "Unique registry listing", victorCommitment2)
	commitHandler(ctx, victorCommitMsg2)

	loserCommitMsg := types.NewCommitMsg(loser, "Unique registry listing", loserCommitment)
	commitHandler(ctx, loserCommitMsg)

	// Create reveal msg's
	victorRevealMsg1 := types.NewRevealMsg(victor1, "Unique registry listing", true, []byte("Victor1 secret nonce"), sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})
	victorRevealMsg2 := types.NewRevealMsg(victor2, "Unique registry listing", true, []byte("Victor2 secret nonce"), sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 300,
	})
	loserRevealMsg := types.NewRevealMsg(loser, "Unique registry listing", false, []byte("Loser secret nonce"), sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	
	// Fast forward to reveal phase
	ctx = ctx.WithBlockHeight(11)

	revealHandler(ctx, victorRevealMsg1)
	revealHandler(ctx, victorRevealMsg2)
	revealHandler(ctx, loserRevealMsg)

	// Fast foward to apply phase
	ctx = ctx.WithBlockHeight(21)

	// Create Claim reward Msg
	claimVictorMsg1 := types.NewClaimRewardMsg(victor1, "Unique registry listing")
	claimVictorMsg2 := types.NewClaimRewardMsg(victor2, "Unique registry listing")
	claimLoserMsg := types.NewClaimRewardMsg(loser, "Unique registry listing")

	// Make sure claimReward fails before being applied
	res := claimRewardHandler(ctx, claimVictorMsg1)
	assert.Equal(t, sdk.ABCICodeType(0x20082), res.Code, "Allowed claim reward to pass before apply")

	// Create Apply msg and handle
	applyMsg := types.NewApplyMsg(addr, "Unique registry listing")
	applyHandler(ctx, applyMsg)

	res1 := claimRewardHandler(ctx, claimVictorMsg1)
	res2 := claimRewardHandler(ctx, claimVictorMsg2)
	res3 := claimRewardHandler(ctx, claimLoserMsg)

	// Check that victor was awarded his initial bond as well as (1 - dispPct) = 0.5, multiplied by ballot's bond mutliplied by victor's ratio of total correct votes
	// 100 + 0.5 * 200 * 100 / 400 = 125
	actual := accountKeeper.HasCoins(ctx, victor1, []sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 125,
	}})
	assert.Equal(t, true, actual, "Victor1 not refunded properly")

	// 300 + 0.5 * 200 * 300 / 400 = 375
	actual = accountKeeper.HasCoins(ctx, victor2, []sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 375,
	}})
	assert.Equal(t, true, actual, "Victor2 not refunded properly")

	// Loser should get back their original bond, since reward for victors comes from challenger bond
	actual = accountKeeper.HasCoins(ctx, loser, []sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	}})
	assert.Equal(t, true, actual, "Loser not refunded properly")

	// Check handler passes
	assert.Equal(t, sdk.Result{}, res1, "Handler did not pass for victor1")
	assert.Equal(t, sdk.Result{}, res2, "Handler did not pass for victor2")
	assert.Equal(t, sdk.Result{}, res3, "Handler did not pass for loser")
}