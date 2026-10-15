package app

import (
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	terraappconfig "github.com/terra-money/core/v2/app/config"
	v2_21 "github.com/terra-money/core/v2/app/upgrades/v2.21"
	v2_22 "github.com/terra-money/core/v2/app/upgrades/v2.22"
	v2_23 "github.com/terra-money/core/v2/app/upgrades/v2.23"
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
	app.Keepers.UpgradeKeeper.SetUpgradeHandler(
		terraappconfig.Upgrade2_22,
		v2_22.CreateUpgradeHandler(
			app.GetModuleManager(),
			app.GetConfigurator(),
			app.Keepers,
		),
	)
	app.Keepers.UpgradeKeeper.SetUpgradeHandler(
		terraappconfig.Upgrade2_23,
		v2_23.CreateUpgradeHandler(app.GetModuleManager(), app.GetConfigurator(), app.Keepers),
	)
}

func (app *TerraApp) RegisterUpgradeStores() {
	upgradeInfo, err := app.Keepers.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// Add stores for new modules
	if upgradeInfo.Name == terraappconfig.Upgrade2_23 && !app.Keepers.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{ratelimittypes.StoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
