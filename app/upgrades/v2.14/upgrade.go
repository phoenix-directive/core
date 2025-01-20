package v2_14

import (
	"slices"

	"github.com/terra-money/core/v2/app/keepers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k keepers.TerraAppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		var phoenixRecoveryAddr sdk.AccAddress
		var assetsToBurnList []string

		if ctx.ChainID() == "phoenix-1" {
			phoenixRecoveryAddr = sdk.MustAccAddressFromBech32("terra19v058dytxxwdwada9zla0v7my9kjr5z4dtqp9efzn96q79uxdersl7wy9g")
			assetsToBurnList = []string{
				"ibc/05D299885B07905B6886F554B39346EA6761246076A1120B1950049B92B922DD", // axlWBTC
				"ibc/CBF67A2BCF6CAE343FDF251E510C8E18C361FC02B23430C121116E0811835DEF", // axlUSDT
				"ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4", // axlUSDC
			}
		} else if ctx.ChainID() == "pisco-1" {
			phoenixRecoveryAddr = sdk.MustAccAddressFromBech32("terra1c3qlvp28tg3v924ta88qrr9a8yswwapacm0lzxg9wmgc4qqrzlmq45sus7")
			assetsToBurnList = []string{
				"factory/terra1azcpczz2jfysyl6m3wx3scpunt4lgcujpkyzt2/test-burn", // fake asset
			}
		} else { // integration tests
			phoenixRecoveryAddr = sdk.MustAccAddressFromBech32("terra14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9ssrc8au")
			assetsToBurnList = []string{
				"factory/terra1v0eee20gjl68fuk0chyrkch2z7suw2mhg3wkxf/test", // fake asset
			}
		}

		// Burn from phoenix recovery contract
		var err error
		if phoenixRecoveryAddr != nil {
			k.BankKeeper.IterateAccountBalances(ctx, phoenixRecoveryAddr, func(balance sdk.Coin) bool {
				if !slices.Contains(assetsToBurnList, balance.Denom) {
					return false
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
		}

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
