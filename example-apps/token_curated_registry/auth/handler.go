package auth

import (
	"crypto/sha256"
	db "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/db"
	tcr "github.com/cosmos/cosmos-academy/example-apps/token_curated_registry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/tendermint/go-amino"
	"reflect"
)

func NewCandidacyHandler(accountKeeper bank.Keeper, ballotKeeper db.BallotKeeper, minBond int64, applyLen int64) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		declareMsg := msg.(tcr.DeclareCandidacyMsg)
		if declareMsg.Identifier == "" || declareMsg.Identifier == "candidateQueue" {
			return tcr.ErrInvalidBallot(tcr.DefaultCodespace, "Cannot use reserved identifiers for ballot").Result()
		}
		if declareMsg.Deposit.Amount < minBond {
			return sdk.ErrInsufficientFunds("Must send at least the minimum bond").Result()
		}
		_, _, err := accountKeeper.SubtractCoins(ctx, declareMsg.Owner, []sdk.Coin{declareMsg.Deposit})

		if err != nil {
			return err.Result()
		}

		ballot := ballotKeeper.GetBallot(ctx, declareMsg.Identifier)
		if !reflect.DeepEqual(ballot, tcr.Ballot{}) {
			return tcr.ErrInvalidBallot(2, "Candidate already exists").Result()
		}

		err2 := ballotKeeper.AddBallot(ctx, declareMsg.Identifier, declareMsg.Owner, applyLen, declareMsg.Deposit.Amount)
		if err2 != nil {
			return err2.Result()
		}

		ballotKeeper.ProposalQueuePush(ctx, declareMsg.Identifier, ctx.BlockHeight() + applyLen)

		return sdk.Result{}
	}
}

func NewChallengeHandler(accountKeeper bank.Keeper, ballotKeeper db.BallotKeeper, commitLen int64, revealLen int64, minBond int64) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		challengeMsg := msg.(tcr.ChallengeMsg)
		_, _, err := accountKeeper.SubtractCoins(ctx, challengeMsg.Owner, []sdk.Coin{challengeMsg.Bond})
		if err != nil {
			return err.Result()
		}

		store := ctx.KVStore(ballotKeeper.BallotKey)
		key := []byte(challengeMsg.Identifier)
		bz := store.Get(key)

		if bz == nil {
			return tcr.ErrInvalidBallot(2,"Candidate with given identifier does not exist").Result()
		}
		ballot := &tcr.Ballot{}
		err2 := ballotKeeper.Cdc.UnmarshalBinary(bz, ballot)
		if err2 != nil {
			panic(err2)
		}

		if ballot.EndCommitBlockStamp != 0 {
			return tcr.ErrInvalidPhase(2, "Candidate has already been challenged").Result()
		}

		if challengeMsg.Bond.Amount < ballot.Bond {
			return tcr.ErrInvalidBond(2, "Must match candidate bond to challenge").Result()
		}

		err3 := ballotKeeper.ActivateBallot(ctx, accountKeeper, ballot.Owner, challengeMsg.Owner, challengeMsg.Identifier, commitLen, revealLen, minBond, challengeMsg.Bond.Amount)
		if err3 != nil {
			return err3.Result()
		}

		if ballotKeeper.ProposalQueueContains(ctx, challengeMsg.Identifier) {
			ballotKeeper.ProposalQueueUpdate(ctx, challengeMsg.Identifier, ctx.BlockHeight() + commitLen + revealLen)
		} else {
			ballotKeeper.ProposalQueuePush(ctx, challengeMsg.Identifier, ctx.BlockHeight() + commitLen + revealLen)
		}

		return sdk.Result{}
	}
}

func NewCommitHandler(cdc *amino.Codec, ballotKeeper db.BallotKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		commitMsg := msg.(tcr.CommitMsg)

		candidate := ballotKeeper.GetBallot(ctx, commitMsg.Identifier)

		if reflect.DeepEqual(candidate, tcr.Ballot{}) {
			return tcr.ErrInvalidBallot(2, "Candidate with given identifier does not exist").Result()
		}

		if candidate.EndCommitBlockStamp == 0 || candidate.EndCommitBlockStamp < ctx.BlockHeight() {
			return tcr.ErrInvalidPhase(2, "Candidate not in commit phase").Result()
		}

		ballotKeeper.CommitBallot(ctx, commitMsg.Owner, commitMsg.Identifier, commitMsg.Commitment)
		return sdk.Result{}
	}
}

func NewRevealHandler(accountKeeper bank.Keeper, ballotKeeper db.BallotKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		revealMsg := msg.(tcr.RevealMsg)
		_, _, err := accountKeeper.SubtractCoins(ctx, revealMsg.Owner, []sdk.Coin{revealMsg.Bond})
		if err != nil {
			return err.Result()
		}

		candidate := ballotKeeper.GetBallot(ctx, revealMsg.Identifier)

		if reflect.DeepEqual(candidate, tcr.Ballot{}) {
			return tcr.ErrInvalidBallot(2, "Candidate with given identifier does not exist").Result()
		}

		if candidate.EndCommitBlockStamp > ctx.BlockHeight() || candidate.EndApplyBlockStamp < ctx.BlockHeight() {
			return tcr.ErrInvalidPhase(2, "Candidate not in reveal phase").Result()
		}

		if !reflect.DeepEqual(ballotKeeper.GetVote(ctx, revealMsg.Owner, revealMsg.Identifier), tcr.Vote{}) {
			return tcr.ErrInvalidVote(2, "Already voted").Result()
		}

		commitment := ballotKeeper.GetCommitment(ctx, revealMsg.Owner, revealMsg.Identifier)

		hasher := sha256.New()
		vz, _ := ballotKeeper.Cdc.MarshalBinary(revealMsg.Vote)
		hasher.Sum(vz)
		val := hasher.Sum(revealMsg.Nonce)

		if !reflect.DeepEqual(val, commitment) {
			return tcr.ErrInvalidVote(2, "Vote does not match commitment").Result()
		}

		ballotKeeper.VoteBallot(ctx, revealMsg.Owner, revealMsg.Identifier, revealMsg.Vote, revealMsg.Bond.Amount)

		ballotKeeper.DeleteCommitment(ctx, revealMsg.Owner, revealMsg.Identifier)

		return sdk.Result{}
	}
}
