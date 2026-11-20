package main

import (
	"context"
	"testing"
	"time"

	"vault-chain/chainprefix"
	"vault-chain/vault"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSignedRefusalsAndReplay(t *testing.T) {
	chainprefix.Apply()
	admin := ed25519.GenPrivKey()
	stranger := ed25519.GenPrivKey()
	app, genesis := applicationFixture(t, admin, fixtureConsensusKeys())
	wrapper := server.NewCometABCIWrapper(app)
	if _, err := wrapper.InitChain(context.Background(), initialRequest(genesis)); err != nil {
		t.Fatal(err)
	}
	owner := sdk.AccAddress(admin.PubKey().Address()).String()
	message := &vault.MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*vault.Location{{Name: "phone", Pointer: "opaque"}}}
	forged := signedTransaction(t, app, stranger, 0, message)
	accepted := signedTransaction(t, app, admin, 0, message)
	for index, transaction := range [][]byte{forged, accepted, accepted} {
		height := int64(index + 1)
		result, err := wrapper.FinalizeBlock(context.Background(), &abci.RequestFinalizeBlock{Height: height, Time: time.Unix(1800000000+height, 0), Hash: make([]byte, 32), Txs: [][]byte{transaction}})
		if err != nil || len(result.TxResults) != 1 {
			t.Fatalf("block failure: %v", err)
		}
		if (result.TxResults[0].Code == 0) != (index == 1) {
			t.Fatalf("signature/replay result %d: %+v", index, result.TxResults[0])
		}
		if _, err := wrapper.Commit(context.Background(), &abci.RequestCommit{}); err != nil {
			t.Fatal(err)
		}
	}
}
