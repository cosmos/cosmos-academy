package auth

import (
	"bytes"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"reflect"
)

func NewAnteHandler(accountMapper auth.AccountMapper) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (_ sdk.Context, _ sdk.Result, abort bool) {
		stdTx, ok := tx.(auth.StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		sigs := stdTx.GetSignatures()
		if len(sigs) != 1 {
			return ctx,
				sdk.ErrUnauthorized("no signers").Result(),
				true
		}

		sig := sigs[0]
		msg := tx.GetMsg()

		signerAddr := msg.GetSigners()[0]

		acc := accountMapper.GetAccount(ctx, signerAddr)

		if acc == nil {
			return ctx, sdk.ErrUnknownAddress(signerAddr.String()).Result(), true
		}
	
		// Check and increment sequence number.
		seq := acc.GetSequence()
		if seq != sig.Sequence {
			return ctx, sdk.ErrInvalidSequence(
				fmt.Sprintf("Invalid sequence. Got %d, expected %d", sig.Sequence, seq)).Result(), true
		}
		acc.SetSequence(seq + 1)
	
		// If pubkey is not known for account,
		// set it from the StdSignature.
		pubKey := acc.GetPubKey()
		if pubKey == nil {
			pubKey = sig.PubKey
			if pubKey == nil {
				return ctx, sdk.ErrInvalidPubKey("PubKey not found").Result(), true
			}
			if !bytes.Equal(pubKey.Address(), signerAddr) {
				return ctx, sdk.ErrInvalidPubKey(
					fmt.Sprintf("PubKey does not match Signer address %v", signerAddr)).Result(), true
			}
			err := acc.SetPubKey(pubKey)
			if err != nil {
				return ctx, sdk.ErrInternal("setting PubKey on signer's account").Result(), true
			}
		}

		if !reflect.DeepEqual(sigs[0].PubKey.Address().Bytes(), signerAddr.Bytes()) {
			return ctx, sdk.ErrInternal("Wrong signer address").Result(), true
		}

		if !sigs[0].PubKey.VerifyBytes(msg.GetSignBytes(), sigs[0].Signature) {
			return ctx, sdk.ErrInternal("Invalid Signature").Result(), true
		}

		if !pubKey.VerifyBytes(msg.GetSignBytes(), sig.Signature) {
			return ctx, sdk.ErrUnauthorized("signature verification failed").Result(), true
		}

		return ctx, sdk.Result{}, false
	}
}