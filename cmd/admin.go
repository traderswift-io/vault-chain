package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"vault-chain/validatorset"
)

func adminTransactionCommand() *cobra.Command {
	parent := &cobra.Command{Use: "validatorset", Short: "Rotate the validator set admin", RunE: client.ValidateCmd}
	propose := &cobra.Command{
		Use:  "set-admin [new-admin]",
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			if _, err := sdk.AccAddressFromBech32(arguments[0]); err != nil {
				return err
			}
			context, err := client.GetClientTxContext(command)
			if err != nil {
				return err
			}
			message := &validatorset.MsgSetAdmin{Admin: context.GetFromAddress().String(), NewAdmin: arguments[0]}
			return tx.GenerateOrBroadcastTxCLI(context, command.Flags(), message)
		},
	}
	accept := &cobra.Command{
		Use:  "accept-admin",
		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			context, err := client.GetClientTxContext(command)
			if err != nil {
				return err
			}
			message := &validatorset.MsgAcceptAdmin{NewAdmin: context.GetFromAddress().String()}
			return tx.GenerateOrBroadcastTxCLI(context, command.Flags(), message)
		},
	}
	flags.AddTxFlagsToCmd(propose)
	flags.AddTxFlagsToCmd(accept)
	parent.AddCommand(propose, accept)
	return parent
}
