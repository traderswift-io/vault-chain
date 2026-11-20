package cmd

import (
	"encoding/json"
	"fmt"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"vault-chain/chainprefix"
)

var GitCommit = "unversioned"
var ArchiveTag = "unversioned"
var ArchiveSource = "unversioned"

func NewRootCommand(base client.Context, basic module.BasicManager, txConfig client.TxConfig, home string, creator servertypes.AppCreator, exporter servertypes.AppExporter, validate func(map[string]json.RawMessage) error) *cobra.Command {
	root := &cobra.Command{
		Use:           chainprefix.BinaryName,
		Short:         chainprefix.ChainID + " metadata chain node",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(command *cobra.Command, _ []string) error {
			if command.Name() == "version" {
				return nil
			}
			command.SetOut(command.OutOrStdout())
			command.SetErr(command.ErrOrStderr())
			context, err := client.ReadPersistentCommandFlags(base.WithCmdContext(command.Context()), command.Flags())
			if err != nil {
				return err
			}
			context, err = clientconfig.ReadFromClientConfig(context)
			if err != nil {
				return err
			}
			if err := client.SetCmdClientContextHandler(context, command); err != nil {
				return err
			}
			application := serverconfig.DefaultConfig()
			application.MinGasPrices = "0nofee"
			application.Pruning = "custom"
			application.PruningKeepRecent = "100000"
			application.PruningInterval = "10"
			application.API.Address = "tcp://127.0.0.1:1317"
			application.GRPC.Address = "127.0.0.1:9090"
			application.Telemetry.Enabled = false
			consensus := cmtcfg.DefaultConfig()
			consensus.RPC.ListenAddress = "tcp://127.0.0.1:26657"
			consensus.RPC.PprofListenAddress = ""
			consensus.P2P.PexReactor = false
			return server.InterceptConfigsPreRunHandler(command, serverconfig.DefaultConfigTemplate, *application, consensus)
		},
	}
	initialize := genutilcli.InitCmd(basic, home)
	initialize.Flags().Lookup(flags.FlagChainID).DefValue = chainprefix.ChainID
	if err := initialize.Flags().Set(flags.FlagChainID, chainprefix.ChainID); err != nil {
		panic(err)
	}
	root.AddCommand(initialize)
	server.AddCommandsWithStartCmdOptions(root, home, creator, exporter, server.StartCmdOptions{})
	genesis := genutilcli.Commands(txConfig, basic, home)
	configureGenesis(genesis, validate)
	query := &cobra.Command{Use: "query", Aliases: []string{"q"}, Short: "Querying subcommands", RunE: client.ValidateCmd}
	query.AddCommand(rpc.WaitTxCmd(), server.QueryBlockCmd(), authcmd.QueryTxCmd(), vaultQueryCommand())
	tx := &cobra.Command{Use: "tx", Short: "Transactions subcommands", RunE: client.ValidateCmd}
	tx.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID (required when signing)")
	tx.AddCommand(authcmd.GetSignCommand(), authcmd.GetMultiSignCommand(), authcmd.GetValidateSignaturesCommand(), authcmd.GetBroadcastCommand(), vaultTransactionCommand(), adminTransactionCommand())
	root.AddCommand(server.StatusCommand(), genesis, query, tx, keys.Commands())
	for _, command := range root.Commands() {
		if command.Name() == "version" {
			root.RemoveCommand(command)
		}
	}
	root.AddCommand(&cobra.Command{
		Use:  "version",
		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(command.OutOrStdout(), "%s\ngit_commit %s\narchive_tag %s\narchive_source_sha256 %s\narchive_recipe trimpath buildvcs=false buildid= tar-mtime=0 gzip-n\n", chainprefix.BinaryName, GitCommit, ArchiveTag, ArchiveSource)
			return err
		},
	})
	return root
}
