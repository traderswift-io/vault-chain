package main

import (
	"encoding/json"
	"testing"

	"vault-chain/chainprefix"
	"vault-chain/validatorset"
	"vault-chain/vault"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func TestR1GenesisProviderCheck(t *testing.T) {
	app, encoded := applicationFixture(t, ed25519.GenPrivKey(), fixtureConsensusKeys())
	modules := map[string]json.RawMessage{}
	if err := json.Unmarshal(encoded, &modules); err != nil {
		t.Fatal(err)
	}
	if err := validateModuleGenesis(modules); err != nil {
		t.Fatal(err)
	}
	original := append([]byte{}, modules[vault.ModuleName]...)
	for _, mutation := range []func(*vault.State){
		func(state *vault.State) { state.Providers = state.Providers[1:] },
		func(state *vault.State) { state.Providers[0].Power++ },
		func(state *vault.State) { state.Providers[0].Founding = false },
		func(state *vault.State) { state.Providers[0].ProviderID = "mismatched" },
		func(state *vault.State) { state.Providers[0].ConsensusPubkey[0] ^= 1 },
	} {
		state := &vault.State{}
		if err := json.Unmarshal(original, state); err != nil {
			t.Fatal(err)
		}
		mutation(state)
		modules[vault.ModuleName], _ = json.Marshal(state)
		if validateModuleGenesis(modules) == nil {
			t.Fatal("unmatched genesis validator/provider accepted")
		}
	}
	modules[vault.ModuleName] = original
	validatorState := &validatorset.GenesisState{}
	app.appCodec.MustUnmarshalJSON(modules[validatorset.ModuleName], validatorState)
	validatorState.Validators[1] = validatorState.Validators[0]
	modules[validatorset.ModuleName] = app.appCodec.MustMarshalJSON(validatorState)
	if validateModuleGenesis(modules) == nil {
		t.Fatal("duplicate validator/provider mapping accepted")
	}
	t.Log("PASS genesis validator provider check")
}

func TestR1GenesisFlagsAtAnyHeight(t *testing.T) {
	for _, height := range []int64{1, 2, 1000000001} {
		app, encoded := applicationFixture(t, ed25519.GenPrivKey(), fixtureConsensusKeys())
		modules := map[string]json.RawMessage{}
		if err := json.Unmarshal(encoded, &modules); err != nil {
			t.Fatal(err)
		}
		state := &vault.State{}
		if err := json.Unmarshal(modules[vault.ModuleName], state); err != nil {
			t.Fatal(err)
		}
		state.Params.BaseEnabled = true
		modules[vault.ModuleName], _ = json.Marshal(state)
		if validateModuleGenesis(modules) != vault.ErrParams {
			t.Fatal("launch storage flags enabled")
		}
		genesis, err := json.Marshal(modules)
		if err != nil {
			t.Fatal(err)
		}
		request := initialRequest(genesis)
		request.InitialHeight = height
		if _, err := app.initChainer(sdk.Context{}.WithBlockHeight(height), request); err != vault.ErrParams {
			t.Fatalf("initial height %d bypassed flag check: %v", height, err)
		}
	}
}

func TestR1GenesisAdmin(t *testing.T) {
	chainprefix.Apply()
	state := validatorset.DefaultGenesisState()
	if state.Admin != chainprefix.GenesisAdmin {
		t.Fatal("genesis default admin is not the required multisig")
	}
	app, genesis := applicationFixture(t, ed25519.GenPrivKey(), fixtureConsensusKeys())
	modules := map[string]json.RawMessage{}
	if err := json.Unmarshal(genesis, &modules); err != nil {
		t.Fatal(err)
	}
	app.appCodec.MustUnmarshalJSON(modules[validatorset.ModuleName], state)
	state.Admin = chainprefix.GenesisAdmin
	modules[validatorset.ModuleName] = app.appCodec.MustMarshalJSON(state)
	genesis, _ = json.Marshal(modules)
	revisionInitialize(t, app, genesis)
	revisionBlock(t, app, 1, 1800000001)
	stored, err := app.validatorSetKeeper.Admin(revisionContext(app))
	if err != nil {
		t.Fatal(err)
	}
	_, actual, err := bech32.DecodeAndConvert(stored)
	if err != nil {
		t.Fatal(err)
	}
	_, expected, err := bech32.DecodeAndConvert(chainprefix.GenesisAdmin)
	if err != nil || string(actual) != string(expected) {
		t.Fatal("multisig address bytes changed")
	}
	t.Log("PASS genesis admin is the multisig")
}
