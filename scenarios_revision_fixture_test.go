package main

import (
	"context"
	"testing"
	"time"

	"vault-chain/chainprefix"
	"vault-chain/vault"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func revisionLink(t *testing.T, private cryptotypes.PrivKey, nonce uint64) *vault.MsgLinkAccount {
	t.Helper()
	owner := sdk.AccAddress(private.PubKey().Address()).String()
	hub, err := bech32.ConvertAndEncode(chainprefix.HubAddressPrefix, private.PubKey().Address())
	if err != nil {
		t.Fatal(err)
	}
	signature, err := private.Sign(vault.LinkTranscript(owner, hub, nonce))
	if err != nil {
		t.Fatal(err)
	}
	return &vault.MsgLinkAccount{Signer: owner, Owner: owner, HubAccount: hub, HubPublicKey: private.PubKey().Bytes(), HubSignature: signature, Nonce: nonce}
}

func revisionBlock(t *testing.T, app *TraderSwiftApp, height, seconds int64, transactions ...[]byte) []*abci.ExecTxResult {
	t.Helper()
	wrapper := server.NewCometABCIWrapper(app)
	result, err := wrapper.FinalizeBlock(context.Background(), &abci.RequestFinalizeBlock{Height: height, Time: time.Unix(seconds, 0), Hash: make([]byte, 32), Txs: transactions})
	if err != nil || len(result.TxResults) != len(transactions) {
		t.Fatalf("block failed: %v", err)
	}
	if _, err := wrapper.Commit(context.Background(), &abci.RequestCommit{}); err != nil {
		t.Fatal(err)
	}
	return result.TxResults
}

func revisionContext(app *TraderSwiftApp) sdk.Context {
	return app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
}

func revisionInitialize(t *testing.T, app *TraderSwiftApp, genesis []byte) {
	t.Helper()
	if _, err := server.NewCometABCIWrapper(app).InitChain(context.Background(), initialRequest(genesis)); err != nil {
		t.Fatal(err)
	}
}
