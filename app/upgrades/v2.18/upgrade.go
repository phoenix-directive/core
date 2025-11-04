package v2_18

import (
	"context"
	"slices"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/terra-money/core/v2/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k keepers.TerraAppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		var phoenixTreasuryAddr sdk.AccAddress
		var assetsToBurnList []string

		if sdkCtx.ChainID() == "phoenix-1" {
			phoenixTreasuryAddr = sdk.MustAccAddressFromBech32("terra12ncurr62xe93xrsh2drp4zvehj0gn32lfnshr8k0p4xfyju2knwq2qgmh2")
			assetsToBurnList = []string{
				"ibc/8D8A7F7253615E5F76CB6252A1E1BD921D5EDB7BBAAF8913FB1C77FF125D9995", // ASTRO
			}
		}

		// Burn from phoenix recovery contract
		var err error
		if phoenixTreasuryAddr != nil {
			k.BankKeeper.IterateAccountBalances(ctx, phoenixTreasuryAddr, func(balance sdk.Coin) bool {
				if !slices.Contains(assetsToBurnList, balance.Denom) {
					return false
				}

				err = k.BankKeeper.SendCoinsFromAccountToModule(ctx, phoenixTreasuryAddr, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
				if err != nil {
					return true
				}
				err = k.BankKeeper.BurnCoins(ctx, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
				if err != nil {
					return true
				}
				return false
			})
			if err != nil {
				return vm, err
			}
		}
		return mm.RunMigrations(ctx, cfg, vm)

	}
}
