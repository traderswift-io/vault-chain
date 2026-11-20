package validatorset

import (
	"vault-chain/chainprefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func genesisAdminAddress(value string) (string, error) {
	if value == chainprefix.GenesisAdmin {
		_, address, err := bech32.DecodeAndConvert(value)
		if err != nil || len(address) != 20 {
			return "", ErrInvalidField
		}
		return sdk.AccAddress(address).String(), nil
	}
	address, err := sdk.AccAddressFromBech32(value)
	if err != nil {
		return "", err
	}
	return address.String(), nil
}
