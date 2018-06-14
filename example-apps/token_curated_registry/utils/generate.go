package utils

import (
	"github.com/tendermint/go-crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenerateAddress() sdk.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

func GeneratePrivKey() crypto.PrivKey {
	return crypto.GenPrivKeyEd25519()
}