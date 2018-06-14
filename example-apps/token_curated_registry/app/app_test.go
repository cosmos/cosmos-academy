package app

import (
	"testing"
	"os"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tmlibs/log"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/AdityaSripal/token_curated_registry/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/abci/types"
	"github.com/AdityaSripal/token_curated_registry/utils"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newRegistryApp() *RegistryApp {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return NewRegistryApp(logger, db, 100, 10, 10, 10, 0.5, 0.5)
}

func setGenesis(rapp *RegistryApp, accs ...auth.BaseAccount) error {
	genaccs := make([]*types.GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = types.NewGenesisAccount(&acc)
	}

	genesisState := types.GenesisState{
		Accounts:  genaccs,
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
		Denom: "RegistryCoin",
		Amount: 50,
	}})

	err := setGenesis(rapp, acc)
	if err != nil {
		panic(err)
	}

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 50,
	})

	sig := privKey.Sign(msg.GetSignBytes())

	assert.Equal(t, true, privKey.PubKey().VerifyBytes(msg.GetSignBytes(), sig), "Sig doesn't work")

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

	// Run a check
	cres := rapp.CheckTx(txBytes)
	assert.Equal(t, sdk.CodeType(5),
		sdk.CodeType(cres.Code), cres.Log)

	// Simulate a Block
	rapp.BeginBlock(abci.RequestBeginBlock{})
	dres := rapp.DeliverTx(txBytes)
	assert.Equal(t, sdk.CodeType(5), sdk.CodeType(dres.Code), dres.Log)
	
}

func TestBadTx(t *testing.T) {
	rapp := newRegistryApp()

	privKey := utils.GeneratePrivKey()
	addr := privKey.PubKey().Address()
	acc := auth.NewBaseAccountWithAddress(addr)

	acc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	}})

	err := setGenesis(rapp, acc)
	if err != nil {
		panic(err)
	}

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
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
	rapp.BeginBlock(abci.RequestBeginBlock{})
	dres := rapp.DeliverTx(txBytes)
	assert.Equal(t, sdk.CodeType(4), sdk.CodeType(dres.Code), dres.Log)
	
}

func TestApplyUnchallengedFlow(t *testing.T) {
	rapp := newRegistryApp()

	privKey := utils.GeneratePrivKey()
	addr := privKey.PubKey().Address()
	acc := auth.NewBaseAccountWithAddress(addr)

	acc.SetCoins([]sdk.Coin{sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	}})

	err := setGenesis(rapp, acc)
	if err != nil {
		panic(err)
	}

	msg := types.NewDeclareCandidacyMsg(addr, "Unique registry listing", sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 100,
	})

	sig := privKey.Sign(msg.GetSignBytes())

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

	// Run a check
	cres := rapp.CheckTx(txBytes)
	assert.Equal(t, sdk.CodeType(0),
		sdk.CodeType(cres.Code), cres.Log)

	// Simulate a Block
	rapp.BeginBlock(abci.RequestBeginBlock{})
	dres := rapp.Deliver(tx)
	assert.Equal(t, sdk.CodeType(0), sdk.CodeType(dres.Code), dres.Log)
	
	rapp.EndBlock(abci.RequestEndBlock{})
	rapp.Commit()

	header := abci.Header{AppHash: []byte("apphash")}

	// Mine 10 empty blocks
	for i := 0; i < 10; i++ {
		header.Height = int64(i + 1)
		rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
		rapp.EndBlock(abci.RequestEndBlock{})
		rapp.Commit()
	}

	applyMsg := types.NewApplyMsg(addr, "Unique registry listing")

	sig = privKey.Sign(applyMsg.GetSignBytes())

	applyTx := auth.NewStdTx(applyMsg, auth.StdFee{}, []auth.StdSignature{auth.StdSignature{
		privKey.PubKey(),
		sig,
		0,
	}})

	rapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	applyRes := rapp.Deliver(applyTx)

	assert.Equal(t, sdk.CodeType(0), sdk.CodeType(applyRes.Code), applyRes.Log)

	ctx := rapp.NewContext(false, header)

	store := ctx.KVStore(rapp.capKeyListings)

	listing := types.Listing{
		Identifier: "Unique registry listing",
		Votes: 0,
	}
	expected, _ := rapp.cdc.MarshalBinary(listing)
	actual := store.Get([]byte("Unique registry listing"))

	assert.Equal(t, expected, actual, "Listing not added correctly to registry")

}