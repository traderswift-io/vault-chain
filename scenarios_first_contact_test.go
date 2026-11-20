package main

import (
	"encoding/json"
	"testing"

	"vault-chain/vault"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestR1FirstContact(t *testing.T) {
	admin, newcomer, impostor := ed25519.GenPrivKey(), ed25519.GenPrivKey(), ed25519.GenPrivKey()
	app, genesis := applicationFixture(t, admin, fixtureConsensusKeys())
	revisionInitialize(t, app, genesis)
	address := sdk.AccAddress(newcomer.PubKey().Address())
	number := FirstContactAccountNumber(address)
	link := revisionLink(t, newcomer, 1)
	choice := &vault.MsgSetStorageChoice{Signer: address.String(), Owner: address.String(), Locations: []*vault.Location{{Name: "phone", Pointer: "opaque"}}}
	badProof := *link
	badProof.HubSignature = make([]byte, 64)
	transactions := [][]byte{
		signedTransactionNumber(t, app, newcomer, number, 0, choice),
		signedTransactionNumber(t, app, newcomer, number, 0, link, choice),
		signedTransactionNumber(t, app, newcomer, number, 0, link, link),
		signedTransactionNumber(t, app, newcomer, number, 1, link),
		signedTransactionNumber(t, app, newcomer, number, 0, &badProof),
		signedTransactionNumber(t, app, impostor, number, 0, link),
		signedTransactionNumber(t, app, newcomer, 0, 0, link),
	}
	height := int64(0)
	for _, transaction := range transactions {
		height++
		result := revisionBlock(t, app, height, 1800000000+height, transaction)
		if result[0].Code == 0 || app.accountKeeper.HasAccount(revisionContext(app), address) {
			t.Fatal("invalid first contact created an account")
		}
	}
	height++
	result := revisionBlock(t, app, height, 1800000000+height, signedTransactionNumber(t, app, newcomer, number, 0, link))
	account := app.accountKeeper.GetAccount(revisionContext(app), address)
	if result[0].Code != 0 || account == nil || account.GetAccountNumber() != number || account.GetSequence() != 1 {
		t.Fatalf("first contact failed: %+v", result[0])
	}
	height++
	result = revisionBlock(t, app, height, 1800000000+height, signedTransactionNumber(t, app, newcomer, number, 1, choice))
	if result[0].Code != 0 {
		t.Fatalf("new owner could not choose: %+v", result[0])
	}
}

func TestR1FirstContactBlockCap(t *testing.T) {
	admin := ed25519.GenPrivKey()
	app, genesis := applicationFixture(t, admin, fixtureConsensusKeys())
	revisionInitialize(t, app, genesis)
	transactions := [][]byte{}
	for index := 0; index <= MaximumAccountsCreatedPerBlock; index++ {
		private := ed25519.GenPrivKey()
		address := sdk.AccAddress(private.PubKey().Address())
		transactions = append(transactions, signedTransactionNumber(t, app, private, FirstContactAccountNumber(address), 0, revisionLink(t, private, 1)))
	}
	results := revisionBlock(t, app, 1, 1800000001, transactions...)
	for index, result := range results {
		if (result.Code == 0) != (index < MaximumAccountsCreatedPerBlock) {
			t.Fatalf("first contact cap result %d: %+v", index, result)
		}
	}
	if result := revisionBlock(t, app, 2, 1800000002, transactions[len(transactions)-1]); result[0].Code != 0 {
		t.Fatalf("next block did not reset cap: %+v", result[0])
	}
}

func TestR1FirstContactNumberCollision(t *testing.T) {
	admin, collider, holder := ed25519.GenPrivKey(), ed25519.GenPrivKey(), ed25519.GenPrivKey()
	app, genesis := applicationFixture(t, admin, fixtureConsensusKeys())
	address := sdk.AccAddress(collider.PubKey().Address())
	number := FirstContactAccountNumber(address)
	modules := map[string]json.RawMessage{}
	if err := json.Unmarshal(genesis, &modules); err != nil {
		t.Fatal(err)
	}
	state := authtypes.DefaultGenesisState()
	app.appCodec.MustUnmarshalJSON(modules[authtypes.ModuleName], state)
	account, err := types.NewAnyWithValue(authtypes.NewBaseAccount(sdk.AccAddress(holder.PubKey().Address()), nil, number, 0))
	if err != nil {
		t.Fatal(err)
	}
	state.Accounts = append(state.Accounts, account)
	modules[authtypes.ModuleName] = app.appCodec.MustMarshalJSON(state)
	genesis, err = json.Marshal(modules)
	if err != nil {
		t.Fatal(err)
	}
	revisionInitialize(t, app, genesis)
	result := revisionBlock(t, app, 1, 1800000001, signedTransactionNumber(t, app, collider, number, 0, revisionLink(t, collider, 1)))
	if result[0].Code == 0 || app.accountKeeper.HasAccount(revisionContext(app), address) {
		t.Fatal("derived account number collision accepted")
	}
}
