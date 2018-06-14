package types

import (
	crypto "github.com/tendermint/go-crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenerateCandidacyMsg() DeclareCandidacyMsg {
	coin := sdk.Coin{
		Denom: "RegistryCoin",
		Amount: 5000,
	}

	return DeclareCandidacyMsg{
		Owner: crypto.GenPrivKeyEd25519().PubKey().Address(),
		Identifier: "Unique registry listing",
		Bond: coin,
	}
}