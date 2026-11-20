package main

import (
	"testing"

	"vault-chain/validatorset"
	"vault-chain/vault"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestR1AdminRotation(t *testing.T) {
	admin, pending, outsider := ed25519.GenPrivKey(), ed25519.GenPrivKey(), ed25519.GenPrivKey()
	app, genesis := applicationFixture(t, admin, fixtureConsensusKeys())
	revisionInitialize(t, app, genesis)
	oldAddress := sdk.AccAddress(admin.PubKey().Address()).String()
	newAddress := sdk.AccAddress(pending.PubKey().Address()).String()
	otherAddress := sdk.AccAddress(outsider.PubKey().Address()).String()
	height := int64(0)
	deliver := func(private cryptotypes.PrivKey, sequence uint64, message sdk.Msg, accepted bool) {
		t.Helper()
		number := uint64(0)
		if private != admin {
			number = FirstContactAccountNumber(sdk.AccAddress(private.PubKey().Address()))
		}
		height++
		result := revisionBlock(t, app, height, 1800000000+height, signedTransactionNumber(t, app, private, number, sequence, message))
		if (result[0].Code == 0) != accepted {
			t.Fatalf("rotation result %T: %+v", message, result[0])
		}
	}
	deliver(pending, 0, revisionLink(t, pending, 1), true)
	deliver(outsider, 0, revisionLink(t, outsider, 1), true)
	deliver(outsider, 1, &validatorset.MsgSetAdmin{Admin: otherAddress, NewAdmin: newAddress}, false)
	t.Log("PASS set admin by non-admin refused")
	deliver(admin, 0, &validatorset.MsgSetAdmin{Admin: oldAddress, NewAdmin: newAddress}, true)
	deliver(outsider, 2, &validatorset.MsgSetAdmin{Admin: otherAddress, NewAdmin: otherAddress}, false)
	proposed, err := app.validatorSetKeeper.PendingAdminAddress.Get(revisionContext(app))
	if err != nil || proposed != newAddress {
		t.Fatal("non-admin replaced pending admin")
	}
	t.Log("PASS non-admin cannot replace pending admin")
	deliver(outsider, 3, &validatorset.MsgAcceptAdmin{NewAdmin: otherAddress}, false)
	t.Log("PASS accept admin by non-pending refused")
	deliver(pending, 1, &validatorset.MsgAcceptAdmin{NewAdmin: newAddress}, true)
	deliver(admin, 1, &vault.MsgUpdateParams{Signer: oldAddress, BaseEnabled: true}, false)
	deliver(admin, 2, &validatorset.MsgSetAdmin{Admin: oldAddress, NewAdmin: otherAddress}, false)
	deliver(pending, 2, &vault.MsgUpdateParams{Signer: newAddress, BaseEnabled: true}, true)
	current, err := app.validatorSetKeeper.Admin(revisionContext(app))
	if err != nil || current != newAddress {
		t.Fatal("accepted admin did not take authority")
	}
	t.Log("PASS full rotation then old admin refused")
}
