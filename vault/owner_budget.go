package vault

import (
	"context"
	"encoding/binary"
)

const MaxOwnerGrants = 256
const MaxOwnerRoots = 65536

func recordBudget(record any) ([]byte, uint64) {
	var owner string
	var family byte
	limit := uint64(1)
	switch entry := record.(type) {
	case *AccountLink:
		owner, family = entry.Owner, linkTable
	case *StorageChoice:
		owner, family = entry.Owner, choiceTable
	case *CommitmentRoot:
		owner, family, limit = entry.Owner, rootTable, MaxOwnerRoots
	case *Grant:
		owner, family, limit = entry.Owner, grantTable, MaxOwnerGrants
	case *Frontend:
		owner, family = entry.ServiceSigner, frontendTable
	case *Provider:
		owner, family = entry.Owner, providerTable
	default:
		return nil, 0
	}
	return recordKey(budgetTable, owner, string([]byte{family})), limit
}

func (keeper *Keeper) updateOwnerBudget(ctx context.Context, record any, adding bool) error {
	key, limit := recordBudget(record)
	if key == nil {
		return nil
	}
	store := keeper.storeService.OpenKVStore(ctx)
	encoded, err := store.Get(key)
	if err != nil {
		return err
	}
	var count uint64
	if encoded != nil {
		if len(encoded) != 8 {
			return ErrBadEncoding
		}
		count = binary.BigEndian.Uint64(encoded)
	}
	if adding {
		if count >= limit {
			return ErrLimitExceeded
		}
		count++
	} else {
		if count == 0 {
			return ErrBadEncoding
		}
		count--
	}
	return store.Set(key, binary.BigEndian.AppendUint64(nil, count))
}
