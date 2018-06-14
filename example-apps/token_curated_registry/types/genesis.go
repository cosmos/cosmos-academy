package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/wire"
)

type GenesisState struct {
	Accounts []*GenesisAccount `json:"accounts"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
}

func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (res auth.Account, err error) {
		if len(accBytes) == 0 {
			return nil, sdk.ErrTxDecode("accBytes are empty")
		}
		acct := new(auth.BaseAccount)
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}
		return acct, err
	}
}

func NewGenesisAccount(aa *auth.BaseAccount) *GenesisAccount {
	return &GenesisAccount{
		Address: aa.Address,
		Coins:   aa.Coins.Sort(),
	}
}

// convert GenesisAccount to AppAccount
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount, err error) {
	baseAcc := auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
	return &baseAcc, nil
}