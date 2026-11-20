package validatorset

import (
	"context"
	"encoding/base64"
	"math/big"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName             = "validatorset"
	StoreKey               = ModuleName
	ed25519PublicKeyLength = 32
	maximumValidatorPower  = int64(1_000_000_000_000)
	maximumMonikerLength   = 70
	maximumValidatorCount  = 32
)

var (
	ErrUnauthorized          = errorsmod.Register(ModuleName, 2, "unauthorized")
	ErrInvalidField          = errorsmod.Register(ModuleName, 3, "invalid field")
	ErrAlreadyExists         = errorsmod.Register(ModuleName, 4, "already exists")
	ErrNotFound              = errorsmod.Register(ModuleName, 5, "not found")
	ErrOutsiderPowerTooLarge = errorsmod.Register(ModuleName, 6, "a non-founding validator would hold one third of voting power or more")
	ErrEmptyValidatorSet     = errorsmod.Register(ModuleName, 7, "the validator set would be empty")
	ErrRecoveryInProgress    = errorsmod.Register(ModuleName, 8, "founder recovery in progress")
)

type Keeper struct {
	storageAdmission             func(context.Context, []byte, int64, string, bool) error
	AdminAddress                 collections.Item[string]
	validators                   collections.Map[string, Validator]
	PowerLastSentToConsensus     collections.Map[string, int64]
	ValidatorSetChangedThisBlock collections.Item[bool]
	PendingAdminAddress          collections.Item[string]
}

func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService) *Keeper {
	schemaBuilder := collections.NewSchemaBuilder(storeService)
	keeper := &Keeper{
		AdminAddress:                 collections.NewItem(schemaBuilder, collections.NewPrefix(1), "admin_address", collections.StringValue),
		validators:                   collections.NewMap(schemaBuilder, collections.NewPrefix(2), "validators", collections.StringKey, codec.CollValue[Validator](cdc)),
		PowerLastSentToConsensus:     collections.NewMap(schemaBuilder, collections.NewPrefix(3), "power_last_sent_to_consensus", collections.StringKey, collections.Int64Value),
		ValidatorSetChangedThisBlock: collections.NewItem(schemaBuilder, collections.NewPrefix(4), "validator_set_changed_this_block", collections.BoolValue),
		PendingAdminAddress:          collections.NewItem(schemaBuilder, collections.NewPrefix(12), "pending_admin_address", collections.StringValue),
	}
	if _, err := schemaBuilder.Build(); err != nil {
		panic(err)
	}
	return keeper
}

func (keeper *Keeper) Admin(ctx sdk.Context) (string, error) {
	return keeper.AdminAddress.Get(ctx)
}

func (keeper *Keeper) requireAdmin(ctx context.Context, submitter string) error {
	admin, err := keeper.AdminAddress.Get(ctx)
	if err != nil {
		return err
	}
	if admin != submitter {
		return errorsmod.Wrapf(ErrUnauthorized, "%s is not the validator-set admin", submitter)
	}
	return nil
}

func decodeConsensusPublicKey(encoded string) ([]byte, error) {
	if len(encoded) != 44 {
		return nil, ErrInvalidField
	}
	publicKeyBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(publicKeyBytes) != ed25519PublicKeyLength || base64.StdEncoding.EncodeToString(publicKeyBytes) != encoded {
		return nil, errorsmod.Wrap(ErrInvalidField, "consensus_public_key must be a 32-byte ed25519 key in standard base64")
	}
	return publicKeyBytes, nil
}

func validatePower(power int64) error {
	if power < 1 || power > maximumValidatorPower {
		return errorsmod.Wrapf(ErrInvalidField, "power must be 1 to %d", maximumValidatorPower)
	}
	return nil
}

func validateMoniker(moniker string) error {
	if moniker == "" || len(moniker) > maximumMonikerLength {
		return errorsmod.Wrapf(ErrInvalidField, "moniker must be 1 to %d bytes", maximumMonikerLength)
	}
	return nil
}

func validateValidator(validator Validator) error {
	if _, err := decodeConsensusPublicKey(validator.ConsensusPublicKey); err != nil {
		return err
	}
	if err := validatePower(validator.Power); err != nil {
		return err
	}
	return validateMoniker(validator.Moniker)
}

func requireSetWithinOutsiderLimit(validators []Validator) error {
	if len(validators) == 0 {
		return ErrEmptyValidatorSet
	}
	if len(validators) > maximumValidatorCount {
		return errorsmod.Wrapf(ErrInvalidField, "at most %d validators", maximumValidatorCount)
	}
	vacancies := 0
	vacantPower := new(big.Int)
	totalPower, outsiderPower, largestFounder := new(big.Int), new(big.Int), new(big.Int)
	for _, validator := range validators {
		if err := validateValidator(validator); err != nil {
			return err
		}
		power := big.NewInt(validator.Power)
		if validator.Vacant {
			vacancies++
			if !validator.Founding || vacancies > 1 {
				return ErrRecoveryInProgress
			}
			vacantPower.Add(vacantPower, power)
		}
		totalPower.Add(totalPower, power)
		if validator.Founding {
			if power.Cmp(largestFounder) > 0 {
				largestFounder.Set(power)
			}
		} else {
			outsiderPower.Add(outsiderPower, power)
		}
	}
	threeTimesPower := new(big.Int).Mul(outsiderPower, big.NewInt(3))
	foundingPower := new(big.Int).Sub(totalPower, outsiderPower)
	remaining := new(big.Int).Sub(foundingPower, new(big.Int).Mul(largestFounder, big.NewInt(3)))
	twiceOutsider := new(big.Int).Mul(outsiderPower, big.NewInt(2))
	activePower := new(big.Int).Sub(totalPower, vacantPower)
	if activePower.Sign() <= 0 || threeTimesPower.Cmp(activePower) >= 0 || twiceOutsider.Cmp(remaining) >= 0 {
		return ErrOutsiderPowerTooLarge
	}
	return nil
}

func (keeper *Keeper) allValidators(ctx context.Context) ([]Validator, error) {
	validators := []Validator{}
	iterator, err := keeper.validators.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		validator, err := iterator.Value()
		if err != nil {
			return nil, err
		}
		if len(validators) >= maximumValidatorCount {
			return nil, ErrInvalidField
		}
		validators = append(validators, validator)
	}
	return validators, nil
}

func (keeper *Keeper) applyChange(ctx context.Context, consensusPublicKey string, replacement *Validator) error {
	current, err := keeper.allValidators(ctx)
	if err != nil {
		return err
	}
	proposed := []Validator{}
	for _, validator := range current {
		if validator.ConsensusPublicKey != consensusPublicKey {
			proposed = append(proposed, validator)
		}
	}
	if replacement != nil {
		proposed = append(proposed, *replacement)
	}
	if err := requireSetWithinOutsiderLimit(proposed); err != nil {
		return err
	}
	if replacement == nil || replacement.ConsensusPublicKey != consensusPublicKey {
		if err := keeper.validators.Remove(ctx, consensusPublicKey); err != nil {
			return err
		}
	}
	if replacement != nil {
		if err := keeper.validators.Set(ctx, replacement.ConsensusPublicKey, *replacement); err != nil {
			return err
		}
	}
	return keeper.ValidatorSetChangedThisBlock.Set(ctx, true)
}

func emitValidatorSetEvent(ctx context.Context, eventType string, attributes ...sdk.Attribute) {
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent("validatorset."+eventType, attributes...))
}
