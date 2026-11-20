package main

import (
	"encoding/json"
	"testing"

	"vault-chain/validatorset"
	"vault-chain/vault"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func revisionProvider(t *testing.T, signer string, label string) *vault.MsgActivateProvider {
	t.Helper()
	private := ed25519.GenPrivKey()
	message := &vault.MsgActivateProvider{Signer: signer, ProviderId: label, ConsensusPubkey: private.PubKey().Bytes()}
	signature, err := private.Sign(vault.ProviderTranscript(message))
	if err != nil {
		t.Fatal(err)
	}
	message.ConsensusSignature = signature
	return message
}

func TestR1OpenJoinPathAndCap(t *testing.T) {
	admin, second := ed25519.GenPrivKey(), ed25519.GenPrivKey()
	app, genesis := applicationFixture(t, admin, fixtureConsensusKeys())
	revisionInitialize(t, app, genesis)
	owner := sdk.AccAddress(admin.PubKey().Address()).String()
	other := sdk.AccAddress(second.PubKey().Address())
	first := revisionProvider(t, owner, "first")
	result := revisionBlock(t, app, 1, 1800000001, signedTransaction(t, app, admin, 0, first))
	if result[0].Code != 0 {
		t.Fatalf("open join disconnected: %+v", result[0])
	}
	number := FirstContactAccountNumber(other)
	result = revisionBlock(t, app, 2, 1800000002, signedTransactionNumber(t, app, second, number, 0, revisionLink(t, second, 1)))
	if result[0].Code != 0 {
		t.Fatal("second provider owner link failed")
	}
	candidate := revisionProvider(t, other.String(), "second")
	result = revisionBlock(t, app, 3, 1800000003, signedTransactionNumber(t, app, second, number, 1, candidate))
	if result[0].Code != vault.ErrJoinCooldown.ABCICode() {
		t.Fatalf("second add within 24 hours not refused by cooldown: %+v", result[0])
	}
	state, err := app.vaultKeeper.ExportToGenesis(revisionContext(app))
	if err != nil || len(state.Providers) != 5 || state.LastAdmissionTime != 1800000001 {
		t.Fatal("cooldown refusal did not roll back provider state")
	}
	result = revisionBlock(t, app, 4, 1800000001+vault.JoinSeconds, signedTransactionNumber(t, app, second, number, 2, candidate))
	if result[0].Code != 0 {
		t.Fatalf("24-hour boundary rejected: %+v", result[0])
	}
	validators, err := app.validatorSetKeeper.ExportToGenesis(revisionContext(app))
	if err != nil || len(validators.Validators) != 6 {
		t.Fatal("activation did not reach validator set")
	}
	for _, power := range []int64{2, 1} {
		limited, encoded := applicationFixture(t, admin, fixtureConsensusKeys())
		encoded = revisionFounderPower(t, limited, encoded, power)
		revisionInitialize(t, limited, encoded)
		result = revisionBlock(t, limited, 1, 1800000001, signedTransaction(t, limited, admin, 0, first))
		if result[0].Code != validatorset.ErrOutsiderPowerTooLarge.ABCICode() || result[0].Codespace != validatorset.ModuleName {
			t.Fatalf("founder-loss cap bypassed: %+v", result[0])
		}
	}
	t.Log("PASS open-join path and cap")
}

func revisionFounderPower(t *testing.T, app *TraderSwiftApp, encoded []byte, power int64) []byte {
	t.Helper()
	modules := map[string]json.RawMessage{}
	if err := json.Unmarshal(encoded, &modules); err != nil {
		t.Fatal(err)
	}
	state := &vault.State{}
	if err := json.Unmarshal(modules[vault.ModuleName], state); err != nil {
		t.Fatal(err)
	}
	validators := &validatorset.GenesisState{}
	app.appCodec.MustUnmarshalJSON(modules[validatorset.ModuleName], validators)
	for _, provider := range state.Providers {
		provider.Power = power
	}
	for _, validator := range validators.Validators {
		validator.Power = power
	}
	modules[vault.ModuleName], _ = json.Marshal(state)
	modules[validatorset.ModuleName] = app.appCodec.MustMarshalJSON(validators)
	result, err := json.Marshal(modules)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
