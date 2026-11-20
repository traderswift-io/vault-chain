package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func addAccountCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-genesis-account [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Add an auth account without coins",
		RunE: func(command *cobra.Command, args []string) error {
			if len(args[0]) > 90 {
				return fmt.Errorf("address exceeds 90 bytes")
			}
			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			path, document, modules, err := genesisDocument(command)
			if err != nil {
				return err
			}
			context := client.GetClientContextFromCmd(command)
			var state authtypes.GenesisState
			if err := context.Codec.UnmarshalJSON(modules[authtypes.ModuleName], &state); err != nil {
				return err
			}
			if len(state.Accounts) >= 128 {
				return fmt.Errorf("genesis account limit reached")
			}
			for _, packed := range state.Accounts {
				var account sdk.AccountI
				if err := context.Codec.UnpackAny(packed, &account); err != nil {
					return err
				}
				if account.GetAddress().Equals(address) {
					return fmt.Errorf("account already exists")
				}
			}
			account := authtypes.NewBaseAccount(address, nil, uint64(len(state.Accounts)), 0)
			packed, err := codectypes.NewAnyWithValue(account)
			if err != nil {
				return err
			}
			state.Accounts = append(state.Accounts, packed)
			modules[authtypes.ModuleName], err = context.Codec.MarshalJSON(&state)
			if err != nil {
				return err
			}
			return saveGenesis(path, document, modules)
		},
	}
}
