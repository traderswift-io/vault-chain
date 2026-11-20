package vault

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"vault-chain/chainprefix"
)

func TestAdditionalMessageBounds(t *testing.T) {
	keeper, ctx, _ := scenarioKeeper(t)
	owner := scenarioAddress("owner")
	grant, _ := scenarioGrant(owner, ctx)
	for _, message := range []any{
		&MsgLinkAccount{Signer: owner, Owner: owner},
		&MsgSetStorageChoice{Signer: owner, Owner: owner},
		&MsgCommitRoot{Signer: owner, Owner: owner, Day: "9999-01-01", Root: bytes.Repeat([]byte{1}, 32)},
		&MsgGrant{Signer: owner, Owner: owner, Scope: chainprefix.NoteScope, Expiry: blockSeconds(ctx) + 1},
		&MsgRevoke{Signer: owner, Owner: owner, GranteeCommitment: grant.GranteeCommitment, Scope: grant.Scope},
		&MsgRegisterFrontend{Signer: owner, ProviderId: strings.Repeat("a", 65)},
		&MsgUpdateParams{Signer: "invalid"},
	} {
		before := stateBytes(t, keeper, ctx)
		if keeper.Execute(ctx, message) == nil {
			t.Fatalf("invalid %T accepted", message)
		}
		if !bytes.Equal(before, stateBytes(t, keeper, ctx)) {
			t.Fatal("invalid message changed state")
		}
	}
	grant.Expiry = blockSeconds(ctx) + MaxGrantLifetime + 1
	requireError(t, keeper.Execute(ctx, grant), ErrGrant)
	grant.Expiry = blockSeconds(ctx) + 1
	requireError(t, keeper.Execute(ctx.WithChainID("wrong-chain"), grant), ErrWrongChain)
	query := NewQueryServer(keeper)
	for _, request := range []*QueryMayReadRequest{nil, {Owner: owner}, {Owner: owner, Grantee: owner, Salt: bytes.Repeat([]byte{1}, 33), Scope: chainprefix.NoteScope}} {
		_, err := query.MayRead(ctx, request)
		requireError(t, err, ErrQuery)
	}
	response, err := query.Where(ctx, &QueryWhereRequest{Owner: owner})
	requireError(t, err, nil)
	if string(response.CanonicalRecord) != `{"link":null,"choice":null}` {
		t.Fatal("missing owner query fabricated records")
	}
	server := NewMsgServer(keeper)
	_, err = server.LinkAccount(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.SetStorageChoice(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.CommitRoot(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.Grant(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.Revoke(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.RegisterFrontend(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.RemoveFrontend(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.ActivateProvider(ctx, nil)
	requireError(t, err, ErrBadEncoding)
	_, err = server.UpdateParams(ctx, nil)
	requireError(t, err, ErrBadEncoding)
}

func TestGrantCommitmentVector(t *testing.T) {
	commitment := GrantCommitment(bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 20))
	if hex.EncodeToString(commitment) != "85d6725f78fe290cafd12640e0473af88e6d7309a41b40775333cdf0c27fbec5" {
		t.Fatal("commitment must hash salt followed by raw address bytes")
	}
	if GrantCommitment(make([]byte, 31), make([]byte, 20)) != nil || GrantCommitment(make([]byte, 32), make([]byte, 19)) != nil {
		t.Fatal("invalid commitment witness length")
	}
}
