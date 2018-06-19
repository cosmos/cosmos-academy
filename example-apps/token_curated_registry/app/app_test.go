package app

import (
	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	"github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/utils"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
	"os"
	"testing"

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

	assert.Equal(t, true, privKey.PubKey().VerifyBytes(signBytes, sig), "Sig doesn't work")

	tx := auth.StdTx{
		Msg: msg,
		Signatures: []auth.StdSignature{auth.StdSignature{
			privKey.PubKey(),
			sig,
			0,
		}},
	}

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
	assert.Equal(t, sdk.CodeType(5),
		sdk.CodeType(cres.Code), cres.Log)

	dres := rapp.DeliverTx(txBytes)
	assert.Equal(t, sdk.CodeType(5), sdk.CodeType(dres.Code), dres.Log)

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
	assert.Equal(t, sdk.CodeType(4),
		sdk.CodeType(cres.Code), cres.Log)

	// Simulate a Block
	rapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: t.Name()}})
	dres := rapp.DeliverTx(txBytes)
	assert.Equal(t, sdk.CodeType(4), sdk.CodeType(dres.Code), dres.Log)

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

	signBytes := auth.StdSignBytes(t.Name(), []int64{0}, auth.StdFee{Gas: 10000000}, msg)
	sig := privKey.Sign(signBytes)

	tx := auth.StdTx{
		Msg: msg,
		Fee: auth.StdFee{Gas: 10000000},
		Signatures: []auth.StdSignature{auth.StdSignature{
			privKey.PubKey(),
			sig,
			0,
		}},
	}

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
	assert.Equal(t, sdk.CodeType(0),
		sdk.CodeType(cres.Code), cres.Log)

	// Simulate a Block
	dres := rapp.Deliver(tx)
	assert.Equal(t, sdk.CodeType(0), sdk.CodeType(dres.Code), dres.Log)

	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()


	// Mine 10 empty blocks
	for i := 0; i < 10; i++ {
		header.Height = int64(i + 1)
		rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
		rapp.EndBlock(abci.RequestEndBlock{})
		rapp.Commit()
	}

	applyMsg := tcr.NewApplyMsg(addr, "Unique registry listing")

	signBytes = auth.StdSignBytes(t.Name(), []int64{1}, auth.StdFee{Gas: 10000000}, applyMsg)
	sig = privKey.Sign(signBytes)

	applyTx := auth.NewStdTx(applyMsg, auth.StdFee{Gas: 10000000}, []auth.StdSignature{auth.StdSignature{
		privKey.PubKey(),
		sig,
		1,
	}})

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	applyRes := rapp.Deliver(applyTx)

	assert.Equal(t, sdk.CodeType(0), sdk.CodeType(applyRes.Code), applyRes.Log)

	ctx := rapp.NewContext(false, header)

	store := ctx.KVStore(rapp.capKeyListings)

	listing := tcr.Listing{
		Identifier: "Unique registry listing",
		Votes:      0,
	}
	expected, _ := rapp.cdc.MarshalBinary(listing)
	actual := store.Get([]byte("Unique registry listing"))

	assert.Equal(t, expected, actual, "Listing not added correctly to registry")

}
