package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"vault-chain/chainprefix"
	"vault-chain/validatorset"
	"vault-chain/vault"

	"cosmossdk.io/log/v2"
	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type localValidatorModule struct {
	validatorset.AppModule
	keeper *validatorset.Keeper
}

func (module localValidatorModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var state validatorset.GenesisState
	cdc.MustUnmarshalJSON(data, &state)
	updates, err := module.keeper.InitializeFromGenesis(ctx, &state)
	if err != nil {
		panic(err)
	}
	return updates[:1]
}

func applicationFixture(t *testing.T, admin cryptotypes.PrivKey, consensusKeys [][]byte) (*TraderSwiftApp, []byte) {
	t.Helper()
	chainprefix.Apply()
	home := *localChainDirectory
	if home == "" {
		home = t.TempDir()
	}
	app := newTraderSwiftApp(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(home), baseapp.SetChainID(chainprefix.ChainID))
	owner := sdk.AccAddress(admin.PubKey().Address()).String()
	localOnly := len(consensusKeys) == 1
	if localOnly {
		consensusKeys = append(consensusKeys, fixtureConsensusKeys()[:3]...)
	}
	vaultGenesis := vault.DefaultGenesisState()
	validatorGenesis := &validatorset.GenesisState{Admin: owner}
	for index, public := range consensusKeys {
		digest := sha256.Sum256(public)
		vaultGenesis.Providers = append(vaultGenesis.Providers, &vault.Provider{ProviderID: fmt.Sprintf("fixture-%d", index), Owner: sdk.AccAddress(digest[:20]).String(), ConsensusPubkey: public, Power: 100, Founding: true})
		validatorGenesis.Validators = append(validatorGenesis.Validators, &validatorset.Validator{ConsensusPublicKey: base64.StdEncoding.EncodeToString(public), Power: 100, Moniker: fmt.Sprintf("fixture-%d", index), Founding: true})
	}
	if localOnly {
		app.moduleManager.Modules[validatorset.ModuleName] = localValidatorModule{AppModule: validatorset.NewAppModule(app.validatorSetKeeper), keeper: app.validatorSetKeeper}
	}
	genesis := app.basicModuleManager.DefaultGenesis(app.appCodec)
	genesis[vault.ModuleName], _ = json.Marshal(vaultGenesis)
	genesis[validatorset.ModuleName] = app.appCodec.MustMarshalJSON(validatorGenesis)
	authGenesis := authtypes.DefaultGenesisState()
	account, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccountWithAddress(sdk.AccAddress(admin.PubKey().Address())))
	if err != nil {
		t.Fatal(err)
	}
	authGenesis.Accounts = []*codectypes.Any{account}
	genesis[authtypes.ModuleName] = app.appCodec.MustMarshalJSON(authGenesis)
	encoded, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	return app, encoded
}

func signedTransaction(t *testing.T, app *TraderSwiftApp, private cryptotypes.PrivKey, sequence uint64, messages ...sdk.Msg) []byte {
	t.Helper()
	return signedTransactionNumber(t, app, private, 0, sequence, messages...)
}

func signedTransactionNumber(t *testing.T, app *TraderSwiftApp, private cryptotypes.PrivKey, number, sequence uint64, messages ...sdk.Msg) []byte {
	t.Helper()
	builder := app.txConfig.NewTxBuilder()
	if err := builder.SetMsgs(messages...); err != nil {
		t.Fatal(err)
	}
	builder.SetGasLimit(10000000)
	placeholder := signing.SignatureV2{PubKey: private.PubKey(), Data: &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT}, Sequence: sequence}
	if err := builder.SetSignatures(placeholder); err != nil {
		t.Fatal(err)
	}
	data := authsigning.SignerData{Address: sdk.AccAddress(private.PubKey().Address()).String(), ChainID: chainprefix.ChainID, AccountNumber: number, Sequence: sequence, PubKey: private.PubKey()}
	signature, err := clienttx.SignWithPrivKey(context.Background(), signing.SignMode_SIGN_MODE_DIRECT, data, builder, private, app.txConfig, sequence)
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.SetSignatures(signature); err != nil {
		t.Fatal(err)
	}
	encoded, err := app.txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func fixtureConsensusKeys() [][]byte {
	keys := [][]byte{}
	for index := 0; index < 4; index++ {
		digest := sha256.Sum256([]byte(fmt.Sprintf("fixture-consensus-%d", index)))
		keys = append(keys, ed25519.GenPrivKeyFromSecret(digest[:]).PubKey().Bytes())
	}
	return keys
}

func initialRequest(encoded []byte) *abci.RequestInitChain {
	consensus := cmttypes.DefaultConsensusParams().ToProto()
	return &abci.RequestInitChain{Time: time.Unix(1800000000, 0), ChainId: chainprefix.ChainID, InitialHeight: 1, ConsensusParams: &consensus, AppStateBytes: encoded}
}
