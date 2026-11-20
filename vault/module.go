package vault

import (
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
)

type AppModule struct{ keeper *Keeper }

var _ module.AppModuleBasic = AppModule{}
var _ appmodule.AppModule = AppModule{}

func NewAppModule(keeper *Keeper) AppModule {
	return AppModule{keeper: keeper}
}

func (AppModule) Name() string {
	return ModuleName
}

func (AppModule) IsAppModule() {}

func (AppModule) IsOnePerModuleType() {}

func (AppModule) ConsensusVersion() uint64 {
	return 1
}

func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

func (AppModule) RegisterGRPCGatewayRoutes(client.Context, *gwruntime.ServeMux) {}

func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLinkAccount{}, &MsgSetStorageChoice{}, &MsgCommitRoot{}, &MsgGrant{}, &MsgRevoke{}, &MsgRegisterFrontend{}, &MsgUpdateParams{}, &MsgRemoveFrontend{}, &MsgActivateProvider{}, &MsgRemoveProvider{}, &MsgReplaceProviderOwner{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &MsgServiceDesc)
}

func (module AppModule) RegisterServices(configurator module.Configurator) {
	configurator.MsgServer().RegisterService(&MsgServiceDesc, msgServer{keeper: module.keeper})
	configurator.QueryServer().RegisterService(&QueryServiceDesc, queryServer{keeper: module.keeper})
}

func (AppModule) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	encoded, err := json.Marshal(DefaultGenesisState())
	if err != nil {
		panic(err)
	}
	return encoded
}

func (AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, data json.RawMessage) error {
	state := &State{}
	if err := DecodeCanonical(data, state, MaxGenesisBytes); err != nil {
		return err
	}
	if err := ValidateGenesisState(state); err != nil {
		return err
	}
	if state.Params.BaseEnabled || state.Params.TraderswiftEnabled {
		return ErrParams
	}
	return ValidateGenesisChoices(state)
}

func (module AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, data json.RawMessage) {
	state := &State{}
	if err := DecodeCanonical(data, state, MaxGenesisBytes); err != nil {
		panic(err)
	}
	if state.Params.BaseEnabled || state.Params.TraderswiftEnabled {
		panic(ErrParams)
	}
	if err := module.keeper.InitializeFromGenesis(ctx, state); err != nil {
		panic(err)
	}
}

func (module AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	state, err := module.keeper.ExportToGenesis(ctx)
	if err != nil {
		panic(err)
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return encoded
}
