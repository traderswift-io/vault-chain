package main

import (
	"errors"
	"fmt"
	"os"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"vault-chain/chainprefix"
	daemoncmd "vault-chain/cmd"
)

var defaultNodeHome string

func main() {
	chainprefix.Apply()
	var err error
	defaultNodeHome, err = clienthelpers.GetNodeHomeDirectory("." + chainprefix.BinaryName)
	if err != nil {
		panic(err)
	}
	rootCommand := buildRootCommand()
	if err := svrcmd.Execute(rootCommand, clienthelpers.EnvPrefix, defaultNodeHome); err != nil {
		_, _ = fmt.Fprintln(rootCommand.OutOrStderr(), err)
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	throwawayApp := newTraderSwiftApp(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(defaultNodeHome))
	baseClientContext := client.Context{}.
		WithCodec(throwawayApp.appCodec).
		WithInterfaceRegistry(throwawayApp.interfaceRegistry).
		WithTxConfig(throwawayApp.txConfig).
		WithLegacyAmino(throwawayApp.legacyAmino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(defaultNodeHome).
		WithViper("")

	rootCommand := daemoncmd.NewRootCommand(baseClientContext, throwawayApp.basicModuleManager, throwawayApp.txConfig, defaultNodeHome, newApplicationForServer, exportApplicationForServer, validateModuleGenesis)

	autoCliOptions := throwawayApp.autoCliOptions()
	autoCliOptions.ClientCtx = baseClientContext
	if err := autoCliOptions.EnhanceRootCommand(rootCommand); err != nil {
		panic(err)
	}
	return rootCommand
}

func newApplicationForServer(logger log.Logger, db dbm.DB, applicationOptions servertypes.AppOptions) servertypes.Application {
	return newTraderSwiftApp(logger, db, true, applicationOptions, server.DefaultBaseappOptions(applicationOptions)...)
}

func exportApplicationForServer(logger log.Logger, db dbm.DB, height int64, forZeroHeight bool, jailAllowedAddresses []string, applicationOptions servertypes.AppOptions, modulesToExport []string) (servertypes.ExportedApp, error) {
	if _, isViper := applicationOptions.(*viper.Viper); !isViper {
		return servertypes.ExportedApp{}, errors.New("application options are not viper")
	}
	if height == -1 {
		return newTraderSwiftApp(logger, db, true, applicationOptions).ExportAppStateAndValidators(forZeroHeight, jailAllowedAddresses, modulesToExport)
	}
	app := newTraderSwiftApp(logger, db, false, applicationOptions)
	if err := app.LoadVersion(height); err != nil {
		return servertypes.ExportedApp{}, err
	}
	return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddresses, modulesToExport)
}
