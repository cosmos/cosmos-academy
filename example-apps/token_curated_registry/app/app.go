package app

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	cmn "github.com/tendermint/tmlibs/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-amino"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	handle "github.com/AdityaSripal/token_curated_registry/auth"
	dbl "github.com/AdityaSripal/token_curated_registry/db"
	"github.com/AdityaSripal/token_curated_registry/types"
	"github.com/tendermint/go-crypto"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	appName = "Registry"
)

// Extended ABCI application
type RegistryApp struct {
	*bam.BaseApp

	cdc *amino.Codec

	minDeposit int64

	applyStage int64

	commitStage int64

	revealStage int64

	dispensationPct float64

	quorum float64

	// keys to access the substores
	capKeyMain *sdk.KVStoreKey
	capKeyAccount *sdk.KVStoreKey
	capKeyListings *sdk.KVStoreKey
	capKeyCommits *sdk.KVStoreKey
	capKeyReveals *sdk.KVStoreKey
	capKeyBallots *sdk.KVStoreKey
	capKeyFees *sdk.KVStoreKey

	ballotMapper dbl.BallotMapper

	// Manage addition and subtraction of account balances
	accountMapper auth.AccountMapper
	accountKeeper bank.Keeper
}

func NewRegistryApp(logger log.Logger, db dbm.DB, mindeposit int64, applystage int64, commitstage int64, revealstage int64, dispensationpct float64, _quorum float64) *RegistryApp {
	cdc := MakeCodec()
	var app = &RegistryApp{
		BaseApp: bam.NewBaseApp(appName, cdc, logger, db),
		cdc: cdc,
		minDeposit: mindeposit,
		applyStage: applystage,
		commitStage: commitstage,
		revealStage: revealstage,
		dispensationPct: dispensationpct,
		quorum: _quorum,
		capKeyMain: sdk.NewKVStoreKey("main"),
		capKeyAccount: sdk.NewKVStoreKey("acc"),
		capKeyFees: sdk.NewKVStoreKey("fee"),
		capKeyListings: sdk.NewKVStoreKey("listings"),
		capKeyCommits: sdk.NewKVStoreKey("commits"),
		capKeyReveals: sdk.NewKVStoreKey("reveals"),
		capKeyBallots: sdk.NewKVStoreKey("ballots"),
	}

	app.ballotMapper = dbl.NewBallotMapper(app.capKeyListings, app.capKeyBallots, app.capKeyCommits, app.capKeyReveals, app.cdc)
	app.accountMapper = auth.NewAccountMapper(app.cdc, app.capKeyAccount, &auth.BaseAccount{})
	app.accountKeeper =  bank.NewKeeper(app.accountMapper)

	app.Router().
		AddRoute("DeclareCandidacy", handle.NewCandidacyHandler(app.accountKeeper, app.ballotMapper, app.minDeposit, app.applyStage)).
		AddRoute("Challenge", handle.NewChallengeHandler(app.accountKeeper, app.ballotMapper, app.commitStage, app.revealStage, app.minDeposit)).
		AddRoute("Commit", handle.NewCommitHandler(app.cdc, app.capKeyBallots, app.capKeyCommits)).
		AddRoute("Reveal", handle.NewRevealHandler(app.accountKeeper, app.ballotMapper)).
		AddRoute("Apply", handle.NewApplyHandler(app.accountKeeper, app.ballotMapper, app.capKeyListings, app.quorum, app.dispensationPct)).
		AddRoute("ClaimReward", handle.NewClaimRewardHandler(app.cdc, app.accountKeeper, app.capKeyBallots, app.capKeyReveals, app.capKeyListings, app.dispensationPct))

	app.SetTxDecoder(app.txDecoder)
	app.SetInitChainer(app.initChainer)
	app.MountStoresIAVL(app.capKeyMain, app.capKeyAccount, app.capKeyFees, app.capKeyListings, app.capKeyCommits, app.capKeyReveals, app.capKeyBallots)
	app.SetAnteHandler(handle.NewAnteHandler(app.accountMapper))

	err := app.LoadLatestVersion(app.capKeyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}


	return app
}

func (app *RegistryApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(types.GenesisState)
	err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAccount()
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}
		app.accountMapper.SetAccount(ctx, acc)
	}
	return abci.ResponseInitChain{}
}

func (app *RegistryApp) txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx = auth.StdTx{}
	err := app.cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxDecode("")
	}
	return tx, nil
}

func MakeCodec() *amino.Codec {
	cdc := amino.NewCodec()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	types.RegisterAmino(cdc)
	crypto.RegisterAmino(cdc)
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	return cdc
}

// Custom logic for state export
func (app *RegistryApp) ExportAppStateJSON() (appState json.RawMessage, err error) {
	ctx := app.NewContext(true, abci.Header{})

	// iterate to get the accounts
	accounts := []*types.GenesisAccount{}
	appendAccount := func(acc auth.Account) (stop bool) {
		account := &types.GenesisAccount{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}
		accounts = append(accounts, account)
		return false
	}
	app.accountMapper.IterateAccounts(ctx, appendAccount)

	genState := types.GenesisState{
		Accounts: accounts,
	}
	return wire.MarshalJSONIndent(app.cdc, genState)
}