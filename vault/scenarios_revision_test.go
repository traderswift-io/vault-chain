package vault

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"vault-chain/chainprefix"

	errorsmod "cosmossdk.io/errors"
)

func TestR1RootDayAndOwnerBounds(t *testing.T) {
	keeper, ctx, _ := scenarioKeeper(t)
	owner, other := scenarioAddress("owner"), scenarioAddress("other")
	root := &MsgCommitRoot{Signer: owner, Owner: owner, Root: bytes.Repeat([]byte{1}, 32)}
	for _, offset := range []int{-1, 1} {
		root.Day = ctx.BlockTime().UTC().AddDate(0, 0, offset).Format("2006-01-02")
		requireError(t, keeper.Execute(ctx, root), ErrRoot)
	}
	root.Day = ctx.BlockTime().UTC().Format("2006-01-02")
	requireError(t, keeper.Execute(ctx, root), nil)
	requireError(t, keeper.Execute(ctx, root), ErrRootExists)
	root.Day = ctx.BlockTime().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	requireError(t, keeper.Execute(ctx.WithBlockTime(ctx.BlockTime().Add(24*time.Hour)), root), nil)
	t.Log("PASS root day is block day")
	grant, _ := scenarioGrant(owner, ctx)
	for index := 0; index < MaxOwnerGrants; index++ {
		grant.Scope = fmt.Sprintf("scope.%d", index)
		requireError(t, keeper.Execute(ctx, grant), nil)
	}
	grant.Scope = "overflow"
	requireError(t, keeper.Execute(ctx, grant), ErrLimitExceeded)
	grant.Scope = "scope.0"
	requireError(t, keeper.Execute(ctx, grant), nil)
	grant.Owner, grant.Signer = other, other
	requireError(t, keeper.Execute(ctx, grant), nil)
	root.Day, root.Owner, root.Signer = ctx.BlockTime().UTC().Format("2006-01-02"), other, other
	requireError(t, keeper.Execute(ctx, root), nil)
	for _, signer := range []string{owner, other} {
		requireError(t, keeper.Execute(ctx, &MsgSetStorageChoice{Signer: signer, Owner: signer, Locations: []*Location{{Name: "phone", Pointer: "opaque"}}}), nil)
	}
	state, err := keeper.ExportToGenesis(ctx)
	requireError(t, err, nil)
	restored, restoredContext, _ := scenarioKeeper(t)
	requireError(t, restored.InitializeFromGenesis(restoredContext, state), nil)
	grant.Owner, grant.Signer, grant.Scope = owner, owner, "overflow"
	requireError(t, restored.Execute(restoredContext, grant), ErrLimitExceeded)
	t.Log("PASS per-owner bound does not block another owner")
}

func TestR1LinkReplay(t *testing.T) {
	keeper, ctx, _ := scenarioKeeper(t)
	owner := scenarioAddress("owner")
	original := scenarioLink(t, owner, "first")
	requireError(t, keeper.Execute(ctx, original), nil)
	requireError(t, keeper.Execute(ctx, scenarioLink(t, owner, "second")), ErrLinkReplay)
	requireError(t, keeper.Execute(ctx, scenarioLink(t, owner, "second", 2)), nil)
	requireError(t, keeper.Execute(ctx, original), ErrLinkReplay)
	forged := *original
	forged.Nonce = 3
	requireError(t, keeper.Execute(ctx, &forged), ErrLink)
	state, err := keeper.ExportToGenesis(ctx)
	requireError(t, err, nil)
	restored, restoredContext, _ := scenarioKeeper(t)
	requireError(t, restored.InitializeFromGenesis(restoredContext, state), nil)
	requireError(t, restored.Execute(restoredContext, original), ErrLinkReplay)
	requireError(t, restored.Execute(restoredContext, scenarioLink(t, owner, "third", 3)), nil)
	t.Log("PASS link replay refused")
}

func revisionFrontend(admin, owner, label string) *MsgRegisterFrontend {
	seed := sha256.Sum256([]byte(label))
	private := ed25519.NewKeyFromSeed(seed[:])
	message := &MsgRegisterFrontend{Signer: admin, ServiceSigner: owner, ProviderId: label, ServicePubkey: private.Public().(ed25519.PublicKey), Endpoint: "https://frontend.invalid"}
	message.ServiceSignature = ed25519.Sign(private, FrontendTranscript(message))
	return message
}

func TestR1FrontendAuthorityAndRemoval(t *testing.T) {
	keeper, ctx, admin := scenarioKeeper(t)
	owner, other := scenarioAddress("owner"), scenarioAddress("other")
	requireError(t, keeper.Execute(ctx, revisionFrontend(owner, owner, "unapproved")), ErrFrontendAuthority)
	message := revisionFrontend(admin.address, owner, "approved")
	requireError(t, keeper.Execute(ctx, message), nil)
	requireError(t, keeper.Execute(ctx, revisionFrontend(admin.address, owner, "second")), ErrLimitExceeded)
	requireError(t, keeper.Execute(ctx, revisionFrontend(admin.address, other, "other")), nil)
	t.Log("PASS frontend needs hub evidence or admin")
	requireError(t, keeper.Execute(ctx, &MsgRemoveFrontend{Signer: owner, ProviderId: message.ProviderId}), ErrFrontendAuthority)
	requireError(t, keeper.Execute(ctx, &MsgRemoveFrontend{Signer: admin.address, ProviderId: message.ProviderId}), nil)
	requireError(t, keeper.Execute(ctx, message), nil)
	requireError(t, keeper.Execute(ctx, &MsgRemoveFrontend{Signer: admin.address, ProviderId: "missing"}), ErrFrontend)
	t.Log("PASS admin removes frontend")
}

func TestR1DistinctRefusalCodes(t *testing.T) {
	keeper, ctx, admin := scenarioKeeper(t)
	owner := scenarioAddress("owner")
	choice := &MsgSetStorageChoice{Signer: admin.address, Owner: owner, Locations: []*Location{{Name: "phone", Pointer: "opaque"}}}
	actual := []error{keeper.Execute(ctx, choice), keeper.Execute(ctx, &MsgUpdateParams{Signer: owner})}
	frontend := scenarioFrontend(admin.address)
	frontend.ServiceSignature[0] ^= 1
	actual = append(actual, keeper.Execute(ctx, frontend))
	choice.Signer = owner
	actual = append(actual, keeper.Execute(ctx.WithChainID("wrong"), choice))
	choice.Locations[0].Name = "unknown"
	actual = append(actual, keeper.Execute(ctx, choice))
	choice.Locations[0].Name, choice.Locations[0].Pointer = "phone", string(bytes.Repeat([]byte{'x'}, 257))
	actual = append(actual, keeper.Execute(ctx, choice))
	expected := []error{ErrWrongSigner, ErrNonAdminParams, ErrFrontendSignature, ErrWrongChain, ErrUnknownLocation, ErrOversizePointer}
	seen := map[uint32]bool{}
	for index, err := range actual {
		requireError(t, err, expected[index])
		space, code, _ := errorsmod.ABCIInfo(err, false)
		if space != chainprefix.ModuleName || code == 0 || seen[code] {
			t.Fatal("refusal codes overlap")
		}
		seen[code] = true
	}
	t.Log("PASS distinct refusal codes")
}

func TestR1GenesisChoices(t *testing.T) {
	for _, location := range []string{"base", "traderswift"} {
		state := DefaultGenesisState()
		state.Choices = []*StorageChoice{{Owner: scenarioAddress("owner"), Locations: []*Location{{Name: location, Pointer: "opaque"}}}}
		requireError(t, ValidateGenesisChoices(state), ErrLocationDisabled)
		state.Params.BaseEnabled, state.Params.TraderswiftEnabled = true, true
		requireError(t, ValidateGenesisChoices(state), nil)
	}
	t.Log("PASS genesis choices checked against params")
}
