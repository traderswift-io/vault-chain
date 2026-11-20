package vault

import (
	"context"
	"encoding/json"
	"sort"

	"vault-chain/chainprefix"
)

func DefaultGenesisState() *State {
	return &State{Providers: []*Provider{}, SchemaVersion: 1, Links: []*AccountLink{}, Choices: []*StorageChoice{}, Roots: []*CommitmentRoot{}, Grants: []*Grant{}, Frontends: []*Frontend{}}
}

func genesisRecords(state *State) (map[string]any, error) {
	if state == nil || state.SchemaVersion != 1 || len(state.Providers) > MaxProviders || state.LastAdmissionTime < 0 {
		return nil, ErrBadEncoding
	}
	records := map[string]any{string([]byte{paramsTable}): &state.Params, string([]byte{admissionTable}): &state.LastAdmissionTime}
	records[string([]byte{retiredTable})] = &state.RetiredPubkeys
	counts := map[string]uint64{}
	add := func(key []byte, value any) error {
		if _, exists := records[string(key)]; exists {
			return ErrBadEncoding
		}
		budget, limit := recordBudget(value)
		if budget != nil {
			counts[string(budget)]++
			if counts[string(budget)] > limit {
				return ErrLimitExceeded
			}
		}
		records[string(key)] = value
		return nil
	}
	for _, entry := range state.Links {
		if entry == nil || !validOwner(entry.Owner) || entry.Nonce == 0 {
			return nil, ErrLink
		}
		if _, err := addressBytes(entry.HubAccount, chainprefix.HubAddressPrefix); err != nil {
			return nil, ErrLink
		}
		if err := add(recordKey(linkTable, entry.Owner), entry); err != nil {
			return nil, err
		}
	}
	for _, entry := range state.Choices {
		if entry == nil || !validOwner(entry.Owner) || !validLocations(entry.Locations) {
			return nil, ErrChoice
		}
		if err := add(recordKey(choiceTable, entry.Owner), entry); err != nil {
			return nil, err
		}
	}
	for _, entry := range state.Roots {
		if entry == nil || !validOwner(entry.Owner) || !validDay(entry.Day) || len(entry.Root) != 32 {
			return nil, ErrRoot
		}
		if err := add(recordKey(rootTable, entry.Owner, entry.Day), entry); err != nil {
			return nil, err
		}
	}
	for _, entry := range state.Grants {
		if entry == nil || !validOwner(entry.Owner) || len(entry.GranteeCommitment) != 32 || !scopeName(entry.Scope) || entry.Expiry == 0 || entry.Expiry > 253402300799 {
			return nil, ErrGrant
		}
		if err := add(recordKey(grantTable, entry.Owner, string(entry.GranteeCommitment), entry.Scope), entry); err != nil {
			return nil, err
		}
	}
	seenSigners, seenKeys := map[string]bool{}, map[string]bool{}
	for _, entry := range state.Frontends {
		if entry == nil || !identifier(entry.ProviderID) || !validOwner(entry.ServiceSigner) || entry.Kind != "frontend" || len(entry.ServicePubkey) != 32 || !validEndpoint(entry.Endpoint) || entry.CapacityClass != "inactive" || entry.CapacityBytes != 0 || entry.Status != "active" {
			return nil, ErrFrontend
		}
		if seenSigners[entry.ServiceSigner] || seenKeys[string(entry.ServicePubkey)] {
			return nil, ErrFrontend
		}
		seenSigners[entry.ServiceSigner], seenKeys[string(entry.ServicePubkey)] = true, true
		if err := add(recordKey(frontendTable, entry.ProviderID), entry); err != nil {
			return nil, err
		}
	}
	if err := validateRecoveryProviders(state.Providers, state.RetiredPubkeys); err != nil {
		return nil, err
	}
	seenOwners, seenConsensus := map[string]bool{}, map[string]bool{}
	for _, entry := range state.Providers {
		if entry == nil || !identifier(entry.ProviderID) || !validOwner(entry.Owner) || len(entry.ConsensusPubkey) != 32 || entry.Power < 1 || entry.Power > 1000000000000 || !entry.Founding && entry.Power != NewcomerPower {
			return nil, ErrProvider
		}
		if seenOwners[entry.Owner] || seenConsensus[string(entry.ConsensusPubkey)] {
			return nil, ErrProvider
		}
		seenOwners[entry.Owner], seenConsensus[string(entry.ConsensusPubkey)] = true, true
		if err := add(recordKey(providerTable, entry.ProviderID), entry); err != nil {
			return nil, err
		}
	}

	return records, validateRecoveryCapacity(state.Providers)
}

func ValidateGenesisState(state *State) error {
	records, err := genesisRecords(state)
	if err != nil {
		return err
	}
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil || len(data) > MaxMessageBytes {
			return ErrLimitExceeded
		}
	}
	return nil
}

func ValidateGenesisChoices(state *State) error {
	for _, entry := range state.Choices {
		if entry == nil {
			return ErrChoice
		}
		for _, location := range entry.Locations {
			if location == nil {
				return ErrChoice
			}
			if location.Name == "base" && !state.Params.BaseEnabled || location.Name == "traderswift" && !state.Params.TraderswiftEnabled {
				return ErrLocationDisabled
			}
		}
	}
	return nil
}

func (keeper *Keeper) InitializeFromGenesis(ctx context.Context, state *State) error {
	if err := ValidateGenesisState(state); err != nil {
		return err
	}
	if err := ValidateGenesisChoices(state); err != nil {
		return err
	}
	records, err := genesisRecords(state)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(records))
	for key := range records {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := keeper.put(ctx, []byte(key), records[key]); err != nil {
			return err
		}
	}
	for _, entry := range state.Frontends {
		if err := keeper.storeService.OpenKVStore(ctx).Set(recordKey(frontendKeyTable, string(entry.ServicePubkey)), []byte(entry.ProviderID)); err != nil {
			return err
		}
	}
	return nil
}

func (keeper *Keeper) ExportToGenesis(ctx context.Context) (*State, error) {
	state := DefaultGenesisState()
	retired, err := keeper.retiredConsensus(ctx)
	if err != nil {
		return nil, err
	}
	state.RetiredPubkeys = retired
	iterator, err := keeper.storeService.OpenKVStore(ctx).Iterator([]byte{linkTable}, []byte{budgetTable})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var target any
		switch iterator.Key()[0] {
		case linkTable:
			entry := &AccountLink{}
			state.Links, target = append(state.Links, entry), entry
		case choiceTable:
			entry := &StorageChoice{}
			state.Choices, target = append(state.Choices, entry), entry
		case rootTable:
			entry := &CommitmentRoot{}
			state.Roots, target = append(state.Roots, entry), entry
		case grantTable:
			entry := &Grant{}
			state.Grants, target = append(state.Grants, entry), entry
		case frontendTable:
			entry := &Frontend{}
			state.Frontends, target = append(state.Frontends, entry), entry
		case providerTable:
			entry := &Provider{}
			state.Providers, target = append(state.Providers, entry), entry
		case admissionTable:
			target = &state.LastAdmissionTime
		case paramsTable:
			target = &state.Params
		default:
			return nil, ErrBadEncoding
		}
		if err := DecodeCanonical(iterator.Value(), target, MaxMessageBytes); err != nil {
			return nil, err
		}
	}
	return state, ValidateGenesisState(state)
}
