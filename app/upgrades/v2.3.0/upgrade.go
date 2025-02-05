package v2_3_0

import (
	"context"

	"github.com/terra-money/core/v2/app/config"
	tokenfactorykeeper "github.com/terra-money/core/v2/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/terra-money/core/v2/x/tokenfactory/types"

	math "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	tokenFactoryKeeper tokenfactorykeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		currentVm := mm.GetVersionMap()

		// Init token factory with the correct denom
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		tokenFactoryKeeper.InitGenesis(sdkCtx, tokenfactorytypes.GenesisState{
			Params: tokenfactorytypes.Params{
				DenomCreationFee: sdk.NewCoins(sdk.NewCoin(config.BondDenom, math.NewInt(10_000_000))),
			},
			FactoryDenoms: []tokenfactorytypes.GenesisDenom{},
		})
		fromVM[tokenfactorytypes.ModuleName] = currentVm[tokenfactorytypes.ModuleName]

		return mm.RunMigrations(ctx, cfg, fromVM)
	}
}
