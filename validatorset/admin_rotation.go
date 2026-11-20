package validatorset

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var ErrNoPendingAdmin = errorsmod.Register(ModuleName, 11, "no admin rotation is pending")

func (keeper *Keeper) pendingAdmin(ctx context.Context) (string, error) {
	pending, err := keeper.PendingAdminAddress.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return "", nil
	}
	return pending, err
}

func (server msgServer) SetAdmin(ctx context.Context, message *MsgSetAdmin) (*MsgSetAdminResponse, error) {
	if err := server.keeper.requireAdmin(ctx, message.Admin); err != nil {
		return nil, err
	}
	if _, err := sdk.AccAddressFromBech32(message.NewAdmin); err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidField, "new_admin %q: %v", message.NewAdmin, err)
	}
	if message.NewAdmin == message.Admin {
		return nil, errorsmod.Wrap(ErrInvalidField, "new_admin is already the admin")
	}
	replaced, err := server.keeper.pendingAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := server.keeper.PendingAdminAddress.Set(ctx, message.NewAdmin); err != nil {
		return nil, err
	}
	emitValidatorSetEvent(ctx, "admin_rotation_proposed",
		sdk.NewAttribute("admin", message.Admin),
		sdk.NewAttribute("pending_admin", message.NewAdmin),
		sdk.NewAttribute("replaced_pending_admin", replaced))
	return &MsgSetAdminResponse{}, nil
}

func (server msgServer) AcceptAdmin(ctx context.Context, message *MsgAcceptAdmin) (*MsgAcceptAdminResponse, error) {
	pending, err := server.keeper.pendingAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if pending == "" {
		return nil, ErrNoPendingAdmin
	}
	if message.NewAdmin != pending {
		return nil, errorsmod.Wrapf(ErrUnauthorized, "%s is not the pending admin", message.NewAdmin)
	}
	previous, err := server.keeper.AdminAddress.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := server.keeper.AdminAddress.Set(ctx, pending); err != nil {
		return nil, err
	}
	if err := server.keeper.PendingAdminAddress.Remove(ctx); err != nil {
		return nil, err
	}
	emitValidatorSetEvent(ctx, "admin_rotation_accepted",
		sdk.NewAttribute("previous_admin", previous),
		sdk.NewAttribute("admin", pending))
	return &MsgAcceptAdminResponse{}, nil
}
