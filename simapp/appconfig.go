package simapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	moduletypes "github.com/cosmos/cosmos-sdk/regen/types/module"
	regenserver "github.com/cosmos/cosmos-sdk/regen/types/module/server"

	"github.com/cosmos/cosmos-sdk/types/module"

	group "github.com/cosmos/cosmos-sdk/x/group/module"
)

// from regen-ledger: https://github.com/regen-network/regen-ledger/blob/v2.0.0-beta1/app/experimental_appconfig.go

func setCustomModuleBasics() []module.AppModuleBasic {
	return []module.AppModuleBasic{
		group.Module{},
	}
}

func setCustomKVStoreKeys() []string {
	return []string{}
}

// setCustomModules registers new modules with the server module manager.
func setCustomModules(app *SimApp, interfaceRegistry types.InterfaceRegistry) *regenserver.Manager {

	/* New Module Wiring START */
	newModuleManager := regenserver.NewManager(app.BaseApp, codec.NewProtoCodec(interfaceRegistry))

	// BEGIN HACK: this is a total, ugly hack until x/auth & x/bank supports ADR 033 or we have a suitable alternative
	groupModule := group.Module{AccountKeeper: app.AccountKeeper, BankKeeper: app.BankKeeper}
	// use a separate newModules from the global NewModules here because we need to pass state into the group module
	newModules := []moduletypes.Module{
		groupModule,
	}
	err := newModuleManager.RegisterModules(newModules)
	if err != nil {
		panic(err)
	}
	// END HACK

	/* New Module Wiring END */
	return newModuleManager
}

func (app *SimApp) setCustomModuleManager() []module.AppModule {
	return []module.AppModule{}
}

func setCustomOrderInitGenesis() []string {
	return []string{}
}

func (app *SimApp) setCustomSimulationManager() []module.AppModuleSimulation {
	return []module.AppModuleSimulation{
		group.Module{
			Registry:      app.interfaceRegistry,
			BankKeeper:    app.BankKeeper,
			AccountKeeper: app.AccountKeeper,
		},
	}
}
