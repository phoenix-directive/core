package app

import (
	terraappconfig "github.com/terra-money/core/v2/app/config"
	v2_21 "github.com/terra-money/core/v2/app/upgrades/v2.21"
)

// RegisterUpgradeHandlers returns upgrade handlers
func (app *TerraApp) RegisterUpgradeHandlers() {
	app.Keepers.UpgradeKeeper.SetUpgradeHandler(
		terraappconfig.Upgrade2_21,
		v2_21.CreateUpgradeHandler(
			app.GetModuleManager(),
			app.GetConfigurator(),
			app.Keepers,
		),
	)
}

func (app *TerraApp) RegisterUpgradeStores() {
	_, err := app.Keepers.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}
}
