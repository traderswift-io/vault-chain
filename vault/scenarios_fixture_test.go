package vault

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"vault-chain/chainprefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

type scenarioAdmin struct{ address string }

func (admin *scenarioAdmin) Admin(sdk.Context) (string, error) {
	return admin.address, nil
}

func scenarioAddress(label string) string {
	chainprefix.Apply()
	digest := sha256.Sum256([]byte(label))
	return sdk.AccAddress(digest[:20]).String()
}

func scenarioKeeper(t *testing.T) (*Keeper, sdk.Context, *scenarioAdmin) {
	t.Helper()
	chainprefix.Apply()
	key := storetypes.NewKVStoreKey(StoreKey)
	ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient"))
	ctx = ctx.WithChainID(chainprefix.ChainID).WithBlockHeight(1).WithBlockTime(time.Unix(1800000000, 0)).WithEventManager(sdk.NewEventManager())
	admin := &scenarioAdmin{address: scenarioAddress("admin")}
	keeper := NewKeeper(runtime.NewKVStoreService(key), admin)
	if err := keeper.InitializeFromGenesis(ctx, DefaultGenesisState()); err != nil {
		t.Fatal(err)
	}
	return keeper, ctx, admin
}

func scenarioLink(t *testing.T, owner, label string, nonces ...uint64) *MsgLinkAccount {
	t.Helper()
	seed := sha256.Sum256([]byte(label))
	private := ed25519.NewKeyFromSeed(seed[:])
	public := private.Public().(ed25519.PublicKey)
	digest := sha256.Sum256(public)
	hub, err := bech32.ConvertAndEncode(chainprefix.HubAddressPrefix, digest[:20])
	if err != nil {
		t.Fatal(err)
	}
	nonce := uint64(1)
	if len(nonces) > 0 {
		nonce = nonces[0]
	}
	return &MsgLinkAccount{Nonce: nonce, Signer: owner, Owner: owner, HubAccount: hub, HubPublicKey: public, HubSignature: ed25519.Sign(private, LinkTranscript(owner, hub, nonce))}
}

func scenarioFrontend(owner string) *MsgRegisterFrontend {
	seed := sha256.Sum256([]byte("frontend-service"))
	private := ed25519.NewKeyFromSeed(seed[:])
	message := &MsgRegisterFrontend{Signer: owner, ProviderId: "skin", ServicePubkey: private.Public().(ed25519.PublicKey), Endpoint: "https://frontend.invalid"}
	message.ServiceSignature = ed25519.Sign(private, FrontendTranscript(message))
	return message
}

func scenarioGrant(owner string, ctx sdk.Context) (*MsgGrant, *QueryMayReadRequest) {
	grantee := scenarioAddress("private-reader")
	salt := bytes.Repeat([]byte{73}, 32)
	address, _ := addressBytes(grantee, chainprefix.AccountAddressPrefix)
	query := &QueryMayReadRequest{Owner: owner, Grantee: grantee, Salt: salt, Scope: chainprefix.NoteScope}
	message := &MsgGrant{Signer: owner, Owner: owner, GranteeCommitment: GrantCommitment(salt, address), Scope: query.Scope, Expiry: blockSeconds(ctx) + 100}
	return message, query
}

func requireError(t *testing.T, actual, expected error) {
	t.Helper()
	if actual != expected {
		t.Fatalf("wanted %v, received %v", expected, actual)
	}
}

func stateBytes(t *testing.T, keeper *Keeper, ctx sdk.Context) []byte {
	t.Helper()
	state, err := keeper.ExportToGenesis(ctx)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
