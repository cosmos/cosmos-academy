package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	amino "github.com/tendermint/go-amino"
)

const (
	TokenName = "RegistryCoin"
)

// ===================================================================================================================================

// DeclareCandidacyMsg is used to propose a new listing to be added to registry.
// Identifier is a unique identifier of the listing
// Deposit is taken and held for entire duration of listing. Awarded to challengers upon successful challenge.
type DeclareCandidacyMsg struct {
	Owner      sdk.Address
	Identifier string
	Details string
	Deposit       sdk.Coin
}

func NewDeclareCandidacyMsg(owner sdk.Address, identifier string, bond sdk.Coin) DeclareCandidacyMsg {
	return DeclareCandidacyMsg{
		Owner:      owner,
		Identifier: identifier,
		Deposit:       bond,
	}
}

func (msg DeclareCandidacyMsg) Type() string {
	return "DeclareCandidacy"
}

func (msg DeclareCandidacyMsg) ValidateBasic() sdk.Error {
	if msg.Owner == nil {
		return sdk.ErrInvalidAddress("Must provide Owner Address")
	}
	if !msg.Deposit.IsPositive() || msg.Deposit.Denom != TokenName {
		return ErrInvalidDeposit(2, "Must provide Deposit in RegistryCoin")
	}
	return nil
}

func (msg DeclareCandidacyMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func (msg DeclareCandidacyMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Owner}
}

// ===================================================================================================================================

// ChallengeMsg is used to challenge a pending or finalized listing
type ChallengeMsg struct {
	Owner      sdk.Address
	Identifier string
	Bond       sdk.Coin
}

func NewChallengeMsg(owner sdk.Address, identifier string, bond sdk.Coin) ChallengeMsg {
	return ChallengeMsg{
		Owner:      owner,
		Identifier: identifier,
		Bond:       bond,
	}
}

func (msg ChallengeMsg) Type() string {
	return "Challenge"
}

func (msg ChallengeMsg) ValidateBasic() sdk.Error {
	if msg.Owner == nil {
		return sdk.ErrInvalidAddress("Must provide Owner Address")
	}
	if  !msg.Bond.IsPositive() || msg.Bond.Denom != TokenName {
		return ErrInvalidDeposit(2, "Must provide deposit in RegistryCoin")
	}
	return nil
}

func (msg ChallengeMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func (msg ChallengeMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Owner}
}

// ===================================================================================================================================

// CommitMsg is used to make a commitment during commit phase on an active challenge to a specific listing identified by Identifier.
type CommitMsg struct {
	Owner      sdk.Address
	Identifier string
	Commitment []byte
}

func NewCommitMsg(owner sdk.Address, identifier string, commitment []byte) CommitMsg {
	return CommitMsg{
		Owner:      owner,
		Identifier: identifier,
		Commitment: commitment,
	}
}

func (msg CommitMsg) Type() string {
	return "Commit"
}

func (msg CommitMsg) ValidateBasic() sdk.Error {
	if msg.Owner == nil {
		return sdk.ErrInvalidAddress("Must provide owner address")
	}
	return nil
}

func (msg CommitMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func (msg CommitMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Owner}
}

// ===================================================================================================================================

// RevealMsg is to reveal vote during reveal phase on active challenge to listing identified by Identifier.
type RevealMsg struct {
	Owner      sdk.Address
	Identifier string
	Vote       bool
	Nonce      []byte
	Bond       sdk.Coin
}

func NewRevealMsg(owner sdk.Address, identifier string, vote bool, nonce []byte, bond sdk.Coin) RevealMsg {
	return RevealMsg{
		Owner:      owner,
		Identifier: identifier,
		Vote:       vote,
		Nonce:      nonce,
		Bond:       bond,
	}
}

func (msg RevealMsg) Type() string {
	return "Reveal"
}

func (msg RevealMsg) ValidateBasic() sdk.Error {
	if msg.Owner == nil {
		return sdk.ErrInvalidAddress("Must provide Owner address")
	}
	if msg.Bond.Amount <= 0 || msg.Bond.Denom != TokenName {
		return ErrInvalidDeposit(2, "Must provide Bond in RegistryCoin")
	}
	return nil
}

func (msg RevealMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func (msg RevealMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Owner}
}

func RegisterAmino(cdc *amino.Codec) {
	cdc.RegisterConcrete(DeclareCandidacyMsg{}, "types/DeclareCandidacyMsg", nil)
	cdc.RegisterConcrete(ChallengeMsg{}, "types/ChallengeMsg", nil)
	cdc.RegisterConcrete(CommitMsg{}, "types/CommintMsg", nil)
	cdc.RegisterConcrete(RevealMsg{}, "types/RevealMsg", nil)
	cdc.RegisterConcrete(Listing{}, "types/Listing", nil)
	cdc.RegisterConcrete(Vote{}, "types/Vote", nil)
	cdc.RegisterConcrete(Ballot{}, "types/Ballot", nil)
}
