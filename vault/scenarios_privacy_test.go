package vault

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"vault-chain/chainprefix"
)

func TestGrantPrivacy(t *testing.T) {
	keeper, ctx, _ := scenarioKeeper(t)
	owner := scenarioAddress("owner")
	message, query := scenarioGrant(owner, ctx)
	requireError(t, keeper.Execute(ctx, message), nil)
	server := NewQueryServer(keeper)
	response, err := server.MayRead(ctx, query)
	requireError(t, err, nil)
	if !response.Allowed {
		t.Fatal("active grant refused")
	}
	wrongSalt := *query
	wrongSalt.Salt = bytes.Repeat([]byte{74}, 32)
	response, err = server.MayRead(ctx, &wrongSalt)
	requireError(t, err, nil)
	if response.Allowed {
		t.Fatal("wrong salt authorized")
	}
	wrongScope := *query
	wrongScope.Scope = chainprefix.FileScope
	response, err = server.MayRead(ctx, &wrongScope)
	requireError(t, err, nil)
	if response.Allowed {
		t.Fatal("wrong scope authorized")
	}
	response, err = server.MayRead(ctx.WithBlockTime(time.Unix(int64(message.Expiry), 0)), query)
	requireError(t, err, nil)
	if response.Allowed {
		t.Fatal("grant valid at exclusive expiry")
	}
	response, err = server.MayRead(ctx.WithBlockTime(time.Unix(int64(message.Expiry)+1, 0)), query)
	requireError(t, err, nil)
	if response.Allowed {
		t.Fatal("expired grant authorized")
	}
	exported := stateBytes(t, keeper, ctx)
	events, err := json.Marshal(ctx.EventManager().Events())
	requireError(t, err, nil)
	address, err := addressBytes(query.Grantee, chainprefix.AccountAddressPrefix)
	requireError(t, err, nil)
	for _, secret := range [][]byte{address, []byte(query.Grantee), []byte(hex.EncodeToString(address)), []byte(base64.StdEncoding.EncodeToString(address)), query.Salt, []byte(base64.StdEncoding.EncodeToString(query.Salt))} {
		if bytes.Contains(exported, secret) || bytes.Contains(events, secret) {
			t.Fatal("private witness in exported state or events")
		}
	}
	iterator, err := keeper.storeService.OpenKVStore(ctx).Iterator(nil, nil)
	requireError(t, err, nil)
	for ; iterator.Valid(); iterator.Next() {
		for _, secret := range [][]byte{address, []byte(query.Grantee), query.Salt} {
			if bytes.Contains(iterator.Key(), secret) || bytes.Contains(iterator.Value(), secret) {
				t.Fatal("private witness in stored rows")
			}
		}
	}
	requireError(t, iterator.Close(), nil)
	restored, restoredContext, _ := scenarioKeeper(t)
	var state State
	requireError(t, DecodeCanonical(exported, &state, MaxGenesisBytes), nil)
	requireError(t, restored.InitializeFromGenesis(restoredContext, &state), nil)
	response, err = NewQueryServer(restored).MayRead(restoredContext, query)
	requireError(t, err, nil)
	if !response.Allowed {
		t.Fatal("export/import lost grant")
	}
	revoke := &MsgRevoke{Signer: owner, Owner: owner, GranteeCommitment: message.GranteeCommitment, Scope: message.Scope}
	requireError(t, keeper.Execute(ctx, revoke), nil)
	response, err = server.MayRead(ctx, query)
	requireError(t, err, nil)
	if response.Allowed {
		t.Fatal("revoked grant authorized")
	}
	t.Log("PASS grant privacy")
}
