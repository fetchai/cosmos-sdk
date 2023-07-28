package verification

import (
	"encoding/json"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	v040 "github.com/cosmos/cosmos-sdk/x/bank/legacy/v040"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/verification/client/cli"
	"github.com/cosmos/cosmos-sdk/x/verification/client/rest"
	"github.com/cosmos/cosmos-sdk/x/verification/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the verification module.
type AppModuleBasic struct {
	cdc codec.Codec
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc}
}

// Name returns the verification module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterRESTRoutes registers the REST routes for the verification module.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterHandlers(clientCtx, rtr)
}

// GetTxCmd returns the root tx command for the verification module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns no root query command for the verification module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements an application module for the bank module.
type AppModule struct {
	AppModuleBasic
}

// RegisterLegacyAminoCodec registers the bank module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)

	// Register legacy interfaces for migration scripts.
	v040.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the bank
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ValidateGenesis performs genesis state validation for the bank module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bank module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	return
}

// InitGenesis performs genesis initialization for the bank module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the bank
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// QuerierRoute returns the bank module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// LegacyQuerierHandler returns the bank module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterInvariants registers the bank module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	return
}

// Route returns the message routing key for the bank module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, nil)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	return
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
	}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }
