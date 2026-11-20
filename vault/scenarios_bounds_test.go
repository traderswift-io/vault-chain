package vault

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/cosmos/gogoproto/proto"
)

func TestGenesisAndDecoderBounds(t *testing.T) {
	keeper, ctx, _ := scenarioKeeper(t)
	owner := scenarioAddress("owner")
	root := &MsgCommitRoot{Signer: owner, Owner: owner, Day: ctx.BlockTime().UTC().Format("2006-01-02"), Root: bytes.Repeat([]byte{2}, 32)}
	requireError(t, keeper.Execute(ctx, root), nil)
	state, err := keeper.ExportToGenesis(ctx)
	requireError(t, err, nil)
	state.Roots = append(state.Roots, state.Roots[0])
	requireError(t, ValidateGenesisState(state), ErrBadEncoding)
	for _, data := range [][]byte{[]byte(`{"x":1,"x":2}`), []byte(`{} {}`), bytes.Repeat([]byte("["), 9)} {
		if CheckJSON(data, MaxGenesisBytes) == nil {
			t.Fatal("unbounded or duplicate JSON accepted")
		}
	}
	valid, err := proto.Marshal(root)
	requireError(t, err, nil)
	requireError(t, ValidateMessageWire(valid, &MsgCommitRoot{}), nil)
	for _, invalid := range [][]byte{append(append([]byte{}, valid...), 0x78, 1), append([]byte{0x0a, 1, 'a'}, valid...), {0x8a, 0, 1, 'a'}, bytes.Repeat([]byte{1}, MaxMessageBytes+1)} {
		if ValidateMessageWire(invalid, &MsgCommitRoot{}) == nil {
			t.Fatal("noncanonical wire accepted")
		}
	}
	random := rand.New(rand.NewSource(18))
	for iteration := 0; iteration < 256; iteration++ {
		data := make([]byte, random.Intn(256))
		_, _ = random.Read(data)
		_ = CheckJSON(data, MaxMessageBytes)
		_ = ValidateMessageWire(data, &MsgGrant{})
	}
	var unknown State
	if DecodeCanonical([]byte(`{"schema_version":1,"unknown":true}`), &unknown, MaxGenesisBytes) == nil {
		t.Fatal("unknown state fields accepted")
	}
	encoded, err := json.Marshal(DefaultGenesisState())
	requireError(t, err, nil)
	requireError(t, DecodeCanonical(encoded, &unknown, MaxGenesisBytes), nil)
	if unknown.Params.BaseEnabled || unknown.Params.TraderswiftEnabled {
		t.Fatal("storage gates enabled by default")
	}
}

func TestAllMessageAndQueryCodecs(t *testing.T) {
	_, ctx, _ := scenarioKeeper(t)
	owner := scenarioAddress("owner")
	grant, query := scenarioGrant(owner, ctx)
	messages := []proto.Message{
		scenarioLink(t, owner, "hub"),
		&MsgSetStorageChoice{Signer: owner, Owner: owner, Locations: []*Location{{Name: "phone", Pointer: "opaque"}}},
		&MsgCommitRoot{Signer: owner, Owner: owner, Day: ctx.BlockTime().UTC().Format("2006-01-02"), Root: bytes.Repeat([]byte{7}, 32)},
		grant,
		&MsgRevoke{Signer: owner, Owner: owner, GranteeCommitment: grant.GranteeCommitment, Scope: grant.Scope},
		scenarioFrontend(owner),
		&MsgUpdateParams{Signer: owner, BaseEnabled: true, TraderswiftEnabled: true},
		&MsgRemoveFrontend{Signer: owner, ProviderId: "skin"},
		&MsgActivateProvider{Signer: owner, ProviderId: "candidate", ConsensusPubkey: bytes.Repeat([]byte{2}, 32), ConsensusSignature: bytes.Repeat([]byte{3}, 64)},
		&QueryWhereRequest{Owner: owner}, query,
	}
	for _, message := range messages {
		encoded, err := proto.Marshal(message)
		requireError(t, err, nil)
		requireError(t, ValidateMessageWire(encoded, message), nil)
		copy := proto.Clone(message)
		copy.Reset()
		requireError(t, proto.Unmarshal(encoded, copy), nil)
		if !proto.Equal(message, copy) {
			t.Fatal("reflection codec lost fields")
		}
	}
}
