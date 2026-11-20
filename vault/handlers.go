package vault

import (
	"context"
	"encoding/json"

	"vault-chain/chainprefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (keeper *Keeper) Execute(context context.Context, message any) error {
	encoded, err := json.Marshal(message)
	if err != nil || len(encoded) > MaxMessageBytes {
		return ErrLimitExceeded
	}
	ctx, commit := sdk.UnwrapSDKContext(context).CacheContext()
	if ctx.ChainID() != chainprefix.ChainID {
		return ErrWrongChain
	}
	if err := keeper.apply(ctx, message); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(ModuleName + ".accepted"))
	commit()
	return nil
}

func (keeper *Keeper) apply(ctx sdk.Context, message any) error {
	switch value := message.(type) {
	case *MsgLinkAccount:
		return keeper.linkAccount(ctx, value)
	case *MsgSetStorageChoice:
		return keeper.setChoice(ctx, value)
	case *MsgCommitRoot:
		return keeper.commitRoot(ctx, value)
	case *MsgGrant:
		return keeper.grant(ctx, value)
	case *MsgRevoke:
		return keeper.revoke(ctx, value)
	case *MsgRegisterFrontend:
		return keeper.registerFrontend(ctx, value)
	case *MsgRemoveFrontend:
		return keeper.removeFrontend(ctx, value)
	case *MsgActivateProvider:
		return keeper.activateProvider(ctx, value)
	case *MsgRemoveProvider:
		return keeper.removeProvider(ctx, value)
	case *MsgReplaceProviderOwner:
		return keeper.replaceProviderOwner(ctx, value)
	case *MsgUpdateParams:
		if keeper.admin == nil || !validOwner(value.Signer) {
			return ErrParams
		}
		admin, err := keeper.admin.Admin(ctx)
		if err != nil || admin != value.Signer {
			return ErrNonAdminParams
		}
		return keeper.put(ctx, []byte{paramsTable}, &Params{BaseEnabled: value.BaseEnabled, TraderswiftEnabled: value.TraderswiftEnabled})
	default:
		return ErrBadEncoding
	}
}

func (keeper *Keeper) setChoice(ctx sdk.Context, value *MsgSetStorageChoice) error {
	if err := authorize(value.Signer, value.Owner); err != nil {
		return err
	}
	for _, location := range value.Locations {
		if location != nil && location.Name != "phone" && location.Name != "base" && location.Name != "traderswift" {
			return ErrUnknownLocation
		}
		if location != nil && len(location.Pointer) > 256 {
			return ErrOversizePointer
		}
	}
	if !validLocations(value.Locations) {
		return ErrChoice
	}
	params, err := keeper.params(ctx)
	if err != nil {
		return err
	}
	for _, location := range value.Locations {
		if location.Name == "base" && !params.BaseEnabled || location.Name == "traderswift" && !params.TraderswiftEnabled {
			return ErrLocationDisabled
		}
	}
	return keeper.put(ctx, recordKey(choiceTable, value.Owner), &StorageChoice{Owner: value.Owner, Locations: value.Locations})
}

func (keeper *Keeper) commitRoot(ctx sdk.Context, value *MsgCommitRoot) error {
	if err := authorize(value.Signer, value.Owner); err != nil {
		return err
	}
	if !validDay(value.Day) || value.Day != ctx.BlockTime().UTC().Format("2006-01-02") || len(value.Root) != 32 {
		return ErrRoot
	}
	day := ctx.BlockTime().UTC().Format("2006-01-02")
	key := recordKey(rootTable, value.Owner, day)
	found, err := keeper.storeService.OpenKVStore(ctx).Has(key)
	if err != nil {
		return err
	}
	if found {
		return ErrRootExists
	}
	return keeper.put(ctx, key, &CommitmentRoot{Owner: value.Owner, Day: day, Root: value.Root})
}

func (keeper *Keeper) grant(ctx sdk.Context, value *MsgGrant) error {
	if err := authorize(value.Signer, value.Owner); err != nil {
		return err
	}
	now := blockSeconds(ctx)
	if len(value.GranteeCommitment) != 32 || !scopeName(value.Scope) || value.Expiry <= now || value.Expiry-now > MaxGrantLifetime {
		return ErrGrant
	}
	entry := &Grant{Owner: value.Owner, GranteeCommitment: value.GranteeCommitment, Scope: value.Scope, Expiry: value.Expiry}
	return keeper.put(ctx, recordKey(grantTable, value.Owner, string(value.GranteeCommitment), value.Scope), entry)
}

func (keeper *Keeper) revoke(ctx sdk.Context, value *MsgRevoke) error {
	if err := authorize(value.Signer, value.Owner); err != nil {
		return err
	}
	if len(value.GranteeCommitment) != 32 || !scopeName(value.Scope) {
		return ErrRevoke
	}
	key := recordKey(grantTable, value.Owner, string(value.GranteeCommitment), value.Scope)
	entry := &Grant{}
	found, err := keeper.read(ctx, key, entry)
	if err != nil {
		return err
	}
	if !found || entry.Revoked {
		return ErrRevoke
	}
	entry.Revoked = true
	return keeper.put(ctx, key, entry)
}
