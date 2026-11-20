package validatorset

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
)

func (keeper *Keeper) initializeValidatedGenesis(ctx context.Context, state *GenesisState) ([]abci.ValidatorUpdate, error) {
	admin, err := genesisAdminAddress(state.Admin)
	if err != nil {
		return nil, err
	}
	if err := keeper.AdminAddress.Set(ctx, admin); err != nil {
		return nil, err
	}
	if state.PendingAdmin != "" {
		if err := keeper.PendingAdminAddress.Set(ctx, state.PendingAdmin); err != nil {
			return nil, err
		}
	}
	updates := []abci.ValidatorUpdate{}
	for _, validator := range state.Validators {
		if err := keeper.validators.Set(ctx, validator.ConsensusPublicKey, *validator); err != nil {
			return nil, err
		}
		if validator.Vacant {
			continue
		}
		if err := keeper.PowerLastSentToConsensus.Set(ctx, validator.ConsensusPublicKey, validator.Power); err != nil {
			return nil, err
		}
		update, err := consensusUpdate(validator.ConsensusPublicKey, validator.Power)
		if err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}
	if err := keeper.ValidatorSetChangedThisBlock.Set(ctx, false); err != nil {
		return nil, err
	}
	return updates, nil
}
