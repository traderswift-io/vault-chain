package vault

import (
	"bytes"
	"math/big"
)

func validateRecoveryProviders(providers []*Provider, independent [][]byte) error {
	vacancies := 0
	retired := map[string]string{}
	for _, publicKey := range independent {
		if len(publicKey) != 32 || retired[string(publicKey)] != "" {
			return ErrProviderRetired
		}
		retired[string(publicKey)] = "\x00"
	}
	for _, entry := range providers {
		if entry == nil || len(entry.RetiredConsensusPubkeys) > MaxRetiredConsensusKeys {
			return ErrProvider
		}
		if entry.Vacant {
			vacancies++
			if !entry.Founding || vacancies > 1 || len(entry.RetiredConsensusPubkeys) == 0 {
				return ErrProviderVacancy
			}
			if !bytes.Equal(entry.ConsensusPubkey, entry.RetiredConsensusPubkeys[len(entry.RetiredConsensusPubkeys)-1]) {
				return ErrProvider
			}
		}
		if !entry.Founding && len(entry.RetiredConsensusPubkeys) != 0 {
			return ErrProvider
		}
		for _, publicKey := range entry.RetiredConsensusPubkeys {
			if len(publicKey) != 32 || retired[string(publicKey)] != "" {
				return ErrProviderRetired
			}
			retired[string(publicKey)] = entry.ProviderID
		}
	}
	for _, entry := range providers {
		if owner := retired[string(entry.ConsensusPubkey)]; owner != "" && (!entry.Vacant || owner != entry.ProviderID) {
			return ErrProviderRetired
		}
	}
	return nil
}

func validateRecoveryCapacity(providers []*Provider) error {
	foundingPower, outsiderPower, largestFounder := new(big.Int), new(big.Int), new(big.Int)
	vacant := false
	for _, entry := range providers {
		power := big.NewInt(entry.Power)
		if entry.Founding {
			foundingPower.Add(foundingPower, power)
			if power.Cmp(largestFounder) > 0 {
				largestFounder.Set(power)
			}
		} else {
			outsiderPower.Add(outsiderPower, power)
		}
		vacant = vacant || entry.Vacant
	}
	remaining := new(big.Int).Sub(foundingPower, new(big.Int).Mul(largestFounder, big.NewInt(3)))
	if vacant && new(big.Int).Mul(outsiderPower, big.NewInt(2)).Cmp(remaining) >= 0 {
		return ErrProviderRecoveryCapacity
	}
	return nil
}
