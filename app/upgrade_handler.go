package app

import (
	terraappconfig "github.com/terra-money/core/v2/app/config"
	v2_20 "github.com/terra-money/core/v2/app/upgrades/v2.20"
)

// RegisterUpgradeHandlers returns upgrade handlers
func (app *TerraApp) RegisterUpgradeHandlers() {
	app.Keepers.UpgradeKeeper.SetUpgradeHandler(
		terraappconfig.Upgrade2_20,
		v2_20.CreateUpgradeHandler(
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
