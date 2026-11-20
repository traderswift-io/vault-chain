package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"vault-chain/chainprefix"
	"vault-chain/validatorset"
	"vault-chain/vault"
)

func readJSONFile(path string, maximum int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, int64(maximum)+1))
	if err != nil {
		return nil, err
	}
	if err := vault.CheckJSON(data, maximum); err != nil {
		return nil, err
	}
	return data, nil
}

func genesisDocument(command *cobra.Command) (string, map[string]json.RawMessage, map[string]json.RawMessage, error) {
	path := server.GetServerContextFromCmd(command).Config.GenesisFile()
	data, err := readJSONFile(path, vault.MaxGenesisBytes)
	if err != nil {
		return path, nil, nil, err
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return path, nil, nil, err
	}
	var modules map[string]json.RawMessage
	err = json.Unmarshal(document["app_state"], &modules)
	return path, document, modules, err
}

func saveGenesis(path string, document, modules map[string]json.RawMessage) error {
	data, err := json.Marshal(modules)
	if err != nil {
		return err
	}
	document["app_state"] = data
	data, err = json.Marshal(document)
	if err != nil {
		return err
	}
	if err := vault.CheckJSON(data, vault.MaxGenesisBytes); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), "genesis-*.json")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func configureGenesis(parent *cobra.Command, validate func(map[string]json.RawMessage) error) {
	var inheritedValidation func(*cobra.Command, []string) error
	for _, command := range parent.Commands() {
		switch command.Name() {
		case "validate-genesis", "validate":
			inheritedValidation = command.RunE
			parent.RemoveCommand(command)
		case "gentx", "collect-gentxs", "add-genesis-account", "add-genesis-accounts":
			parent.RemoveCommand(command)
		}
	}
	validateFounding := func(command *cobra.Command, modules map[string]json.RawMessage) error {
		context := client.GetClientContextFromCmd(command)
		var state validatorset.GenesisState
		if err := context.Codec.UnmarshalJSON(modules[validatorset.ModuleName], &state); err != nil {
			return err
		}
		_, expected, err := bech32.DecodeAndConvert(chainprefix.GenesisAdmin)
		if err != nil {
			return err
		}
		_, actual, err := bech32.DecodeAndConvert(state.Admin)
		if err != nil || !bytes.Equal(expected, actual) || state.PendingAdmin != "" {
			return fmt.Errorf("founding admin must be GenesisAdmin with no pending rotation")
		}
		if state.Admin != chainprefix.GenesisAdmin && state.Admin != sdk.AccAddress(expected).String() {
			return fmt.Errorf("founding admin has an invalid prefix")
		}
		if err := validatorset.ValidateGenesisState(&state); err != nil {
			return err
		}
		if len(state.Validators) != 4 {
			return fmt.Errorf("founding genesis requires exactly four validators")
		}
		for _, validator := range state.Validators {
			if !validator.Founding || validator.Vacant || validator.Power != state.Validators[0].Power {
				return fmt.Errorf("founding genesis requires equal power and only founding validators")
			}
		}
		if err := validate(modules); err != nil {
			return err
		}
		if inheritedValidation == nil {
			return fmt.Errorf("SDK genesis validator is unavailable")
		}
		return inheritedValidation(command, nil)
	}
	validateDocument := func(command *cobra.Command, modules map[string]json.RawMessage) error {
		snapshot, err := command.Flags().GetBool("snapshot")
		if err != nil {
			return err
		}
		if !snapshot {
			return validateFounding(command, modules)
		}
		if err := validate(modules); err != nil {
			return err
		}
		if inheritedValidation == nil {
			return fmt.Errorf("SDK genesis validator is unavailable")
		}
		return inheritedValidation(command, nil)
	}
	validationCommand := &cobra.Command{
		Use:     "validate",
		Aliases: []string{"validate-genesis"},
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, _, modules, err := genesisDocument(command)
			if err != nil {
				return err
			}
			return validateDocument(command, modules)
		},
	}
	collectionCommand := &cobra.Command{
		Use:   "collect",
		Args:  cobra.NoArgs,
		Short: "Canonicalize the assembled genesis without staking gentxs",
		RunE: func(command *cobra.Command, _ []string) error {
			path, document, modules, err := genesisDocument(command)
			if err != nil {
				return err
			}
			if err := validateDocument(command, modules); err != nil {
				return err
			}
			return saveGenesis(path, document, modules)
		},
	}
	validationCommand.Flags().Bool("snapshot", false, "Validate an exported snapshot instead of an initial founding genesis")
	collectionCommand.Flags().Bool("snapshot", false, "Collect an exported snapshot instead of an initial founding genesis")
	parent.AddCommand(validationCommand, collectionCommand, addValidatorCommand(), addAccountCommand())
}

func addValidatorCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "add-validator [consensus-public-key-base64] [power] [moniker]",
		Args: cobra.ExactArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			if len(args[0]) != 44 || len(args[1]) > 13 || len(args[2]) > 64 {
				return vault.ErrLimitExceeded
			}
			public, err := base64.StdEncoding.DecodeString(args[0])
			if err != nil || len(public) != 32 || base64.StdEncoding.EncodeToString(public) != args[0] {
				return vault.ErrBadEncoding
			}
			power, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			owner, err := command.Flags().GetString("owner")
			if err != nil {
				return err
			}
			founding, err := command.Flags().GetBool("founding")
			if err != nil {
				return err
			}
			path, document, modules, err := genesisDocument(command)
			if err != nil {
				return err
			}
			context := client.GetClientContextFromCmd(command)
			var validators validatorset.GenesisState
			if err := context.Codec.UnmarshalJSON(modules[validatorset.ModuleName], &validators); err != nil {
				return err
			}
			var state vault.State
			if err := json.Unmarshal(modules[vault.ModuleName], &state); err != nil {
				return err
			}
			state.Providers = append(state.Providers, &vault.Provider{ProviderID: args[2], Owner: owner, ConsensusPubkey: public, Power: power, Founding: founding})
			if err := vault.ValidateGenesisState(&state); err != nil {
				return err
			}
			validators.Validators = append(validators.Validators, &validatorset.Validator{ConsensusPublicKey: args[0], Power: power, Moniker: args[2], Founding: founding})
			modules[validatorset.ModuleName], err = context.Codec.MarshalJSON(&validators)
			if err != nil {
				return err
			}
			modules[vault.ModuleName], err = json.Marshal(&state)
			if err != nil {
				return err
			}
			return saveGenesis(path, document, modules)
		},
	}
	command.Flags().String("owner", "", "Provider owner in the chain address prefix")
	command.Flags().Bool("founding", false, "Include this validator in the founding set")
	return command
}
