package vault

import (
	"bytes"
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type providerRemovalAuthorization struct{}

type ProviderRemoval interface {
	RemoveProvider(context.Context, []byte) error
}

func (keeper *Keeper) recoveryProvider(ctx sdk.Context, signer, id string) (*Provider, error) {
	if !keeper.isAdmin(ctx, signer) {
		return nil, ErrProviderAuthority
	}
	if !identifier(id) {
		return nil, ErrProvider
	}
	entry := &Provider{}
	found, err := keeper.read(ctx, recordKey(providerTable, id), entry)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrProviderMissing
	}
	return entry, nil
}

func (keeper *Keeper) authorizeProviderRemoval(ctx context.Context, publicKey []byte, power int64, id string) error {
	if ctx.Value(providerRemovalAuthorization{}) != id {
		return ErrProviderAuthority
	}
	entry := &Provider{}
	found, err := keeper.read(ctx, recordKey(providerTable, id), entry)
	if err != nil {
		return err
	}
	if !found || entry.Power != power || !bytes.Equal(entry.ConsensusPubkey, publicKey) {
		return ErrProvider
	}
	return nil
}

func (keeper *Keeper) removeProvider(ctx sdk.Context, value *MsgRemoveProvider) error {
	entry, err := keeper.recoveryProvider(ctx, value.Signer, value.ProviderId)
	if err != nil {
		return err
	}
	providers, err := keeper.providers(ctx)
	if err != nil {
		return err
	}
	for _, provider := range providers {
		if provider.Vacant {
			return ErrProviderVacancy
		}
	}
	if entry.Founding && len(entry.RetiredConsensusPubkeys) >= MaxRetiredConsensusKeys {
		return ErrLimitExceeded
	}
	membership, ok := keeper.admin.(ProviderRemoval)
	if !ok {
		return ErrProvider
	}
	authorized := context.WithValue(ctx, providerRemovalAuthorization{}, entry.ProviderID)
	if err := membership.RemoveProvider(authorized, entry.ConsensusPubkey); err != nil {
		return err
	}
	if entry.Founding {
		entry.Vacant = true
		entry.RetiredConsensusPubkeys = append(entry.RetiredConsensusPubkeys, entry.ConsensusPubkey)
		return keeper.put(ctx, recordKey(providerTable, entry.ProviderID), entry)
	}
	retired, err := keeper.retiredConsensus(ctx)
	if err != nil {
		return err
	}
	retired = append(retired, entry.ConsensusPubkey)
	if err := keeper.put(ctx, []byte{retiredTable}, &retired); err != nil {
		return err
	}
	if err := keeper.updateOwnerBudget(ctx, entry, false); err != nil {
		return err
	}
	return keeper.storeService.OpenKVStore(ctx).Delete(recordKey(providerTable, entry.ProviderID))
}

func (keeper *Keeper) replaceProviderOwner(ctx sdk.Context, value *MsgReplaceProviderOwner) error {
	entry, err := keeper.recoveryProvider(ctx, value.Signer, value.ProviderId)
	if err != nil {
		return err
	}
	if !validOwner(value.NewOwner) {
		return ErrProviderOwner
	}
	providers, err := keeper.providers(ctx)
	if err != nil {
		return err
	}
	for _, provider := range providers {
		if provider.Owner == value.NewOwner {
			return ErrProviderOwnerUsed
		}
	}
	if err := keeper.updateOwnerBudget(ctx, entry, false); err != nil {
		return err
	}
	entry.Owner = value.NewOwner
	if err := keeper.updateOwnerBudget(ctx, entry, true); err != nil {
		return err
	}
	return keeper.put(ctx, recordKey(providerTable, entry.ProviderID), entry)
}
