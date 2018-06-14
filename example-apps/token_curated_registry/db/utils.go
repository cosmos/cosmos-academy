package db

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/tendermint/go-amino"
	"github.com/AdityaSripal/token_curated_registry/types"
	"github.com/tendermint/go-crypto"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func SetupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey, *sdk.KVStoreKey, *sdk.KVStoreKey, *sdk.KVStoreKey, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	listKey := sdk.NewKVStoreKey("ListKey")
	ballotKey := sdk.NewKVStoreKey("BallotKey")
	commitKey := sdk.NewKVStoreKey("CommitKey")
	revealKey := sdk.NewKVStoreKey("RevealKey")
	accountKey := sdk.NewKVStoreKey("AccountKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(listKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(ballotKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(commitKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(revealKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(accountKey, sdk.StoreTypeIAVL, db)

	ms.LoadLatestVersion()
	return ms, listKey, ballotKey, commitKey, revealKey, accountKey
}

func MakeCodec() *amino.Codec {
	cdc := amino.NewCodec()
	types.RegisterAmino(cdc)
	crypto.RegisterAmino(cdc)
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	return cdc
}