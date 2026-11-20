package main

import (
	"vault-chain/vault"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
)

func validateTransaction(data []byte) error {
	var raw txtypes.TxRaw
	if err := proto.Unmarshal(data, &raw); err != nil {
		return err
	}
	var body txtypes.TxBody
	if err := proto.Unmarshal(raw.BodyBytes, &body); err != nil {
		return err
	}
	if len(body.Messages) == 0 {
		return vault.ErrBadEncoding
	}
	for _, message := range body.Messages {
		if message == nil {
			return vault.ErrBadEncoding
		}
		if err := validateMessage(message.TypeUrl, message.Value); err != nil {
			return err
		}
	}
	return nil
}
