package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"vault-chain/vault"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestOfflineApplicationAgreement(t *testing.T) {
	admin := ed25519.GenPrivKeyFromSecret([]byte("offline-application-fixture"))
	var expected []byte
	for node := 0; node < 4; node++ {
		app, genesis := applicationFixture(t, admin, fixtureConsensusKeys())
		wrapper := server.NewCometABCIWrapper(app)
		initial, err := wrapper.InitChain(context.Background(), initialRequest(genesis))
		if err != nil || len(initial.Validators) != 4 {
			t.Fatalf("init chain: %v", err)
		}
		owner := sdk.AccAddress(admin.PubKey().Address()).String()
		message := &vault.MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*vault.Location{{Name: "phone", Pointer: "opaque"}}}
		tx := signedTransaction(t, app, admin, 0, message)
		var hash []byte
		for height := int64(1); height <= 3; height++ {
			blockHash := sha256.Sum256([]byte(fmt.Sprintf("synthetic-block-%d", height)))
			request := &abci.RequestFinalizeBlock{Height: height, Time: time.Unix(1800000000+height, 0), Hash: blockHash[:]}
			if height == 1 {
				request.Txs = [][]byte{tx}
			}
			result, err := wrapper.FinalizeBlock(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if height == 1 && (len(result.TxResults) != 1 || result.TxResults[0].Code != 0) {
				t.Fatalf("signed transaction failed: %+v", result.TxResults)
			}
			hash = result.AppHash
			if _, err := wrapper.Commit(context.Background(), &abci.RequestCommit{}); err != nil {
				t.Fatal(err)
			}
		}
		if len(hash) == 0 {
			t.Fatal("empty app hash")
		}
		if node == 0 {
			expected = append([]byte{}, hash...)
		} else if !bytes.Equal(expected, hash) {
			t.Fatal("four application hashes differ")
		}
	}
}

func TestTransactionDecoderBounds(t *testing.T) {
	if validateTransaction(make([]byte, vault.MaxTransactionBytes+1)) == nil {
		t.Fatal("oversized tx accepted")
	}
	for index := 0; index < 256; index++ {
		data := sha256.Sum256([]byte(fmt.Sprint(index)))
		if validateTransaction(data[:]) == nil {
			t.Fatal("malformed tx accepted")
		}
	}
}
