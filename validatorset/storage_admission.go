package validatorset

import (
	"context"
	"encoding/base64"
)

func (keeper *Keeper) ActivateProvider(ctx context.Context, publicKey []byte, power int64, id string) error {
	if len(publicKey) != 32 {
		return ErrInvalidField
	}
	key := base64.StdEncoding.EncodeToString(publicKey)
	present, err := keeper.validators.Has(ctx, key)
	if err != nil {
		return err
	}
	if present {
		return ErrAlreadyExists
	}
	if keeper.storageAdmission == nil {
		return ErrUnauthorized
	}
	if err := keeper.storageAdmission(ctx, publicKey, power, id, false); err != nil {
		return err
	}
	value := &Validator{ConsensusPublicKey: key, Power: power, Moniker: id}
	if err := validateValidator(*value); err != nil {
		return err
	}
	current, err := keeper.allValidators(ctx)
	if err != nil {
		return err
	}
	previous := key
	for _, entry := range current {
		if entry.Moniker == id {
			if !entry.Vacant || !entry.Founding || entry.Power != power {
				return ErrInvalidField
			}
			previous = entry.ConsensusPublicKey
			value.Founding = true
		}
	}
	if err := keeper.applyChange(ctx, previous, value); err != nil {
		return err
	}
	return nil
}

func (keeper *Keeper) RemoveProvider(ctx context.Context, publicKey []byte) error {
	if len(publicKey) != 32 {
		return ErrInvalidField
	}
	key := base64.StdEncoding.EncodeToString(publicKey)
	entry, err := keeper.validators.Get(ctx, key)
	if err != nil {
		return err
	}
	if keeper.storageAdmission == nil {
		return ErrUnauthorized
	}
	if err := keeper.storageAdmission(ctx, publicKey, entry.Power, entry.Moniker, true); err != nil {
		return err
	}
	current, err := keeper.allValidators(ctx)
	if err != nil {
		return err
	}
	for _, validator := range current {
		if validator.Vacant {
			return ErrRecoveryInProgress
		}
	}
	if entry.Founding {
		entry.Vacant = true
		return keeper.applyChange(ctx, key, &entry)
	}
	return keeper.applyChange(ctx, key, nil)
}

func (keeper *Keeper) SetStorageAdmission(policy func(context.Context, []byte, int64, string, bool) error) {
	if keeper.storageAdmission != nil {
		panic("admission policy already installed")
	}
	keeper.storageAdmission = policy
}
