package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vault-chain/chainprefix"
	"vault-chain/vault"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/rpc/client/local"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/gogoproto/proto"
)

func encodeEvents(t *testing.T, transaction []abci.Event, block cmttypes.EventDataNewBlock) []byte {
	t.Helper()
	encoded, err := json.Marshal(struct {
		Transaction []abci.Event               `json:"transaction"`
		Block       cmttypes.EventDataNewBlock `json:"block"`
	}{transaction, block})
	if err != nil {
		t.Fatal(err)
	}
	return append(encoded, '\n')
}

func runLocalMessages(t *testing.T, app *TraderSwiftApp, admin cryptotypes.PrivKey, client *local.Local, ctx context.Context, broadcast func(sdk.Msg) []byte, root string) {
	t.Helper()
	owner := sdk.AccAddress(admin.PubKey().Address()).String()
	hub, err := bech32.ConvertAndEncode(chainprefix.HubAddressPrefix, admin.PubKey().Address())
	if err != nil {
		t.Fatal(err)
	}
	signature, err := admin.Sign(vault.LinkTranscript(owner, hub, 1))
	if err != nil {
		t.Fatal(err)
	}
	events := broadcast(&vault.MsgLinkAccount{Nonce: 1, Signer: owner, Owner: owner, HubAccount: hub, HubPublicKey: admin.PubKey().Bytes(), HubSignature: signature})
	events = append(events, broadcast(&vault.MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*vault.Location{{Name: "phone", Pointer: "opaque-local"}}})...)
	events = append(events, broadcast(&vault.MsgCommitRoot{Signer: owner, Owner: owner, Day: time.Now().UTC().Format("2006-01-02"), Root: bytes.Repeat([]byte{4}, 32)})...)
	reader := ed25519.GenPrivKey().PubKey().Address()
	grantee := sdk.AccAddress(reader).String()
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}
	commitment := vault.GrantCommitment(salt, reader)
	events = append(events, broadcast(&vault.MsgGrant{Signer: owner, Owner: owner, GranteeCommitment: commitment, Scope: chainprefix.NoteScope, Expiry: uint64(time.Now().Add(time.Hour).Unix())})...)
	query := func(method string, request, response proto.Message) {
		t.Helper()
		data, err := proto.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		result, err := client.ABCIQuery(ctx, "/traderswift."+chainprefix.ModuleName+".v1.Query/"+method, data)
		if err != nil {
			t.Fatal(err)
		}
		if result.Response.Code != 0 {
			t.Fatalf("query refused: %s", result.Response.Log)
		}
		if err := proto.Unmarshal(result.Response.Value, response); err != nil {
			t.Fatal(err)
		}
	}
	where := &vault.QueryWhereResponse{}
	query("Where", &vault.QueryWhereRequest{Owner: owner}, where)
	var record vault.WhereRecord
	if err := json.Unmarshal(where.CanonicalRecord, &record); err != nil || record.Link == nil || record.Link.HubAccount != hub || record.Choice == nil || record.Choice.Locations[0].Pointer != "opaque-local" {
		t.Fatalf("where mismatch: %v", err)
	}
	request := &vault.QueryMayReadRequest{Owner: owner, Grantee: grantee, Salt: salt, Scope: chainprefix.NoteScope}
	allowed := &vault.QueryMayReadResponse{}
	query("MayRead", request, allowed)
	if !allowed.Allowed {
		t.Fatal("grant did not authorize reader")
	}
	request.Salt = bytes.Repeat([]byte{0}, 32)
	query("MayRead", request, allowed)
	if allowed.Allowed {
		t.Fatal("wrong salt authorized reader")
	}
	request.Salt = salt
	events = append(events, broadcast(&vault.MsgRevoke{Signer: owner, Owner: owner, GranteeCommitment: commitment, Scope: chainprefix.NoteScope})...)
	query("MayRead", request, allowed)
	if allowed.Allowed {
		t.Fatal("revoked grant authorized reader")
	}
	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range [][]byte{reader, []byte(grantee), []byte(base64.StdEncoding.EncodeToString(reader)), []byte(hex.EncodeToString(reader)), salt, []byte(base64.StdEncoding.EncodeToString(salt)), []byte(hex.EncodeToString(salt))} {
		if bytes.Contains(exported.AppState, private) || bytes.Contains(events, private) {
			t.Fatal("grant privacy violated in export or consensus events")
		}
	}
	if err := os.WriteFile(filepath.Join(root, "export.json"), exported.AppState, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "events.ndjson"), events, 0600); err != nil {
		t.Fatal(err)
	}

	t.Log("PASS local export and event privacy")
	t.Log("PASS round trip")
}
