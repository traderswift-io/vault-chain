package vault

import (
	"bytes"
	"context"
	"crypto/ed25519"

	"vault-chain/chainprefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const JoinSeconds = int64(24 * 60 * 60)
const NewcomerPower = int64(1)
const MaxRetiredConsensusKeys = 256

func (keeper *Keeper) retiredConsensus(ctx context.Context) ([][]byte, error) {
	var retired [][]byte
	_, err := keeper.read(ctx, []byte{retiredTable}, &retired)
	return retired, err
}

type providerActivationAuthorization struct{}

type ValidatorMembership interface {
	ActivateProvider(context.Context, []byte, int64, string) error
}

func ProviderTranscript(message *MsgActivateProvider) []byte {
	return framedDigest(ModuleName+"-provider-v1", []byte(chainprefix.ChainID), []byte(message.Signer), []byte(message.ProviderId), message.ConsensusPubkey)
}

func (keeper *Keeper) providers(ctx context.Context) ([]*Provider, error) {
	iterator, err := keeper.storeService.OpenKVStore(ctx).Iterator([]byte{providerTable}, []byte{providerTable + 1})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	result := []*Provider{}
	for ; iterator.Valid(); iterator.Next() {
		if len(result) >= MaxProviders {
			return nil, ErrLimitExceeded
		}
		entry := &Provider{}
		if err := DecodeCanonical(iterator.Value(), entry, MaxMessageBytes); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, nil
}

func (keeper *Keeper) AuthorizeConsensusChange(ctx context.Context, publicKey []byte, power int64, id string, removal bool) error {
	if removal {
		return keeper.authorizeProviderRemoval(ctx, publicKey, power, id)
	}
	if ctx.Value(providerActivationAuthorization{}) != id {
		return ErrProviderAuthority
	}
	entry := &Provider{}
	found, err := keeper.read(ctx, recordKey(providerTable, id), entry)
	if err != nil {
		return err
	}
	if !found || entry.Vacant || entry.Power != power || !bytes.Equal(entry.ConsensusPubkey, publicKey) {
		return ErrProvider
	}
	if entry.Founding {
		return nil
	}
	if power != NewcomerPower {
		return ErrProvider
	}
	providers, err := keeper.providers(ctx)
	if err != nil {
		return err
	}
	if err := validateRecoveryCapacity(providers); err != nil {
		return err
	}
	var previous int64
	_, err = keeper.read(ctx, []byte{admissionTable}, &previous)
	if err != nil {
		return err
	}
	now := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
	if now < previous || previous != 0 && now-previous < JoinSeconds {
		return ErrJoinCooldown
	}
	return nil
}

func (keeper *Keeper) activateProvider(ctx sdk.Context, value *MsgActivateProvider) error {
	if !validOwner(value.Signer) || !identifier(value.ProviderId) || len(value.ConsensusPubkey) != 32 || len(value.ConsensusSignature) != 64 {
		return ErrProvider
	}
	if !ed25519.Verify(value.ConsensusPubkey, ProviderTranscript(value), value.ConsensusSignature) {
		return ErrProvider
	}
	retired, err := keeper.retiredConsensus(ctx)
	if err != nil {
		return err
	}
	for _, removed := range retired {
		if bytes.Equal(removed, value.ConsensusPubkey) {
			return ErrProviderRetired
		}
	}
	membership, ok := keeper.admin.(ValidatorMembership)
	if !ok {
		return ErrProvider
	}
	providers, err := keeper.providers(ctx)
	if err != nil {
		return err
	}
	var slot *Provider
	for _, entry := range providers {
		for _, retired := range entry.RetiredConsensusPubkeys {
			if bytes.Equal(retired, value.ConsensusPubkey) {
				return ErrProviderRetired
			}
		}
		if entry.ProviderID == value.ProviderId {
			slot = entry
			continue
		}
		if entry.Owner == value.Signer || bytes.Equal(entry.ConsensusPubkey, value.ConsensusPubkey) {
			return ErrProvider
		}
	}
	if slot != nil {
		if !slot.Vacant {
			return ErrProviderLive
		}
		if slot.Owner != value.Signer {
			return ErrProviderOwner
		}
		if !slot.Founding {
			return ErrProvider
		}
	} else {
		if len(providers) >= MaxProviders {
			return ErrLimitExceeded
		}
		slot = &Provider{ProviderID: value.ProviderId, Owner: value.Signer, Power: NewcomerPower}
	}
	slot.ConsensusPubkey = value.ConsensusPubkey
	slot.Vacant = false
	if err := keeper.put(ctx, recordKey(providerTable, value.ProviderId), slot); err != nil {
		return err
	}
	authorized := context.WithValue(ctx, providerActivationAuthorization{}, slot.ProviderID)
	if err := membership.ActivateProvider(authorized, value.ConsensusPubkey, slot.Power, slot.ProviderID); err != nil {
		return err
	}
	if slot.Founding {
		return nil
	}
	now := ctx.BlockTime().Unix()
	return keeper.put(ctx, []byte{admissionTable}, &now)
}
