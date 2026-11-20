package main

import (
	"vault-chain/validatorset"
	"vault-chain/vault"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func boundedTxDecoder(decoder sdk.TxDecoder) sdk.TxDecoder {
	return func(data []byte) (sdk.Tx, error) {
		if err := validateTransaction(data); err != nil {
			return nil, err
		}
		return decoder(data)
	}
}

func validateMessage(typeURL string, data []byte) error {
	var message any
	switch typeURL {
	case "/traderswift.vault.v1.MsgLinkAccount":
		message = &vault.MsgLinkAccount{}
	case "/traderswift.vault.v1.MsgSetStorageChoice":
		message = &vault.MsgSetStorageChoice{}
	case "/traderswift.vault.v1.MsgCommitRoot":
		message = &vault.MsgCommitRoot{}
	case "/traderswift.vault.v1.MsgGrant":
		message = &vault.MsgGrant{}
	case "/traderswift.vault.v1.MsgRevoke":
		message = &vault.MsgRevoke{}
	case "/traderswift.vault.v1.MsgRegisterFrontend":
		message = &vault.MsgRegisterFrontend{}
	case "/traderswift.vault.v1.MsgUpdateParams":
		message = &vault.MsgUpdateParams{}
	case "/traderswift.vault.v1.MsgRemoveFrontend":
		message = &vault.MsgRemoveFrontend{}
	case "/traderswift.vault.v1.MsgActivateProvider":
		message = &vault.MsgActivateProvider{}
	case "/traderswift.vault.v1.MsgRemoveProvider":
		message = &vault.MsgRemoveProvider{}
	case "/traderswift.vault.v1.MsgReplaceProviderOwner":
		message = &vault.MsgReplaceProviderOwner{}
	case "/traderswift.validatorset.v1.MsgSetAdmin":
		message = &validatorset.MsgSetAdmin{}
	case "/traderswift.validatorset.v1.MsgAcceptAdmin":
		message = &validatorset.MsgAcceptAdmin{}
	default:
		return vault.ErrUnauthorized
	}
	return vault.ValidateMessageWire(data, message)
}
