package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"vault-chain/vault"
)

func vaultTransactionCommand() *cobra.Command {
	parent := &cobra.Command{Use: "vault", Short: "Submit vault messages from bounded JSON files", RunE: client.ValidateCmd}
	messages := map[string]func() sdk.Msg{
		"link-account":           func() sdk.Msg { return &vault.MsgLinkAccount{} },
		"set-storage-choice":     func() sdk.Msg { return &vault.MsgSetStorageChoice{} },
		"commit-root":            func() sdk.Msg { return &vault.MsgCommitRoot{} },
		"grant":                  func() sdk.Msg { return &vault.MsgGrant{} },
		"revoke":                 func() sdk.Msg { return &vault.MsgRevoke{} },
		"register-frontend":      func() sdk.Msg { return &vault.MsgRegisterFrontend{} },
		"remove-frontend":        func() sdk.Msg { return &vault.MsgRemoveFrontend{} },
		"update-params":          func() sdk.Msg { return &vault.MsgUpdateParams{} },
		"activate-provider":      func() sdk.Msg { return &vault.MsgActivateProvider{} },
		"remove-provider":        func() sdk.Msg { return &vault.MsgRemoveProvider{} },
		"replace-provider-owner": func() sdk.Msg { return &vault.MsgReplaceProviderOwner{} },
	}
	for name, create := range messages {
		command := &cobra.Command{
			Use:  name + " [message.json]",
			Args: cobra.ExactArgs(1),
			RunE: func(command *cobra.Command, args []string) error {
				clientContext, err := client.GetClientTxContext(command)
				if err != nil {
					return err
				}
				data, err := readJSONFile(args[0], vault.MaxMessageBytes)
				if err != nil {
					return err
				}
				message := create()
				if err := clientContext.Codec.UnmarshalJSON(data, message); err != nil {
					return err
				}
				return tx.GenerateOrBroadcastTxCLI(clientContext, command.Flags(), message)
			},
		}
		flags.AddTxFlagsToCmd(command)
		parent.AddCommand(command)
	}
	return parent
}

func vaultQueryCommand() *cobra.Command {
	parent := &cobra.Command{Use: "vault", Short: "Query vault metadata", RunE: client.ValidateCmd}
	where := &cobra.Command{
		Use:  "where [owner]",
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if len(args[0]) > 90 {
				return vault.ErrLimitExceeded
			}
			return invokeVaultQuery(command, "Where", &vault.QueryWhereRequest{Owner: args[0]}, &vault.QueryWhereResponse{})
		},
	}
	mayRead := &cobra.Command{
		Use:  "may-read [owner] [grantee] [salt-base64] [scope]",
		Args: cobra.ExactArgs(4),
		RunE: func(command *cobra.Command, args []string) error {
			if len(args[0]) > 90 || len(args[1]) > 90 || len(args[2]) != 44 || len(args[3]) > 64 {
				return vault.ErrLimitExceeded
			}
			salt, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil || len(salt) != 32 {
				return fmt.Errorf("salt must be 32 bytes encoded as base64")
			}
			return invokeVaultQuery(command, "MayRead", &vault.QueryMayReadRequest{Owner: args[0], Grantee: args[1], Salt: salt, Scope: args[3]}, &vault.QueryMayReadResponse{})
		},
	}
	flags.AddQueryFlagsToCmd(where)
	flags.AddQueryFlagsToCmd(mayRead)
	parent.AddCommand(where, mayRead)
	return parent
}

func invokeVaultQuery(command *cobra.Command, method string, request, response proto.Message) error {
	clientContext, err := client.GetClientQueryContext(command)
	if err != nil {
		return err
	}
	if err := clientContext.Invoke(command.Context(), "/traderswift.vault.v1.Query/"+method, request, response); err != nil {
		return err
	}
	return clientContext.PrintProto(response)
}
