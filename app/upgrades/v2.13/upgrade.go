package v2_13

import (
	"errors"
	"slices"

	"github.com/terra-money/core/v2/app/keepers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type EscrowUpdate struct {
	EscrowAddress sdk.AccAddress
	Assets        []sdk.Coin
}

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k keepers.TerraAppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		var phoenixRecoveryAddr sdk.AccAddress
		var axelarAssetList []string

		if ctx.ChainID() == "phoenix-1" {
			phoenixRecoveryAddr = sdk.MustAccAddressFromBech32("terra19v058dytxxwdwada9zla0v7my9kjr5z4dtqp9efzn96q79uxdersl7wy9g")
			axelarAssetList = []string{
				"ibc/05D299885B07905B6886F554B39346EA6761246076A1120B1950049B92B922DD", // axlWBTC
				"ibc/CBF67A2BCF6CAE343FDF251E510C8E18C361FC02B23430C121116E0811835DEF", // axlUSDT
				"ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4", // axlUSDC
			}
		} else {
			phoenixRecoveryAddr = sdk.MustAccAddressFromBech32("")
			axelarAssetList = []string{}
		}

		// Burn from distribution module + update community pool
		communityPoolCoins := k.DistrKeeper.GetFeePoolCommunityCoins(ctx)
		for _, coin := range communityPoolCoins {
			if !slices.Contains(axelarAssetList, coin.Denom) {
				continue
			}

			// Update community pool to reflect the burn
			communityPool := k.DistrKeeper.GetFeePool(ctx)
			newPool, negative := communityPool.CommunityPool.SafeSub(sdk.NewDecCoins(coin))
			if negative {
				return vm, errors.New("negative community pool")
			}
			communityPool.CommunityPool = newPool
			k.DistrKeeper.SetFeePool(ctx, communityPool)

			// Burn from distribution module
			distributionModuleAddr := k.AccountKeeper.GetModuleAddress(distributiontypes.ModuleName)
			balance := k.BankKeeper.GetBalance(ctx, distributionModuleAddr, coin.Denom)
			err := k.BankKeeper.SendCoinsFromModuleToModule(ctx, distributiontypes.ModuleName, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
			if err != nil {
				return vm, err
			}
			err = k.BankKeeper.BurnCoins(ctx, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
			if err != nil {
				return vm, err
			}
		}

		// Burn from phoenix recovery contract
		var err error
		k.BankKeeper.IterateAccountBalances(ctx, phoenixRecoveryAddr, func(balance sdk.Coin) bool {
			if !slices.Contains(axelarAssetList, balance.Denom) {
				return true
			}

			err = k.BankKeeper.SendCoinsFromAccountToModule(ctx, phoenixRecoveryAddr, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
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

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
