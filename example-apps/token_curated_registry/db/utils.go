package db

import (
	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
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
	tcr.RegisterAmino(cdc)
	crypto.RegisterAmino(cdc)
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	return cdc
}
