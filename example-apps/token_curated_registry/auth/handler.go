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

func NewCandidacyHandler(accountKeeper bank.Keeper, ballotMapper db.BallotMapper, minBond int64, applyLen int64) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		declareMsg := msg.(tcr.DeclareCandidacyMsg)
		if declareMsg.Deposit.Amount < minBond {
			return sdk.ErrInsufficientFunds("Must send at least the minimum bond").Result()
		}
		_, _, err := accountKeeper.SubtractCoins(ctx, declareMsg.Owner, []sdk.Coin{declareMsg.Deposit})

		if err != nil {
			return err.Result()
		}

		ballot := ballotMapper.GetBallot(ctx, declareMsg.Identifier)
		if !reflect.DeepEqual(ballot, tcr.Ballot{}) {
			return tcr.ErrInvalidBallot(2, "Candidate already exists").Result()
		}

		err2 := ballotMapper.AddBallot(ctx, declareMsg.Identifier, declareMsg.Owner, applyLen, declareMsg.Deposit.Amount)
		if err2 != nil {
			return err2.Result()
		}
		return sdk.Result{}
	}
}

func NewChallengeHandler(accountKeeper bank.Keeper, ballotMapper db.BallotMapper, commitLen int64, revealLen int64, minBond int64) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		challengeMsg := msg.(tcr.ChallengeMsg)
		_, _, err := accountKeeper.SubtractCoins(ctx, challengeMsg.Owner, []sdk.Coin{challengeMsg.Bond})
		if err != nil {
			return err.Result()
		}

		store := ctx.KVStore(ballotMapper.BallotKey)
		key := []byte(challengeMsg.Identifier)
		bz := store.Get(key)

		if bz == nil {
			return tcr.ErrInvalidBallot(2,"Candidate with given identifier does not exist").Result()
		}
		ballot := &tcr.Ballot{}
		err2 := ballotMapper.Cdc.UnmarshalBinary(bz, ballot)
		if err2 != nil {
			panic(err2)
		}

		if ballot.EndCommitBlockStamp != 0 {
			return tcr.ErrInvalidPhase(2, "Candidate has already been challenged").Result()
		}

		if challengeMsg.Bond.Amount < ballot.Bond {
			return tcr.ErrInvalidBond(2, "Must match candidate bond to challenge").Result()
		}

		err3 := ballotMapper.ActivateBallot(ctx, accountKeeper, ballot.Owner, challengeMsg.Owner, challengeMsg.Identifier, commitLen, revealLen, minBond, challengeMsg.Bond.Amount)
		if err3 != nil {
			return err3.Result()
		}
		return sdk.Result{}
	}
}

func NewCommitHandler(cdc *amino.Codec, ballotKey, commitKey sdk.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		commitMsg := msg.(tcr.CommitMsg)

		store := ctx.KVStore(ballotKey)
		key := []byte(commitMsg.Identifier)
		bz := store.Get(key)

		if bz == nil {
			return tcr.ErrInvalidBallot(2, "Candidate with given identifier does not exist").Result()
		}
		candidate := &tcr.Ballot{}
		err2 := cdc.UnmarshalBinary(bz, candidate)
		if err2 != nil {
			panic(err2)
		}

		if candidate.EndCommitBlockStamp == 0 || candidate.EndCommitBlockStamp < ctx.BlockHeight() {
			return tcr.ErrInvalidPhase(2, "Candidate not in commit phase").Result()
		}

		commitStore := ctx.KVStore(commitKey)

		voter := tcr.Voter{
			Owner:      commitMsg.Owner,
			Identifier: commitMsg.Identifier,
		}
		voterKey, _ := cdc.MarshalBinary(voter)
		commitStore.Set(voterKey, commitMsg.Commitment)
		return sdk.Result{}
	}
}

func NewRevealHandler(accountKeeper bank.Keeper, ballotMapper db.BallotMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		revealMsg := msg.(tcr.RevealMsg)
		_, _, err := accountKeeper.SubtractCoins(ctx, revealMsg.Owner, []sdk.Coin{revealMsg.Bond})
		if err != nil {
			return err.Result()
		}

		store := ctx.KVStore(ballotMapper.BallotKey)
		key := []byte(revealMsg.Identifier)
		bz := store.Get(key)

		if bz == nil {
			return tcr.ErrInvalidBallot(2, "Candidate with given identifier does not exist").Result()
		}
		candidate := &tcr.Ballot{}
		err2 := ballotMapper.Cdc.UnmarshalBinary(bz, candidate)
		if err2 != nil {
			panic(err2)
		}

		if candidate.EndCommitBlockStamp > ctx.BlockHeight() || candidate.EndRevealBlockStamp < ctx.BlockHeight() {
			return tcr.ErrInvalidPhase(2, "Candidate not in reveal phase").Result()
		}

		commitStore := ctx.KVStore(ballotMapper.CommitKey)
		revealStore := ctx.KVStore(ballotMapper.RevealKey)

		voter := tcr.Voter{
			Owner:      revealMsg.Owner,
			Identifier: revealMsg.Identifier,
		}
		voterKey, _ := ballotMapper.Cdc.MarshalBinary(voter)
		if revealStore.Get(voterKey) != nil {
			return tcr.ErrInvalidVote(2, "Cannot vote more than once").Result()
		}

		commitment := commitStore.Get(voterKey)

		hasher := sha256.New()
		vz, _ := ballotMapper.Cdc.MarshalBinary(revealMsg.Vote)
		hasher.Sum(vz)
		val := hasher.Sum(revealMsg.Nonce)

		if !reflect.DeepEqual(val, commitment) {
			return tcr.ErrInvalidVote(2, "Vote does not match commitment").Result()
		}

		reveal := tcr.Vote{
			Choice: revealMsg.Vote,
			Power:  revealMsg.Bond.Amount,
		}
		revealVal, _ := ballotMapper.Cdc.MarshalBinary(reveal)

		commitStore.Delete(voterKey)
		revealStore.Set(voterKey, revealVal)

		err3 := ballotMapper.VoteBallot(ctx, revealMsg.Owner, revealMsg.Identifier, revealMsg.Vote, revealMsg.Bond.Amount)
		if err2 != nil {
			return err3.Result()
		}
		return sdk.Result{}
	}
}

func NewApplyHandler(accountKeeper bank.Keeper, ballotMapper db.BallotMapper, listingKey sdk.StoreKey, quorum float64, dispPct float64) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		applyMsg := msg.(tcr.ApplyMsg)

		store := ctx.KVStore(ballotMapper.BallotKey)
		key := []byte(applyMsg.Identifier)
		val := store.Get(key)
		ballot := &tcr.Ballot{}
		err := ballotMapper.Cdc.UnmarshalBinary(val, ballot)
		if err != nil {
			panic(err)
		}

		registry := ctx.KVStore(listingKey)

		if ballot.Active {
			if ctx.BlockHeight() < ballot.EndRevealBlockStamp {
				return tcr.ErrInvalidPhase(2, "Cannot apply until reveal phase ends").Result()
			}
		} else {
			if ctx.BlockHeight() < ballot.EndApplyBlockStamp {
				return tcr.ErrInvalidPhase(2, "Cannot apply until application phase ends").Result()
			} else {
				listing := tcr.Listing{
					Identifier: ballot.Identifier,
					Votes:      0,
				}
				val, _ := ballotMapper.Cdc.MarshalBinary(listing)
				registry.Set(key, val)
				return sdk.Result{}
			}
		}

		total := ballot.Approve + ballot.Deny

		if float64(ballot.Approve)/float64(total) > quorum {
			listing := tcr.Listing{
				Identifier: ballot.Identifier,
				Votes:      ballot.Approve,
			}
			entry, _ := ballotMapper.Cdc.MarshalBinary(listing)
			registry.Set(key, entry)

			reward := sdk.Coin{
				Denom:  "RegistryCoin",
				Amount: int64(float64(ballot.Bond) * dispPct),
			}
			_, _, err := accountKeeper.AddCoins(ctx, ballot.Owner, []sdk.Coin{reward})

			if err != nil {
				return err.Result()
			}

		} else {
			registry.Delete(key)

			// Challenger receives his original bond as well as dispPct of applier bond
			reward := sdk.Coin{
				Denom:  "RegistryCoin",
				Amount: int64(float64(ballot.Bond)*dispPct) + ballot.Bond,
			}
			_, _, err := accountKeeper.AddCoins(ctx, ballot.Challenger, []sdk.Coin{reward})

			if err != nil {
				return err.Result()
			}
		}

		ballot.Active = false
		val, _ = ballotMapper.Cdc.MarshalBinary(ballot)
		store.Set(key, val)

		return sdk.Result{}
	}
}

func NewClaimRewardHandler(cdc *amino.Codec, accountKeeper bank.Keeper, ballotKey sdk.StoreKey, revealKey sdk.StoreKey, listingKey sdk.StoreKey, dispPct float64) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		claimMsg := msg.(tcr.ClaimRewardMsg)
		revealStore := ctx.KVStore(revealKey)

		voter := tcr.Voter{
			Owner:      claimMsg.Owner,
			Identifier: claimMsg.Identifier,
		}
		key, _ := cdc.MarshalBinary(voter)
		bz := revealStore.Get(key)

		vote := &tcr.Vote{}
		err := cdc.UnmarshalBinary(bz, vote)
		if err != nil {
			panic(err)
		}

		registry := ctx.KVStore(listingKey)
		val := registry.Get([]byte(claimMsg.Identifier))

		ballotStore := ctx.KVStore(ballotKey)
		listKey := []byte(claimMsg.Identifier)
		ballot := &tcr.Ballot{}
		lz := ballotStore.Get(listKey)

		err = cdc.UnmarshalBinary(lz, ballot)
		if err != nil {
			panic(err)
		}

		if ballot.Active {
			return tcr.ErrInvalidPhase(2, "Cannot claim reward until after ballot vote is applied").Result()
		}

		var decision bool
		if val == nil {
			decision = false
		} else {
			decision = true
		}

		if vote.Choice != decision {
			refund := sdk.Coin{
				Denom:  "RegistryCoin",
				Amount: vote.Power,
			}
			accountKeeper.AddCoins(ctx, claimMsg.Owner, []sdk.Coin{refund})
			return sdk.Result{}
		}

		var pool, total int64
		pool = int64(float64(ballot.Bond) * (float64(1.0) - dispPct))
		if decision {
			total = ballot.Approve
		} else {
			total = ballot.Deny
		}

		reward := sdk.Coin{
			Denom:  "RegistryCoin",
			Amount: vote.Power + int64(float64(pool)*float64(vote.Power)/float64(total)),
		}
		_, _, accErr := accountKeeper.AddCoins(ctx, claimMsg.Owner, []sdk.Coin{reward})

		if accErr != nil {
			return accErr.Result()
		}

		return sdk.Result{}
	}
}
