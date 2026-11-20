package validatorset

import "context"

func (keeper *Keeper) ActiveValidatorPowers(ctx context.Context) (map[string]int64, error) {
	validators, err := keeper.allValidators(ctx)
	if err != nil {
		return nil, err
	}
	powers := map[string]int64{}
	for _, validator := range validators {
		if validator.Vacant {
			continue
		}
		publicKey, err := decodeConsensusPublicKey(validator.ConsensusPublicKey)
		if err != nil {
			return nil, err
		}
		powers[string(publicKey)] = validator.Power
	}
	return powers, nil
}
