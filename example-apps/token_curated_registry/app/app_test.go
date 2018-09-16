package app

import (
	"github.com/tendermint/go-amino"
	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	"github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/utils"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
	"github.com/tendermint/go-crypto"
	"os"
	"testing"
	"crypto/sha256"
	"math/rand"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newRegistryApp() *RegistryApp {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return NewRegistryApp(logger, db, 100, 10, 10, 10, 0.5, 0.5)
}

func setGenesis(rapp *RegistryApp, accs ...auth.BaseAccount) error {
	genaccs := make([]*tcr.GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = tcr.NewGenesisAccount(&acc)
	}

	genesisState := tcr.GenesisState{
		Accounts: genaccs,
	}

	stateBytes, err := wire.MarshalJSONIndent(rapp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	rapp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	rapp.Commit()

	return nil
}

func TestBadMsg(t *testing.T) {
	rapp := newRegistryApp()

	privKey := utils.GeneratePrivKey()
	addr := privKey.PubKey().Address()
	acc := auth.NewBaseAccountWithAddress(addr)

	acc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 50,
	}})

	err := setGenesis(rapp, acc)
	if err != nil {
		panic(err)
	}

	msg := tcr.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 50,
	})

	signBytes := auth.StdSignBytes(t.Name(), []int64{0}, auth.StdFee{}, msg)
	sig := privKey.Sign(signBytes)

	require.Equal(t, true, privKey.PubKey().VerifyBytes(signBytes, sig), "Sig doesn't work")

	tx := GenTx(t.Name(), msg, auth.StdFee{}, privKey, 0)

	cdc := MakeCodec()

	txBytes, encodeErr := cdc.MarshalBinary(tx)

	require.NoError(t, encodeErr)

	header := abci.Header{ChainID: t.Name()}

	// Simulate a Block
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	// Must commit to setCheckState
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Run a check
	cres := rapp.CheckTx(txBytes)
	require.Equal(t, sdk.CodeType(5),
		sdk.CodeType(cres.Code), cres.Log)

	dres := rapp.DeliverTx(txBytes)
	require.Equal(t, sdk.CodeType(5), sdk.CodeType(dres.Code), dres.Log)

}

func TestBadTx(t *testing.T) {
	rapp := newRegistryApp()

	privKey := utils.GeneratePrivKey()
	addr := privKey.PubKey().Address()
	acc := auth.NewBaseAccountWithAddress(addr)

	acc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	}})

	err := setGenesis(rapp, acc)
	if err != nil {
		panic(err)
	}

	msg := tcr.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})

	tx := auth.StdTx{
		Msg: msg,
	}

	cdc := MakeCodec()

	txBytes, encodeErr := cdc.MarshalBinary(tx)

	require.NoError(t, encodeErr)

	// Run a check
	cres := rapp.CheckTx(txBytes)
	require.Equal(t, sdk.CodeType(4),
		sdk.CodeType(cres.Code), cres.Log)

	// Simulate a Block
	rapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: t.Name()}})
	dres := rapp.DeliverTx(txBytes)
	require.Equal(t, sdk.CodeType(4), sdk.CodeType(dres.Code), dres.Log)

}

func TestApplyUnchallengedFlow(t *testing.T) {
	rapp := newRegistryApp()

	privKey := utils.GeneratePrivKey()
	addr := privKey.PubKey().Address()
	acc := auth.NewBaseAccountWithAddress(addr)

	acc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	}})

	err := setGenesis(rapp, acc)
	if err != nil {
		panic(err)
	}

	msg := tcr.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	})

	fee := auth.StdFee{Gas: 10000000}

	tx := GenTx(t.Name(), msg, fee, privKey, 0)

	cdc := MakeCodec()

	txBytes, encodeErr := cdc.MarshalBinary(tx)

	require.NoError(t, encodeErr)

	header := abci.Header{AppHash: []byte("apphash"), ChainID: t.Name()}
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	// Must commit a block to setCheckState
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Run a check
	cres := rapp.CheckTx(txBytes)
	require.Equal(t, sdk.CodeType(0),
		sdk.CodeType(cres.Code), cres.Log)

	// Simulate a Block
	dres := rapp.Deliver(tx)
	require.Equal(t, sdk.CodeType(0), sdk.CodeType(dres.Code), dres.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Mine 10 empty blocks
	for i := 0; i < 10; i++ {
		header.Height = int64(i + 1)
		rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
		rapp.EndBlock(abci.RequestEndBlock{})
		rapp.Commit()
	}

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := rapp.NewContext(false, header)

	store := ctx.KVStore(rapp.capKeyListings)

	listing := tcr.Listing{
		Identifier: "Unique registry listing",
		Votes:      0,
	}
	expected, _ := rapp.cdc.MarshalBinary(listing)
	actual := store.Get([]byte("Unique registry listing"))

	require.Equal(t, expected, actual, "Listing not added correctly to registry")

}

func TestChallengeWonFlow(t *testing.T) {
	rapp := newRegistryApp()

	priv1 := utils.GeneratePrivKey()
	addr1 := priv1.PubKey().Address()
	acc1 := auth.NewBaseAccountWithAddress(addr1)
	acc1.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	priv2 := utils.GeneratePrivKey()
	addr2 := priv2.PubKey().Address()
	acc2 := auth.NewBaseAccountWithAddress(addr2)
	acc2.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	}})

	priv3 := utils.GeneratePrivKey()
	addr3 := priv3.PubKey().Address()
	acc3 := auth.NewBaseAccountWithAddress(addr3)
	acc3.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	}})

	priv4 := utils.GeneratePrivKey()
	addr4 := priv4.PubKey().Address()
	acc4 := auth.NewBaseAccountWithAddress(addr4)
	acc4.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	priv5 := utils.GeneratePrivKey()
	addr5 := priv5.PubKey().Address()
	acc5 := auth.NewBaseAccountWithAddress(addr5)
	acc5.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	}})

	err := setGenesis(rapp, acc1, acc2, acc3, acc4, acc5)
	if err != nil {
		panic(err)
	}

	// Set two declare tx. First one will be challenged, second one will be unchallenged.
	msg1 := tcr.NewDeclareCandidacyMsg(addr1, "Unique registry listing 1", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})
	msg2 := tcr.NewDeclareCandidacyMsg(addr2, "Unique registry listing 2", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})

	fee := auth.StdFee{Gas: 10000000}

	declareTx1 := GenTx(t.Name(), msg1, fee, priv1, 0)
	declareTx2 := GenTx(t.Name(), msg2, fee, priv2, 0)


	header := abci.Header{AppHash: []byte("apphash"), ChainID: t.Name()}
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx1 in Block 0
	res := rapp.Deliver(declareTx1)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// deliver tx2 in Block 1
	header.Height += 1
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	res = rapp.Deliver(declareTx2)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := rapp.NewContext(false, header)

	// Check that ApplyEndBlockStamp is correct
	endStamp1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 1").EndApplyBlockStamp
	endStamp2 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 2").EndApplyBlockStamp

	require.Equal(t, int64(10), endStamp1, "End stamp for candidate 1 is wrong")
	require.Equal(t, int64(11), endStamp2, "End stamp for candidate 2 is wrong")

	// Check that head of queue is candidate 1
	head := rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing 1", head, "Head of queue is wrong")

	// Check that candidate 2 is in queue
	require.True(t, rapp.ballotKeeper.ProposalQueueContains(ctx, "Unique registry listing 2"), "Candidate 2 is not in queue")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Challenge candidate 1
	challengeMsg := tcr.NewChallengeMsg(addr3, "Unique registry listing 1", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})
	challengeTx := GenTx(t.Name(), challengeMsg, fee, priv3, 0)

	header.Height = 5
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	res = rapp.Deliver(challengeTx)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)

	// Check ballot 1 updated correctly
	ballot1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 1")
	require.True(t, ballot1.Active, "Ballot 1 unactivated")
	require.Equal(t, int64(25), ballot1.EndApplyBlockStamp, "Ballot 1 end blockstamp not updated after challenge")

	// Check that candidate 2 is now at head of queue
	head = rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing 2", head, "Head of queue is wrong")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	cdc := MakeCodec()

	commitMsg1, nonce1 := makeCommitment(cdc, addr4, "Unique registry listing 1", true)
	commitMsg2, nonce2 := makeCommitment(cdc, addr5, "Unique registry listing 1", false)
	commitMsg3, nonce3 := makeCommitment(cdc, addr1, "Unique registry listing 1", true)

	commitTx1 := GenTx(t.Name(), commitMsg1, fee, priv4, 0)
	commitTx2 := GenTx(t.Name(), commitMsg2, fee, priv5, 0)
	commitTx3 := GenTx(t.Name(), commitMsg3, fee, priv1, 1)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver commit Msgs
	res = rapp.Deliver(commitTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Move to reveal phase
	header.Height = 16

	revealMsg1 := tcr.NewRevealMsg(addr4, "Unique registry listing 1", true, nonce1, sdk.Coin{"RegistryCoin", 300})
	revealMsg2 := tcr.NewRevealMsg(addr5, "Unique registry listing 1", false, nonce2, sdk.Coin{"RegistryCoin", 100})
	revealMsg3 := tcr.NewRevealMsg(addr1, "Unique registry listing 1", true, nonce3, sdk.Coin{"RegistryCoin", 100})

	
	revealTx1 := GenTx(t.Name(), revealMsg1, fee, priv4, 1)
	revealTx2 := GenTx(t.Name(), revealMsg2, fee, priv5, 1)
	revealTx3 := GenTx(t.Name(), revealMsg3, fee, priv1, 2)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver reveal Msgs
	res = rapp.Deliver(revealTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Candidate 2 should have been added to queue by now
	require.Equal(t, "Unique registry listing 2", rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 2").Identifier, "Candidate 2 not added to registry after application phase")
	// Candidate 1 should be at head now
	require.Equal(t, "Unique registry listing 1", rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier, "Ballot queue not updated after pop")
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Move to apply phase
	header.Height = 30
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Candidate 1 passed the challenge. Should be added to registry.
	require.Equal(t, "Unique registry listing 1", rapp.ballotKeeper.GetListing(ctx, "Unique registry listing 1").Identifier, "Candidate 1 not added to registry after beating challenge")

	// Voters who vote wrong should be refunded: 100
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr5, sdk.Coins{{"RegistryCoin", 100}}), fmt.Sprintf("Losing voter has: %+v", rapp.accountKeeper.GetCoins(ctx, addr5)))

	// Proposer wins refund along with dispensation percent (50%) of challenger bond as well as his share of the correct vote pool (1 / 4) * rest of challenger bond (50%)
	// 100 + 100 + 25 = 225
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr1, sdk.Coins{{"RegistryCoin", 225}}), fmt.Sprintf("Proposer and winning voter has: %+v", rapp.accountKeeper.GetCoins(ctx, addr1)))

	// Winning voter gets refund along with his share of correct vote pool (3 / 4) * rest of challenger bond (50%)
	// 300 + 75
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr4, sdk.Coins{{"RegistryCoin", 375}}), fmt.Sprintf("Winning voter has: %v", rapp.accountKeeper.GetCoins(ctx, addr4)))

	// Challenger loses bond
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr3, sdk.Coins(nil)), fmt.Sprintf("Challenge has: %+v", rapp.accountKeeper.GetCoins(ctx, addr3)))
}


func TestChallengeLostFlow(t *testing.T) {
	rapp := newRegistryApp()

	priv1 := utils.GeneratePrivKey()
	addr1 := priv1.PubKey().Address()
	acc1 := auth.NewBaseAccountWithAddress(addr1)
	acc1.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	priv2 := utils.GeneratePrivKey()
	addr2 := priv2.PubKey().Address()
	acc2 := auth.NewBaseAccountWithAddress(addr2)
	acc2.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	}})

	priv3 := utils.GeneratePrivKey()
	addr3 := priv3.PubKey().Address()
	acc3 := auth.NewBaseAccountWithAddress(addr3)
	acc3.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	}})

	priv4 := utils.GeneratePrivKey()
	addr4 := priv4.PubKey().Address()
	acc4 := auth.NewBaseAccountWithAddress(addr4)
	acc4.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	}})

	priv5 := utils.GeneratePrivKey()
	addr5 := priv5.PubKey().Address()
	acc5 := auth.NewBaseAccountWithAddress(addr5)
	acc5.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	err := setGenesis(rapp, acc1, acc2, acc3, acc4, acc5)
	if err != nil {
		panic(err)
	}

	// Set two declare tx. First one will be challenged, second one will be unchallenged.
	msg1 := tcr.NewDeclareCandidacyMsg(addr1, "Unique registry listing 1", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})
	msg2 := tcr.NewDeclareCandidacyMsg(addr2, "Unique registry listing 2", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})

	fee := auth.StdFee{Gas: 10000000}

	declareTx1 := GenTx(t.Name(), msg1, fee, priv1, 0)
	declareTx2 := GenTx(t.Name(), msg2, fee, priv2, 0)

	header := abci.Header{AppHash: []byte("apphash"), ChainID: t.Name()}
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx1 in Block 0
	res := rapp.Deliver(declareTx1)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// deliver tx2 in Block 1
	header.Height += 1
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	res = rapp.Deliver(declareTx2)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := rapp.NewContext(false, header)

	// Check that ApplyEndBlockStamp is correct
	endStamp1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 1").EndApplyBlockStamp
	endStamp2 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 2").EndApplyBlockStamp

	require.Equal(t, int64(10), endStamp1, "End stamp for candidate 1 is wrong")
	require.Equal(t, int64(11), endStamp2, "End stamp for candidate 2 is wrong")

	// Check that head of queue is candidate 1
	head := rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing 1", head, "Head of queue is wrong")

	// Check that candidate 2 is in queue
	require.True(t, rapp.ballotKeeper.ProposalQueueContains(ctx, "Unique registry listing 2"), "Candidate 2 is not in queue")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Challenge candidate 1
	challengeMsg := tcr.NewChallengeMsg(addr3, "Unique registry listing 1", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})
	challengeTx := GenTx(t.Name(), challengeMsg, fee, priv3, 0)

	header.Height = 5
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	res = rapp.Deliver(challengeTx)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)

	// Check ballot 1 updated correctly
	ballot1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 1")
	require.True(t, ballot1.Active, "Ballot 1 unactivated")
	require.Equal(t, int64(25), ballot1.EndApplyBlockStamp, "Ballot 1 end blockstamp not updated after challenge")

	// Check that candidate 2 is now at head of queue
	head = rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing 2", head, "Head of queue is wrong")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	cdc := MakeCodec()

	commitMsg1, nonce1 := makeCommitment(cdc, addr4, "Unique registry listing 1", false)
	commitMsg2, nonce2 := makeCommitment(cdc, addr5, "Unique registry listing 1", false)
	commitMsg3, nonce3 := makeCommitment(cdc, addr1, "Unique registry listing 1", true)

	commitTx1 := GenTx(t.Name(), commitMsg1, fee, priv4, 0)
	commitTx2 := GenTx(t.Name(), commitMsg2, fee, priv5, 0)
	commitTx3 := GenTx(t.Name(), commitMsg3, fee, priv1, 1)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver commit Msgs
	res = rapp.Deliver(commitTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Move to reveal phase
	header.Height = 16

	revealMsg1 := tcr.NewRevealMsg(addr4, "Unique registry listing 1", false, nonce1, sdk.Coin{"RegistryCoin", 100})
	revealMsg2 := tcr.NewRevealMsg(addr5, "Unique registry listing 1", false, nonce2, sdk.Coin{"RegistryCoin", 300})
	revealMsg3 := tcr.NewRevealMsg(addr1, "Unique registry listing 1", true, nonce3, sdk.Coin{"RegistryCoin", 100})

	revealTx1 := GenTx(t.Name(), revealMsg1, fee, priv4, 1)
	revealTx2 := GenTx(t.Name(), revealMsg2, fee, priv5, 1)
	revealTx3 := GenTx(t.Name(), revealMsg3, fee, priv1, 2)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver reveal Msgs
	res = rapp.Deliver(revealTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Candidate 2 should have been added to queue by now
	require.Equal(t, "Unique registry listing 2", rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing 2").Identifier, "Candidate 2 not added to registry after application phase")
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Move to apply phase
	header.Height = 30
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Candidate 1 passed the challenge. Should be added to registry.
	require.Equal(t, tcr.Listing{}, rapp.ballotKeeper.GetListing(ctx, "Unique registry listing 1"), "Candidate 1 added to registry after losing challenge")

	// Proposer wins refund on vote 100 but loses bond
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr1, sdk.Coins{{"RegistryCoin", 100}}), fmt.Sprintf("Proposer and winning voter has: %+v", rapp.accountKeeper.GetCoins(ctx, addr1)))

	// Winning voter gets refund along with his share of correct vote pool (1 / 4) * rest of challenger bond (50%)
	// 100 + 25
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr4, sdk.Coins{{"RegistryCoin", 125}}), fmt.Sprintf("Winning voter 1 has: %v", rapp.accountKeeper.GetCoins(ctx, addr4)))

	// Winning voter gets refund along with his share of correct vote pool (3 / 4) * rest of challenger bond (50%)
	// 300 + 75
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr5, sdk.Coins{{"RegistryCoin", 375}}), fmt.Sprintf("Winning voter 2 has: %v", rapp.accountKeeper.GetCoins(ctx, addr4)))

	// Challenger wins refund on his bond (200) and dispensationPct of owner bond
	// 200 + 100
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr3, sdk.Coins{{"RegistryCoin", 300}}), fmt.Sprintf("Challenger has: %+v", rapp.accountKeeper.GetCoins(ctx, addr3)))
}

// -----------------------------------------------------------------------------------------------------------------------------------
// Tests for challenging existing listing

func TestChallengeExistWonFlow(t *testing.T) {
	rapp := newRegistryApp()

	priv1 := utils.GeneratePrivKey()
	addr1 := priv1.PubKey().Address()
	acc1 := auth.NewBaseAccountWithAddress(addr1)
	acc1.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	priv2 := utils.GeneratePrivKey()
	addr2 := priv2.PubKey().Address()
	acc2 := auth.NewBaseAccountWithAddress(addr2)
	acc2.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	}})

	priv3 := utils.GeneratePrivKey()
	addr3 := priv3.PubKey().Address()
	acc3 := auth.NewBaseAccountWithAddress(addr3)
	acc3.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	priv4 := utils.GeneratePrivKey()
	addr4 := priv4.PubKey().Address()
	acc4 := auth.NewBaseAccountWithAddress(addr4)
	acc4.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	}})

	err := setGenesis(rapp, acc1, acc2, acc3, acc4)
	if err != nil {
		panic(err)
	}

	// Set two declare tx. First one will be challenged, second one will be unchallenged.
	msg1 := tcr.NewDeclareCandidacyMsg(addr1, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})

	fee := auth.StdFee{Gas: 10000000}

	declareTx1 := GenTx(t.Name(), msg1, fee, priv1, 0)

	header := abci.Header{AppHash: []byte("apphash"), ChainID: t.Name()}
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx1 in Block 0
	res := rapp.Deliver(declareTx1)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := rapp.NewContext(false, header)

	// Check that ApplyEndBlockStamp is correct
	endStamp1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing").EndApplyBlockStamp
	require.Equal(t, int64(10), endStamp1, "End stamp for candidate is wrong")

	// Check that head of queue is candidate
	head := rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing", head, "Head of queue is wrong")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Fast forward to apply phase
	header.Height = 10
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Check that candidate added to listing
	require.Equal(t, "Unique registry listing", rapp.ballotKeeper.GetListing(ctx, "Unique registry listing").Identifier, "Candidate not added to registry after application phase")

	// Challenge candidate after it was already added to registry
	challengeMsg := tcr.NewChallengeMsg(addr2, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})
	challengeTx := GenTx(t.Name(), challengeMsg, fee, priv2, 0)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	res = rapp.Deliver(challengeTx)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)

	// Check ballot 1 updated correctly
	ballot1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing")
	require.True(t, ballot1.Active, "Ballot unactivated")
	require.Equal(t, int64(30), ballot1.EndApplyBlockStamp, "Ballot end blockstamp not updated after challenge")
	// Listing should not be removed from registry after challenge
	require.Equal(t, "Unique registry listing", rapp.ballotKeeper.GetListing(ctx, "Unique registry listing").Identifier, "Candidate not added to registry after application phase")

	// Check that candidate is now at head of queue
	head = rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing", head, "Head of queue is wrong")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	cdc := MakeCodec()

	commitMsg1, nonce1 := makeCommitment(cdc, addr3, "Unique registry listing", true)
	commitMsg2, nonce2 := makeCommitment(cdc, addr4, "Unique registry listing", false)
	commitMsg3, nonce3 := makeCommitment(cdc, addr1, "Unique registry listing", true)

	commitTx1 := GenTx(t.Name(), commitMsg1, fee, priv3, 0)
	commitTx2 := GenTx(t.Name(), commitMsg2, fee, priv4, 0)
	commitTx3 := GenTx(t.Name(), commitMsg3, fee, priv1, 1)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver commit Msgs
	res = rapp.Deliver(commitTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Move to reveal phase
	header.Height = 30

	revealMsg1 := tcr.NewRevealMsg(addr3, "Unique registry listing", true, nonce1, sdk.Coin{"RegistryCoin", 300})
	revealMsg2 := tcr.NewRevealMsg(addr4, "Unique registry listing", false, nonce2, sdk.Coin{"RegistryCoin", 100})
	revealMsg3 := tcr.NewRevealMsg(addr1, "Unique registry listing", true, nonce3, sdk.Coin{"RegistryCoin", 100})

	
	revealTx1 := GenTx(t.Name(), revealMsg1, fee, priv3, 1)
	revealTx2 := GenTx(t.Name(), revealMsg2, fee, priv4, 1)
	revealTx3 := GenTx(t.Name(), revealMsg3, fee, priv1, 2)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver reveal Msgs and apply
	res = rapp.Deliver(revealTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Candidate 1 passed the challenge. Should be added to registry.
	require.Equal(t, "Unique registry listing", rapp.ballotKeeper.GetListing(ctx, "Unique registry listing").Identifier, "Candidate not added to registry after beating challenge")
	require.Equal(t, int64(400), rapp.ballotKeeper.GetListing(ctx, "Unique registry listing").Votes, "Candidate votes did not update")

	// Voters who vote wrong should be refunded: 100
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr4, sdk.Coins{{"RegistryCoin", 100}}), fmt.Sprintf("Losing voter has: %+v", rapp.accountKeeper.GetCoins(ctx, addr4)))

	// Proposer wins refund of vote and dispensation percent (50%) of challenger bond as well as his share of the correct vote pool (1 / 4) * rest of challenger bond (50%)
	// 100 + 100 + 25 = 225
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr1, sdk.Coins{{"RegistryCoin", 225}}), fmt.Sprintf("Proposer and winning voter has: %+v", rapp.accountKeeper.GetCoins(ctx, addr1)))

	// Winning voter gets refund along with his share of correct vote pool (3 / 4) * rest of challenger bond (50%)
	// 300 + 75
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr3, sdk.Coins{{"RegistryCoin", 375}}), fmt.Sprintf("Winning voter has: %v", rapp.accountKeeper.GetCoins(ctx, addr3)))

	// Challenger loses bond
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr3, sdk.Coins(nil)), fmt.Sprintf("Challenge has: %+v", rapp.accountKeeper.GetCoins(ctx, addr3)))
}

func TestChallengeExistLostFlow(t *testing.T) {
	rapp := newRegistryApp()

	priv1 := utils.GeneratePrivKey()
	addr1 := priv1.PubKey().Address()
	acc1 := auth.NewBaseAccountWithAddress(addr1)
	acc1.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	priv2 := utils.GeneratePrivKey()
	addr2 := priv2.PubKey().Address()
	acc2 := auth.NewBaseAccountWithAddress(addr2)
	acc2.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	}})

	priv3 := utils.GeneratePrivKey()
	addr3 := priv3.PubKey().Address()
	acc3 := auth.NewBaseAccountWithAddress(addr3)
	acc3.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 300,
	}})

	priv4 := utils.GeneratePrivKey()
	addr4 := priv4.PubKey().Address()
	acc4 := auth.NewBaseAccountWithAddress(addr4)
	acc4.SetCoins([]sdk.Coin{sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 100,
	}})

	err := setGenesis(rapp, acc1, acc2, acc3, acc4)
	if err != nil {
		panic(err)
	}

	// Set two declare tx. First one will be challenged, second one will be unchallenged.
	msg1 := tcr.NewDeclareCandidacyMsg(addr1, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})

	fee := auth.StdFee{Gas: 10000000}

	declareTx1 := GenTx(t.Name(), msg1, fee, priv1, 0)

	header := abci.Header{AppHash: []byte("apphash"), ChainID: t.Name()}
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx1 in Block 0
	res := rapp.Deliver(declareTx1)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := rapp.NewContext(false, header)

	// Check that ApplyEndBlockStamp is correct
	endStamp1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing").EndApplyBlockStamp
	require.Equal(t, int64(10), endStamp1, "End stamp for candidate is wrong")

	// Check that head of queue is candidate
	head := rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing", head, "Head of queue is wrong")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Fast forward to apply phase
	header.Height = 10
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Check that candidate added to listing
	require.Equal(t, "Unique registry listing", rapp.ballotKeeper.GetListing(ctx, "Unique registry listing").Identifier, "Candidate not added to registry after application phase")

	// Challenge candidate after it was already added to registry
	challengeMsg := tcr.NewChallengeMsg(addr2, "Unique registry listing", sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 200,
	})
	challengeTx := GenTx(t.Name(), challengeMsg, fee, priv2, 0)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	res = rapp.Deliver(challengeTx)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)

	// Check ballot 1 updated correctly
	ballot1 := rapp.ballotKeeper.GetBallot(ctx, "Unique registry listing")
	require.True(t, ballot1.Active, "Ballot unactivated")
	require.Equal(t, int64(30), ballot1.EndApplyBlockStamp, "Ballot end blockstamp not updated after challenge")
	// Listing should not be removed from registry after challenge
	require.Equal(t, "Unique registry listing", rapp.ballotKeeper.GetListing(ctx, "Unique registry listing").Identifier, "Candidate not added to registry after application phase")

	// Check that candidate is now at head of queue
	head = rapp.ballotKeeper.ProposalQueueHead(ctx).Identifier
	require.Equal(t, "Unique registry listing", head, "Head of queue is wrong")

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	cdc := MakeCodec()

	commitMsg1, nonce1 := makeCommitment(cdc, addr3, "Unique registry listing", false)
	commitMsg2, nonce2 := makeCommitment(cdc, addr4, "Unique registry listing", false)
	commitMsg3, nonce3 := makeCommitment(cdc, addr1, "Unique registry listing", true)

	commitTx1 := GenTx(t.Name(), commitMsg1, fee, priv3, 0)
	commitTx2 := GenTx(t.Name(), commitMsg2, fee, priv4, 0)
	commitTx3 := GenTx(t.Name(), commitMsg3, fee, priv1, 1)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver commit Msgs
	res = rapp.Deliver(commitTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(commitTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	// Move to reveal phase
	header.Height = 30

	revealMsg1 := tcr.NewRevealMsg(addr3, "Unique registry listing", false, nonce1, sdk.Coin{"RegistryCoin", 300})
	revealMsg2 := tcr.NewRevealMsg(addr4, "Unique registry listing", false, nonce2, sdk.Coin{"RegistryCoin", 100})
	revealMsg3 := tcr.NewRevealMsg(addr1, "Unique registry listing", true, nonce3, sdk.Coin{"RegistryCoin", 100})

	
	revealTx1 := GenTx(t.Name(), revealMsg1, fee, priv3, 1)
	revealTx2 := GenTx(t.Name(), revealMsg2, fee, priv4, 1)
	revealTx3 := GenTx(t.Name(), revealMsg3, fee, priv1, 2)

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Deliver reveal Msgs and apply
	res = rapp.Deliver(revealTx1)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx2)
	require.True(t, res.IsOK(), res.Log)
	res = rapp.Deliver(revealTx3)
	require.True(t, res.IsOK(), res.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()
	
	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx = rapp.NewContext(false, header)
	// Candidate failed the challenge. Should be removed from registry.
	require.Equal(t, tcr.Listing{}, rapp.ballotKeeper.GetListing(ctx, "Unique registry listing"), "Listing not removed from registry after losing challenge")

	// Proposer who voted wrong should be refunded: 100
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr1, sdk.Coins{{"RegistryCoin", 100}}), fmt.Sprintf("Losing proposer has: %+v", rapp.accountKeeper.GetCoins(ctx, addr4)))

	// Winning voter gets refund along with his share of correct vote pool (3 / 4) * rest of challenger bond (50%)
	// 300 + 75
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr3, sdk.Coins{{"RegistryCoin", 375}}), fmt.Sprintf("Winning voter 1 has: %v", rapp.accountKeeper.GetCoins(ctx, addr3)))

	// Winning voter gets refund along with his share of correct vote pool (1 / 4) * rest of challenger bond (50%)
	// 100 + 25
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr3, sdk.Coins{{"RegistryCoin", 125}}), fmt.Sprintf("Winning voter 2 has: %v", rapp.accountKeeper.GetCoins(ctx, addr3)))

	// Challenger wins refund on bond + dispensationPct of Owner Bond
	// 200 + 100 = 300
	require.True(t, rapp.accountKeeper.HasCoins(ctx, addr3, sdk.Coins{{"RegistryCoin", 300}}), fmt.Sprintf("Challenge has: %+v", rapp.accountKeeper.GetCoins(ctx, addr3)))
}

func makeCommitment(cdc *amino.Codec, owner sdk.Address, identifier string, vote bool) (commitMsg tcr.CommitMsg, nonce []byte) {
	hasher := sha256.New()
	vz, _ := cdc.MarshalBinary(vote)
	hasher.Sum(vz)

	hasher2 := sha256.New()
	bz, _ := cdc.MarshalBinary(rand.Int())
	nonce = hasher2.Sum(bz)

	commitment := hasher.Sum(nonce)

	commitMsg = tcr.NewCommitMsg(owner, identifier, commitment)
	return
}

func GenTx(chainID string, msg sdk.Msg, fee auth.StdFee, priv crypto.PrivKey, seq int64) auth.StdTx {
	signBytes := auth.StdSignBytes(chainID, []int64{seq}, fee, msg)
	sig := priv.Sign(signBytes)

	return auth.NewStdTx(msg, fee, []auth.StdSignature{auth.StdSignature{
		priv.PubKey(),
		sig,
		seq,
	}})
}
