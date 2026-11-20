package main

import (
	"encoding/base64"
	"encoding/json"

	"vault-chain/validatorset"
	"vault-chain/vault"
)

func validateModuleGenesis(modules map[string]json.RawMessage) error {
	for _, name := range []string{"auth", "genutil", "validatorset", vault.ModuleName} {
		if _, exists := modules[name]; !exists {
			return vault.ErrBadEncoding
		}
	}
	for name, data := range modules {
		switch name {
		case "auth", "consensus", "genutil", "validatorset", vault.ModuleName:
		default:
			return vault.ErrBadEncoding
		}
		limit := vault.MaxGenesisBytes
		if name != vault.ModuleName {
			limit = 32768
		}
		if err := vault.CheckJSON(data, limit); err != nil {
			return err
		}
	}
	state := &vault.State{}
	if err := vault.DecodeCanonical(modules[vault.ModuleName], state, vault.MaxGenesisBytes); err != nil {
		return err
	}
	if err := vault.ValidateGenesisState(state); err != nil {
		return err
	}
	if state.Params.BaseEnabled || state.Params.TraderswiftEnabled {
		return vault.ErrParams
	}
	if err := vault.ValidateGenesisChoices(state); err != nil {
		return err
	}
	return validateGenesisProviders(modules[validatorset.ModuleName], state)
}

func validateGenesisProviders(data []byte, state *vault.State) error {
	var validatorState struct {
		Validators []struct {
			ConsensusPublicKey string      `json:"consensus_public_key"`
			Power              json.Number `json:"power"`
			Moniker            string      `json:"moniker"`
			Founding           bool        `json:"founding"`
			Vacant             bool        `json:"vacant"`
		} `json:"validators"`
	}
	if len(data) > 32768 {
		return vault.ErrLimitExceeded
	}
	if err := json.Unmarshal(data, &validatorState); err != nil {
		return vault.ErrBadEncoding
	}
	if len(validatorState.Validators) != len(state.Providers) {
		return vault.ErrProvider
	}
	seen := map[string]bool{}
	for _, entry := range validatorState.Validators {
		if seen[entry.ConsensusPublicKey] {
			return vault.ErrProvider
		}
		seen[entry.ConsensusPublicKey] = true
		matched := false
		for _, provider := range state.Providers {
			if entry.ConsensusPublicKey == base64.StdEncoding.EncodeToString(provider.ConsensusPubkey) {
				power, err := entry.Power.Int64()
				if err != nil || power != provider.Power || entry.Founding != provider.Founding || entry.Vacant != provider.Vacant || entry.Moniker != provider.ProviderID {
					return vault.ErrProvider
				}
				matched = true
			}
		}
		if !matched {
			return vault.ErrProvider
		}
	}
	return nil
}
