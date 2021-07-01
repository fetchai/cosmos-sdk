package airdrop

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)


//____________________________________________________________________________

type AppModuleBasic struct{}

func (a AppModuleBasic) Name() string {
	return types.ModuleName
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	//panic("implement me")
}

func (a AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	//panic("implement me")
}

func (a AppModuleBasic) DefaultGenesis(marshaler codec.JSONMarshaler) json.RawMessage {
	return []byte{}
}

func (a AppModuleBasic) ValidateGenesis(marshaler codec.JSONMarshaler, config client.TxEncodingConfig, message json.RawMessage) error {
	return nil
}

func (a AppModuleBasic) RegisterRESTRoutes(context client.Context, router *mux.Router) {
	//panic("implement me")
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(context client.Context, serveMux *runtime.ServeMux) {
	//panic("implement me")
}

func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
	//panic("implement me")
}

func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	//panic("implement me")
	return nil
}

// AppModule implements an application module for the crisis module.
type AppModule struct {
	AppModuleBasic

	//bankKeeper    types.BankKeeper
}

func NewAppModule() AppModule {
	return AppModule{}
}

func (a AppModule) Name() string {
	return types.ModuleName
}

func (a AppModule) InitGenesis(context sdk.Context, marshaler codec.JSONMarshaler, message json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (a AppModule) ExportGenesis(context sdk.Context, marshaler codec.JSONMarshaler) json.RawMessage {
	return []byte{}
}

func (a AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {
}

func (a AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.ModuleName, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		return nil, nil
	})
}

func (a AppModule) QuerierRoute() string {
	return ""
}

func (a AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		return nil, nil
	}
}

func (a AppModule) RegisterServices(configurator module.Configurator) {
}

func (a AppModule) BeginBlock(context sdk.Context, block abci.RequestBeginBlock) {
	fmt.Println("Whoop whoop this is a block...")
}

func (a AppModule) EndBlock(context sdk.Context, block abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}