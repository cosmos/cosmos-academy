package types

import (
	"testing"
	"github.com/stretchr/testify/assert"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidMsg(t *testing.T) {
	msg := GenerateCandidacyMsg()

	err := msg.ValidateBasic()
	
	assert.Nil(t, err) 
}

func TestInvalidDenom(t *testing.T) {
	msg := GenerateCandidacyMsg()

	msg.Bond.Denom = "FakeCoin"
	err := msg.ValidateBasic()

	assert.Equal(t, sdk.CodeType(101), err.Code(), err.Error())
}

func TestInvalidAmount(t *testing.T) {
	msg := GenerateCandidacyMsg()

	msg.Bond.Amount = 0
	err := msg.ValidateBasic()

	assert.Equal(t, sdk.CodeType(101), err.Code(), err.Error())

	msg.Bond.Amount = -100
	err = msg.ValidateBasic()

	assert.Equal(t, sdk.CodeType(101), err.Code(), err.Error())
}

