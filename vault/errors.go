package vault

import (
	"vault-chain/chainprefix"

	errorsmod "cosmossdk.io/errors"
)

const (
	ModuleName = chainprefix.ModuleName
	StoreKey   = ModuleName
)

var (
	ErrBadEncoding              = errorsmod.Register(ModuleName, 2, "bad_encoding")
	ErrLimitExceeded            = errorsmod.Register(ModuleName, 3, "limit_exceeded")
	ErrUnauthorized             = errorsmod.Register(ModuleName, 4, "unauthorized")
	ErrLink                     = errorsmod.Register(ModuleName, 10, "invalid_link")
	ErrChoice                   = errorsmod.Register(ModuleName, 11, "invalid_choice")
	ErrLocationDisabled         = errorsmod.Register(ModuleName, 12, "location_disabled")
	ErrRoot                     = errorsmod.Register(ModuleName, 13, "invalid_root")
	ErrRootExists               = errorsmod.Register(ModuleName, 14, "root_exists")
	ErrGrant                    = errorsmod.Register(ModuleName, 15, "invalid_grant")
	ErrRevoke                   = errorsmod.Register(ModuleName, 16, "invalid_revoke")
	ErrFrontend                 = errorsmod.Register(ModuleName, 17, "invalid_frontend")
	ErrParams                   = errorsmod.Register(ModuleName, 18, "invalid_params")
	ErrWrongSigner              = errorsmod.Register(ModuleName, 20, "wrong_signer")
	ErrNonAdminParams           = errorsmod.Register(ModuleName, 21, "non_admin_params")
	ErrFrontendSignature        = errorsmod.Register(ModuleName, 22, "bad_frontend_signature")
	ErrWrongChain               = errorsmod.Register(ModuleName, 23, "wrong_chain_id")
	ErrUnknownLocation          = errorsmod.Register(ModuleName, 24, "unknown_location")
	ErrOversizePointer          = errorsmod.Register(ModuleName, 25, "oversize_pointer")
	ErrLinkReplay               = errorsmod.Register(ModuleName, 26, "link_replay")
	ErrFrontendAuthority        = errorsmod.Register(ModuleName, 27, "frontend_admin_required")
	ErrProvider                 = errorsmod.Register(ModuleName, 28, "invalid_provider")
	ErrJoinCooldown             = errorsmod.Register(ModuleName, 29, "join_cooldown")
	ErrAccountCap               = errorsmod.Register(ModuleName, 30, "account_cap")
	ErrNumberTaken              = errorsmod.Register(ModuleName, 31, "account_number_taken")
	ErrProviderAuthority        = errorsmod.Register(ModuleName, 32, "provider_admin_required")
	ErrProviderMissing          = errorsmod.Register(ModuleName, 33, "provider_not_found")
	ErrProviderOwner            = errorsmod.Register(ModuleName, 34, "invalid_provider_owner")
	ErrProviderOwnerUsed        = errorsmod.Register(ModuleName, 35, "provider_owner_in_use")
	ErrQuery                    = errorsmod.Register(ModuleName, 19, "invalid_query")
	ErrProviderVacancy          = errorsmod.Register(ModuleName, 36, "provider_slot_already_vacant")
	ErrProviderLive             = errorsmod.Register(ModuleName, 37, "provider_slot_live")
	ErrProviderRetired          = errorsmod.Register(ModuleName, 38, "provider_consensus_retired")
	ErrProviderRecoveryCapacity = errorsmod.Register(ModuleName, 39, "provider_recovery_capacity")
)
