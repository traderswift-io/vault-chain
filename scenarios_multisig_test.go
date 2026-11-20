package main

import (
	"context"
	"encoding/json"
	"testing"

	"vault-chain/chainprefix"
	"vault-chain/validatorset"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	multisigkey "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestR1MultisigAdmin(t *testing.T) {
	first, second, pending := ed25519.GenPrivKey(), ed25519.GenPrivKey(), ed25519.GenPrivKey()
	public := multisigkey.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{first.PubKey(), second.PubKey()})
	address := sdk.AccAddress(public.Address())
	app, encoded := applicationFixture(t, pending, fixtureConsensusKeys())
	modules := map[string]json.RawMessage{}
	if err := json.Unmarshal(encoded, &modules); err != nil {
		t.Fatal(err)
	}
	validators := &validatorset.GenesisState{}
	app.appCodec.MustUnmarshalJSON(modules[validatorset.ModuleName], validators)
	validators.Admin = address.String()
	modules[validatorset.ModuleName] = app.appCodec.MustMarshalJSON(validators)
	auth := authtypes.DefaultGenesisState()
	account, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccountWithAddress(address))
	if err != nil {
		t.Fatal(err)
	}
	auth.Accounts = []*codectypes.Any{account}
	modules[authtypes.ModuleName] = app.appCodec.MustMarshalJSON(auth)
	encoded, err = json.Marshal(modules)
	if err != nil {
		t.Fatal(err)
	}
	revisionInitialize(t, app, encoded)
	builder := app.txConfig.NewTxBuilder()
	pendingAddress := sdk.AccAddress(pending.PubKey().Address())
	if err := builder.SetMsgs(&validatorset.MsgSetAdmin{Admin: address.String(), NewAdmin: pendingAddress.String()}); err != nil {
		t.Fatal(err)
	}
	builder.SetGasLimit(10000000)
	data := multisig.NewMultisig(2)
	for _, private := range []cryptotypes.PrivKey{first, second} {
		placeholder := signing.SignatureV2{PubKey: private.PubKey(), Data: &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT}}
		if err := multisig.AddSignatureV2(data, placeholder, public.GetPubKeys()); err != nil {
			t.Fatal(err)
		}
	}
	if err := builder.SetSignatures(signing.SignatureV2{PubKey: public, Data: data}); err != nil {
		t.Fatal(err)
	}
	signer := authsigning.SignerData{Address: address.String(), ChainID: chainprefix.ChainID, AccountNumber: 0, PubKey: public}
	for _, private := range []cryptotypes.PrivKey{first, second} {
		signature, err := clienttx.SignWithPrivKey(context.Background(), signing.SignMode_SIGN_MODE_DIRECT, signer, builder, private, app.txConfig, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := multisig.AddSignatureV2(data, signature, public.GetPubKeys()); err != nil {
			t.Fatal(err)
		}
	}
	if err := builder.SetSignatures(signing.SignatureV2{PubKey: public, Data: data}); err != nil {
		t.Fatal(err)
	}
	transaction, err := app.txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateTransaction(transaction); err != nil {
		t.Fatal(err)
	}
	forged := append([]byte{}, transaction...)
	forged[len(forged)-1] ^= 1
	result := revisionBlock(t, app, 1, 1800000001, forged)
	if result[0].Code == 0 {
		t.Fatal("forged multisig signature accepted")
	}
	result = revisionBlock(t, app, 2, 1800000002, transaction)
	if result[0].Code != 0 {
		t.Fatalf("multisig proposal failed: %+v", result[0])
	}
	proposed, err := app.validatorSetKeeper.PendingAdminAddress.Get(revisionContext(app))
	if err != nil || proposed != pendingAddress.String() {
		t.Fatal("multisig proposal not recorded")
	}
	t.Log("PASS throwaway multisig admin proposal")
}
