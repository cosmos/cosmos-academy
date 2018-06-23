package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidMsg(t *testing.T) {
	msg := GenerateCandidacyMsg()

	err := msg.ValidateBasic()

	assert.Nil(t, err)
}

func TestInvalidDenom(t *testing.T) {
	msg := GenerateCandidacyMsg()

	msg.Deposit.Denom = "FakeCoin"
	err := msg.ValidateBasic()

	assert.Equal(t, sdk.CodeType(101), err.Code(), err.Error())
}

func TestInvalidAmount(t *testing.T) {
	msg := GenerateCandidacyMsg()

	msg.Deposit.Amount = 0
	err := msg.ValidateBasic()

	assert.Equal(t, sdk.CodeType(101), err.Code(), err.Error())

	msg.Deposit.Amount = -100
	err = msg.ValidateBasic()

	assert.Equal(t, sdk.CodeType(101), err.Code(), err.Error())
}
