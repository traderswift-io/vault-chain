package main

import (
	"bytes"
	"testing"

	"vault-chain/chainprefix"
	"vault-chain/validatorset"
	"vault-chain/vault"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBoundedTransactionRoundtrip(t *testing.T) {
	chainprefix.Apply()
	private := ed25519.GenPrivKey()
	app, _ := applicationFixture(t, private, fixtureConsensusKeys())
	owner := sdk.AccAddress(private.PubKey().Address()).String()
	messages := []sdk.Msg{
		&vault.MsgLinkAccount{Nonce: 1, Signer: owner, Owner: owner, HubAccount: owner, HubPublicKey: private.PubKey().Bytes(), HubSignature: bytes.Repeat([]byte{1}, 64)},
		&vault.MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*vault.Location{{Name: "phone", Pointer: "opaque"}}},
		&vault.MsgCommitRoot{Signer: owner, Owner: owner, Day: "2026-01-01", Root: bytes.Repeat([]byte{1}, 32)},
		&vault.MsgGrant{Signer: owner, Owner: owner, GranteeCommitment: bytes.Repeat([]byte{2}, 32), Scope: chainprefix.NoteScope, Expiry: 1800000100},
		&vault.MsgRevoke{Signer: owner, Owner: owner, GranteeCommitment: bytes.Repeat([]byte{2}, 32), Scope: chainprefix.NoteScope},
		&vault.MsgRegisterFrontend{Signer: owner, ProviderId: "frontend", ServicePubkey: private.PubKey().Bytes(), Endpoint: "https://frontend.invalid", ServiceSignature: bytes.Repeat([]byte{1}, 64)},
		&vault.MsgUpdateParams{Signer: owner, BaseEnabled: true, TraderswiftEnabled: true},
		&vault.MsgRemoveFrontend{Signer: owner, ProviderId: "frontend"},
		revisionProvider(t, owner, "candidate"),
		&validatorset.MsgSetAdmin{Admin: owner, NewAdmin: owner},
		&validatorset.MsgAcceptAdmin{NewAdmin: owner},
	}
	for _, message := range messages {
		encoded := signedTransaction(t, app, private, 0, message)
		decoded, err := boundedTxDecoder(app.txConfig.TxDecoder())(encoded)
		if err != nil || len(decoded.GetMsgs()) != 1 {
			t.Fatalf("decode %T: %v", message, err)
		}
	}
}
