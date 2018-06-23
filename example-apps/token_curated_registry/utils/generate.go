package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-crypto"
)

func GenerateAddress() sdk.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

func GeneratePrivKey() crypto.PrivKey {
	return crypto.GenPrivKeyEd25519()
}
