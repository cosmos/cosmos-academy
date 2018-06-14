package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"
	"github.com/AdityaSripal/token_curated_registry/db"
	"github.com/AdityaSripal/token_curated_registry/types"
	"github.com/AdityaSripal/token_curated_registry/utils"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"testing"
)

func setup() (sdk.Context, auth.AccountMapper) {
	ms, _, _, _, _, accountKey := db.SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	cdc := db.MakeCodec()

	mapper := auth.NewAccountMapper(cdc, accountKey, &auth.BaseAccount{})

	return ctx, mapper

}

func TestBadTx(t *testing.T) {
	ctx, mapper := setup()

	ante := NewAnteHandler(mapper)

	msg := types.GenerateCandidacyMsg()

	privKey := utils.GeneratePrivKey()
	mapper.NewAccountWithAddress(ctx, privKey.PubKey().Address())

	sig := privKey.Sign(msg.GetSignBytes())

	tx := auth.StdTx{
		Msg: msg,
		Signatures: []auth.StdSignature{auth.StdSignature{
			privKey.PubKey(),
			sig,
			0,
		}},
	}

	_, _, abort  := ante(ctx, tx)

	assert.Equal(t, true, abort, "Bad Tx allowed to pass")
}

func TestGoodTx(t *testing.T) {
	ctx, mapper := setup()

	ante := NewAnteHandler(mapper)

	privKey := utils.GeneratePrivKey()

	acc := mapper.NewAccountWithAddress(ctx, privKey.PubKey().Address())
	mapper.SetAccount(ctx, acc)

	msg := types.GenerateCandidacyMsg()

	msg.Owner = privKey.PubKey().Address()

	sig := privKey.Sign(msg.GetSignBytes())

	tx := auth.StdTx{
		Msg: msg,
		Signatures: []auth.StdSignature{auth.StdSignature{
			privKey.PubKey(),
			sig,
			0,
		}},
	}

	_, res, abort  := ante(ctx, tx)

	assert.Equal(t, sdk.Result{}, res, "Good tx failed")

	assert.Equal(t, false, abort, "Good tx failed")
}

