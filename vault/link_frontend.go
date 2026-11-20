package vault

import (
	"crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (keeper *Keeper) linkAccount(ctx sdk.Context, value *MsgLinkAccount) error {
	if err := authorize(value.Signer, value.Owner); err != nil {
		return err
	}
	if !verifyHubLink(value) {
		return ErrLink
	}
	previous := &AccountLink{}
	_, err := keeper.read(ctx, recordKey(linkTable, value.Owner), previous)
	if err != nil {
		return err
	}
	if previous.Nonce == ^uint64(0) || value.Nonce != previous.Nonce+1 {
		return ErrLinkReplay
	}
	return keeper.put(ctx, recordKey(linkTable, value.Owner), &AccountLink{Owner: value.Owner, HubAccount: value.HubAccount, Nonce: value.Nonce})
}

func (keeper *Keeper) ValidateFirstLink(ctx sdk.Context, value *MsgLinkAccount) error {
	present, err := keeper.storeService.OpenKVStore(ctx).Has(recordKey(linkTable, value.Owner))
	if err != nil {
		return err
	}
	if present {
		return ErrLinkReplay
	}
	cached, _ := ctx.CacheContext()
	return keeper.Execute(cached, value)
}

func (keeper *Keeper) isAdmin(ctx sdk.Context, signer string) bool {
	if keeper.admin == nil || !validOwner(signer) {
		return false
	}
	admin, err := keeper.admin.Admin(ctx)
	return err == nil && signer == admin
}

func (keeper *Keeper) registerFrontend(ctx sdk.Context, value *MsgRegisterFrontend) error {
	if !identifier(value.ProviderId) || !validEndpoint(value.Endpoint) || len(value.ServicePubkey) != 32 || len(value.ServiceSignature) != 64 {
		return ErrFrontend
	}
	if !ed25519.Verify(value.ServicePubkey, FrontendTranscript(value), value.ServiceSignature) {
		return ErrFrontendSignature
	}
	if !keeper.isAdmin(ctx, value.Signer) {
		return ErrFrontendAuthority
	}
	owner := value.ServiceSigner
	if owner == "" {
		owner = value.Signer
	}
	if !validOwner(owner) {
		return ErrFrontend
	}
	store := keeper.storeService.OpenKVStore(ctx)
	key := recordKey(frontendTable, value.ProviderId)
	for _, unique := range [][]byte{key, recordKey(frontendKeyTable, string(value.ServicePubkey))} {
		present, err := store.Has(unique)
		if err != nil {
			return err
		}
		if present {
			return ErrFrontend
		}
	}
	entry := &Frontend{ProviderID: value.ProviderId, Kind: "frontend", ServiceSigner: owner, ServicePubkey: value.ServicePubkey, Endpoint: value.Endpoint, CapacityClass: "inactive", Status: "active"}
	if err := keeper.put(ctx, key, entry); err != nil {
		return err
	}
	return store.Set(recordKey(frontendKeyTable, string(value.ServicePubkey)), []byte(value.ProviderId))
}

func (keeper *Keeper) removeFrontend(ctx sdk.Context, value *MsgRemoveFrontend) error {
	if !keeper.isAdmin(ctx, value.Signer) {
		return ErrFrontendAuthority
	}
	if !identifier(value.ProviderId) {
		return ErrFrontend
	}
	entry := &Frontend{}
	key := recordKey(frontendTable, value.ProviderId)
	present, err := keeper.read(ctx, key, entry)
	if err != nil {
		return err
	}
	if !present {
		return ErrFrontend
	}
	if err := keeper.updateOwnerBudget(ctx, entry, false); err != nil {
		return err
	}
	store := keeper.storeService.OpenKVStore(ctx)
	if err := store.Delete(recordKey(frontendKeyTable, string(entry.ServicePubkey))); err != nil {
		return err
	}
	return store.Delete(key)
}
