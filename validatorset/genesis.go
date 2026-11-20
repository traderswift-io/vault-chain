package validatorset

import (
	"context"

	"vault-chain/chainprefix"

	abci "github.com/cometbft/cometbft/abci/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Admin: chainprefix.GenesisAdmin}
}

func ValidateGenesisState(state *GenesisState) error {
	if state == nil || len(state.Admin) > 90 || len(state.PendingAdmin) > 90 || len(state.Validators) > 32 {
		return ErrInvalidField
	}
	admin, err := genesisAdminAddress(state.Admin)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidField, "admin %q: %v", state.Admin, err)
	}
	if state.PendingAdmin != "" {
		if _, err := sdk.AccAddressFromBech32(state.PendingAdmin); err != nil {
			return errorsmod.Wrapf(ErrInvalidField, "pending_admin %q: %v", state.PendingAdmin, err)
		}
		if state.PendingAdmin == admin {
			return errorsmod.Wrap(ErrInvalidField, "pending_admin is already the admin")
		}
	}
	seenKeys := map[string]bool{}
	validators := make([]Validator, 0, len(state.Validators))
	for _, validator := range state.Validators {
		if validator == nil {
			return errorsmod.Wrap(ErrInvalidField, "empty validator entry")
		}
		if err := validateValidator(*validator); err != nil {
			return err
		}
		if seenKeys[validator.ConsensusPublicKey] {
			return errorsmod.Wrapf(ErrAlreadyExists, "validator %s twice in genesis", validator.ConsensusPublicKey)
		}
		seenKeys[validator.ConsensusPublicKey] = true
		validators = append(validators, *validator)
	}
	return requireSetWithinOutsiderLimit(validators)
}

func (keeper *Keeper) InitializeFromGenesis(ctx context.Context, state *GenesisState) ([]abci.ValidatorUpdate, error) {
	if err := ValidateGenesisState(state); err != nil {
		return nil, err
	}
	return keeper.initializeValidatedGenesis(ctx, state)
}

func (keeper *Keeper) ExportToGenesis(ctx context.Context) (*GenesisState, error) {
	admin, err := keeper.AdminAddress.Get(ctx)
	if err != nil {
		return nil, err
	}
	validators, err := keeper.allValidators(ctx)
	if err != nil {
		return nil, err
	}
	state := &GenesisState{Admin: admin}
	state.PendingAdmin, err = keeper.pendingAdmin(ctx)
	if err != nil {
		return nil, err
	}
	for index := range validators {
		state.Validators = append(state.Validators, &validators[index])
	}
	return state, nil
}
