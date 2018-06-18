package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Reserve errors 100 ~ 199
const (
	DefaultCodespace sdk.CodespaceType = 2

	CodeInvalidDeposit      sdk.CodeType = 101
	CodeInvalidBond         sdk.CodeType = 102
	CodeInvalidBallot       sdk.CodeType = 103
	CodeInvalidPhase        sdk.CodeType = 104
	CodeInvalidVote         sdk.CodeType = 105
	CodeInvalidUTXO         sdk.CodeType = 106
	CodeInvalidTransaction  sdk.CodeType = 107
)

func codeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

func ErrInvalidDeposit(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDeposit, msg)
}

func ErrInvalidBallot(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBallot, msg)
}

func ErrInvalidBond(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBond, msg)
}

func ErrInvalidPhase(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPhase, msg)
}

func ErrInvalidVote(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidVote, msg)
}