package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

func GenerateCandidacyMsg() DeclareCandidacyMsg {
	coin := sdk.Coin{
		Denom:  "RegistryCoin",
		Amount: 5000,
	}

	return DeclareCandidacyMsg{
		Owner:      crypto.GenPrivKeyEd25519().PubKey().Address(),
		Identifier: "Unique registry listing",
		Deposit:       coin,
	}
}
