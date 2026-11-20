package validatorset

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
)

func TestActiveAttestationPowerFollowsMembership(t *testing.T) {
	key := storetypes.NewKVStoreKey(StoreKey)
	ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient_validator_power"))
	keeper := NewKeeper(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), runtime.NewKVStoreService(key))
	entries := syntheticValidators()
	for _, entry := range entries {
		if err := keeper.validators.Set(ctx, entry.ConsensusPublicKey, entry); err != nil {
			t.Fatal(err)
		}
	}
	powers, err := keeper.ActiveValidatorPowers(ctx)
	if err != nil || len(powers) != 4 {
		t.Fatalf("initial power: %v", err)
	}
	for _, entry := range entries[2:] {
		if err := keeper.validators.Remove(ctx, entry.ConsensusPublicKey); err != nil {
			t.Fatal(err)
		}
	}
	powers, err = keeper.ActiveValidatorPowers(ctx)
	if err != nil || len(powers) != 2 {
		t.Fatalf("removal was not reflected: %v", err)
	}
	for _, entry := range entries[2:] {
		entry.Power = 200
		if err := keeper.validators.Set(ctx, entry.ConsensusPublicKey, entry); err != nil {
			t.Fatal(err)
		}
	}
	powers, err = keeper.ActiveValidatorPowers(ctx)
	publicKey, keyError := decodeConsensusPublicKey(entries[3].ConsensusPublicKey)
	if err != nil || keyError != nil || len(powers) != 4 || powers[string(publicKey)] != 200 {
		t.Fatalf("re-add/reweight was not reflected: %v %v", err, keyError)
	}
}
