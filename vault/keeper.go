package vault

import (
	"context"
	"encoding/binary"
	"encoding/json"

	"cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	linkTable        = byte(1)
	choiceTable      = byte(2)
	rootTable        = byte(3)
	grantTable       = byte(4)
	frontendTable    = byte(5)
	paramsTable      = byte(6)
	providerTable    = byte(7)
	admissionTable   = byte(8)
	budgetTable      = byte(9)
	frontendKeyTable = byte(10)
	retiredTable     = byte(11)
)

type AdminSource interface {
	Admin(sdk.Context) (string, error)
}

type Keeper struct {
	storeService store.KVStoreService
	admin        AdminSource
}

func NewKeeper(service store.KVStoreService, admin AdminSource) *Keeper {
	return &Keeper{storeService: service, admin: admin}
}

func recordKey(table byte, components ...string) []byte {
	key := []byte{table}
	for _, component := range components {
		key = binary.BigEndian.AppendUint32(key, uint32(len(component)))
		key = append(key, component...)
	}
	return key
}

func (keeper *Keeper) read(ctx context.Context, key []byte, target any) (bool, error) {
	data, err := keeper.storeService.OpenKVStore(ctx).Get(key)
	if err != nil || data == nil {
		return false, err
	}
	return true, DecodeCanonical(data, target, MaxMessageBytes)
}

func (keeper *Keeper) put(ctx context.Context, key []byte, record any) error {
	data, err := json.Marshal(record)
	if err != nil || len(data) > MaxMessageBytes {
		return ErrLimitExceeded
	}
	store := keeper.storeService.OpenKVStore(ctx)
	previous, err := store.Get(key)
	if err != nil {
		return err
	}
	if previous == nil {
		if err := keeper.updateOwnerBudget(ctx, record, true); err != nil {
			return err
		}
	}
	return store.Set(key, data)
}

func (keeper *Keeper) params(ctx context.Context) (Params, error) {
	var result Params
	found, err := keeper.read(ctx, []byte{paramsTable}, &result)
	if err == nil && !found {
		err = ErrParams
	}
	return result, err
}
