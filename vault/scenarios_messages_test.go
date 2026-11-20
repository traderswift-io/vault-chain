package vault

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"vault-chain/chainprefix"
)

func TestMessageScenarios(t *testing.T) {
	keeper, ctx, admin := scenarioKeeper(t)
	owner, stranger := scenarioAddress("owner"), scenarioAddress("stranger")
	link := scenarioLink(t, owner, "hub-owner")
	choice := &MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*Location{{Name: "phone", Pointer: "opaque"}}}
	root := &MsgCommitRoot{Signer: owner, Owner: owner, Day: ctx.BlockTime().UTC().Format("2006-01-02"), Root: bytes.Repeat([]byte{1}, 32)}
	grant, query := scenarioGrant(owner, ctx)
	revoke := &MsgRevoke{Signer: owner, Owner: owner, GranteeCommitment: grant.GranteeCommitment, Scope: grant.Scope}
	frontend := scenarioFrontend(admin.address)
	for _, message := range []any{link, choice, root, grant, frontend} {
		requireError(t, keeper.Execute(ctx, message), nil)
	}
	response, err := NewQueryServer(keeper).Where(ctx, &QueryWhereRequest{Owner: owner})
	requireError(t, err, nil)
	var where WhereRecord
	requireError(t, json.Unmarshal(response.CanonicalRecord, &where), nil)
	if where.Link.HubAccount != link.HubAccount || where.Choice.Locations[0].Pointer != "opaque" {
		t.Fatal("where query lost records")
	}
	updated := scenarioLink(t, owner, "replacement-hub", 2)
	requireError(t, keeper.Execute(ctx, updated), nil)
	link.Signer, choice.Signer, root.Signer, grant.Signer, revoke.Signer = stranger, stranger, stranger, stranger, stranger
	frontend.Signer = stranger
	for _, message := range []any{link, choice, root, grant, revoke} {
		before := stateBytes(t, keeper, ctx)
		requireError(t, keeper.Execute(ctx, message), ErrWrongSigner)
		if !bytes.Equal(before, stateBytes(t, keeper, ctx)) {
			t.Fatal("refusal mutated state")
		}
	}
	requireError(t, keeper.Execute(ctx, frontend), ErrFrontendSignature)
	root.Signer = owner
	requireError(t, keeper.Execute(ctx, root), ErrRootExists)
	root.Day = "2026-02-30"
	requireError(t, keeper.Execute(ctx, root), ErrRoot)
	root.Day, root.Root = "2026-01-01", []byte{1}
	requireError(t, keeper.Execute(ctx, root), ErrRoot)
	choice.Signer = owner
	choice.Locations[0].Name = "unknown"
	requireError(t, keeper.Execute(ctx, choice), ErrUnknownLocation)
	choice.Locations[0].Name, choice.Locations[0].Pointer = "phone", strings.Repeat("x", 257)
	requireError(t, keeper.Execute(ctx, choice), ErrOversizePointer)
	choice.Locations = []*Location{{Name: "phone", Pointer: "a"}, {Name: "phone", Pointer: "b"}}
	requireError(t, keeper.Execute(ctx, choice), ErrChoice)
	grant.Signer, grant.Scope = owner, strings.Repeat("s", 65)
	requireError(t, keeper.Execute(ctx, grant), ErrGrant)
	grant.Scope, grant.Expiry = query.Scope, blockSeconds(ctx)
	requireError(t, keeper.Execute(ctx, grant), ErrGrant)
	revoke.Signer = owner
	requireError(t, keeper.Execute(ctx, revoke), nil)
	requireError(t, keeper.Execute(ctx, revoke), ErrRevoke)
	requireError(t, keeper.Execute(ctx, &MsgUpdateParams{Signer: admin.address}), nil)
	link.Signer, link.HubSignature = owner, []byte{1}
	requireError(t, keeper.Execute(ctx, link), ErrLink)
	frontend = scenarioFrontend(admin.address)
	requireError(t, keeper.Execute(ctx, frontend), ErrFrontend)
	frontend.Endpoint = "http://frontend.invalid"
	requireError(t, keeper.Execute(ctx, frontend), ErrFrontend)
	if _, err := NewQueryServer(keeper).Where(ctx, &QueryWhereRequest{Owner: "invalid"}); err != ErrQuery {
		t.Fatal("invalid query accepted")
	}
	t.Log("PASS scenarios")
}

func TestStorageGates(t *testing.T) {
	keeper, ctx, admin := scenarioKeeper(t)
	owner := scenarioAddress("owner")
	for _, location := range []string{"base", "traderswift"} {
		message := &MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*Location{{Name: location, Pointer: "opaque"}}}
		requireError(t, keeper.Execute(ctx, message), ErrLocationDisabled)
	}
	phone := &MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*Location{{Name: "phone", Pointer: "opaque"}}}
	requireError(t, keeper.Execute(ctx, phone), nil)
	params := &MsgUpdateParams{Signer: owner, BaseEnabled: true, TraderswiftEnabled: true}
	requireError(t, keeper.Execute(ctx, params), ErrNonAdminParams)
	params.Signer, params.TraderswiftEnabled = admin.address, false
	requireError(t, keeper.Execute(ctx, params), nil)
	phone.Locations[0].Name = "base"
	requireError(t, keeper.Execute(ctx, phone), nil)
	phone.Locations[0].Name = "traderswift"
	requireError(t, keeper.Execute(ctx, phone), ErrLocationDisabled)
	params.TraderswiftEnabled = true
	requireError(t, keeper.Execute(ctx, params), nil)
	requireError(t, keeper.Execute(ctx, phone), nil)
	params.BaseEnabled, params.TraderswiftEnabled = false, false
	requireError(t, keeper.Execute(ctx, params), nil)
	requireError(t, keeper.Execute(ctx, phone), ErrLocationDisabled)
	phone.Locations[0].Name = "base"
	requireError(t, keeper.Execute(ctx, phone), ErrLocationDisabled)
	admin.address = scenarioAddress("new-admin")
	requireError(t, keeper.Execute(ctx, params), ErrNonAdminParams)
	params.Signer = admin.address
	requireError(t, keeper.Execute(ctx, params), nil)
	t.Log("PASS storage gates")
}

func TestScopesAreShapeOnly(t *testing.T) {
	for _, scope := range []string{chainprefix.NoteScope, chainprefix.FileScope, "custom:v2_read-all"} {
		if !scopeName(scope) {
			t.Fatal("valid free scope refused")
		}
	}
	for _, scope := range []string{"", "with space", "path/scope", "é", strings.Repeat("a", 65)} {
		if scopeName(scope) {
			t.Fatal("invalid scope accepted")
		}
	}
}
